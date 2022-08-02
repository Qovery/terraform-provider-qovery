package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

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

func (t organizationDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return organizationDataSource{
		organizationService: p.(*provider).organizationService,
	}, nil
}

type organizationDataSource struct {
	organizationService organization.Service
}

// Read qovery organization data source
func (d organizationDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data Organization
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization from API
	orga, err := d.organizationService.Get(ctx, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on organization read", err.Error())
		return
	}

	state := convertDomainOrganizationToTerraform(orga)
	tflog.Trace(ctx, "read organization", map[string]interface{}{"organization_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
