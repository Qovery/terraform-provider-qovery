package qoveryapi

import (
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

// newDomainCredentialsFromQovery takes a qovery.EnvironmentVariable returned by the API client and turns it into the domain model variable.Variable.
func newDomainJobFromQovery(j *qovery.JobResponse, deploymentStageID string, advancedSettingsJson string) (*job.Job, error) {
	if j == nil {
		return nil, variable.ErrNilVariable
	}

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
	if j.Source.JobResponseAllOfSourceOneOf != nil {
		imageFrom := j.Source.JobResponseAllOfSourceOneOf.Image
		var registryID = ""
		if imageFrom.RegistryId != nil {
			registryID = *imageFrom.RegistryId
		}
		sourceImage = &image.NewImageParams{
			RegistryID: registryID,
			Name:       imageFrom.ImageName,
			Tag:        imageFrom.Tag,
		}
	}

	var sourceDocker *docker.NewDockerParams

	if j.Source.JobResponseAllOfSourceOneOf1 != nil {
		dockerFrom := j.Source.JobResponseAllOfSourceOneOf1.Docker
		var gitRepository = git_repository.NewGitRepositoryParams{
			Url:        "",
			Branch:     nil,
			CommitID:   nil,
			RootPath:   nil,
			GitTokenId: nil,
		}

		if gitRepositoryFrom := dockerFrom.GitRepository; gitRepositoryFrom != nil {
			if gitRepositoryFrom.Url != nil {
				gitRepository.Url = *gitRepositoryFrom.Url
			}
			if gitRepositoryFrom.Branch != nil {
				gitRepository.Branch = gitRepositoryFrom.Branch
			}
			if gitRepositoryFrom.DeployedCommitId != nil {
				gitRepository.CommitID = gitRepositoryFrom.DeployedCommitId
			}
			if gitRepositoryFrom.RootPath != nil {
				gitRepository.RootPath = gitRepositoryFrom.RootPath
			}
			if gitRepositoryFrom.GitTokenId.Get() != nil {
				gitRepository.GitTokenId = gitRepositoryFrom.GitTokenId.Get()
			}
		}

		sourceDocker = &docker.NewDockerParams{
			GitRepository:  gitRepository,
			DockerFilePath: dockerFrom.DockerfilePath.Get(),
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
	if j.Schedule != nil {
		if j.Schedule.OnStart != nil {
			onStart = &execution_command.NewExecutionCommandParams{
				Entrypoint: j.Schedule.OnStart.Entrypoint,
				Arguments:  j.Schedule.OnStart.Arguments,
			}
		}
		if j.Schedule.OnStop != nil {
			onStop = &execution_command.NewExecutionCommandParams{
				Entrypoint: j.Schedule.OnStop.Entrypoint,
				Arguments:  j.Schedule.OnStop.Arguments,
			}
		}
		if j.Schedule.OnDelete != nil {
			onDelete = &execution_command.NewExecutionCommandParams{
				Entrypoint: j.Schedule.OnDelete.Entrypoint,
				Arguments:  j.Schedule.OnDelete.Arguments,
			}
		}
		if j.Schedule.Cronjob != nil {
			cronJob = &job.NewJobScheduleCronParams{
				Schedule: j.Schedule.Cronjob.ScheduledAt,
				Command: execution_command.NewExecutionCommandParams{
					Entrypoint: j.Schedule.Cronjob.Entrypoint,
					Arguments:  j.Schedule.Cronjob.Arguments,
				},
			}
		}
	}

	var jobSchedule = job.NewJobScheduleParams{
		OnStart:  onStart,
		OnStop:   onStop,
		OnDelete: onDelete,
		CronJob:  cronJob,
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
		EnvironmentID:        j.Environment.Id,
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
	})
}

// newQoveryJobRequestFromDomain takes the domain request job.UpsertRequest and turns it into a qovery.JobRequest to make the api call.
func newQoveryJobRequestFromDomain(request job.UpsertRepositoryRequest) (*qovery.JobRequest, error) {
	var docker *qovery.JobRequestAllOfSourceDocker = nil
	if request.Source.Docker != nil {
		docker = &qovery.JobRequestAllOfSourceDocker{
			DockerfilePath: *qovery.NewNullableString(request.Source.Docker.DockerFilePath),
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
			OnStart:  scheduleOnStart,
			OnStop:   scheduleOnStop,
			OnDelete: scheduleOnDelete,
			Cronjob:  scheduleCron,
		},
		AutoDeploy: request.AutoDeploy,
	}, nil
}
