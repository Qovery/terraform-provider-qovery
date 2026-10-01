package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// newDomainRegistryFromQovery takes a qovery.ContainerRegistryResponse returned by the API client and turns it into the domain model registry.Registry.
func newDomainRegistryFromQovery(v *qovery.ContainerRegistryResponse, organizationID string) (*registry.Registry, error) {
	if v == nil {
		return nil, registry.ErrNilRegistry
	}

	return registry.NewRegistry(registry.NewRegistryParams{
		RegistryID:     v.GetId(),
		OrganizationID: organizationID,
		Name:           v.GetName(),
		Kind:           string(v.GetKind()),
		URL:            v.GetUrl(),
		Description:    v.Description,
		Config:         newDomainRegistryConfigFromQovery(v.Config),
	})
}

// newDomainRegistryConfigFromQovery keeps the keys of the response config that the provider manages.
func newDomainRegistryConfigFromQovery(config *qovery.ContainerRegistryResponseAllOfConfig) registry.Config {
	if config == nil {
		return registry.Config{}
	}
	return registry.Config{
		AccessKeyID:                      config.AccessKeyId,
		Region:                           config.Region,
		ScalewayAccessKey:                config.ScalewayAccessKey,
		ScalewayProjectID:                config.ScalewayProjectId,
		GcpCredentialsType:               config.GcpCredentialsType,
		ProjectID:                        config.ProjectId,
		ServiceAccountEmail:              config.ServiceAccountEmail,
		WorkloadIdentityProviderResource: config.WorkloadIdentityProviderResource,
		TokenLifetimeSeconds:             config.TokenLifetimeSeconds,
		Username:                         config.Username,
	}
}

// newQoveryContainerRegistryRequestFromDomain takes the domain request registry.UpsertRequest and turns it into a qovery.ContainerRegistryRequest to make the api call.
func newQoveryContainerRegistryRequestFromDomain(request registry.UpsertRequest) (*qovery.ContainerRegistryRequest, error) {
	kind, err := qovery.NewContainerRegistryKindEnumFromValue(request.Kind)
	if err != nil {
		return nil, registry.ErrInvalidKindParam
	}

	return &qovery.ContainerRegistryRequest{
		Name:        request.Name,
		Kind:        *kind,
		Url:         new(request.URL),
		Description: request.Description,
		Config: qovery.ContainerRegistryRequestConfig{
			AccessKeyId:                      request.Config.AccessKeyID,
			SecretAccessKey:                  request.Config.SecretAccessKey,
			Region:                           request.Config.Region,
			ScalewayAccessKey:                request.Config.ScalewayAccessKey,
			ScalewaySecretKey:                request.Config.ScalewaySecretKey,
			ScalewayProjectId:                request.Config.ScalewayProjectId,
			JsonCredentials:                  request.Config.JsonCredentials,
			GcpCredentialsType:               request.Config.GcpCredentialsType,
			ProjectId:                        request.Config.ProjectId,
			ServiceAccountEmail:              request.Config.ServiceAccountEmail,
			WorkloadIdentityProviderResource: request.Config.WorkloadIdentityProviderResource,
			TokenLifetimeSeconds:             request.Config.TokenLifetimeSeconds,
			Username:                         request.Config.Username,
			Password:                         request.Config.Password,
		},
	}, nil
}
