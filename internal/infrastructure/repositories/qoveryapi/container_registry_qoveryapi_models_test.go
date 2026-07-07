package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

func TestNewDomainRegistryFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Registry      *qovery.ContainerRegistryResponse
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_registry",
			Registry:      nil,
			ExpectedError: registry.ErrNilRegistry,
		},
		{
			TestName: "success",
			Registry: &qovery.ContainerRegistryResponse{
				Id:          gofakeit.UUID(),
				Name:        new(gofakeit.Name()),
				Kind:        qovery.CONTAINERREGISTRYKINDENUM_DOCKER_HUB.Ptr(),
				Url:         new(gofakeit.URL()),
				Description: new(gofakeit.Name()),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			organizationID := gofakeit.UUID()
			reg, err := newDomainRegistryFromQovery(tc.Registry, organizationID)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, reg)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, reg)
			assert.True(t, reg.IsValid())
			assert.Equal(t, tc.Registry.Id, reg.ID.String())
			assert.Equal(t, organizationID, reg.OrganizationID.String())
			assert.Equal(t, tc.Registry.GetName(), reg.Name)
			assert.Equal(t, string(tc.Registry.GetKind()), reg.Kind.String())
			assert.Equal(t, tc.Registry.GetUrl(), reg.URL.String())
			assert.Equal(t, tc.Registry.Description, reg.Description)
		})
	}
}

func TestNewQoveryRegistryEditRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  registry.UpsertRequest
	}{
		{
			TestName: "success_without_description",
			Request: registry.UpsertRequest{
				Name: gofakeit.Name(),
				Kind: registry.KindDockerHub.String(),
				URL:  gofakeit.URL(),
			},
		},
		{
			TestName: "success_with_description",
			Request: registry.UpsertRequest{
				Name:        gofakeit.Name(),
				Kind:        registry.KindDockerHub.String(),
				URL:         gofakeit.URL(),
				Description: new(gofakeit.Word()),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			req, err := newQoveryContainerRegistryRequestFromDomain(tc.Request)
			require.NoError(t, err)

			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.Kind, string(req.Kind))
			assert.Equal(t, tc.Request.URL, *req.Url)
			assert.Equal(t, tc.Request.Description, req.Description)
		})
	}
}

func TestNewQoveryRegistryRequestFromDomain_Config(t *testing.T) {
	t.Parallel()

	jsonCredentials := gofakeit.UUID()
	region := gofakeit.Word()

	req, err := newQoveryContainerRegistryRequestFromDomain(registry.UpsertRequest{
		Name: gofakeit.Name(),
		Kind: registry.KindGcpArtifactRegistry.String(),
		URL:  gofakeit.URL(),
		Config: registry.UpsertRequestConfig{
			Region:          &region,
			JsonCredentials: &jsonCredentials,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, &region, req.Config.Region)
	assert.Equal(t, &jsonCredentials, req.Config.JsonCredentials)
}

func TestNewQoveryRegistryRequestFromDomain_Config_GcpWorkloadIdentityFederation(t *testing.T) {
	t.Parallel()

	region := gofakeit.Word()
	gcpCredentialsType := "workload_identity_federation"
	projectID := gofakeit.UUID()
	serviceAccountEmail := gofakeit.Email()
	workloadIdentityProviderResource := gofakeit.Word()
	tokenLifetimeSeconds := int32(14400)

	req, err := newQoveryContainerRegistryRequestFromDomain(registry.UpsertRequest{
		Name: gofakeit.Name(),
		Kind: registry.KindGcpArtifactRegistry.String(),
		URL:  gofakeit.URL(),
		Config: registry.UpsertRequestConfig{
			Region:                           &region,
			GcpCredentialsType:               &gcpCredentialsType,
			ProjectId:                        &projectID,
			ServiceAccountEmail:              &serviceAccountEmail,
			WorkloadIdentityProviderResource: &workloadIdentityProviderResource,
			TokenLifetimeSeconds:             &tokenLifetimeSeconds,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, &region, req.Config.Region)
	assert.Equal(t, &gcpCredentialsType, req.Config.GcpCredentialsType)
	assert.Equal(t, &projectID, req.Config.ProjectId)
	assert.Equal(t, &serviceAccountEmail, req.Config.ServiceAccountEmail)
	assert.Equal(t, &workloadIdentityProviderResource, req.Config.WorkloadIdentityProviderResource)
	assert.Equal(t, &tokenLifetimeSeconds, req.Config.TokenLifetimeSeconds)
	assert.Nil(t, req.Config.JsonCredentials)
}
