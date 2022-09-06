package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

type ContainerRegistry struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Kind           types.String `tfsdk:"kind"`
	URL            types.String `tfsdk:"url"`
	Description    types.String `tfsdk:"description"`
	Config         types.Map    `tfsdk:"config"`
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
		Config:      toMapStringString(p.Config),
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
