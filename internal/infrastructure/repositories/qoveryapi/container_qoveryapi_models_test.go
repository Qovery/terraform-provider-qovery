package qoveryapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
)

const (
	minContainerInt32Range = 1
	maxContainerInt32Range = 100.000
)

func TestNewDomainContainerFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Container     *qovery.ContainerResponse
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_container",
			Container:     nil,
			ExpectedError: container.ErrNilContainer,
		},
		{
			TestName: "success",
			Container: &qovery.ContainerResponse{
				Id: gofakeit.UUID(),
				Environment: qovery.ReferenceObject{
					Id: gofakeit.UUID(),
				},
				Registry: qovery.ContainerRegistryProviderDetailsResponse{
					Id:   gofakeit.UUID(),
					Name: gofakeit.Name(),
					Url:  gofakeit.URL(),
					Kind: qovery.CONTAINERREGISTRYKINDENUM_DOCKER_HUB,
				},
				Arguments: []string{
					gofakeit.Word(),
				},
				Name:                  gofakeit.Name(),
				IconUri:               fmt.Sprintf("app://qovery-console/%v", gofakeit.Name()),
				ImageName:             gofakeit.Name(),
				Tag:                   gofakeit.Word(),
				Entrypoint:            new(gofakeit.Word()),
				AutoPreview:           gofakeit.Bool(),
				Cpu:                   int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
				Memory:                int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
				EphemeralStorageInGib: new(int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range))),
				MaximumCpu:            int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
				MaximumMemory:         int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
				MinRunningInstances:   int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
				MaxRunningInstances:   int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
				Healthchecks: qovery.Healthcheck{
					ReadinessProbe: *qovery.NewNullableProbe(nil),
					LivenessProbe:  *qovery.NewNullableProbe(nil),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			fakeDeploymentStageId := uuid.NewString()
			customDomains := qovery.CustomDomainResponseList{}
			cont, err := newDomainContainerFromQovery(tc.Container, fakeDeploymentStageId, false, "", &customDomains)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, cont)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, cont)
			assert.True(t, cont.IsValid())
			assert.Equal(t, tc.Container.Id, cont.ID.String())
			assert.Equal(t, tc.Container.Environment.Id, cont.EnvironmentID.String())
			s := cont.RegistryID.String()
			assert.Equal(t, tc.Container.Registry.Id, s)
			assert.Equal(t, tc.Container.Name, cont.Name)
			assert.Equal(t, tc.Container.ImageName, cont.ImageName)
			assert.Equal(t, tc.Container.Tag, cont.Tag)
			assert.Equal(t, tc.Container.Cpu, cont.CPU)
			assert.Equal(t, tc.Container.Memory, cont.Memory)
			assert.Equal(t, tc.Container.EphemeralStorageInGib, cont.EphemeralStorage)
			assert.Equal(t, tc.Container.MinRunningInstances, cont.MinRunningInstances)
			assert.Equal(t, tc.Container.MaxRunningInstances, cont.MaxRunningInstances)
			assert.Equal(t, tc.Container.AutoPreview, cont.AutoPreview)
			assert.Equal(t, tc.Container.Entrypoint, cont.Entrypoint)
			assert.Equal(t, fakeDeploymentStageId, cont.DeploymentStageID)

			assert.Len(t, tc.Container.Ports, len(cont.Ports))
			for idx, p := range cont.Ports {
				assert.Equal(t, tc.Container.Ports[idx].Name, p.Name)
				assert.Equal(t, tc.Container.Ports[idx].InternalPort, p.InternalPort)
				assert.Equal(t, tc.Container.Ports[idx].ExternalPort, p.ExternalPort)
				assert.Equal(t, string(tc.Container.Ports[idx].Protocol), p.Protocol.String())
				assert.Equal(t, tc.Container.Ports[idx].PubliclyAccessible, p.PubliclyAccessible)
			}

			assert.Len(t, tc.Container.Storage, len(cont.Storages))
			for idx, s := range cont.Storages {
				assert.Equal(t, tc.Container.Storage[idx].Id, s.ID.String())
				assert.Equal(t, string(tc.Container.Storage[idx].Type), s.Type.String())
				assert.Equal(t, tc.Container.Storage[idx].Size, s.Size)
				assert.Equal(t, tc.Container.Storage[idx].MountPoint, s.MountPoint)
			}

			assert.Len(t, tc.Container.Arguments, len(cont.Arguments))
			for _, arg := range cont.Arguments {
				assert.Contains(t, tc.Container.Arguments, arg)
			}
		})
	}
}

func TestNewQoveryContainerRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  container.UpsertRepositoryRequest
	}{
		{
			TestName: "success_required_fields_only",
			Request: container.UpsertRepositoryRequest{
				RegistryID: gofakeit.UUID(),
				Name:       gofakeit.Name(),
				ImageName:  gofakeit.Name(),
				Tag:        gofakeit.Word(),
			},
		},
		{
			TestName: "success",
			Request: container.UpsertRepositoryRequest{
				RegistryID:          gofakeit.UUID(),
				Name:                gofakeit.Name(),
				ImageName:           gofakeit.Name(),
				Tag:                 gofakeit.Word(),
				Entrypoint:          new(gofakeit.Word()),
				CPU:                 new(int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range))),
				Memory:              new(int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range))),
				EphemeralStorage:    new(int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range))),
				MinRunningInstances: new(int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range))),
				MaxRunningInstances: new(int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range))),
				AutoPreview:         new(gofakeit.Bool()),
				Arguments: []string{
					gofakeit.Word(),
				},
				Ports: []port.UpsertRequest{
					{
						Name:               new(gofakeit.Name()),
						InternalPort:       int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
						ExternalPort:       new(int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range))),
						Protocol:           new(port.ProtocolHTTP.String()),
						PubliclyAccessible: gofakeit.Bool(),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req, err := newQoveryContainerRequestFromDomain(tc.Request)
			assert.NoError(t, err)

			assert.Equal(t, tc.Request.RegistryID, *req.RegistryId)
			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.ImageName, *req.ImageName)
			assert.Equal(t, tc.Request.Tag, *req.Tag)
			assert.Equal(t, tc.Request.Entrypoint, req.Entrypoint)
			assert.Equal(t, tc.Request.CPU, req.Cpu)
			assert.Equal(t, tc.Request.Memory, req.Memory)
			assert.Equal(t, tc.Request.EphemeralStorage, req.EphemeralStorageInGib)
			assert.Equal(t, tc.Request.MinRunningInstances, req.MinRunningInstances)
			assert.Equal(t, tc.Request.MaxRunningInstances, req.MaxRunningInstances)

			assert.Len(t, tc.Request.Ports, len(req.Ports))
			for idx, p := range req.Ports {
				assert.Equal(t, tc.Request.Ports[idx].Name, p.Name)
				assert.Equal(t, tc.Request.Ports[idx].InternalPort, p.InternalPort)
				assert.Equal(t, tc.Request.Ports[idx].ExternalPort, p.ExternalPort)
				assert.Equal(t, *tc.Request.Ports[idx].Protocol, string(*p.Protocol))
				assert.Equal(t, tc.Request.Ports[idx].PubliclyAccessible, p.PubliclyAccessible)
			}

			assert.Len(t, tc.Request.Storages, len(req.Storage))
			for idx, s := range req.Storage {
				assert.Equal(t, tc.Request.Storages[idx].ID, s.Id)
				assert.Equal(t, tc.Request.Storages[idx].Type, string(s.Type))
				assert.Equal(t, tc.Request.Storages[idx].Size, s.Size)
				assert.Equal(t, tc.Request.Storages[idx].MountPoint, s.MountPoint)
			}

			assert.Len(t, tc.Request.Arguments, len(req.Arguments))
			for _, arg := range req.Arguments {
				assert.Contains(t, tc.Request.Arguments, arg)
			}
		})
	}
}

func TestNewQoveryContainerRequestFromDomain_ImageContainerJSON(t *testing.T) {
	t.Parallel()

	req, err := newQoveryContainerRequestFromDomain(container.UpsertRepositoryRequest{
		RegistryID:          "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10",
		Name:                "my-container",
		IconUri:             new("app://qovery-console/container"),
		ImageName:           "qovery/nginx",
		Tag:                 "1.27",
		AutoPreview:         new(false),
		Entrypoint:          new("/docker-entrypoint.sh"),
		CPU:                 new(int32(500)),
		Memory:              new(int32(512)),
		EphemeralStorage:    new(int32(0)),
		MinRunningInstances: new(int32(1)),
		MaxRunningInstances: new(int32(2)),
		Arguments:           []string{"nginx", "-g", "daemon off;"},
		Ports: []port.UpsertRequest{
			{
				Name:               new("http"),
				InternalPort:       80,
				ExternalPort:       new(int32(443)),
				Protocol:           new(port.ProtocolHTTP.String()),
				PubliclyAccessible: true,
			},
		},
		Healthchecks: qovery.Healthcheck{
			ReadinessProbe: *qovery.NewNullableProbe(nil),
			LivenessProbe:  *qovery.NewNullableProbe(nil),
		},
		AutoDeploy: *qovery.NewNullableBool(new(true)),
	})
	require.NoError(t, err)

	body, err := json.Marshal(req)
	require.NoError(t, err)

	// Bytes captured with the client before the flat image fields became optional and `source` was added.
	assert.Equal(t, `{"annotations_groups":[],"arguments":["nginx","-g","daemon off;"],"auto_deploy":true,"auto_preview":false,"cpu":500,"entrypoint":"/docker-entrypoint.sh","ephemeral_storage_in_gib":0,"healthchecks":{"liveness_probe":null,"readiness_probe":null},"icon_uri":"app://qovery-console/container","image_name":"qovery/nginx","labels_groups":[],"max_running_instances":2,"memory":512,"min_running_instances":1,"name":"my-container","ports":[{"external_port":443,"internal_port":80,"is_default":false,"name":"http","protocol":"HTTP","publicly_accessible":true}],"registry_id":"2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10","storage":[],"tag":"1.27"}`, string(body))
}

func TestNewDomainContainerFromQovery_Source(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Registry      string
		ImageName     string
		Tag           string
		SourceAndMore string
		ExpectGit     bool
	}{
		{
			TestName:  "no_source",
			Registry:  `{"id": "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10", "name": "Docker Hub", "url": "https://docker.io", "kind": "DOCKER_HUB"}`,
			ImageName: "qovery/nginx",
			Tag:       "1.27",
		},
		{
			TestName:      "image_source",
			Registry:      `{"id": "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10", "name": "Docker Hub", "url": "https://docker.io", "kind": "DOCKER_HUB"}`,
			ImageName:     "qovery/nginx",
			Tag:           "1.27",
			SourceAndMore: `"source": {"image": {"image_name": "qovery/nginx", "tag": "1.27", "registry": {"id": "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10", "name": "Docker Hub", "url": "https://docker.io", "kind": "DOCKER_HUB"}}},`,
		},
		{
			TestName:  "git_repository_source",
			Registry:  `{"id": "7c3e1a90-5b2d-4e8f-a1c6-0d9b8e7f6a54", "name": "cluster registry", "url": "https://rg.fr-par.scw.cloud", "kind": "SCALEWAY_CR"}`,
			ImageName: "z6f0c8a52-built",
			Tag:       "4f2a9c1",
			SourceAndMore: `"source": {"git_repository": {"provider": "GITHUB", "owner": "acme", "url": "https://github.com/acme/api.git", "name": "acme/api", "branch": "main", "root_path": "/api",` +
				` "git_token_id": "9d1c2b3a-4e5f-4a6b-8c7d-0e1f2a3b4c5d", "git_token_name": "acme token", "deployed_commit_id": "4f2a9c1", "dockerfile_path": "Dockerfile", "docker_target_build_stage": "runtime"}},` +
				` "build_settings": {"timeout_max_sec": 1800, "cpu_max_in_milli": 4000, "ram_max_in_gib": 8, "disable_buildkit_cache": false, "skip_git_submodules": true},`,
			ExpectGit: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			body := fmt.Sprintf(`{
				"id": "6f0c8a52-3e0f-4c43-a1a8-5d6d1c0e9b21",
				"created_at": "2026-11-18T10:00:00Z",
				"environment": {"id": "0e8f9b0c-7d0a-4f5e-9c1e-3a2b1c0d9e8f"},
				"registry": %s,
				"image_name": %q,
				"tag": %q,
				%s
				"name": "api",
				"icon_uri": "app://qovery-console/container",
				"service_type": "CONTAINER",
				"maximum_cpu": 4000, "maximum_memory": 8192, "maximum_gpu": 0,
				"cpu": 500, "memory": 512, "gpu": 0,
				"min_running_instances": 1, "max_running_instances": 1,
				"healthchecks": {"readiness_probe": null, "liveness_probe": null},
				"auto_preview": false,
				"auto_deploy": false
			}`, tc.Registry, tc.ImageName, tc.Tag, tc.SourceAndMore)

			var resp qovery.ContainerResponse
			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			if tc.SourceAndMore == "" {
				assert.Nil(t, resp.Source)
			} else {
				require.NotNil(t, resp.Source)
				assert.Equal(t, !tc.ExpectGit, resp.Source.ContainerSourceResponseOneOf != nil)
				assert.Equal(t, tc.ExpectGit, resp.Source.ContainerSourceResponseOneOf1 != nil)
			}

			var registry qovery.ContainerRegistryProviderDetailsResponse
			require.NoError(t, json.Unmarshal([]byte(tc.Registry), &registry))

			cont, err := newDomainContainerFromQovery(&resp, uuid.NewString(), false, "", &qovery.CustomDomainResponseList{})
			require.NoError(t, err)
			assert.Equal(t, registry.Id, cont.RegistryID.String())
			assert.Equal(t, tc.ImageName, cont.ImageName)
			assert.Equal(t, tc.Tag, cont.Tag)

			if !tc.ExpectGit {
				assert.Nil(t, cont.GitSource)
				assert.Nil(t, cont.BuildSettings)
				return
			}

			require.NotNil(t, cont.GitSource)
			assert.Equal(t, "https://github.com/acme/api.git", cont.GitSource.GitRepository.Url)
			assert.Equal(t, new("main"), cont.GitSource.GitRepository.Branch)
			assert.Equal(t, new("/api"), cont.GitSource.GitRepository.RootPath)
			assert.Equal(t, new("9d1c2b3a-4e5f-4a6b-8c7d-0e1f2a3b4c5d"), cont.GitSource.GitRepository.GitTokenId)
			assert.Equal(t, new("GITHUB"), cont.GitSource.GitRepository.Provider)
			assert.Equal(t, new("4f2a9c1"), cont.GitSource.GitRepository.CommitID)
			assert.Equal(t, new("Dockerfile"), cont.GitSource.DockerFilePath)
			assert.Equal(t, new("runtime"), cont.GitSource.DockerTargetBuildStage)
			require.NotNil(t, cont.BuildSettings)
			assert.Equal(t, int32(1800), cont.BuildSettings.GetTimeoutMaxSec())
			assert.True(t, cont.BuildSettings.GetSkipGitSubmodules())
		})
	}
}

func TestNewQoveryContainerRequestFromDomain_Source(t *testing.T) {
	t.Parallel()

	baseRequest := func() container.UpsertRepositoryRequest {
		return container.UpsertRepositoryRequest{
			Name: "api",
			Healthchecks: qovery.Healthcheck{
				ReadinessProbe: *qovery.NewNullableProbe(nil),
				LivenessProbe:  *qovery.NewNullableProbe(nil),
			},
			AutoDeploy: *qovery.NewNullableBool(new(false)),
		}
	}

	buildSettings := &qovery.BuildSettings{TimeoutMaxSec: new(int32(900))}

	gitRequest := baseRequest()
	gitRequest.GitSource = &docker.Docker{
		GitRepository: git_repository.GitRepository{
			Url:        "https://github.com/acme/api.git",
			Branch:     new("main"),
			RootPath:   new("/api"),
			GitTokenId: new("9d1c2b3a-4e5f-4a6b-8c7d-0e1f2a3b4c5d"),
		},
		DockerFilePath:         new("Dockerfile"),
		DockerTargetBuildStage: new("runtime"),
	}
	gitRequest.BuildSettings = buildSettings

	switchBackRequest := baseRequest()
	switchBackRequest.RegistryID = "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10"
	switchBackRequest.ImageName = "qovery/nginx"
	switchBackRequest.Tag = "1.27"
	switchBackRequest.ImageAsSource = true

	testCases := []struct {
		TestName       string
		Request        container.UpsertRepositoryRequest
		ExpectedSource string
		ExpectedBuild  string
	}{
		{
			TestName: "git_repository_source",
			Request:  gitRequest,
			ExpectedSource: `{"git_repository": {"url": "https://github.com/acme/api.git", "branch": "main", "root_path": "/api", "git_token_id": "9d1c2b3a-4e5f-4a6b-8c7d-0e1f2a3b4c5d",` +
				` "provider": "GITHUB", "dockerfile_path": "Dockerfile", "docker_target_build_stage": "runtime"}}`,
			ExpectedBuild: `{"timeout_max_sec": 900}`,
		},
		{
			TestName:       "switch_back_to_image",
			Request:        switchBackRequest,
			ExpectedSource: `{"image": {"registry_id": "2b1f5d0e-2c7a-4f43-9d3a-8f1c6b1a2e10", "image_name": "qovery/nginx", "tag": "1.27"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			req, err := newQoveryContainerRequestFromDomain(tc.Request)
			require.NoError(t, err)

			body, err := json.Marshal(req)
			require.NoError(t, err)

			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(body, &fields))
			assert.NotContains(t, fields, "registry_id")
			assert.NotContains(t, fields, "image_name")
			assert.NotContains(t, fields, "tag")
			require.Contains(t, fields, "source")
			assert.JSONEq(t, tc.ExpectedSource, string(fields["source"]))
			if tc.ExpectedBuild == "" {
				assert.NotContains(t, fields, "build_settings")
			} else {
				assert.JSONEq(t, tc.ExpectedBuild, string(fields["build_settings"]))
			}
		})
	}
}

func TestNewQoveryContainerRequestFromDomain_GitSourceProviderConflict(t *testing.T) {
	t.Parallel()

	_, err := newQoveryContainerRequestFromDomain(container.UpsertRepositoryRequest{
		Name: "api",
		GitSource: &docker.Docker{
			GitRepository:  git_repository.GitRepository{Url: "https://github.com/acme/api.git", Provider: new("GITLAB")},
			DockerFilePath: new("Dockerfile"),
		},
	})
	assert.ErrorContains(t, err, "conflicts with the repository URL")
}

func TestNewQoveryContainerRequestFromDomain_GitSourceSendsAutoFlags(t *testing.T) {
	t.Parallel()

	for _, enabled := range []bool{true, false} {
		t.Run(strconv.FormatBool(enabled), func(t *testing.T) {
			t.Parallel()
			req, err := newQoveryContainerRequestFromDomain(container.UpsertRepositoryRequest{
				Name: "api",
				Healthchecks: qovery.Healthcheck{
					ReadinessProbe: *qovery.NewNullableProbe(nil),
					LivenessProbe:  *qovery.NewNullableProbe(nil),
				},
				AutoDeploy:  *qovery.NewNullableBool(new(enabled)),
				AutoPreview: new(enabled),
				GitSource: &docker.Docker{
					GitRepository:  git_repository.GitRepository{Url: "https://github.com/acme/api.git"},
					DockerFilePath: new("Dockerfile"),
				},
			})
			require.NoError(t, err)

			body, err := json.Marshal(req)
			require.NoError(t, err)

			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(body, &fields))
			assert.JSONEq(t, strconv.FormatBool(enabled), string(fields["auto_deploy"]))
			assert.JSONEq(t, strconv.FormatBool(enabled), string(fields["auto_preview"]))
		})
	}
}

// TestNewDomainContainerFromQovery_Autoscaling covers the autoscaling glue in
// the response→domain direction: the KEDA policy is unwrapped and response-only
// fields are dropped.
func TestNewDomainContainerFromQovery_Autoscaling(t *testing.T) {
	t.Parallel()

	resp := &qovery.ContainerResponse{
		Id:          gofakeit.UUID(),
		Environment: qovery.ReferenceObject{Id: gofakeit.UUID()},
		Registry: qovery.ContainerRegistryProviderDetailsResponse{
			Id:   gofakeit.UUID(),
			Name: gofakeit.Name(),
			Url:  gofakeit.URL(),
			Kind: qovery.CONTAINERREGISTRYKINDENUM_DOCKER_HUB,
		},
		Name:                gofakeit.Name(),
		IconUri:             fmt.Sprintf("app://qovery-console/%v", gofakeit.Name()),
		ImageName:           gofakeit.Name(),
		Tag:                 gofakeit.Word(),
		Cpu:                 int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
		Memory:              int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
		MinRunningInstances: int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
		MaxRunningInstances: int32(gofakeit.IntRange(minContainerInt32Range, maxContainerInt32Range)),
		Healthchecks: qovery.Healthcheck{
			ReadinessProbe: *qovery.NewNullableProbe(nil),
			LivenessProbe:  *qovery.NewNullableProbe(nil),
		},
		Autoscaling: &qovery.AutoscalingPolicyResponse{
			KedaAutoscalingResponse: &qovery.KedaAutoscalingResponse{
				Id:                     gofakeit.UUID(),
				ServiceId:              gofakeit.UUID(),
				Mode:                   qovery.AUTOSCALINGMODE_KEDA,
				PollingIntervalSeconds: 30,
				CooldownPeriodSeconds:  300,
				Scalers: []qovery.KedaScalerResponse{
					{
						Id:         gofakeit.UUID(),
						ScalerType: "prometheus",
						Enabled:    true,
						Role:       qovery.KEDASCALERROLE_PRIMARY,
						ConfigJson: map[string]any{"query": "up"},
					},
				},
			},
		},
	}

	cont, err := newDomainContainerFromQovery(resp, uuid.NewString(), false, "", &qovery.CustomDomainResponseList{})
	assert.NoError(t, err)
	assert.NotNil(t, cont)

	assert.NotNil(t, cont.Autoscaling)
	assert.Equal(t, int32(30), *cont.Autoscaling.PollingIntervalSeconds)
	assert.Equal(t, int32(300), *cont.Autoscaling.CooldownPeriodSeconds)
	assert.Len(t, cont.Autoscaling.Scalers, 1)

	scaler := cont.Autoscaling.Scalers[0]
	assert.Equal(t, "prometheus", scaler.ScalerType)
	assert.True(t, scaler.Enabled)
	assert.Equal(t, autoscaling.RolePrimary, scaler.Role)
	assert.JSONEq(t, `{"query":"up"}`, scaler.Config.ConfigJSON)
}

// TestNewQoveryContainerRequestFromDomain_Autoscaling covers the autoscaling
// glue in the domain→request direction: the policy is wrapped into a KEDA
// request and exposed on the container request.
func TestNewQoveryContainerRequestFromDomain_Autoscaling(t *testing.T) {
	t.Parallel()

	request := container.UpsertRepositoryRequest{
		RegistryID: gofakeit.UUID(),
		Name:       gofakeit.Name(),
		ImageName:  gofakeit.Name(),
		Tag:        gofakeit.Word(),
		Autoscaling: &autoscaling.AutoscalingPolicy{
			Scalers: []autoscaling.Scaler{
				{
					ScalerType: "cpu",
					Enabled:    true,
					Role:       autoscaling.RolePrimary,
					Config:     autoscaling.Config{ConfigJSON: `{"value":"80"}`},
				},
			},
		},
	}

	req, err := newQoveryContainerRequestFromDomain(request)
	assert.NoError(t, err)
	assert.NotNil(t, req.Autoscaling)
	assert.NotNil(t, req.Autoscaling.KedaAutoscalingRequest)
	assert.Equal(t, qovery.AUTOSCALINGMODE_KEDA, req.Autoscaling.KedaAutoscalingRequest.Mode)
	assert.Len(t, req.Autoscaling.KedaAutoscalingRequest.Scalers, 1)
	assert.Equal(t, "cpu", req.Autoscaling.KedaAutoscalingRequest.Scalers[0].ScalerType)
}

// TestNewQoveryContainerRequestFromDomain_NoAutoscaling guards that an absent
// policy leaves the request field nil (omitempty → no KEDA payload sent).
func TestNewQoveryContainerRequestFromDomain_NoAutoscaling(t *testing.T) {
	t.Parallel()

	req, err := newQoveryContainerRequestFromDomain(container.UpsertRepositoryRequest{
		RegistryID: gofakeit.UUID(),
		Name:       gofakeit.Name(),
		ImageName:  gofakeit.Name(),
		Tag:        gofakeit.Word(),
	})
	assert.NoError(t, err)
	assert.Nil(t, req.Autoscaling)
}
