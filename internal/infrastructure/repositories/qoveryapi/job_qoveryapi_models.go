package qoveryapi

import (
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
func newDomainJobFromQovery(j *qovery.JobResponse, deploymentStageID string) (*job.Job, error) {
	if j == nil {
		return nil, variable.ErrNilVariable
	}

	var prt *port.NewPortParams = nil
	if j.Port.IsSet() {
		rawPort := *j.Port.Get()
		prt = &port.NewPortParams{
			PortID:             string(rawPort),
			InternalPort:       rawPort,
			PubliclyAccessible: false,
			Protocol:           port.ProtocolHTTP.String(),
		}
	}

	var sourceImage *image.NewImageParams
	if imageFrom := j.Source.Image.Get(); imageFrom != nil {
		var registryID = ""
		if imageFrom.RegistryId != nil {
			registryID = *imageFrom.RegistryId
		}
		var imageName = ""
		if imageFrom.ImageName != nil {
			imageName = *imageFrom.ImageName
		}
		var imageTag = ""
		if imageFrom.Tag != nil {
			imageTag = *imageFrom.Tag
		}

		sourceImage = &image.NewImageParams{
			RegistryID: registryID,
			Name:       imageName,
			Tag:        imageTag,
		}
	}

	var sourceDocker *docker.NewDockerParams
	if dockerFrom := j.Source.Docker.Get(); dockerFrom != nil {
		var gitRepository = git_repository.NewGitRepositoryParams{
			Url:      "",
			Branch:   nil,
			CommitID: nil,
			RootPath: nil,
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
			if gitRepositoryFrom.Url != nil {
				gitRepository.Url = *gitRepositoryFrom.Url
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
		maxNbRestart = uint32(*j.MaxNbRestart)
	}

	var maxDurationSeconds = job.DefaultMaxDurationSeconds
	if j.MaxDurationSeconds != nil {
		maxDurationSeconds = uint32(*j.MaxDurationSeconds)
	}

	return job.NewJob(job.NewJobParams{
		JobID:              j.Id,
		EnvironmentID:      j.Environment.Id,
		Name:               j.Name,
		AutoPreview:        j.AutoPreview,
		CPU:                j.Cpu,
		Memory:             j.Memory,
		MaxNbRestart:       &maxNbRestart,
		MaxDurationSeconds: &maxDurationSeconds,
		Port:               prt,
		Source:             jobSource,
		Schedule:           jobSchedule,
		DeploymentStageID:  deploymentStageID,
	})
}

// newQoveryJobRequestFromDomain takes the domain request job.UpsertRequest and turns it into a qovery.JobRequest to make the api call.
func newQoveryJobRequestFromDomain(request job.UpsertRepositoryRequest) (*qovery.JobRequest, error) {
	var docker *qovery.JobRequestAllOfSourceDocker = nil
	if request.Source.Docker != nil {
		docker = &qovery.JobRequestAllOfSourceDocker{
			DockerfilePath: *qovery.NewNullableString(request.Source.Docker.DockerFilePath),
			GitRepository: &qovery.ApplicationGitRepositoryRequest{
				Url:      request.Source.Docker.GitRepository.Url,
				Branch:   request.Source.Docker.GitRepository.Branch,
				RootPath: request.Source.Docker.GitRepository.RootPath,
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
	}, nil
}
