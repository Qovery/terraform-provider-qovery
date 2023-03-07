package job

import (
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
)

var (
	ErrInvalidJobScheduleOnStartParam          = errors.New("invalid `on start` param")
	ErrInvalidJobScheduleOnStopParam           = errors.New("invalid `on stop` param")
	ErrInvalidJobScheduleOnDeleteParam         = errors.New("invalid `on delete` param")
	ErrInvalidJobScheduleScheduledAtParam      = errors.New("invalid `scheduled at` param, cannot be blank")
	ErrInvalidJobScheduleMissingRequiredParams = errors.New("invalid job schedule: at least one of `OnStart`,  `OnStop`, `OnDelete` or `ScheduledAt` should be set")
	ErrInvalidJobScheduleTooManyParams         = errors.New("invalid job schedule: only one of `OnStart`,  `OnStop`, `OnDelete` or `ScheduledAt` should be set")
)

type JobSchedule struct {
	OnStart     *execution_command.ExecutionCommand
	OnStop      *execution_command.ExecutionCommand
	OnDelete    *execution_command.ExecutionCommand
	ScheduledAt *string
}

func (s JobSchedule) Validate() error {
	if (s.OnStart != nil && (s.OnStop != nil || s.OnDelete != nil || s.ScheduledAt != nil)) ||
		(s.OnStop != nil && (s.OnStart != nil || s.OnDelete != nil || s.ScheduledAt != nil)) ||
		(s.OnDelete != nil && (s.OnStart != nil || s.OnStop != nil || s.ScheduledAt != nil)) ||
		(s.ScheduledAt != nil && (s.OnStart != nil || s.OnStop != nil || s.OnDelete != nil)) {
		return ErrInvalidJobScheduleTooManyParams
	}

	if s.OnStart == nil && s.OnStop == nil && s.OnDelete == nil && s.ScheduledAt == nil {
		return ErrInvalidJobScheduleMissingRequiredParams
	}

	if s.OnStart != nil {
		if err := s.OnStart.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidJobScheduleOnStartParam.Error())
		}
	}

	if s.OnStop != nil {
		if err := s.OnStop.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidJobScheduleOnStopParam.Error())
		}
	}

	if s.OnDelete != nil {
		if err := s.OnDelete.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidJobScheduleOnDeleteParam.Error())
		}
	}

	if s.ScheduledAt != nil && *s.ScheduledAt == "" {
		// TODO(benjaminch): validate cron format
		return ErrInvalidJobScheduleScheduledAtParam
	}

	return nil
}

type NewJobScheduleParams struct {
	OnStart     *execution_command.NewExecutionCommandParams
	OnStop      *execution_command.NewExecutionCommandParams
	OnDelete    *execution_command.NewExecutionCommandParams
	ScheduledAt *string
}

func NewJobSchedule(params NewJobScheduleParams) (*JobSchedule, error) {
	var err error = nil

	var onStart *execution_command.ExecutionCommand = nil
	if params.OnStart != nil {
		onStart, err = execution_command.NewExecutionCommand(*params.OnStart)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobScheduleOnStartParam.Error())
		}
	}

	var onStop *execution_command.ExecutionCommand = nil
	if params.OnStop != nil {
		onStop, err = execution_command.NewExecutionCommand(*params.OnStop)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobScheduleOnStopParam.Error())
		}
	}

	var onDelete *execution_command.ExecutionCommand = nil
	if params.OnDelete != nil {
		onDelete, err = execution_command.NewExecutionCommand(*params.OnDelete)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobScheduleOnDeleteParam.Error())
		}
	}

	// TODO(benjaminch): validate cron format

	newSchedule := &JobSchedule{
		OnStart:     onStart,
		OnStop:      onStop,
		OnDelete:    onDelete,
		ScheduledAt: params.ScheduledAt,
	}

	if err := newSchedule.Validate(); err != nil {
		return nil, err
	}

	return newSchedule, nil
}
