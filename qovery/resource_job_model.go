package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type JobSource struct {
	Image  *Image  `tfsdk:"image"`
	Docker *Docker `tfsdk:"docker"`
}

func (s JobSource) toUpsertRequest() job.JobSource {
	var img *image.Image = nil
	if s.Image != nil {
		img = s.Image.toUpsertRequest()
	}

	var dkr *docker.Docker = nil
	if s.Docker != nil {
		dkr = s.Docker.toUpsertRequest()
	}

	return job.JobSource{
		Image:  img,
		Docker: dkr,
	}
}

func JobSourceFromDomainJobSource(j job.JobSource) JobSource {
	var dkr *Docker = nil
	if j.Docker != nil {
		dkr = &Docker{
			GitRepository: GitRepository{
				Url:      j.Docker.GitRepository.Url,
				Branch:   j.Docker.GitRepository.Branch,
				RootPath: j.Docker.GitRepository.RootPath,
			},
			DockerFilePath: j.Docker.DockerFilePath,
		}
	}

	var img *Image = nil
	if j.Image != nil {
		img = &Image{
			RegistryID: fromString(j.Image.RegistryID),
			Name:       fromString(j.Image.Name),
			Tag:        fromString(j.Image.Tag),
		}
	}

	return JobSource{
		Docker: dkr,
		Image:  img,
	}
}

type JobSchedule struct {
	OnStart  *ExecutionCommand `tfsdk:"on_start"`
	OnStop   *ExecutionCommand `tfsdk:"on_stop"`
	OnDelete *ExecutionCommand `tfsdk:"on_delete"`
	CronJob  *JobScheduleCron  `tfsdk:"cronjob"`
}

func (s JobSchedule) toUpsertRequest() job.JobSchedule {
	var onStart *execution_command.ExecutionCommand = nil
	if s.OnStart != nil {
		onStart = &execution_command.ExecutionCommand{
			Entrypoint: s.OnStart.Entrypoint,
			Arguments:  s.OnStart.Arguments,
		}
	}

	var onStop *execution_command.ExecutionCommand = nil
	if s.OnStop != nil {
		onStop = &execution_command.ExecutionCommand{
			Entrypoint: s.OnStop.Entrypoint,
			Arguments:  s.OnStop.Arguments,
		}
	}

	var onDelete *execution_command.ExecutionCommand = nil
	if s.OnDelete != nil {
		onDelete = &execution_command.ExecutionCommand{
			Entrypoint: s.OnDelete.Entrypoint,
			Arguments:  s.OnDelete.Arguments,
		}
	}

	var scheduledAt *job.JobScheduleCron = nil
	if s.CronJob != nil {
		s := s.CronJob.toUpsertRequest()
		scheduledAt = &s
	}

	return job.JobSchedule{
		OnStart:  onStart,
		OnStop:   onStop,
		OnDelete: onDelete,
		CronJob:  scheduledAt,
	}
}

func JobScheduleFromDomainJobSchedule(s job.JobSchedule) JobSchedule {
	var onStart *ExecutionCommand = nil
	if s.OnStart != nil {
		onStart = &ExecutionCommand{
			Entrypoint: s.OnStart.Entrypoint,
			Arguments:  s.OnStart.Arguments,
		}
	}

	var onStop *ExecutionCommand = nil
	if s.OnStop != nil {
		onStop = &ExecutionCommand{
			Entrypoint: s.OnStop.Entrypoint,
			Arguments:  s.OnStop.Arguments,
		}
	}

	var onDelete *ExecutionCommand = nil
	if s.OnDelete != nil {
		onDelete = &ExecutionCommand{
			Entrypoint: s.OnDelete.Entrypoint,
			Arguments:  s.OnDelete.Arguments,
		}
	}

	var cronJob *JobScheduleCron = nil
	if s.CronJob != nil {
		c := JobScheduleCronFromDomainJobScheduleCron(*s.CronJob)
		cronJob = &c
	}

	return JobSchedule{
		OnStart:  onStart,
		OnStop:   onStop,
		OnDelete: onDelete,
		CronJob:  cronJob,
	}
}

type JobScheduleCron struct {
	Command  ExecutionCommand `tfsdk:"command"`
	Schedule string           `tfsdk:"schedule"`
}

func (s JobScheduleCron) toUpsertRequest() job.JobScheduleCron {
	return job.JobScheduleCron{
		Command: execution_command.ExecutionCommand{
			Entrypoint: s.Command.Entrypoint,
			Arguments:  s.Command.Arguments,
		},
		Schedule: s.Schedule,
	}
}

func JobScheduleCronFromDomainJobScheduleCron(s job.JobScheduleCron) JobScheduleCron {
	return JobScheduleCron{
		Schedule: s.Schedule,
		Command: ExecutionCommand{
			Entrypoint: s.Command.Entrypoint,
			Arguments:  s.Command.Arguments,
		},
	}
}

type Job struct {
	ID                 types.String `tfsdk:"id"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	Name               types.String `tfsdk:"name"`
	CPU                types.Int64  `tfsdk:"cpu"`
	Memory             types.Int64  `tfsdk:"memory"`
	MaxDurationSeconds types.Int64  `tfsdk:"max_duration_seconds"`
	MaxNbRestart       types.Int64  `tfsdk:"max_nb_restart"`
	AutoPreview        types.Bool   `tfsdk:"auto_preview"`

	Source   *JobSource   `tfsdk:"source"`
	Schedule *JobSchedule `tfsdk:"schedule"`

	BuiltInEnvironmentVariables types.Set    `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables        types.Set    `tfsdk:"environment_variables"`
	Secrets                     types.Set    `tfsdk:"secrets"`
	Port                        types.Int64  `tfsdk:"port"`
	ExternalHost                types.String `tfsdk:"external_host"`
	InternalHost                types.String `tfsdk:"internal_host"`
	DeploymentStageId           types.String `tfsdk:"deployment_stage_id"`
}

func (j Job) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(j.EnvironmentVariables)
}

func (j Job) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(j.BuiltInEnvironmentVariables)
}

func (j Job) SecretList() SecretList {
	return toSecretList(j.Secrets)
}

func (j Job) toUpsertServiceRequest(state *Job) (*job.UpsertServiceRequest, error) {
	var stateEnvironmentVariables EnvironmentVariableList
	if state != nil {
		stateEnvironmentVariables = state.EnvironmentVariableList()
	}

	var stateSecrets SecretList
	if state != nil {
		stateSecrets = state.SecretList()
	}

	return &job.UpsertServiceRequest{
		JobUpsertRequest:     j.toUpsertRepositoryRequest(),
		EnvironmentVariables: j.EnvironmentVariableList().diffRequest(stateEnvironmentVariables),
		Secrets:              j.SecretList().diffRequest(stateSecrets),
	}, nil
}

func (j Job) toUpsertRepositoryRequest() job.UpsertRepositoryRequest {
	return job.UpsertRepositoryRequest{
		Name:               toString(j.Name),
		AutoPreview:        toBoolPointer(j.AutoPreview),
		CPU:                toInt32Pointer(j.CPU),
		Memory:             toInt32Pointer(j.Memory),
		MaxNbRestart:       toInt32Pointer(j.MaxNbRestart),
		MaxDurationSeconds: toInt32Pointer(j.MaxDurationSeconds),
		DeploymentStageID:  toString(j.DeploymentStageId),
		Port:               toInt64Pointer(j.Port),

		Source:   j.Source.toUpsertRequest(),
		Schedule: j.Schedule.toUpsertRequest(),
	}
}

func convertDomainJobToJob(state Job, job *job.Job) Job {
	var prt *int32 = nil
	if job.Port != nil {
		prt = &job.Port.InternalPort
	}

	source := JobSourceFromDomainJobSource(job.Source)
	schedule := JobScheduleFromDomainJobSchedule(job.Schedule)

	return Job{
		ID:                          fromString(job.ID.String()),
		EnvironmentID:               fromString(job.EnvironmentID.String()),
		Name:                        fromString(job.Name),
		CPU:                         fromInt32(job.CPU),
		Memory:                      fromInt32(job.Memory),
		MaxNbRestart:                fromUInt32(job.MaxNbRestart),
		MaxDurationSeconds:          fromUInt32(job.MaxDurationSeconds),
		AutoPreview:                 fromBool(job.AutoPreview),
		Port:                        fromInt32Pointer(prt),
		Source:                      &source,
		Schedule:                    &schedule,
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableList(job.EnvironmentVariables, variable.ScopeJob).toTerraformSet(),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(job.BuiltInEnvironmentVariables, variable.ScopeBuiltIn).toTerraformSet(),
		InternalHost:                fromStringPointer(job.InternalHost),
		ExternalHost:                fromStringPointer(job.ExternalHost),
		Secrets:                     convertDomainSecretsToSecretList(state.SecretList(), job.Secrets, variable.ScopeJob).toTerraformSet(),
		DeploymentStageId:           fromString(job.DeploymentStageID),
	}
}
