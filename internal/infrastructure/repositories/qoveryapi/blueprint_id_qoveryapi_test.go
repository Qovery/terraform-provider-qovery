package qoveryapi

import (
	"testing"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

func TestNewQoveryHelmRequestFromDomain_SetsBlueprintId(t *testing.T) {
	t.Parallel()

	blueprintID := uuid.NewString()
	req := helm.UpsertRepositoryRequest{
		Name:       "test-helm",
		AutoDeploy: false,
		Source: helm.Source{
			HelmRepository: &helm.SourceHelmRepository{
				RepositoryId: uuid.NewString(),
				ChartName:    "redis",
				ChartVersion: "1.0.0",
			},
		},
		ValuesOverride: helm.ValuesOverride{},
		BlueprintID:    &blueprintID,
	}

	got, err := newQoveryHelmRequestFromDomain(req)
	require.NoError(t, err)
	require.True(t, got.BlueprintId.IsSet(), "BlueprintId should be set on the API request")
	assert.Equal(t, blueprintID, *got.BlueprintId.Get())
}

func TestNewQoveryHelmRequestFromDomain_OmitsBlueprintIdWhenNil(t *testing.T) {
	t.Parallel()

	req := helm.UpsertRepositoryRequest{
		Name:       "test-helm",
		AutoDeploy: false,
		Source: helm.Source{
			HelmRepository: &helm.SourceHelmRepository{
				RepositoryId: uuid.NewString(),
				ChartName:    "redis",
				ChartVersion: "1.0.0",
			},
		},
		ValuesOverride: helm.ValuesOverride{},
	}

	got, err := newQoveryHelmRequestFromDomain(req)
	require.NoError(t, err)
	assert.False(t, got.BlueprintId.IsSet(), "BlueprintId should be unset when domain has nil BlueprintID")
}

func TestGetAggregateHelmResponse_ReadsBlueprintId(t *testing.T) {
	t.Parallel()

	blueprintID := uuid.NewString()
	helmResp := &qovery.HelmResponse{
		Id: uuid.NewString(),
		Environment: qovery.ReferenceObject{
			Id: uuid.NewString(),
		},
		BlueprintId: *qovery.NewNullableString(&blueprintID),
	}

	agg := getAggregateHelmResponse(helmResp)
	require.NotNil(t, agg.BlueprintID)
	assert.Equal(t, blueprintID, *agg.BlueprintID)
}

func TestGetAggregateHelmResponse_NilBlueprintIdWhenUnset(t *testing.T) {
	t.Parallel()

	helmResp := &qovery.HelmResponse{
		Id: uuid.NewString(),
		Environment: qovery.ReferenceObject{
			Id: uuid.NewString(),
		},
	}

	agg := getAggregateHelmResponse(helmResp)
	assert.Nil(t, agg.BlueprintID)
}

func TestNewQoveryTerraformRequestFromDomain_SetsBlueprintId(t *testing.T) {
	t.Parallel()

	blueprintID := uuid.NewString()
	req := validTerraformUpsertRequest()
	req.BlueprintID = &blueprintID

	got, err := newQoveryTerraformRequestFromDomain(req)
	require.NoError(t, err)
	require.True(t, got.BlueprintId.IsSet(), "BlueprintId should be set on the API request")
	assert.Equal(t, blueprintID, *got.BlueprintId.Get())
}

func TestNewQoveryTerraformRequestFromDomain_OmitsBlueprintIdWhenNil(t *testing.T) {
	t.Parallel()

	req := validTerraformUpsertRequest()
	got, err := newQoveryTerraformRequestFromDomain(req)
	require.NoError(t, err)
	assert.False(t, got.BlueprintId.IsSet())
}

func TestNewDomainTerraformServiceFromQovery_ReadsBlueprintId(t *testing.T) {
	t.Parallel()

	blueprintID := uuid.NewString()
	resp := minimalTerraformResponse()
	resp.BlueprintId = *qovery.NewNullableString(&blueprintID)

	got, err := newDomainTerraformServiceFromQovery(resp, "", false, "")
	require.NoError(t, err)
	require.NotNil(t, got.BlueprintID)
	assert.Equal(t, blueprintID, *got.BlueprintID)
}

func TestNewDomainTerraformServiceFromQovery_NilBlueprintIdWhenUnset(t *testing.T) {
	t.Parallel()

	resp := minimalTerraformResponse()
	got, err := newDomainTerraformServiceFromQovery(resp, "", false, "")
	require.NoError(t, err)
	assert.Nil(t, got.BlueprintID)
}

func validTerraformUpsertRequest() terraformservice.UpsertRepositoryRequest {
	return terraformservice.UpsertRepositoryRequest{
		Name:       "test-tf",
		AutoDeploy: false,
		GitRepository: terraformservice.GitRepository{
			URL:    "https://github.com/org/repo",
			Branch: "main",
		},
		Backend: terraformservice.Backend{
			Kubernetes: &terraformservice.KubernetesBackend{},
		},
		Engine: terraformservice.EngineTerraform,
		EngineVersion: terraformservice.EngineVersion{
			ExplicitVersion: "1.5.7",
		},
		JobResources: terraformservice.JobResources{
			CPUMilli:   1000,
			RAMMiB:     1024,
			GPU:        0,
			StorageGiB: 20,
		},
	}
}

func minimalTerraformResponse() *qovery.TerraformResponse {
	kubernetesBackend := qovery.NewTerraformBackendOneOf(map[string]any{})
	backend := qovery.TerraformBackendOneOfAsTerraformBackend(kubernetesBackend)

	return &qovery.TerraformResponse{
		Id: uuid.NewString(),
		Environment: qovery.ReferenceObject{
			Id: uuid.NewString(),
		},
		Name:                     "test-tf",
		Backend:                  backend,
		Engine:                   qovery.TerraformEngineEnum(terraformservice.EngineTerraform),
		TerraformVariablesSource: qovery.TerraformVariablesSourceResponse{},
		ProviderVersion: qovery.TerraformProviderVersion{
			ExplicitVersion: "1.5.7",
		},
		JobResources: qovery.TerraformJobResourcesResponse{
			CpuMilli:   1000,
			RamMib:     1024,
			Gpu:        0,
			StorageGib: 20,
		},
		ActionExtraArguments: map[string][]string{},
	}
}
