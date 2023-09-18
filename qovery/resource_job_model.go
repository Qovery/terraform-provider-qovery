package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

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
				Url:      FromString(j.Docker.GitRepository.Url),
				Branch:   FromStringPointer(j.Docker.GitRepository.Branch),
				RootPath: FromStringPointer(j.Docker.GitRepository.RootPath),
			},
			DockerFilePath: FromStringPointer(j.Docker.DockerFilePath),
		}
	}

	var img *Image = nil
	if j.Image != nil {
		img = &Image{
			RegistryID: FromString(j.Image.RegistryID),
			Name:       FromString(j.Image.Name),
			Tag:        FromString(j.Image.Tag),
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
		args := make([]string, len(s.OnStart.Arguments))
		for i, arg := range s.OnStart.Arguments {
			args[i] = ToString(arg)
		}
		onStart = &execution_command.ExecutionCommand{
			Entrypoint: ToStringPointer(s.OnStart.Entrypoint),
			Arguments:  args,
		}
	}

	var onStop *execution_command.ExecutionCommand = nil
	if s.OnStop != nil {
		args := make([]string, len(s.OnStop.Arguments))
		for i, arg := range s.OnStop.Arguments {
			args[i] = ToString(arg)
		}
		onStop = &execution_command.ExecutionCommand{
			Entrypoint: ToStringPointer(s.OnStop.Entrypoint),
			Arguments:  args,
		}
	}

	var onDelete *execution_command.ExecutionCommand = nil
	if s.OnDelete != nil {
		args := make([]string, len(s.OnDelete.Arguments))
		for i, arg := range s.OnDelete.Arguments {
			args[i] = ToString(arg)
		}
		onDelete = &execution_command.ExecutionCommand{
			Entrypoint: ToStringPointer(s.OnDelete.Entrypoint),
			Arguments:  args,
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
		args := make([]types.String, len(s.OnStart.Arguments))
		for i, arg := range s.OnStart.Arguments {
			args[i] = FromString(arg)
		}
		onStart = &ExecutionCommand{
			Entrypoint: FromStringPointer(s.OnStart.Entrypoint),
			Arguments:  args,
		}
	}

	var onStop *ExecutionCommand = nil
	if s.OnStop != nil {
		args := make([]types.String, len(s.OnStop.Arguments))
		for i, arg := range s.OnStop.Arguments {
			args[i] = FromString(arg)
		}
		onStop = &ExecutionCommand{
			Entrypoint: FromStringPointer(s.OnStop.Entrypoint),
			Arguments:  args,
		}
	}

	var onDelete *ExecutionCommand = nil
	if s.OnDelete != nil {
		args := make([]types.String, len(s.OnDelete.Arguments))
		for i, arg := range s.OnDelete.Arguments {
			args[i] = FromString(arg)
		}
		onDelete = &ExecutionCommand{
			Entrypoint: FromStringPointer(s.OnDelete.Entrypoint),
			Arguments:  args,
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
	Schedule types.String     `tfsdk:"schedule"`
}

func (s JobScheduleCron) toUpsertRequest() job.JobScheduleCron {
	args := make([]string, len(s.Command.Arguments))
	for i, arg := range s.Command.Arguments {
		args[i] = ToString(arg)
	}

	return job.JobScheduleCron{
		Command: execution_command.ExecutionCommand{
			Entrypoint: ToStringPointer(s.Command.Entrypoint),
			Arguments:  args,
		},
		Schedule: s.Schedule.Value,
	}
}

func JobScheduleCronFromDomainJobScheduleCron(s job.JobScheduleCron) JobScheduleCron {
	args := make([]types.String, len(s.Command.Arguments))
	for i, arg := range s.Command.Arguments {
		args[i] = FromString(arg)
	}

	return JobScheduleCron{
		Schedule: FromString(s.Schedule),
		Command: ExecutionCommand{
			Entrypoint: FromStringPointer(s.Command.Entrypoint),
			Arguments:  args,
		},
	}
}

type Job struct {
	ID                           types.String `tfsdk:"id"`
	EnvironmentID                types.String `tfsdk:"environment_id"`
	Name                         types.String `tfsdk:"name"`
	CPU                          types.Int64  `tfsdk:"cpu"`
	Memory                       types.Int64  `tfsdk:"memory"`
	MaxDurationSeconds           types.Int64  `tfsdk:"max_duration_seconds"`
	MaxNbRestart                 types.Int64  `tfsdk:"max_nb_restart"`
	AutoPreview                  types.Bool   `tfsdk:"auto_preview"`
	Source                       *JobSource   `tfsdk:"source"`
	Schedule                     *JobSchedule `tfsdk:"schedule"`
	HealthChecks                 HealthChecks `tfsdk:"healthchecks"`
	BuiltInEnvironmentVariables  types.Set    `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables         types.Set    `tfsdk:"environment_variables"`
	EnvironmentVariableAliases   types.Set    `tfsdk:"environment_variable_aliases"`
	EnvironmentVariableOverrides types.Set    `tfsdk:"environment_variable_overrides"`
	Secrets                      types.Set    `tfsdk:"secrets"`
	SecretAliases                types.Set    `tfsdk:"secret_aliases"`
	SecretOverrides              types.Set    `tfsdk:"secret_overrides"`
	Port                         types.Int64  `tfsdk:"port"`
	ExternalHost                 types.String `tfsdk:"external_host"`
	InternalHost                 types.String `tfsdk:"internal_host"`
	DeploymentStageId            types.String `tfsdk:"deployment_stage_id"`
	AdvancedSettingsJson         types.String `tfsdk:"advanced_settings_json"`
	AutoDeploy                   types.Bool   `tfsdk:"auto_deploy"`
}

func (j Job) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(j.EnvironmentVariables)
}
func (j Job) EnvironmentVariableAliasesList() EnvironmentVariableList {
	return toEnvironmentVariableList(j.EnvironmentVariableAliases)
}
func (j Job) EnvironmentVariableOverridesList() EnvironmentVariableList {
	return toEnvironmentVariableList(j.EnvironmentVariableOverrides)
}

func (j Job) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(j.BuiltInEnvironmentVariables)
}

func (j Job) SecretList() SecretList {
	return ToSecretList(j.Secrets)
}
func (j Job) SecretAliasesList() SecretList {
	return ToSecretList(j.SecretAliases)
}
func (j Job) SecretOverridesList() SecretList {
	return ToSecretList(j.SecretOverrides)
}

func (j Job) toUpsertServiceRequest(state *Job) (*job.UpsertServiceRequest, error) {
	var stateEnvironmentVariables EnvironmentVariableList
	var stateEnvironmentVariableAliases EnvironmentVariableList
	var stateEnvironmentVariableOverrides EnvironmentVariableList
	var stateSecrets SecretList
	var stateSecretAliases SecretList
	var stateSecretOverrides SecretList

	if state != nil {
		stateEnvironmentVariables = state.EnvironmentVariableList()
		stateEnvironmentVariableAliases = state.EnvironmentVariableAliasesList()
		stateEnvironmentVariableOverrides = state.EnvironmentVariableOverridesList()
		stateSecrets = state.SecretList()
		stateSecretAliases = state.SecretAliasesList()
		stateSecretOverrides = state.SecretOverridesList()
	}

	return &job.UpsertServiceRequest{
		JobUpsertRequest:             j.toUpsertRepositoryRequest(),
		EnvironmentVariables:         j.EnvironmentVariableList().diffRequest(stateEnvironmentVariables),
		EnvironmentVariableAliases:   j.EnvironmentVariableAliasesList().diffRequest(stateEnvironmentVariableAliases),
		EnvironmentVariableOverrides: j.EnvironmentVariableOverridesList().diffRequest(stateEnvironmentVariableOverrides),
		Secrets:                      j.SecretList().diffRequest(stateSecrets),
		SecretAliases:                j.SecretAliasesList().diffRequest(stateSecretAliases),
		SecretOverrides:              j.SecretOverridesList().diffRequest(stateSecretOverrides),
	}, nil
}

func (j Job) toUpsertRepositoryRequest() job.UpsertRepositoryRequest {
	return job.UpsertRepositoryRequest{
		Name:                 ToString(j.Name),
		AutoPreview:          ToBoolPointer(j.AutoPreview),
		CPU:                  ToInt32Pointer(j.CPU),
		Memory:               ToInt32Pointer(j.Memory),
		MaxNbRestart:         ToInt32Pointer(j.MaxNbRestart),
		MaxDurationSeconds:   ToInt32Pointer(j.MaxDurationSeconds),
		DeploymentStageID:    ToString(j.DeploymentStageId),
		Port:                 ToInt64Pointer(j.Port),
		Healthchecks:         j.HealthChecks.toHealthchecksRequest(),
		Source:               j.Source.toUpsertRequest(),
		Schedule:             j.Schedule.toUpsertRequest(),
		AdvancedSettingsJson: ToString(j.AdvancedSettingsJson),
		AutoDeploy:           *qovery.NewNullableBool(ToBoolPointer(j.AutoDeploy)),
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
		ID:                           FromString(job.ID.String()),
		EnvironmentID:                FromString(job.EnvironmentID.String()),
		Name:                         FromString(job.Name),
		CPU:                          FromInt32(job.CPU),
		Memory:                       FromInt32(job.Memory),
		MaxNbRestart:                 FromUInt32(job.MaxNbRestart),
		MaxDurationSeconds:           FromUInt32(job.MaxDurationSeconds),
		AutoPreview:                  FromBool(job.AutoPreview),
		Port:                         FromInt32Pointer(prt),
		Source:                       &source,
		Schedule:                     &schedule,
		BuiltInEnvironmentVariables:  convertDomainVariablesToEnvironmentVariableList(job.BuiltInEnvironmentVariables, variable.ScopeBuiltIn, "BUILT_IN").toTerraformSet(),
		EnvironmentVariables:         convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariables, job.EnvironmentVariables, variable.ScopeJob, "VALUE").toTerraformSet(),
		EnvironmentVariableAliases:   convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariableAliases, job.EnvironmentVariables, variable.ScopeJob, "ALIAS").toTerraformSet(),
		EnvironmentVariableOverrides: convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariableOverrides, job.EnvironmentVariables, variable.ScopeJob, "OVERRIDE").toTerraformSet(),
		Secrets:                      convertDomainSecretsToSecretList(state.Secrets, job.Secrets, variable.ScopeJob, "VALUE").toTerraformSet(),
		SecretAliases:                convertDomainSecretsToSecretList(state.SecretAliases, job.Secrets, variable.ScopeJob, "ALIAS").toTerraformSet(),
		SecretOverrides:              convertDomainSecretsToSecretList(state.SecretOverrides, job.Secrets, variable.ScopeJob, "OVERRIDE").toTerraformSet(),
		InternalHost:                 FromStringPointer(job.InternalHost),
		ExternalHost:                 FromStringPointer(job.ExternalHost),
		DeploymentStageId:            FromString(job.DeploymentStageID),
		HealthChecks:                 convertHealthchecksResponseToDomain(job.HealthChecks),
		AdvancedSettingsJson:         FromString(job.AdvancedSettingsJson),
		AutoDeploy:                   FromBoolPointer(job.AutoDeploy),
	}
}
