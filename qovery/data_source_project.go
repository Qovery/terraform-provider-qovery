package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
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
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this project.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"scope": {
						Description: "Scope of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}, tfsdk.SetNestedAttributesOptions{}),
			},
			"environment_variables": {
				Description: "List of environment variables linked to this project.",
				Optional:    true,
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"scope": {
						Description: "Scope of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}, tfsdk.SetNestedAttributesOptions{}),
			},
			"secrets": {
				Description: "List of secrets linked to this project.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the secret [NOTE: will always be empty].",
						Type:        types.StringType,
						Computed:    true,
						Sensitive:   true,
					},
				}, tfsdk.SetNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (t projectDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return projectDataSource{
		client: p.(*provider).client,
	}, nil
}

type projectDataSource struct {
	client *client.Client
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
	project, apiErr := d.client.GetProject(ctx, data.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToProject(data, project)
	tflog.Trace(ctx, "read project", map[string]interface{}{"project_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
