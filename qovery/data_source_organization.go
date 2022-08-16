package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = organizationDataSourceType{}
var _ datasource.DataSource = organizationDataSource{}

type organizationDataSourceType struct{}

func (t organizationDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing organization.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the organization.",
				Type:        types.StringType,
				Computed:    true,
			},
			"plan": {
				Description: "Plan of the organization.",
				Type:        types.StringType,
				Computed:    true,
			},
			"description": {
				Description: "Description of the organization.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (t organizationDataSourceType) NewDataSource(_ context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return organizationDataSource{
		client: p.(*qProvider).client,
	}, nil
}

type organizationDataSource struct {
	client *client.Client
}

// Read qovery organization data source
func (d organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Organization
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization from API
	organization, apiErr := d.client.GetOrganization(ctx, data.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToOrganization(organization)
	tflog.Trace(ctx, "read organization", map[string]interface{}{"organization_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
