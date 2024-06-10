package job_test

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	job_helper "github.com/qovery/terraform-provider-qovery/internal/domain/job/test_helper"
	port_helper "github.com/qovery/terraform-provider-qovery/internal/domain/port/test_helper"
	secret_helper "github.com/qovery/terraform-provider-qovery/internal/domain/secret/test_helper"
	variable_helper "github.com/qovery/terraform-provider-qovery/internal/domain/variable/test_helper"
)

func TestJobValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description          string
		name                 string
		cpu                  int32
		memory               int32
		maxNbRestart         uint32
		maxDurationSeconds   uint32
		autoPreview          bool
		schedule             job.JobSchedule
		source               job.Source
		environmentVariables []variable.Variable
		environmentSecrets   []secret.Secret
		port                 *port.Port
		expectedError        error
	}{
		{
			description:          "case 1: nominal case, all fields valid",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        nil,
		},
		{
			description:          "case 2: invalid job name",
			name:                 job_helper.DefaultInvalidJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        job.ErrInvalidJobNameParam,
		},
		{
			description:          "case 3: invalid job CPU (too low)",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultInvalidJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        job.ErrInvalidJobCPUTooLowParam,
		},
		{
			description:          "case 4: invalid job memory (too low)",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultInvalidJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        job.ErrInvalidJobMemoryTooLowParam,
		},
		{
			description:          "case 5: invalid job source",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultInvalidJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        errors.Wrap(job_helper.DefaultInvalidNewJobSourceParamsError, job.ErrInvalidJobSourceParam.Error()),
		},
		{
			description:          "case 6: invalid job schedule",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultInvalidJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        errors.Wrap(job_helper.DefaultInvalidJobScheduleParamsError, job.ErrInvalidJobScheduleParam.Error()),
		},
		{
			description:          "case 7: invalid environment variable",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultInvalidJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        errors.Wrap(variable_helper.DefaultInvalidVariableParamsError, job.ErrInvalidJobEnvironmentVariablesParam.Error()),
		},
		{
			description:          "case 8: invalid secret",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultInvalidJobEnvironmentSecrets,
			expectedError:        errors.Wrap(secret_helper.DefaultInvalidSecretParamsError, job.ErrInvalidJobSecretsParam.Error()),
		},
		{
			description:          "case 9: invalid port",
			name:                 job_helper.DefaultJobName,
			cpu:                  job_helper.DefaultJobCPU,
			memory:               job_helper.DefaultJobMemory,
			maxNbRestart:         job_helper.DefaultJobMaxNbRestart,
			maxDurationSeconds:   job_helper.DefaultJobMaxDurationSeconds,
			autoPreview:          job_helper.DefaultJobAutoPreview,
			port:                 job_helper.DefaultInvalidJobPort,
			source:               job_helper.DefaultJobSource,
			schedule:             job_helper.DefaultJobSchedule,
			environmentVariables: job_helper.DefaultJobEnvironmentVariables,
			environmentSecrets:   job_helper.DefaultJobEnvironmentSecrets,
			expectedError:        errors.Wrap(port_helper.DefaultInvalidPortParamsError, job.ErrInvalidJobPortParam.Error()),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			s := job.Job{
				ID:                   job_helper.DefaultJobID,
				EnvironmentID:        job_helper.DefaultJobEnvironmentID,
				Name:                 tc.name,
				CPU:                  tc.cpu,
				MaxNbRestart:         int32(tc.maxNbRestart),
				MaxDurationSeconds:   int32(tc.maxDurationSeconds),
				Memory:               tc.memory,
				Schedule:             tc.schedule,
				Source:               tc.source,
				EnvironmentVariables: tc.environmentVariables,
				Secrets:              tc.environmentSecrets,
				Port:                 tc.port,
				DeploymentStageID:    job_helper.DefaultJobDeploymentStageID.String(),
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

// TODO(benjaminch): Implement NewJob test here
// func TestNewJob(t *testing.T) {}
