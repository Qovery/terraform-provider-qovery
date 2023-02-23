package execution_command

import (
	"github.com/pkg/errors"
)

var (
	// ErrInvalidArgumentsParam is returned if the URL param is invalid.
	ErrInvalidArgumentsParam = errors.New("invalid arguments param")
)

type ExecutionCommand struct {
	Entrypoint *string
	Arguments  []string `validate:"required"`
}

func (e ExecutionCommand) Validate() error {
	if len(e.Arguments) == 0 {
		return ErrInvalidArgumentsParam
	}

	return nil
}

type NewExecutionCommandParams struct {
	Entrypoint *string
	Arguments  []string
}

func NewExecutionCommand(params NewExecutionCommandParams) (*ExecutionCommand, error) {
	newExecutionCommand := &ExecutionCommand{
		Entrypoint: params.Entrypoint,
		Arguments:  params.Arguments,
	}

	if err := newExecutionCommand.Validate(); err != nil {
		return nil, err
	}

	return newExecutionCommand, nil
}
