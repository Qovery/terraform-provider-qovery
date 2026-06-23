package qoveryapi

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
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

			assert.Equal(t, tc.Request.RegistryID, req.RegistryId)
			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.ImageName, req.ImageName)
			assert.Equal(t, tc.Request.Tag, req.Tag)
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
