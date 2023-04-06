package test_helper

import (
	"github.com/pkg/errors"
	execution_command_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/execution_command/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

var (
	DefaultJobScheduledCronCronString        = "*/30 * * * *"
	DefaultJobScheduledCronInvalidCronString = ""
	DefaultValidJobScheduledCronCronParams   = job.NewJobScheduleCronParams{
		Command:  execution_command_test_helper.DefaultValidNewExecutionCommandParams,
		Schedule: DefaultJobScheduledCronCronString,
	}
	DefaultValidJobScheduledCronCron = job.JobScheduleCron{
		Command:  execution_command_test_helper.DefaultValidExecutionCommand,
		Schedule: DefaultJobScheduledCronCronString,
	}
	DefaultInvalidJobScheduledCronCronParams = job.NewJobScheduleCronParams{
		Command:  execution_command_test_helper.DefaultValidNewExecutionCommandParams,
		Schedule: DefaultJobScheduledCronInvalidCronString,
	}
	DefaultInvalidJobScheduledCronCron = job.JobScheduleCron{
		Command:  execution_command_test_helper.DefaultValidExecutionCommand,
		Schedule: DefaultJobScheduledCronInvalidCronString,
	}
	DefaultInvalidNewInvalidJobScheduledCronCronParamsError = errors.Wrap(errors.New("cron string format is invalid"), job.ErrInvalidJobScheduleCronScheduleParam.Error())
)
