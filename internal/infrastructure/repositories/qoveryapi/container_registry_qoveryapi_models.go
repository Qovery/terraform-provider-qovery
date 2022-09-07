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
	})
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
		Url:         request.URL,
		Description: request.Description,
		Config: qovery.ContainerRegistryRequestConfig{
			AccessKeyId:       request.Config.AccessKeyID,
			SecretAccessKey:   request.Config.SecretAccessKey,
			Region:            request.Config.Region,
			ScalewayAccessKey: request.Config.ScalewayAccessKey,
			ScalewaySecretKey: request.Config.ScalewaySecretKey,
			Username:          request.Config.Username,
			Password:          request.Config.Password,
		},
	}, nil
}
