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
	var configRequest registry.UpsertRequestConfig
	if p.Config == nil {
		configRequest = registry.UpsertRequestConfig{}
	} else {
		configRequest = registry.UpsertRequestConfig{
			AccessKeyID:       ToStringPointer(p.Config.AccessKeyID),
			SecretAccessKey:   ToStringPointer(p.Config.SecretAccessKey),
			Region:            ToStringPointer(p.Config.Region),
			ScalewayAccessKey: ToStringPointer(p.Config.ScalewayAccessKey),
			ScalewaySecretKey: ToStringPointer(p.Config.ScalewaySecretKey),
			Username:          ToStringPointer(p.Config.Username),
			Password:          ToStringPointer(p.Config.Password),
		}
	}
	return registry.UpsertRequest{
		Name:        ToString(p.Name),
		Kind:        ToString(p.Kind),
		URL:         ToString(p.URL),
		Description: ToStringPointer(p.Description),
		Config:      configRequest,
	}
}

func convertDomainRegistryToContainerRegistry(state ContainerRegistry, res *registry.Registry) ContainerRegistry {
	return ContainerRegistry{
		Id:             FromString(res.ID.String()),
		OrganizationId: FromString(res.OrganizationID.String()),
		Name:           FromString(res.Name),
		Kind:           FromString(res.Kind.String()),
		URL:            FromString(res.URL.String()),
		Description:    FromStringPointer(res.Description),
		Config:         state.Config,
	}
}

func convertDomainRegistryToContainerRegistryDataSource(res *registry.Registry) ContainerRegistryDataSource {
	return ContainerRegistryDataSource{
		Id:             FromString(res.ID.String()),
		OrganizationId: FromString(res.OrganizationID.String()),
		Name:           FromString(res.Name),
		Kind:           FromString(res.Kind.String()),
		URL:            FromString(res.URL.String()),
		Description:    FromStringPointer(res.Description),
	}
}
