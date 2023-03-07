package execution_command_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command/test_helper"
)

func TestExecutionCommandValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description   string
		entrypoint    *string
		arguments     []string
		expectedError error
	}{
		{description: "case 1: entrypoint is nil", entrypoint: nil, arguments: test_helper.DefaultArguments, expectedError: nil},
		{description: "case 2: arguments is empty", entrypoint: &test_helper.DefaultEntrypoint, arguments: make([]string, 0), expectedError: execution_command.ErrInvalidArgumentsParam},
		{description: "case 3: all fields are set", entrypoint: &test_helper.DefaultEntrypoint, arguments: test_helper.DefaultArguments, expectedError: nil},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i := execution_command.ExecutionCommand{
				Entrypoint: tc.entrypoint,
				Arguments:  tc.arguments,
			}

			// verify:
			assert.Equal(t, tc.expectedError, i.Validate())
		})
	}
}

func TestNewExecutionCommand(t *testing.T) {
	// setup:
	testCases := []struct {
		description    string
		params         execution_command.NewExecutionCommandParams
		expectedResult *execution_command.ExecutionCommand
		expectedError  error
	}{
		{
			description: "case 1: all params blanks",
			params: execution_command.NewExecutionCommandParams{
				Entrypoint: nil,
				Arguments:  make([]string, 0),
			},
			expectedError:  execution_command.ErrInvalidArgumentsParam,
			expectedResult: nil,
		},
		{
			description: "case 2: entrypoint is nil",
			params: execution_command.NewExecutionCommandParams{
				Entrypoint: nil,
				Arguments:  test_helper.DefaultArguments,
			},
			expectedError: nil,
			expectedResult: &execution_command.ExecutionCommand{
				Entrypoint: nil,
				Arguments:  test_helper.DefaultArguments,
			},
		},
		{
			description: "case 3: arguments is empty",
			params: execution_command.NewExecutionCommandParams{
				Entrypoint: &test_helper.DefaultEntrypoint,
				Arguments:  make([]string, 0),
			},
			expectedError:  execution_command.ErrInvalidArgumentsParam,
			expectedResult: nil,
		},
		{
			description: "case 5: all properly set",
			params: execution_command.NewExecutionCommandParams{
				Entrypoint: &test_helper.DefaultEntrypoint,
				Arguments:  test_helper.DefaultArguments,
			},
			expectedError: nil,
			expectedResult: &execution_command.ExecutionCommand{
				Entrypoint: &test_helper.DefaultEntrypoint,
				Arguments:  test_helper.DefaultArguments,
			},
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i, err := execution_command.NewExecutionCommand(tc.params)

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
