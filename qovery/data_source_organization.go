package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type organizationDataSourceData struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Plan        types.String `tfsdk:"plan"`
	Description types.String `tfsdk:"description"`
}

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
		client: p.(*provider).GetClient(),
	}, nil
}

type organizationDataSource struct {
	client *qovery.APIClient
}

// Read qovery organization data source
func (d organizationDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data organizationDataSourceData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization from API
	organization, res, err := d.client.OrganizationMainCallsApi.
		GetOrganization(ctx, data.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := organizationReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := organizationDataSourceData{
		Id:          data.Id,
		Name:        types.String{Value: organization.Name},
		Plan:        types.String{Value: organization.Plan},
		Description: types.String{Null: true},
	}
	if organization.Description != nil {
		state.Description = types.String{Value: *organization.Description}
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
