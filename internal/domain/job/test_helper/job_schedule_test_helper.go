package test_helper

import (
	"github.com/pkg/errors"
	execution_command_helper "github.com/qovery/terraform-provider-qovery/internal/domain/execution_command/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

var (
	DefaultJobSchedule = job.JobSchedule{
		OnStart:  &execution_command_helper.DefaultValidExecutionCommand,
		OnStop:   nil,
		OnDelete: nil,
		CronJob:  nil,
	}

	DefaultJobScheduleParams = job.NewJobScheduleParams{
		OnStart:  &execution_command_helper.DefaultValidNewExecutionCommandParams,
		OnStop:   nil,
		OnDelete: nil,
		CronJob:  nil,
	}

	DefaultInvalidJobSchedule = job.JobSchedule{
		OnStart:  nil,
		OnStop:   nil,
		OnDelete: nil,
		CronJob:  &DefaultInvalidJobScheduledCronCron,
	}

	DefaultInvalidJobScheduleParams = job.NewJobScheduleParams{
		OnStart:  nil,
		OnStop:   nil,
		OnDelete: nil,
		CronJob:  &DefaultInvalidJobScheduledCronCronParams,
	}

	DefaultInvalidJobScheduleParamsError = errors.Wrap(DefaultInvalidNewInvalidJobScheduledCronCronParamsError, job.ErrInvalidJobScheduleCronParam.Error())
)
