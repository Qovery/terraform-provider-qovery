package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

type ContainerRegistry struct {
	Id             types.String             `tfsdk:"id"`
	OrganizationId types.String             `tfsdk:"organization_id"`
	Name           types.String             `tfsdk:"name"`
	Kind           types.String             `tfsdk:"kind"`
	URL            types.String             `tfsdk:"url"`
	Description    types.String             `tfsdk:"description"`
	Config         *ContainerRegistryConfig `tfsdk:"config"`
}

type ContainerRegistryConfig struct {
	AccessKeyID       types.String `tfsdk:"access_key_id"`
	SecretAccessKey   types.String `tfsdk:"secret_access_key"`
	Region            types.String `tfsdk:"region"`
	ScalewayAccessKey types.String `tfsdk:"scaleway_access_key"`
	ScalewaySecretKey types.String `tfsdk:"scaleway_secret_key"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
}

type ContainerRegistryDataSource struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Kind           types.String `tfsdk:"kind"`
	URL            types.String `tfsdk:"url"`
	Description    types.String `tfsdk:"description"`
}

func (p ContainerRegistry) toUpsertRequest() registry.UpsertRequest {
	return registry.UpsertRequest{
		Name:        toString(p.Name),
		Kind:        toString(p.Kind),
		URL:         toString(p.URL),
		Description: toStringPointer(p.Description),
		Config: registry.UpsertRequestConfig{
			AccessKeyID:       toStringPointer(p.Config.AccessKeyID),
			SecretAccessKey:   toStringPointer(p.Config.SecretAccessKey),
			Region:            toStringPointer(p.Config.Region),
			ScalewayAccessKey: toStringPointer(p.Config.ScalewayAccessKey),
			ScalewaySecretKey: toStringPointer(p.Config.ScalewaySecretKey),
			Username:          toStringPointer(p.Config.Username),
			Password:          toStringPointer(p.Config.Password),
		},
	}
}

func convertDomainRegistryToContainerRegistry(state ContainerRegistry, res *registry.Registry) ContainerRegistry {
	return ContainerRegistry{
		Id:             fromString(res.ID.String()),
		OrganizationId: fromString(res.OrganizationID.String()),
		Name:           fromString(res.Name),
		Kind:           fromString(res.Kind.String()),
		URL:            fromString(res.URL.String()),
		Description:    fromStringPointer(res.Description),
		Config:         state.Config,
	}
}

func convertDomainRegistryToContainerRegistryDataSource(res *registry.Registry) ContainerRegistryDataSource {
	return ContainerRegistryDataSource{
		Id:             fromString(res.ID.String()),
		OrganizationId: fromString(res.OrganizationID.String()),
		Name:           fromString(res.Name),
		Kind:           fromString(res.Kind.String()),
		URL:            fromString(res.URL.String()),
		Description:    fromStringPointer(res.Description),
	}
}
