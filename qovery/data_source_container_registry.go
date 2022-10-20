package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &containerRegistryDataSource{}

type containerRegistryDataSource struct {
	containerRegistryService registry.Service
}

func newContainerRegistryDataSource() datasource.DataSource {
	return &containerRegistryDataSource{}
}
func (d containerRegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_registry"
}

func (d *containerRegistryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.containerRegistryService = provider.containerRegistryService
}

func (d containerRegistryDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
