package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &projectDataSource{}

type projectDataSource struct {
	projectService project.Service
}

func newProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.projectService = provider.projectService
}

func (r projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "project")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "project")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "project")
	secrets := variableListDescriptions("secrets", "project")
	secretAliases := variableListDescriptions("secret_aliases", "project")
	envVarFiles := variableListDescriptions("environment_variable_files", "project")
	secretFiles := variableListDescriptions("secret_files", "project")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("project"),
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("project"),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("project"),
				Optional:            true,
				Computed:            true,
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: builtInEnvVars.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.Description,
							Computed:            true,
						},
					},
				},
			},
			"environment_variables": schema.SetNestedAttribute{
				MarkdownDescription: envVars.List,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVars.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVars.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVars.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVars.Description,
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				MarkdownDescription: envVarAliases.List,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVarAliases.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Description,
							Computed:            true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				MarkdownDescription: secrets.List,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secrets.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secrets.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secrets.Value,
							Computed:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secrets.Description,
							Computed:            true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				MarkdownDescription: secretAliases.List,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secretAliases.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secretAliases.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretAliases.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretAliases.Description,
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_files": schema.SetNestedAttribute{
				MarkdownDescription: envVarFiles.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVarFiles.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVarFiles.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarFiles.Value,
							Computed:            true,
						},
						"mount_path": schema.StringAttribute{
							MarkdownDescription: envVarFiles.MountPath,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarFiles.Description,
							Computed:            true,
						},
					},
				},
			},
			"secret_files": schema.SetNestedAttribute{
				MarkdownDescription: secretFiles.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secretFiles.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secretFiles.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretFiles.Value,
							Computed:            true,
							Sensitive:           true,
						},
						"mount_path": schema.StringAttribute{
							MarkdownDescription: secretFiles.MountPath,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretFiles.Description,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Read qovery project data source
func (d projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Project
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project from API
	proj, err := d.projectService.Get(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on project read", err.Error())
		return
	}

	state := convertDomainProjectToProject(ctx, data, proj)
	tflog.Trace(ctx, "read project", map[string]any{"project_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
