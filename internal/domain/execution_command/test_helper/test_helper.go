package test_helper

import "github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"

var (
	DefaultArguments  = []string{"./app", "run"}
	DefaultEntrypoint = "/"

	/// Exposed to tests needing to get such object without having to know internal sauce magic
	DefaultValidNewExecutionCommandParams = execution_command.NewExecutionCommandParams{
		Arguments:  DefaultArguments,
		Entrypoint: &DefaultEntrypoint,
	}
	DefaultValidExecutionCommand = execution_command.ExecutionCommand{
		Arguments:  DefaultArguments,
		Entrypoint: &DefaultEntrypoint,
	}
)
