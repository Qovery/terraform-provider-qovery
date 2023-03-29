package job_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	execution_command_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/execution_command/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	job_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/job/test_helper"
)

func TestJobScheduleValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description   string
		onStart       *execution_command.ExecutionCommand
		onStop        *execution_command.ExecutionCommand
		onDelete      *execution_command.ExecutionCommand
		scheduledAt   *job.JobScheduleCron
		expectedError error
	}{
		{description: "case 1: all fields are nil", onStart: nil, onStop: nil, onDelete: nil, scheduledAt: nil, expectedError: job.ErrInvalidJobScheduleMissingRequiredParams},
		{description: "case 2: all fields are set", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: &execution_command_test_helper.DefaultValidExecutionCommand, onDelete: &execution_command_test_helper.DefaultValidExecutionCommand, scheduledAt: &job_test_helper.DefaultValidJobScheduledCronCron, expectedError: job.ErrInvalidJobScheduleTooManyParams},
		{description: "case 3: invalid `scheduled at` param", onStart: nil, onStop: nil, onDelete: nil, scheduledAt: &job_test_helper.DefaultInvalidJobScheduledCronCron, expectedError: errors.Wrap(job_test_helper.DefaultInvalidNewInvalidJobScheduledCronCronParamsError, job.ErrInvalidJobScheduleCronParam.Error())},
		{description: "case 4: valid `on start` param", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: nil, expectedError: nil},
		{description: "case 5: valid `on stop` param", onStart: nil, onStop: &execution_command_test_helper.DefaultValidExecutionCommand, onDelete: nil, scheduledAt: nil, expectedError: nil},
		{description: "case 6: valid `on delete` param", onStart: nil, onStop: nil, onDelete: &execution_command_test_helper.DefaultValidExecutionCommand, scheduledAt: nil, expectedError: nil},
		{description: "case 7: valid `scheduled at` param", onStart: nil, onStop: nil, onDelete: nil, scheduledAt: &job_test_helper.DefaultValidJobScheduledCronCron, expectedError: nil},
		{description: "case 8: several valid exclusive fields set", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: &job_test_helper.DefaultValidJobScheduledCronCron, expectedError: job.ErrInvalidJobScheduleTooManyParams},
		{description: "case 9: several invalid & valid exclusive fields set", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: &job_test_helper.DefaultInvalidJobScheduledCronCron, expectedError: job.ErrInvalidJobScheduleTooManyParams},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			s := job.JobSchedule{
				OnStart:  tc.onStart,
				OnStop:   tc.onStop,
				OnDelete: tc.onDelete,
				CronJob:  tc.scheduledAt,
			}

			// verify:
			if err := s.Validate(); err != nil {
				assert.Equal(t, tc.expectedError.Error(), s.Validate().Error())
			} else {
				assert.Equal(t, tc.expectedError, s.Validate()) // <- should be nil
			}
		})
	}
}

func TestNewJobSchedule(t *testing.T) {
	// setup:
	testCases := []struct {
		description    string
		params         job.NewJobScheduleParams
		expectedResult *job.JobSchedule
		expectedError  error
	}{
		{
			description: "case 1: all fields are nil",
			params: job.NewJobScheduleParams{
				OnStart:  nil,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  nil,
			},
			expectedError:  job.ErrInvalidJobScheduleMissingRequiredParams,
			expectedResult: nil,
		},
		{
			description: "case 2: all fields are set",
			params: job.NewJobScheduleParams{
				OnStart:  &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:   &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnDelete: &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				CronJob:  &job_test_helper.DefaultValidJobScheduledCronCronParams,
			},
			expectedError:  job.ErrInvalidJobScheduleTooManyParams,
			expectedResult: nil,
		},
		{
			description: "case 3: invalid `scheduled at` param",
			params: job.NewJobScheduleParams{
				OnStart:  nil,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  &job_test_helper.DefaultInvalidJobScheduledCronCronParams,
			},
			expectedError:  errors.Wrap(job_test_helper.DefaultInvalidNewInvalidJobScheduledCronCronParamsError, job.ErrInvalidJobScheduleCronParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 4: valid `on start` param",
			params: job.NewJobScheduleParams{
				OnStart:  &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  nil,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:  &execution_command_test_helper.DefaultValidExecutionCommand,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  nil,
			},
		},
		{
			description: "case 5: valid `on stop` param",
			params: job.NewJobScheduleParams{
				OnStart:  nil,
				OnStop:   &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnDelete: nil,
				CronJob:  nil,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:  nil,
				OnStop:   &execution_command_test_helper.DefaultValidExecutionCommand,
				OnDelete: nil,
				CronJob:  nil,
			},
		},
		{
			description: "case 6: valid `on delete` param",
			params: job.NewJobScheduleParams{
				OnStart:  nil,
				OnStop:   nil,
				OnDelete: &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				CronJob:  nil,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:  nil,
				OnStop:   nil,
				OnDelete: &execution_command_test_helper.DefaultValidExecutionCommand,
				CronJob:  nil,
			},
		},
		{
			description: "case 7: valid `scheduled at` param",
			params: job.NewJobScheduleParams{
				OnStart:  nil,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  &job_test_helper.DefaultValidJobScheduledCronCronParams,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:  nil,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  &job_test_helper.DefaultValidJobScheduledCronCron,
			},
		},
		{
			description: "case 8: several valid exclusive fields set",
			params: job.NewJobScheduleParams{
				OnStart:  &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  &job_test_helper.DefaultValidJobScheduledCronCronParams,
			},
			expectedError:  job.ErrInvalidJobScheduleTooManyParams,
			expectedResult: nil,
		},
		{
			description: "case 9: several invalid & valid exclusive fields set", params: job.NewJobScheduleParams{
				OnStart:  &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:   nil,
				OnDelete: nil,
				CronJob:  &job_test_helper.DefaultValidJobScheduledCronCronParams,
			},
			expectedError:  job.ErrInvalidJobScheduleTooManyParams,
			expectedResult: nil,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i, err := job.NewJobSchedule(tc.params)

			// verify:
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.Equal(t, nil, err)
			}
			assert.Equal(t, tc.expectedResult, i)
		})
	}
}
