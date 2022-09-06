package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ provider.DataSourceType = containerRegistryDataSourceType{}
var _ datasource.DataSource = containerRegistryDataSource{}

type containerRegistryDataSourceType struct{}

func (t containerRegistryDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing container registry.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the container registry.",
				Type:        types.StringType,
				Required:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the container registry.",
				Type:        types.StringType,
				Computed:    true},
			"kind": {
				Description: "Kind of the container registry.",
				Type:        types.StringType,
				Computed:    true,
			},
			"url": {
				Description: "URL of the container registry.",
				Type:        types.StringType,
				Computed:    true,
			},
			"description": {
				Description: "Description of the container registry.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (t containerRegistryDataSourceType) NewDataSource(_ context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return containerRegistryDataSource{
		containerRegistryService: p.(*qProvider).containerRegistryService,
	}, nil
}

type containerRegistryDataSource struct {
	containerRegistryService registry.Service
}

// Read qovery container registry data source
func (d containerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data ContainerRegistryDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get container registry from API
	reg, err := d.containerRegistryService.Get(ctx, data.OrganizationId.Value, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on container registry read", err.Error())
		return
	}

	state := convertDomainRegistryToContainerRegistryDataSource(reg)
	tflog.Trace(ctx, "read container registry", map[string]interface{}{"container_registry_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
