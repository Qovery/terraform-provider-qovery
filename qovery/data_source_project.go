package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"
)

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
			"environment_variables": {
				Description: "List of environment variables linked to this project.",
				Optional:    true,
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
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
	var data Project
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

	projectVariables, res, err := d.client.ProjectEnvironmentVariableApi.
		ListProjectEnvironmentVariable(ctx, project.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := projectEnvironmentVariableReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToProject(project, projectVariables)
	tflog.Trace(ctx, "read project", "project_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
