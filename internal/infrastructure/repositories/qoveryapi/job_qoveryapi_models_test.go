package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestNewDomainJobFromQovery_NilResponse(t *testing.T) {
	t.Parallel()

	result, err := newDomainJobFromQovery(nil, "deployment-stage-id", "{}")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, variable.ErrNilVariable)
}

func TestNewQoveryJobRequestFromDomain(t *testing.T) {
	t.Parallel()

	iconURI := "app://qovery-console/job"
	registryID := gofakeit.UUID()
	port := int32(8080)
	cpu := int32(1000)
	memory := int32(512)
	autoPreviewFalse := false
	autoPreviewTrue := true

	testCases := []struct {
		TestName    string
		Request     job.UpsertRepositoryRequest
		ExpectError bool
	}{
		{
			TestName: "success_cron_job_with_image_source",
			Request: job.UpsertRepositoryRequest{
				Name:               "cron-job-test",
				IconUri:            &iconURI,
				AutoPreview:        &autoPreviewFalse,
				CPU:                &cpu,
				Memory:             &memory,
				MaxNbRestart:       func() *int32 { v := int32(3); return &v }(),
				MaxDurationSeconds: func() *int32 { v := int32(300); return &v }(),
				Port:               &port,
				Source: job.Source{
					Image: &image.Image{
						RegistryID: registryID,
						Name:       "my-job-image",
						Tag:        "latest",
					},
				},
				Schedule: job.JobSchedule{
					CronJob: &job.JobScheduleCron{
						Schedule: "0 * * * *",
						Command: execution_command.ExecutionCommand{
							Entrypoint: func() *string { s := "/bin/sh"; return &s }(),
							Arguments:  []string{"-c", "echo hello"},
						},
					},
				},
				Healthchecks: qovery.Healthcheck{},
			},
		},
		{
			TestName: "success_lifecycle_job_with_docker_source",
			Request: job.UpsertRepositoryRequest{
				Name:        "lifecycle-job-test",
				AutoPreview: &autoPreviewTrue,
				CPU:         func() *int32 { v := int32(500); return &v }(),
				Memory:      func() *int32 { v := int32(256); return &v }(),
				Source: job.Source{
					Docker: &docker.Docker{
						GitRepository: git_repository.GitRepository{
							Url:      "https://github.com/example/job-repo.git",
							Branch:   func() *string { s := "main"; return &s }(),
							RootPath: func() *string { s := "/"; return &s }(),
						},
						DockerFilePath: func() *string { s := "Dockerfile"; return &s }(),
					},
				},
				Schedule: job.JobSchedule{
					OnStart: &execution_command.ExecutionCommand{
						Entrypoint: func() *string { s := "/bin/bash"; return &s }(),
						Arguments:  []string{"-c", "echo starting"},
					},
					OnStop: &execution_command.ExecutionCommand{
						Entrypoint: func() *string { s := "/bin/bash"; return &s }(),
						Arguments:  []string{"-c", "echo stopping"},
					},
					OnDelete: &execution_command.ExecutionCommand{
						Entrypoint: func() *string { s := "/bin/bash"; return &s }(),
						Arguments:  []string{"-c", "echo cleanup"},
					},
					LifecycleType: func() *qovery.JobLifecycleTypeEnum { t := qovery.JOBLIFECYCLETYPEENUM_GENERIC; return &t }(),
				},
				Healthchecks: qovery.Healthcheck{},
			},
		},
		{
			TestName: "success_job_with_annotations_and_labels",
			Request: job.UpsertRepositoryRequest{
				Name:        "annotated-job",
				AutoPreview: &autoPreviewFalse,
				CPU:         &cpu,
				Memory:      &memory,
				Source: job.Source{
					Image: &image.Image{
						RegistryID: registryID,
						Name:       "job-image",
						Tag:        "v1",
					},
				},
				Schedule: job.JobSchedule{
					CronJob: &job.JobScheduleCron{
						Schedule: "*/5 * * * *",
						Command: execution_command.ExecutionCommand{
							Arguments: []string{"run-job"},
						},
					},
				},
				AnnotationsGroupIds: []string{gofakeit.UUID(), gofakeit.UUID()},
				LabelsGroupIds:      []string{gofakeit.UUID()},
				Healthchecks:        qovery.Healthcheck{},
			},
		},
		{
			TestName: "success_job_with_docker_target_build_stage",
			Request: job.UpsertRepositoryRequest{
				Name:        "multistage-job",
				AutoPreview: &autoPreviewFalse,
				CPU:         func() *int32 { v := int32(2000); return &v }(),
				Memory:      func() *int32 { v := int32(1024); return &v }(),
				Source: job.Source{
					Docker: &docker.Docker{
						GitRepository: git_repository.GitRepository{
							Url:    "https://github.com/example/multistage.git",
							Branch: func() *string { s := "main"; return &s }(),
						},
						DockerFilePath:         func() *string { s := "Dockerfile.prod"; return &s }(),
						DockerTargetBuildStage: func() *string { s := "production"; return &s }(),
					},
				},
				Schedule: job.JobSchedule{
					OnStart: &execution_command.ExecutionCommand{
						Arguments: []string{"start"},
					},
				},
				Healthchecks: qovery.Healthcheck{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newQoveryJobRequestFromDomain(tc.Request)
			if tc.ExpectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.Request.Name, result.Name)
		})
	}
}

func TestGetAggregateJobResponse_CronJob(t *testing.T) {
	t.Parallel()

	cronJobResponse := &qovery.JobResponse{
		CronJobResponse: &qovery.CronJobResponse{
			Id:            gofakeit.UUID(),
			Name:          "test-cron-job",
			Cpu:           1000,
			Memory:        512,
			MaximumCpu:    2000,
			MaximumMemory: 1024,
			AutoPreview:   false,
			Environment: qovery.ReferenceObject{
				Id: gofakeit.UUID(),
			},
			Schedule: qovery.CronJobResponseAllOfSchedule{
				Cronjob: qovery.CronJobResponseAllOfScheduleCronjob{
					ScheduledAt: "0 * * * *",
					Entrypoint:  func() *string { s := "/bin/sh"; return &s }(),
					Arguments:   []string{"-c", "echo hello"},
				},
			},
			Source: qovery.BaseJobResponseAllOfSource{
				BaseJobResponseAllOfSourceOneOf: &qovery.BaseJobResponseAllOfSourceOneOf{
					Image: qovery.ContainerSource{
						ImageName:  "my-image",
						Tag:        "latest",
						RegistryId: func() *string { s := gofakeit.UUID(); return &s }(),
					},
				},
			},
			Healthchecks: qovery.Healthcheck{},
		},
	}

	result := getAggregateJobResponse(cronJobResponse)

	assert.Equal(t, cronJobResponse.CronJobResponse.Id, result.Id)
	assert.Equal(t, cronJobResponse.CronJobResponse.Name, result.Name)
	assert.Equal(t, cronJobResponse.CronJobResponse.Cpu, result.Cpu)
	assert.Equal(t, cronJobResponse.CronJobResponse.Memory, result.Memory)
	assert.NotNil(t, result.ScheduleCron)
	assert.Nil(t, result.ScheduleLifecycle)
	assert.NotNil(t, result.Source.Image)
	assert.Nil(t, result.Source.Docker)
}

func TestGetAggregateJobResponse_LifecycleJob(t *testing.T) {
	t.Parallel()

	lifecycleJobResponse := &qovery.JobResponse{
		LifecycleJobResponse: &qovery.LifecycleJobResponse{
			Id:            gofakeit.UUID(),
			Name:          "test-lifecycle-job",
			Cpu:           500,
			Memory:        256,
			MaximumCpu:    1000,
			MaximumMemory: 512,
			AutoPreview:   true,
			Environment: qovery.ReferenceObject{
				Id: gofakeit.UUID(),
			},
			Schedule: qovery.LifecycleJobResponseAllOfSchedule{
				OnStart: &qovery.JobRequestAllOfScheduleOnStart{
					Entrypoint: func() *string { s := "/start.sh"; return &s }(),
					Arguments:  []string{},
				},
				OnStop: &qovery.JobRequestAllOfScheduleOnStart{
					Entrypoint: func() *string { s := "/stop.sh"; return &s }(),
					Arguments:  []string{},
				},
				LifecycleType: func() *qovery.JobLifecycleTypeEnum { t := qovery.JOBLIFECYCLETYPEENUM_TERRAFORM; return &t }(),
			},
			Source: qovery.BaseJobResponseAllOfSource{
				BaseJobResponseAllOfSourceOneOf1: &qovery.BaseJobResponseAllOfSourceOneOf1{
					Docker: qovery.JobSourceDockerResponse{
						GitRepository: &qovery.ApplicationGitRepository{
							Url:    "https://github.com/example/repo.git",
							Branch: func() *string { s := "main"; return &s }(),
						},
						DockerfilePath: *qovery.NewNullableString(func() *string { s := "Dockerfile"; return &s }()),
					},
				},
			},
			Healthchecks: qovery.Healthcheck{},
		},
	}

	result := getAggregateJobResponse(lifecycleJobResponse)

	assert.Equal(t, lifecycleJobResponse.LifecycleJobResponse.Id, result.Id)
	assert.Equal(t, lifecycleJobResponse.LifecycleJobResponse.Name, result.Name)
	assert.Equal(t, lifecycleJobResponse.LifecycleJobResponse.Cpu, result.Cpu)
	assert.Equal(t, lifecycleJobResponse.LifecycleJobResponse.Memory, result.Memory)
	assert.Nil(t, result.ScheduleCron)
	assert.NotNil(t, result.ScheduleLifecycle)
	assert.Nil(t, result.Source.Image)
	assert.NotNil(t, result.Source.Docker)
}
