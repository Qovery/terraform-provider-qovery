package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"time"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	image "github.com/qovery/terraform-provider-qovery/internal/domain/image"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type AggregateJobResponse struct {
	Id                  string
	EnvironmentId       string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	MaximumCpu          int32
	MaximumMemory       int32
	Name                string
	Description         *string
	Cpu                 int32
	Memory              int32
	MaxNbRestart        *int32
	MaxDurationSeconds  *int32
	AutoPreview         bool
	Port                qovery.NullableInt32
	Source              job.SourceResponse
	Healthchecks        qovery.Healthcheck
	AutoDeploy          *bool
	JobType             string
	ScheduleCron        *qovery.CronJobResponseAllOfSchedule
	ScheduleLifecycle   *qovery.LifecycleJobResponseAllOfSchedule
	AnnotationsGroupIds []string
	LabelsGroupIds      []string
}

func getAggregateJobResponse(jobResponse *qovery.JobResponse) AggregateJobResponse {
	source := job.SourceResponse{}
	if jobResponse.CronJobResponse != nil {
		if jobResponse.CronJobResponse.Source.BaseJobResponseAllOfSourceOneOf != nil {
			source.Image = &jobResponse.CronJobResponse.Source.BaseJobResponseAllOfSourceOneOf.Image
		} else if jobResponse.CronJobResponse.Source.BaseJobResponseAllOfSourceOneOf1 != nil {
			source.Docker = &jobResponse.CronJobResponse.Source.BaseJobResponseAllOfSourceOneOf1.Docker
		}

		var annotations = make([]string, 0, len(jobResponse.CronJobResponse.AnnotationsGroups))
		for _, v := range jobResponse.CronJobResponse.AnnotationsGroups {
			annotations = append(annotations, v.Id)
		}

		var labels = make([]string, 0, len(jobResponse.CronJobResponse.LabelsGroups))
		for _, v := range jobResponse.CronJobResponse.LabelsGroups {
			labels = append(labels, v.Id)
		}

		return AggregateJobResponse{
			Id:                  jobResponse.CronJobResponse.Id,
			EnvironmentId:       jobResponse.CronJobResponse.Environment.Id,
			CreatedAt:           jobResponse.CronJobResponse.CreatedAt,
			UpdatedAt:           jobResponse.CronJobResponse.UpdatedAt,
			MaximumCpu:          jobResponse.CronJobResponse.MaximumCpu,
			MaximumMemory:       jobResponse.CronJobResponse.MaximumMemory,
			Name:                jobResponse.CronJobResponse.Name,
			Description:         jobResponse.CronJobResponse.Description,
			Cpu:                 jobResponse.CronJobResponse.Cpu,
			Memory:              jobResponse.CronJobResponse.Memory,
			MaxNbRestart:        jobResponse.CronJobResponse.MaxNbRestart,
			MaxDurationSeconds:  jobResponse.CronJobResponse.MaxDurationSeconds,
			AutoPreview:         jobResponse.CronJobResponse.AutoPreview,
			Port:                jobResponse.CronJobResponse.Port,
			Source:              source,
			Healthchecks:        jobResponse.CronJobResponse.Healthchecks,
			AutoDeploy:          jobResponse.CronJobResponse.AutoDeploy,
			ScheduleLifecycle:   nil,
			ScheduleCron:        &jobResponse.CronJobResponse.Schedule,
			AnnotationsGroupIds: annotations,
			LabelsGroupIds:      labels,
		}
	} else {
		if jobResponse.LifecycleJobResponse.Source.BaseJobResponseAllOfSourceOneOf != nil {
			source.Image = &jobResponse.LifecycleJobResponse.Source.BaseJobResponseAllOfSourceOneOf.Image
		} else if jobResponse.LifecycleJobResponse.Source.BaseJobResponseAllOfSourceOneOf1 != nil {
			source.Docker = &jobResponse.LifecycleJobResponse.Source.BaseJobResponseAllOfSourceOneOf1.Docker
		}

		var annotations = make([]string, 0, len(jobResponse.LifecycleJobResponse.AnnotationsGroups))
		for _, v := range jobResponse.LifecycleJobResponse.AnnotationsGroups {
			annotations = append(annotations, v.Id)
		}

		var labels = make([]string, 0, len(jobResponse.LifecycleJobResponse.LabelsGroups))
		for _, v := range jobResponse.LifecycleJobResponse.LabelsGroups {
			labels = append(labels, v.Id)
		}

		return AggregateJobResponse{
			Id:                  jobResponse.LifecycleJobResponse.Id,
			EnvironmentId:       jobResponse.LifecycleJobResponse.Environment.Id,
			CreatedAt:           jobResponse.LifecycleJobResponse.CreatedAt,
			UpdatedAt:           jobResponse.LifecycleJobResponse.UpdatedAt,
			MaximumCpu:          jobResponse.LifecycleJobResponse.MaximumCpu,
			MaximumMemory:       jobResponse.LifecycleJobResponse.MaximumMemory,
			Name:                jobResponse.LifecycleJobResponse.Name,
			Description:         jobResponse.LifecycleJobResponse.Description,
			Cpu:                 jobResponse.LifecycleJobResponse.Cpu,
			Memory:              jobResponse.LifecycleJobResponse.Memory,
			MaxNbRestart:        jobResponse.LifecycleJobResponse.MaxNbRestart,
			MaxDurationSeconds:  jobResponse.LifecycleJobResponse.MaxDurationSeconds,
			AutoPreview:         jobResponse.LifecycleJobResponse.AutoPreview,
			Port:                jobResponse.LifecycleJobResponse.Port,
			Source:              source,
			Healthchecks:        jobResponse.LifecycleJobResponse.Healthchecks,
			AutoDeploy:          jobResponse.LifecycleJobResponse.AutoDeploy,
			ScheduleCron:        nil,
			ScheduleLifecycle:   &jobResponse.LifecycleJobResponse.Schedule,
			AnnotationsGroupIds: annotations,
			LabelsGroupIds:      labels,
		}
	}
}

// newDomainCredentialsFromQovery takes a qovery.EnvironmentVariable returned by the API client and turns it into the domain model variable.Variable.
func newDomainJobFromQovery(jobResponse *qovery.JobResponse, deploymentStageID string, advancedSettingsJson string) (*job.Job, error) {
	if jobResponse == nil {
		return nil, variable.ErrNilVariable
	}

	var j = getAggregateJobResponse(jobResponse)

	var prt *port.NewPortParams = nil
	if j.Port.IsSet() {
		if rawPort := j.Port.Get(); rawPort != nil {
			prt = &port.NewPortParams{
				PortID:             uuid.New().String(),
				InternalPort:       *rawPort,
				PubliclyAccessible: false,
				Protocol:           port.ProtocolHTTP.String(),
			}
		}
	}

	var sourceImage *image.NewImageParams
	if j.Source.Image != nil {
		var registryID = ""
		if j.Source.Image.RegistryId != nil {
			registryID = *j.Source.Image.RegistryId
		}
		sourceImage = &image.NewImageParams{
			RegistryID: registryID,
			Name:       j.Source.Image.ImageName,
			Tag:        j.Source.Image.Tag,
		}
	}

	var sourceDocker *docker.NewDockerParams
	if j.Source.Docker != nil {
		sourceDocker = &docker.NewDockerParams{
			GitRepository: git_repository.NewGitRepositoryParams{
				Url:        j.Source.Docker.GitRepository.Url,
				Branch:     j.Source.Docker.GitRepository.Branch,
				CommitID:   j.Source.Docker.GitRepository.DeployedCommitId,
				RootPath:   j.Source.Docker.GitRepository.RootPath,
				GitTokenId: j.Source.Docker.GitRepository.GitTokenId.Get(),
			},
			DockerFilePath: j.Source.Docker.DockerfilePath.Get(),
			DockerFileRaw:  j.Source.Docker.DockerfileRaw.Get(),
		}
	}

	var jobSource = job.NewJobSourceParams{
		Image:  sourceImage,
		Docker: sourceDocker,
	}

	var onStart *execution_command.NewExecutionCommandParams = nil
	var onStop *execution_command.NewExecutionCommandParams = nil
	var onDelete *execution_command.NewExecutionCommandParams = nil
	var cronJob *job.NewJobScheduleCronParams = nil
	var lifecycleType *qovery.JobLifecycleTypeEnum = nil
	if j.ScheduleLifecycle != nil {
		if j.ScheduleLifecycle.OnStart != nil {
			onStart = &execution_command.NewExecutionCommandParams{
				Entrypoint: j.ScheduleLifecycle.OnStart.Entrypoint,
				Arguments:  j.ScheduleLifecycle.OnStart.Arguments,
			}
		}
		if j.ScheduleLifecycle.OnStop != nil {
			onStop = &execution_command.NewExecutionCommandParams{
				Entrypoint: j.ScheduleLifecycle.OnStop.Entrypoint,
				Arguments:  j.ScheduleLifecycle.OnStop.Arguments,
			}
		}
		if j.ScheduleLifecycle.OnDelete != nil {
			onDelete = &execution_command.NewExecutionCommandParams{
				Entrypoint: j.ScheduleLifecycle.OnDelete.Entrypoint,
				Arguments:  j.ScheduleLifecycle.OnDelete.Arguments,
			}
		}

		lifecycleType = j.ScheduleLifecycle.LifecycleType
	}
	if j.ScheduleCron != nil {
		cronJob = &job.NewJobScheduleCronParams{
			Schedule: j.ScheduleCron.Cronjob.ScheduledAt,
			Command: execution_command.NewExecutionCommandParams{
				Entrypoint: j.ScheduleCron.Cronjob.Entrypoint,
				Arguments:  j.ScheduleCron.Cronjob.Arguments,
			},
		}
	}

	var jobSchedule = job.NewJobScheduleParams{
		OnStart:       onStart,
		OnStop:        onStop,
		OnDelete:      onDelete,
		LifecycleType: lifecycleType,
		CronJob:       cronJob,
	}

	var maxNbRestart = job.DefaultMaxNbRestart
	if j.MaxNbRestart != nil {
		maxNbRestart = int64(*j.MaxNbRestart)
	}

	var maxDurationSeconds = job.DefaultMaxDurationSeconds
	if j.MaxDurationSeconds != nil {
		maxDurationSeconds = int64(*j.MaxDurationSeconds)
	}

	paramsMaxNbRestart := int32(maxNbRestart)
	paramsMaxDurationSeconds := int32(maxDurationSeconds)
	return job.NewJob(job.NewJobParams{
		JobID:                j.Id,
		EnvironmentID:        j.EnvironmentId,
		Name:                 j.Name,
		AutoPreview:          j.AutoPreview,
		CPU:                  j.Cpu,
		Memory:               j.Memory,
		MaxNbRestart:         &paramsMaxNbRestart,
		MaxDurationSeconds:   &paramsMaxDurationSeconds,
		Port:                 prt,
		Source:               jobSource,
		Schedule:             jobSchedule,
		DeploymentStageID:    deploymentStageID,
		AdvancedSettingsJson: advancedSettingsJson,
		AutoDeploy:           j.AutoDeploy,
		Healthchecks:         j.Healthchecks,
		AnnotationsGroupIds:  j.AnnotationsGroupIds,
		LabelsGroupIds:       j.LabelsGroupIds,
	})
}

// newQoveryJobRequestFromDomain takes the domain request job.UpsertRequest and turns it into a qovery.JobRequest to make the api call.
func newQoveryJobRequestFromDomain(request job.UpsertRepositoryRequest) (*qovery.JobRequest, error) {
	var docker *qovery.JobRequestAllOfSourceDocker = nil
	if request.Source.Docker != nil {
		docker = &qovery.JobRequestAllOfSourceDocker{
			DockerfilePath: *qovery.NewNullableString(request.Source.Docker.DockerFilePath),
			DockerfileRaw:  *qovery.NewNullableString(request.Source.Docker.DockerFileRaw),
			GitRepository: &qovery.ApplicationGitRepositoryRequest{
				Url:        request.Source.Docker.GitRepository.Url,
				Branch:     request.Source.Docker.GitRepository.Branch,
				RootPath:   request.Source.Docker.GitRepository.RootPath,
				GitTokenId: *qovery.NewNullableString(request.Source.Docker.GitRepository.GitTokenId),
			},
		}
	}

	var image *qovery.JobRequestAllOfSourceImage = nil
	if request.Source.Image != nil {
		var registryID = request.Source.Image.RegistryID
		image = &qovery.JobRequestAllOfSourceImage{
			ImageName:  &request.Source.Image.Name,
			Tag:        &request.Source.Image.Tag,
			RegistryId: &registryID,
		}
	}

	var scheduleOnStart *qovery.JobRequestAllOfScheduleOnStart = nil
	if request.Schedule.OnStart != nil {
		scheduleOnStart = &qovery.JobRequestAllOfScheduleOnStart{
			Arguments:  request.Schedule.OnStart.Arguments,
			Entrypoint: request.Schedule.OnStart.Entrypoint,
		}
	}

	var scheduleOnStop *qovery.JobRequestAllOfScheduleOnStart = nil // Note(benjaminch): open-api-generator reused the `onStart` for all types
	if request.Schedule.OnStop != nil {
		scheduleOnStop = &qovery.JobRequestAllOfScheduleOnStart{
			Arguments:  request.Schedule.OnStop.Arguments,
			Entrypoint: request.Schedule.OnStop.Entrypoint,
		}
	}

	var scheduleOnDelete *qovery.JobRequestAllOfScheduleOnStart = nil // Note(benjaminch): open-api-generator reused the `onStart` for all types
	if request.Schedule.OnDelete != nil {
		scheduleOnDelete = &qovery.JobRequestAllOfScheduleOnStart{
			Arguments:  request.Schedule.OnDelete.Arguments,
			Entrypoint: request.Schedule.OnDelete.Entrypoint,
		}
	}

	var scheduleCron *qovery.JobRequestAllOfScheduleCronjob = nil
	if request.Schedule.CronJob != nil {
		scheduleCron = &qovery.JobRequestAllOfScheduleCronjob{
			Arguments:   request.Schedule.CronJob.Command.Arguments,
			Entrypoint:  request.Schedule.CronJob.Command.Entrypoint,
			ScheduledAt: request.Schedule.CronJob.Schedule,
		}
	}

	annotationsGroups, err := NewQoveryServiceAnnotationsGroupRequestFromDomain(request.AnnotationsGroupIds)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	labelsGroups, err := NewQoveryServiceLabelsGroupRequestFromDomain(request.LabelsGroupIds)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	return &qovery.JobRequest{
		Name:               request.Name,
		AutoPreview:        request.AutoPreview,
		Cpu:                request.CPU,
		Memory:             request.Memory,
		MaxNbRestart:       request.MaxNbRestart,
		MaxDurationSeconds: request.MaxDurationSeconds,
		Port:               *qovery.NewNullableInt32(request.Port),
		Source: &qovery.JobRequestAllOfSource{
			Docker: *qovery.NewNullableJobRequestAllOfSourceDocker(docker),
			Image:  *qovery.NewNullableJobRequestAllOfSourceImage(image),
		},
		Schedule: &qovery.JobRequestAllOfSchedule{
			OnStart:       scheduleOnStart,
			OnStop:        scheduleOnStop,
			OnDelete:      scheduleOnDelete,
			LifecycleType: request.Schedule.LifecycleType,
			Cronjob:       scheduleCron,
		},
		AutoDeploy:        request.AutoDeploy,
		Healthchecks:      request.Healthchecks,
		AnnotationsGroups: annotationsGroups,
		LabelsGroups:      labelsGroups,
	}, nil
}
