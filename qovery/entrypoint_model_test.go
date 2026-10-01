//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

// entrypointReadCases are the reads of an entrypoint: a Console save of the service settings
// stores "", which means the image's entrypoint, like null. It reads back as the null or ""
// the plan or the state holds, and any other value follows the API.
var entrypointReadCases = []struct {
	TestName string
	Prior    types.String
	API      *string
	Expect   types.String
}{
	{TestName: "console_empty_reads_as_null_when_omitted", Prior: types.StringNull(), API: strPtr(""), Expect: types.StringNull()},
	{TestName: "console_empty_stays_empty_when_declared_empty", Prior: types.StringValue(""), API: strPtr(""), Expect: types.StringValue("")},
	{TestName: "console_empty_shows_when_declared", Prior: types.StringValue("/bin/sh"), API: strPtr(""), Expect: types.StringValue("")},
	{TestName: "console_value_shows_when_omitted", Prior: types.StringNull(), API: strPtr("/bin/sh"), Expect: types.StringValue("/bin/sh")},
	{TestName: "absent_reads_as_null", Prior: types.StringNull(), API: nil, Expect: types.StringNull()},
	{TestName: "absent_shows_when_declared_empty", Prior: types.StringValue(""), API: nil, Expect: types.StringNull()},
}

func TestConvertDomainContainerToContainer_Entrypoint(t *testing.T) {
	t.Parallel()

	for _, tc := range entrypointReadCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			plan := nullPlanFor[Container](t, &containerResource{})
			plan.Entrypoint = tc.Prior
			api := &container.Container{ID: uuid.New(), EnvironmentID: uuid.New(), Name: "container", Entrypoint: tc.API}

			state := convertDomainContainerToContainer(context.Background(), plan, api)

			assert.Equal(t, tc.Expect, state.Entrypoint)
		})
	}
}

func TestConvertResponseToApplication_Entrypoint(t *testing.T) {
	t.Parallel()

	for _, tc := range entrypointReadCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			plan := nullPlanFor[Application](t, &applicationResource{})
			plan.Entrypoint = tc.Prior
			api := &client.ApplicationResponse{ApplicationResponse: &qovery.Application{
				Id:          uuid.NewString(),
				Name:        "application",
				Environment: qovery.ReferenceObject{Id: uuid.NewString()},
				Entrypoint:  tc.API,
			}}

			state := convertResponseToApplication(context.Background(), plan, api)

			assert.Equal(t, tc.Expect, state.Entrypoint)
		})
	}
}

func TestJobScheduleFromDomainJobSchedule_Entrypoint(t *testing.T) {
	t.Parallel()

	for _, tc := range entrypointReadCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			command := func() *execution_command.ExecutionCommand {
				return &execution_command.ExecutionCommand{Entrypoint: tc.API}
			}
			prior := &JobSchedule{
				OnStart:  &ExecutionCommand{Entrypoint: tc.Prior},
				OnStop:   &ExecutionCommand{Entrypoint: tc.Prior},
				OnDelete: &ExecutionCommand{Entrypoint: tc.Prior},
				CronJob:  &JobScheduleCron{Command: ExecutionCommand{Entrypoint: tc.Prior}},
			}
			api := job.JobSchedule{
				OnStart:  command(),
				OnStop:   command(),
				OnDelete: command(),
				CronJob:  &job.JobScheduleCron{Schedule: "0 * * * *", Command: *command()},
			}

			schedule := JobScheduleFromDomainJobSchedule(api, prior)

			assert.Equal(t, tc.Expect, schedule.OnStart.Entrypoint, "on_start")
			assert.Equal(t, tc.Expect, schedule.OnStop.Entrypoint, "on_stop")
			assert.Equal(t, tc.Expect, schedule.OnDelete.Entrypoint, "on_delete")
			assert.Equal(t, tc.Expect, schedule.CronJob.Command.Entrypoint, "cronjob.command")
		})
	}

	t.Run("import_reads_console_empty_as_null", func(t *testing.T) {
		t.Parallel()

		api := job.JobSchedule{
			OnStart: &execution_command.ExecutionCommand{Entrypoint: strPtr("")},
			CronJob: &job.JobScheduleCron{Schedule: "0 * * * *", Command: execution_command.ExecutionCommand{Entrypoint: strPtr("")}},
		}

		schedule := JobScheduleFromDomainJobSchedule(api, nil)

		assert.True(t, schedule.OnStart.Entrypoint.IsNull(), "on_start")
		assert.True(t, schedule.CronJob.Command.Entrypoint.IsNull(), "cronjob.command")
	})
}

// TestDataSourceJobSchedulePrior pins the prior the job data source reads with: an empty API
// entrypoint reports as "" and a missing one as null, and the arguments read as without a prior.
func TestDataSourceJobSchedulePrior(t *testing.T) {
	t.Parallel()

	api := job.JobSchedule{
		OnStart:  &execution_command.ExecutionCommand{Entrypoint: strPtr(""), Arguments: []string{}},
		OnStop:   &execution_command.ExecutionCommand{Entrypoint: nil},
		OnDelete: &execution_command.ExecutionCommand{Entrypoint: strPtr("/bin/sh"), Arguments: []string{"-c"}},
		CronJob:  &job.JobScheduleCron{Schedule: "0 * * * *", Command: execution_command.ExecutionCommand{Entrypoint: strPtr("")}},
	}

	schedule := JobScheduleFromDomainJobSchedule(api, dataSourceJobSchedulePrior())

	assert.Equal(t, types.StringValue(""), schedule.OnStart.Entrypoint, "on_start")
	assert.Nil(t, schedule.OnStart.Arguments, "on_start arguments")
	assert.True(t, schedule.OnStop.Entrypoint.IsNull(), "on_stop")
	assert.Equal(t, types.StringValue("/bin/sh"), schedule.OnDelete.Entrypoint, "on_delete")
	assert.Equal(t, []types.String{types.StringValue("-c")}, schedule.OnDelete.Arguments, "on_delete arguments")
	assert.Equal(t, types.StringValue(""), schedule.CronJob.Command.Entrypoint, "cronjob.command")
}
