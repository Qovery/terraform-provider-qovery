package job_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	execution_command_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/execution_command/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

var (
	DefaultScheduledAt        = "*/30 * * * *"
	DefaultInvalidScheduledAt = ""
)

func TestJobScheduleValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description   string
		onStart       *execution_command.ExecutionCommand
		onStop        *execution_command.ExecutionCommand
		onDelete      *execution_command.ExecutionCommand
		scheduledAt   *string
		expectedError error
	}{
		{description: "case 1: all fields are nil", onStart: nil, onStop: nil, onDelete: nil, scheduledAt: nil, expectedError: job.ErrInvalidJobScheduleMissingRequiredParams},
		{description: "case 2: all fields are set", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: &execution_command_test_helper.DefaultValidExecutionCommand, onDelete: &execution_command_test_helper.DefaultValidExecutionCommand, scheduledAt: &DefaultScheduledAt, expectedError: job.ErrInvalidJobScheduleTooManyParams},
		{description: "case 3: invalid `on start` param", onStart: &execution_command_test_helper.DefaultInvalidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: nil, expectedError: errors.Wrap(execution_command_test_helper.DefaultInvalidExecutionCommandParamsError, job.ErrInvalidJobScheduleOnStartParam.Error())},
		{description: "case 4: invalid `on stop` param", onStart: nil, onStop: &execution_command_test_helper.DefaultInvalidExecutionCommand, onDelete: nil, scheduledAt: nil, expectedError: errors.Wrap(execution_command_test_helper.DefaultInvalidExecutionCommandParamsError, job.ErrInvalidJobScheduleOnStopParam.Error())},
		{description: "case 5: invalid `on delete` param", onStart: nil, onStop: nil, onDelete: &execution_command_test_helper.DefaultInvalidExecutionCommand, scheduledAt: nil, expectedError: errors.Wrap(execution_command_test_helper.DefaultInvalidExecutionCommandParamsError, job.ErrInvalidJobScheduleOnDeleteParam.Error())},
		{description: "case 6: invalid `scheduled at` param", onStart: nil, onStop: nil, onDelete: nil, scheduledAt: &DefaultInvalidScheduledAt, expectedError: job.ErrInvalidJobScheduleScheduledAtParam},
		{description: "case 7: valid `on start` param", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: nil, expectedError: nil},
		{description: "case 8: valid `on stop` param", onStart: nil, onStop: &execution_command_test_helper.DefaultValidExecutionCommand, onDelete: nil, scheduledAt: nil, expectedError: nil},
		{description: "case 9: valid `on delete` param", onStart: nil, onStop: nil, onDelete: &execution_command_test_helper.DefaultValidExecutionCommand, scheduledAt: nil, expectedError: nil},
		{description: "case 10: valid `scheduled at` param", onStart: nil, onStop: nil, onDelete: nil, scheduledAt: &DefaultScheduledAt, expectedError: nil},
		{description: "case 11: several valid exclusive fields set", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: &DefaultScheduledAt, expectedError: job.ErrInvalidJobScheduleTooManyParams},
		{description: "case 12: several invalid exclusive fields set", onStart: &execution_command_test_helper.DefaultInvalidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: &DefaultInvalidScheduledAt, expectedError: job.ErrInvalidJobScheduleTooManyParams},
		{description: "case 13: several invalid & valid exclusive fields set", onStart: &execution_command_test_helper.DefaultValidExecutionCommand, onStop: nil, onDelete: nil, scheduledAt: &DefaultInvalidScheduledAt, expectedError: job.ErrInvalidJobScheduleTooManyParams},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			s := job.JobSchedule{
				OnStart:     tc.onStart,
				OnStop:      tc.onStop,
				OnDelete:    tc.onDelete,
				ScheduledAt: tc.scheduledAt,
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
				OnStart:     nil,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: nil,
			},
			expectedError:  job.ErrInvalidJobScheduleMissingRequiredParams,
			expectedResult: nil,
		},
		{
			description: "case 2: all fields are set",
			params: job.NewJobScheduleParams{
				OnStart:     &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:      &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnDelete:    &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				ScheduledAt: &DefaultInvalidScheduledAt,
			},
			expectedError:  job.ErrInvalidJobScheduleTooManyParams,
			expectedResult: nil,
		},
		{
			description: "case 3: invalid `on start` param",
			params: job.NewJobScheduleParams{
				OnStart:     &execution_command_test_helper.DefaultInvalidNewExecutionCommandParams,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: nil,
			},
			expectedError:  errors.Wrap(execution_command_test_helper.DefaultInvalidExecutionCommandParamsError, job.ErrInvalidJobScheduleOnStartParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 4: invalid `on stop` param",
			params: job.NewJobScheduleParams{
				OnStart:     nil,
				OnStop:      &execution_command_test_helper.DefaultInvalidNewExecutionCommandParams,
				OnDelete:    nil,
				ScheduledAt: nil,
			},
			expectedError:  errors.Wrap(execution_command_test_helper.DefaultInvalidExecutionCommandParamsError, job.ErrInvalidJobScheduleOnStopParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 5: invalid `on delete` param",
			params: job.NewJobScheduleParams{
				OnStart:     nil,
				OnStop:      nil,
				OnDelete:    &execution_command_test_helper.DefaultInvalidNewExecutionCommandParams,
				ScheduledAt: nil,
			},
			expectedError:  errors.Wrap(execution_command_test_helper.DefaultInvalidExecutionCommandParamsError, job.ErrInvalidJobScheduleOnDeleteParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 6: invalid `scheduled at` param",
			params: job.NewJobScheduleParams{
				OnStart:     nil,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: &DefaultInvalidScheduledAt,
			},
			expectedError:  job.ErrInvalidJobScheduleScheduledAtParam,
			expectedResult: nil,
		},
		{
			description: "case 7: valid `on start` param",
			params: job.NewJobScheduleParams{
				OnStart:     &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: nil,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:     &execution_command_test_helper.DefaultValidExecutionCommand,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: nil,
			},
		},
		{
			description: "case 8: valid `on stop` param",
			params: job.NewJobScheduleParams{
				OnStart:     nil,
				OnStop:      &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnDelete:    nil,
				ScheduledAt: nil,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:     nil,
				OnStop:      &execution_command_test_helper.DefaultValidExecutionCommand,
				OnDelete:    nil,
				ScheduledAt: nil,
			},
		},
		{
			description: "case 9: valid `on delete` param",
			params: job.NewJobScheduleParams{
				OnStart:     nil,
				OnStop:      nil,
				OnDelete:    &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				ScheduledAt: nil,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:     nil,
				OnStop:      nil,
				OnDelete:    &execution_command_test_helper.DefaultValidExecutionCommand,
				ScheduledAt: nil,
			},
		},
		{
			description: "case 10: valid `scheduled at` param",
			params: job.NewJobScheduleParams{
				OnStart:     nil,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: &DefaultScheduledAt,
			},
			expectedError: nil,
			expectedResult: &job.JobSchedule{
				OnStart:     nil,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: &DefaultScheduledAt,
			},
		},
		{
			description: "case 11: several valid exclusive fields set",
			params: job.NewJobScheduleParams{
				OnStart:     &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: &DefaultScheduledAt,
			},
			expectedError:  job.ErrInvalidJobScheduleTooManyParams,
			expectedResult: nil,
		},
		{
			description: "case 12: several invalid exclusive fields set",
			params: job.NewJobScheduleParams{
				OnStart:     &execution_command_test_helper.DefaultInvalidNewExecutionCommandParams,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: &DefaultInvalidScheduledAt,
			},
			expectedError:  errors.Wrap(execution_command_test_helper.DefaultInvalidExecutionCommandParamsError, job.ErrInvalidJobScheduleOnStartParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 13: several invalid & valid exclusive fields set", params: job.NewJobScheduleParams{
				OnStart:     &execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				OnStop:      nil,
				OnDelete:    nil,
				ScheduledAt: &DefaultInvalidScheduledAt,
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
