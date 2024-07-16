package job

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
)

var (
	ErrInvalidJobScheduleOnStartParam          = errors.New("invalid `on start` param")
	ErrInvalidJobScheduleOnStopParam           = errors.New("invalid `on stop` param")
	ErrInvalidJobScheduleOnDeleteParam         = errors.New("invalid `on delete` param")
	ErrInvalidJobScheduleCronParam             = errors.New("invalid `cron job` param")
	ErrInvalidJobLifecycleType                 = errors.New("invalid `cron job` lifecycle type cannot be set for cron schedule")
	ErrInvalidJobScheduleMissingRequiredParams = errors.New("invalid job schedule: at least one of `OnStart`, `OnStop`, `OnDelete` or `CronJob` should be set")
	ErrInvalidJobScheduleWrongScheduleParams   = errors.New("invalid job schedule: either `CronJob` OR at least one of `OnStart`, `OnStop`, `OnDelete` or should be set")
)

type JobSchedule struct {
	OnStart       *execution_command.ExecutionCommand
	OnStop        *execution_command.ExecutionCommand
	OnDelete      *execution_command.ExecutionCommand
	LifecycleType *qovery.JobLifecycleTypeEnum
	CronJob       *JobScheduleCron
}

func (s JobSchedule) Validate() error {
	// Either CronJob OR at least one of OnStart, OnStop, OnDelete
	if s.CronJob != nil && (s.OnStart != nil || s.OnStop != nil || s.OnDelete != nil) {
		return ErrInvalidJobScheduleWrongScheduleParams
	}

	if s.OnStart == nil && s.OnStop == nil && s.OnDelete == nil && s.CronJob == nil {
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

	if s.CronJob != nil {
		if err := s.CronJob.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidJobScheduleCronParam.Error())
		}

		if s.LifecycleType != nil {
			return ErrInvalidJobLifecycleType
		}
	}

	return nil
}

type NewJobScheduleParams struct {
	OnStart       *execution_command.NewExecutionCommandParams
	OnStop        *execution_command.NewExecutionCommandParams
	OnDelete      *execution_command.NewExecutionCommandParams
	LifecycleType *qovery.JobLifecycleTypeEnum
	CronJob       *NewJobScheduleCronParams
}

func NewJobSchedule(params NewJobScheduleParams) (*JobSchedule, error) {
	var err error

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

	var cronJob *JobScheduleCron = nil
	if params.CronJob != nil {
		cronJob, err = NewJobScheduleCron(*params.CronJob)
		if err != nil {
			return nil, errors.Wrap(err, ErrInvalidJobScheduleCronParam.Error())
		}

		if params.LifecycleType != nil {
			return nil, ErrInvalidJobLifecycleType
		}
	}

	newSchedule := &JobSchedule{
		OnStart:       onStart,
		OnStop:        onStop,
		OnDelete:      onDelete,
		LifecycleType: params.LifecycleType,
		CronJob:       cronJob,
	}

	if err := newSchedule.Validate(); err != nil {
		return nil, err
	}

	return newSchedule, nil
}
