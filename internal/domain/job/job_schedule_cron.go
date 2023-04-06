package job

import (
	"github.com/adhocore/gronx"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
)

var (
	ErrInvalidJobScheduleCronCommandParam  = errors.New("invalid `command` param")
	ErrInvalidJobScheduleCronScheduleParam = errors.New("invalid `schedule` param")
)

type JobScheduleCron struct {
	Command  execution_command.ExecutionCommand
	Schedule string
}

func (c JobScheduleCron) Validate() error {
	gron := gronx.New()
	if !gron.IsValid(c.Schedule) {
		return errors.Wrap(errors.New("cron string format is invalid"), ErrInvalidJobScheduleCronScheduleParam.Error())
	}

	if err := c.Command.Validate(); err != nil {
		{
			return errors.Wrap(err, ErrInvalidJobScheduleOnStartParam.Error())
		}
		return ErrInvalidJobScheduleCronCommandParam
	}

	return nil
}

type NewJobScheduleCronParams struct {
	Command  execution_command.NewExecutionCommandParams
	Schedule string
}

func NewJobScheduleCron(params NewJobScheduleCronParams) (*JobScheduleCron, error) {
	command, err := execution_command.NewExecutionCommand(params.Command)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidJobScheduleCronCommandParam.Error())
	}

	newJobScheduleCron := &JobScheduleCron{
		Command:  *command,
		Schedule: params.Schedule,
	}

	if err := newJobScheduleCron.Validate(); err != nil {
		return nil, err
	}

	return newJobScheduleCron, nil
}
