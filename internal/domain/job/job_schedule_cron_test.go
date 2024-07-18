package job_test

import (
	"testing"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	execution_command_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/execution_command/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job/test_helper"
)

func TestJobScheduleCronValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description   string
		command       execution_command.ExecutionCommand
		schedule      string
		expectedError error
	}{
		{description: "case 1: all fields are set", command: execution_command_test_helper.DefaultValidExecutionCommand, schedule: test_helper.DefaultJobScheduledCronCronString, expectedError: nil},
		{description: "case 2: schedule cron string is invalid", command: execution_command_test_helper.DefaultValidExecutionCommand, schedule: test_helper.DefaultJobScheduledCronInvalidCronString, expectedError: test_helper.DefaultInvalidNewInvalidJobScheduledCronCronParamsError},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			s := job.JobScheduleCron{
				Command:  tc.command,
				Schedule: tc.schedule,
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

func TestNewJobScheduleCron(t *testing.T) {
	// setup:
	testCases := []struct {
		description    string
		params         job.NewJobScheduleCronParams
		expectedResult *job.JobScheduleCron
		expectedError  error
	}{
		{
			description: "case 1: all fields are set",
			params: job.NewJobScheduleCronParams{
				Command:  execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				Schedule: test_helper.DefaultJobScheduledCronCronString,
			},
			expectedResult: &job.JobScheduleCron{
				Command:  execution_command_test_helper.DefaultValidExecutionCommand,
				Schedule: test_helper.DefaultJobScheduledCronCronString,
			},
			expectedError: nil,
		},
		{
			description: "case 2: schedule cron string is invalid",
			params: job.NewJobScheduleCronParams{
				Command:  execution_command_test_helper.DefaultValidNewExecutionCommandParams,
				Schedule: test_helper.DefaultJobScheduledCronInvalidCronString,
			},
			expectedResult: nil,
			expectedError:  test_helper.DefaultInvalidNewInvalidJobScheduledCronCronParamsError,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i, err := job.NewJobScheduleCron(tc.params)

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
