//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

func TestJobScheduleFromDomainJobSchedule(t *testing.T) {
	t.Parallel()

	t.Run("nil_arguments_stays_nil", func(t *testing.T) {
		t.Parallel()
		schedule := job.JobSchedule{
			OnStart: &execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  nil,
			},
			OnStop: &execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  nil,
			},
			OnDelete: &execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  nil,
			},
		}

		result := JobScheduleFromDomainJobSchedule(schedule)

		assert.NotNil(t, result.OnStart)
		assert.Nil(t, result.OnStart.Arguments, "nil arguments should stay nil, not become empty slice")
		assert.NotNil(t, result.OnStop)
		assert.Nil(t, result.OnStop.Arguments, "nil arguments should stay nil, not become empty slice")
		assert.NotNil(t, result.OnDelete)
		assert.Nil(t, result.OnDelete.Arguments, "nil arguments should stay nil, not become empty slice")
	})

	t.Run("empty_arguments_stays_nil", func(t *testing.T) {
		t.Parallel()
		schedule := job.JobSchedule{
			OnStart: &execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  []string{},
			},
			OnStop: &execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  []string{},
			},
			OnDelete: &execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  []string{},
			},
		}

		result := JobScheduleFromDomainJobSchedule(schedule)

		assert.NotNil(t, result.OnStart)
		assert.Nil(t, result.OnStart.Arguments, "empty arguments should become nil, not empty slice")
		assert.NotNil(t, result.OnStop)
		assert.Nil(t, result.OnStop.Arguments, "empty arguments should become nil, not empty slice")
		assert.NotNil(t, result.OnDelete)
		assert.Nil(t, result.OnDelete.Arguments, "empty arguments should become nil, not empty slice")
	})

	t.Run("non_empty_arguments_preserved", func(t *testing.T) {
		t.Parallel()
		schedule := job.JobSchedule{
			OnStart: &execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  []string{"-c", "echo hello"},
			},
		}

		result := JobScheduleFromDomainJobSchedule(schedule)

		assert.NotNil(t, result.OnStart)
		assert.Len(t, result.OnStart.Arguments, 2)
		assert.Equal(t, types.StringValue("-c"), result.OnStart.Arguments[0])
		assert.Equal(t, types.StringValue("echo hello"), result.OnStart.Arguments[1])
	})

	t.Run("nil_events_stay_nil", func(t *testing.T) {
		t.Parallel()
		schedule := job.JobSchedule{
			OnStart:  nil,
			OnStop:   nil,
			OnDelete: nil,
		}

		result := JobScheduleFromDomainJobSchedule(schedule)

		assert.Nil(t, result.OnStart)
		assert.Nil(t, result.OnStop)
		assert.Nil(t, result.OnDelete)
	})

	t.Run("nil_entrypoint_with_arguments", func(t *testing.T) {
		t.Parallel()
		schedule := job.JobSchedule{
			OnStart: &execution_command.ExecutionCommand{
				Entrypoint: nil,
				Arguments:  []string{"arg1"},
			},
		}

		result := JobScheduleFromDomainJobSchedule(schedule)

		assert.NotNil(t, result.OnStart)
		assert.True(t, result.OnStart.Entrypoint.IsNull())
		assert.Len(t, result.OnStart.Arguments, 1)
		assert.Equal(t, types.StringValue("arg1"), result.OnStart.Arguments[0])
	})
}

func TestJobScheduleCronFromDomainJobScheduleCron(t *testing.T) {
	t.Parallel()

	t.Run("nil_arguments_stays_nil", func(t *testing.T) {
		t.Parallel()
		cron := job.JobScheduleCron{
			Schedule: "0 * * * *",
			Command: execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  nil,
			},
		}

		result := JobScheduleCronFromDomainJobScheduleCron(cron)

		assert.Equal(t, types.StringValue("0 * * * *"), result.Schedule)
		assert.Nil(t, result.Command.Arguments, "nil arguments should stay nil, not become empty slice")
	})

	t.Run("empty_arguments_stays_nil", func(t *testing.T) {
		t.Parallel()
		cron := job.JobScheduleCron{
			Schedule: "0 * * * *",
			Command: execution_command.ExecutionCommand{
				Entrypoint: strPtr("/bin/sh"),
				Arguments:  []string{},
			},
		}

		result := JobScheduleCronFromDomainJobScheduleCron(cron)

		assert.Nil(t, result.Command.Arguments, "empty arguments should become nil, not empty slice")
	})

	t.Run("non_empty_arguments_preserved", func(t *testing.T) {
		t.Parallel()
		cron := job.JobScheduleCron{
			Schedule: "0 * * * *",
			Command: execution_command.ExecutionCommand{
				Entrypoint: strPtr("test.sh"),
				Arguments:  []string{"arg1", "arg2"},
			},
		}

		result := JobScheduleCronFromDomainJobScheduleCron(cron)

		assert.Equal(t, types.StringValue("0 * * * *"), result.Schedule)
		assert.Len(t, result.Command.Arguments, 2)
		assert.Equal(t, types.StringValue("arg1"), result.Command.Arguments[0])
		assert.Equal(t, types.StringValue("arg2"), result.Command.Arguments[1])
	})
}
