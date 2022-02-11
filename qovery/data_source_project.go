package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type projectDataSourceData struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

type projectDataSourceType struct{}

func (t projectDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing project.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Required:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the project.",
				Type:        types.StringType,
				Computed:    true,
			},
			"description": {
				Description: "Description of the project.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (t projectDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return projectDataSource{
		client: p.(*provider).GetClient(),
	}, nil
}

type projectDataSource struct {
	client *qovery.APIClient
}

// Read qovery project data source
func (d projectDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data projectDataSourceData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project from API
	project, res, err := d.client.ProjectMainCallsApi.
		GetProject(ctx, data.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := projectReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := projectDataSourceData{
		Id:             data.Id,
		OrganizationId: types.String{Value: project.Organization.Id},
		Name:           types.String{Value: project.Name},
		Description:    types.String{Null: true},
	}
	if project.Description != nil {
		state.Description = types.String{Value: *project.Description}
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
