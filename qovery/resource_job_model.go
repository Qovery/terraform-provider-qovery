package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
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

func (s JobSource) toUpsertRequest() job.Source {
	var img *image.Image = nil
	if s.Image != nil {
		img = s.Image.toUpsertRequest()
	}

	var dkr *docker.Docker = nil
	if s.Docker != nil {
		dkr = s.Docker.toUpsertRequest()
	}

	return job.Source{
		Image:  img,
		Docker: dkr,
	}
}

func JobSourceFromDomainJobSource(j job.Source) JobSource {
	var dkr *Docker = nil
	if j.Docker != nil {
		dkr = &Docker{
			GitRepository: GitRepository{
				Url:        FromString(j.Docker.GitRepository.Url),
				Branch:     FromStringPointer(j.Docker.GitRepository.Branch),
				RootPath:   FromStringPointer(j.Docker.GitRepository.RootPath),
				GitTokenId: FromStringPointer(j.Docker.GitRepository.GitTokenId),
			},
			DockerFilePath:         FromStringPointer(j.Docker.DockerFilePath),
			DockerfileRaw:          FromStringPointer(j.Docker.DockerFileRaw),
			DockerTargetBuildStage: FromStringPointer(j.Docker.DockerTargetBuildStage),
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
	OnStart       *ExecutionCommand `tfsdk:"on_start"`
	OnStop        *ExecutionCommand `tfsdk:"on_stop"`
	OnDelete      *ExecutionCommand `tfsdk:"on_delete"`
	LifecycleType types.String      `tfsdk:"lifecycle_type"`
	CronJob       *JobScheduleCron  `tfsdk:"cronjob"`
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

	var lifecycleType *qovery.JobLifecycleTypeEnum = nil
	if !s.LifecycleType.IsNull() {
		lfType, _ := qovery.NewJobLifecycleTypeEnumFromValue(ToString(s.LifecycleType))
		lifecycleType = lfType
	}

	return job.JobSchedule{
		OnStart:       onStart,
		OnStop:        onStop,
		OnDelete:      onDelete,
		LifecycleType: lifecycleType,
		CronJob:       scheduledAt,
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

	var lifecycleType types.String
	if s.LifecycleType == nil {
		lifecycleType = types.StringNull()
	} else {
		lifecycleType = FromString(string(*s.LifecycleType))
	}
	return JobSchedule{
		OnStart:       onStart,
		OnStop:        onStop,
		OnDelete:      onDelete,
		LifecycleType: lifecycleType,
		CronJob:       cronJob,
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
		Schedule: s.Schedule.ValueString(),
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
	ID                           types.String  `tfsdk:"id"`
	EnvironmentID                types.String  `tfsdk:"environment_id"`
	Name                         types.String  `tfsdk:"name"`
	IconUri                      types.String  `tfsdk:"icon_uri"`
	CPU                          types.Int64   `tfsdk:"cpu"`
	Memory                       types.Int64   `tfsdk:"memory"`
	MaxDurationSeconds           types.Int64   `tfsdk:"max_duration_seconds"`
	MaxNbRestart                 types.Int64   `tfsdk:"max_nb_restart"`
	AutoPreview                  types.Bool    `tfsdk:"auto_preview"`
	Source                       *JobSource    `tfsdk:"source"`
	Schedule                     *JobSchedule  `tfsdk:"schedule"`
	HealthChecks                 *HealthChecks `tfsdk:"healthchecks"`
	BuiltInEnvironmentVariables  types.Set     `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables         types.Set     `tfsdk:"environment_variables"`
	EnvironmentVariableAliases   types.Set     `tfsdk:"environment_variable_aliases"`
	EnvironmentVariableOverrides types.Set     `tfsdk:"environment_variable_overrides"`
	Secrets                      types.Set     `tfsdk:"secrets"`
	SecretAliases                types.Set     `tfsdk:"secret_aliases"`
	SecretOverrides              types.Set     `tfsdk:"secret_overrides"`
	Port                         types.Int64   `tfsdk:"port"`
	ExternalHost                 types.String  `tfsdk:"external_host"`
	InternalHost                 types.String  `tfsdk:"internal_host"`
	DeploymentStageId            types.String  `tfsdk:"deployment_stage_id"`
	AdvancedSettingsJson         types.String  `tfsdk:"advanced_settings_json"`
	AutoDeploy                   types.Bool    `tfsdk:"auto_deploy"`
	DeploymentRestrictions       types.Set     `tfsdk:"deployment_restrictions"`
	AnnotationsGroupIds          types.Set     `tfsdk:"annotations_group_ids"`
	LabelssGroupIds              types.Set     `tfsdk:"labels_group_ids"`
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

func (j Job) DeploymentRestrictionDiff(deploymentRestrictionsState *types.Set) (*deploymentrestriction.ServiceDeploymentRestrictionsDiff, error) {
	return deploymentrestriction.ToDeploymentRestrictionDiff(j.DeploymentRestrictions, deploymentRestrictionsState)
}

func (j Job) toUpsertServiceRequest(state *Job) (*job.UpsertServiceRequest, error) {
	var stateEnvironmentVariables EnvironmentVariableList
	var stateEnvironmentVariableAliases EnvironmentVariableList
	var stateEnvironmentVariableOverrides EnvironmentVariableList
	var stateSecrets SecretList
	var stateSecretAliases SecretList
	var stateSecretOverrides SecretList
	var stateDeploymentRestrictions types.Set

	if state != nil {
		stateEnvironmentVariables = state.EnvironmentVariableList()
		stateEnvironmentVariableAliases = state.EnvironmentVariableAliasesList()
		stateEnvironmentVariableOverrides = state.EnvironmentVariableOverridesList()
		stateSecrets = state.SecretList()
		stateSecretAliases = state.SecretAliasesList()
		stateSecretOverrides = state.SecretOverridesList()
		stateDeploymentRestrictions = state.DeploymentRestrictions
	}

	deploymentRestrictionsDiff, err := j.DeploymentRestrictionDiff(&stateDeploymentRestrictions)
	if err != nil {
		return nil, err
	}

	return &job.UpsertServiceRequest{
		JobUpsertRequest:             j.toUpsertRepositoryRequest(),
		EnvironmentVariables:         j.EnvironmentVariableList().diffRequest(stateEnvironmentVariables),
		EnvironmentVariableAliases:   j.EnvironmentVariableAliasesList().diffRequest(stateEnvironmentVariableAliases),
		EnvironmentVariableOverrides: j.EnvironmentVariableOverridesList().diffRequest(stateEnvironmentVariableOverrides),
		Secrets:                      j.SecretList().diffRequest(stateSecrets),
		SecretAliases:                j.SecretAliasesList().diffRequest(stateSecretAliases),
		SecretOverrides:              j.SecretOverridesList().diffRequest(stateSecretOverrides),
		DeploymentRestrictionsDiff:   *deploymentRestrictionsDiff,
	}, nil
}

func (j Job) toUpsertRepositoryRequest() job.UpsertRepositoryRequest {
	annotationsGroupIds := make([]string, 0, len(j.AnnotationsGroupIds.Elements()))
	for _, id := range j.AnnotationsGroupIds.Elements() {
		id := id.(types.String)
		annotationsGroupIds = append(annotationsGroupIds, id.ValueString())
	}

	labelsGroupIds := make([]string, 0, len(j.LabelssGroupIds.Elements()))
	for _, id := range j.LabelssGroupIds.Elements() {
		id := id.(types.String)
		labelsGroupIds = append(labelsGroupIds, id.ValueString())
	}

	return job.UpsertRepositoryRequest{
		Name:                 ToString(j.Name),
		IconUri:              ToStringPointer(j.IconUri),
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
		AnnotationsGroupIds:  annotationsGroupIds,
		LabelsGroupIds:       labelsGroupIds,
	}
}

func convertDomainJobToJob(ctx context.Context, state Job, job *job.Job) Job {
	var prt *int32 = nil
	if job.Port != nil {
		prt = &job.Port.InternalPort
	}

	source := JobSourceFromDomainJobSource(job.Source)
	schedule := JobScheduleFromDomainJobSchedule(job.Schedule)

	healthchecks := convertHealthchecksResponseToDomain(job.HealthChecks)
	return Job{
		ID:                           FromString(job.ID.String()),
		EnvironmentID:                FromString(job.EnvironmentID.String()),
		Name:                         FromString(job.Name),
		IconUri:                      FromString(job.IconUri),
		CPU:                          FromInt32(job.CPU),
		Memory:                       FromInt32(job.Memory),
		MaxNbRestart:                 FromInt32(job.MaxNbRestart),
		MaxDurationSeconds:           FromInt32(job.MaxDurationSeconds),
		AutoPreview:                  FromBool(job.AutoPreview),
		Port:                         FromInt32Pointer(prt),
		Source:                       &source,
		Schedule:                     &schedule,
		BuiltInEnvironmentVariables:  convertDomainVariablesToEnvironmentVariableList(job.BuiltInEnvironmentVariables, variable.ScopeBuiltIn, "BUILT_IN").toTerraformSet(ctx),
		EnvironmentVariables:         convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariables, job.EnvironmentVariables, variable.ScopeJob, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:   convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableAliases, job.EnvironmentVariables, variable.ScopeJob, "ALIAS").toTerraformSet(ctx),
		EnvironmentVariableOverrides: convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableOverrides, job.EnvironmentVariables, variable.ScopeJob, "OVERRIDE").toTerraformSet(ctx),
		Secrets:                      convertDomainSecretsToSecretList(state.Secrets, job.Secrets, variable.ScopeJob, "VALUE").toTerraformSet(ctx),
		SecretAliases:                convertDomainSecretsToSecretList(state.SecretAliases, job.Secrets, variable.ScopeJob, "ALIAS").toTerraformSet(ctx),
		SecretOverrides:              convertDomainSecretsToSecretList(state.SecretOverrides, job.Secrets, variable.ScopeJob, "OVERRIDE").toTerraformSet(ctx),
		InternalHost:                 FromStringPointer(job.InternalHost),
		ExternalHost:                 FromStringPointer(job.ExternalHost),
		DeploymentStageId:            FromString(job.DeploymentStageID),
		HealthChecks:                 &healthchecks,
		AdvancedSettingsJson:         FromString(job.AdvancedSettingsJson),
		AutoDeploy:                   FromBoolPointer(job.AutoDeploy),
		DeploymentRestrictions:       FromDeploymentRestrictionList(state.DeploymentRestrictions, job.JobDeploymentRestrictions),
		AnnotationsGroupIds:          fromAnnotationsGroupList(ctx, state.AnnotationsGroupIds, job.AnnotationsGroupIds),
		LabelssGroupIds:              fromLabelsGroupList(ctx, state.LabelssGroupIds, job.LabelsGroupIds),
	}
}
