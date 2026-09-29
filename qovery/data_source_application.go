package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &applicationDataSource{}

type applicationDataSource struct {
	client *client.Client
}

func newApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

func (d applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *applicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = provider.client
}

func (r applicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "application")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "application")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "application")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "application")
	secrets := variableListDescriptions("secrets", "application")
	secretAliases := variableListDescriptions("secret_aliases", "application")
	secretOverrides := variableListDescriptions("secret_overrides", "application")
	envVarFiles := variableListDescriptions("environment_variable_files", "application")
	secretFiles := variableListDescriptions("secret_files", "application")
	externalSecrets := variableListDescriptions("external_secrets", "application")
	externalSecretFiles := variableListDescriptions("external_secret_files", "application")
	ports := portDescriptions("application")
	storages := storageDescriptions("application")
	customDomains := customDomainDescriptions("application")
	restrictions := deploymentRestrictionDescriptions("application")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("application"),
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("application"),
				Computed:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: iconURIDescription("application"),
				Optional:            true,
				Computed:            true,
			},
			"git_repository": schema.SingleNestedAttribute{
				MarkdownDescription: "Git repository the application is built from.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: gitRepositoryURLDescription,
						Computed:            true,
					},
					"branch": schema.StringAttribute{
						MarkdownDescription: "Branch to build.",
						Optional:            true,
						Computed:            true,
					},
					"root_path": schema.StringAttribute{
						MarkdownDescription: gitRepositoryRootPathDescription,
						Optional:            true,
						Computed:            true,
					},
					"git_token_id": schema.StringAttribute{
						MarkdownDescription: gitRepositoryTokenIDDescription,
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"build_mode": schema.StringAttribute{
				MarkdownDescription: "How Qovery builds the application: `DOCKER` or `BUILDPACKS`.",
				Computed:            true,
			},
			"dockerfile_path": schema.StringAttribute{
				MarkdownDescription: dockerfilePathDescription,
				Optional:            true,
				Computed:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: cpuDescription("application"),
				Optional:            true,
				Computed:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: memoryDescription("application"),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMemoryMin},
				},
			},
			"ephemeral_storage": schema.Int64Attribute{
				MarkdownDescription: ephemeralStorageDescription("application"),
				Computed:            true,
			},
			"min_running_instances": schema.Int64Attribute{
				MarkdownDescription: minRunningInstancesDescription("application"),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMinRunningInstancesMin},
				},
			},
			"max_running_instances": schema.Int64Attribute{
				MarkdownDescription: maxRunningInstancesDescription("application"),
				Optional:            true,
				Computed:            true,
			},
			"autoscaling":    autoscalingDataSourceSchema(),
			"build_settings": buildSettingsDataSourceSchemaAttributes(),
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: autoPreviewDescription("application"),
				Optional:            true,
				Computed:            true,
			},
			"entrypoint": schema.StringAttribute{
				MarkdownDescription: entrypointDescription,
				Optional:            true,
				Computed:            true,
			},
			"arguments": schema.ListAttribute{
				MarkdownDescription: argumentsDescription,
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"storage": schema.SetNestedAttribute{
				MarkdownDescription: storages.List,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: storages.ID,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: storages.Type,
							Computed:            true,
						},
						"size": schema.Int64Attribute{
							MarkdownDescription: storages.Size,
							Computed:            true,
						},
						"mount_point": schema.StringAttribute{
							MarkdownDescription: storages.MountPoint,
							Computed:            true,
						},
					},
				},
			},
			"ports": schema.SetNestedAttribute{
				MarkdownDescription: ports.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: ports.ID,
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: ports.Name,
							Optional:            true,
							Computed:            true,
						},
						"internal_port": schema.Int64Attribute{
							MarkdownDescription: ports.InternalPort,
							Computed:            true,
						},
						"external_port": schema.Int64Attribute{
							MarkdownDescription: ports.ExternalPort,
							Computed:            true,
						},
						"publicly_accessible": schema.BoolAttribute{
							MarkdownDescription: ports.PubliclyAccessible,
							Computed:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: ports.Protocol,
							Optional:            true,
							Computed:            true,
						},
						"is_default": schema.BoolAttribute{
							MarkdownDescription: dataSourcePortIsDefaultDescription("application"),
							Computed:            true,
						},
					},
				},
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
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				MarkdownDescription: envVars.List,
				Optional:            true,
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
			"environment_variable_overrides": schema.SetNestedAttribute{
				MarkdownDescription: envVarOverrides.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Description,
							Computed:            true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				MarkdownDescription: secrets.List,
				Optional:            true,
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
			"secret_overrides": schema.SetNestedAttribute{
				MarkdownDescription: secretOverrides.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secretOverrides.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Value,
							Computed:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Description,
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
			"external_secrets": schema.SetNestedAttribute{
				MarkdownDescription: externalSecrets.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: externalSecrets.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: externalSecrets.Key,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: externalSecrets.Description,
							Computed:            true,
						},
						"reference": schema.StringAttribute{
							MarkdownDescription: externalSecrets.Reference,
							Computed:            true,
						},
						"secret_manager_access_id": schema.StringAttribute{
							MarkdownDescription: externalSecrets.SecretManagerAccessID,
							Computed:            true,
						},
					},
				},
			},
			"external_secret_files": schema.SetNestedAttribute{
				MarkdownDescription: externalSecretFiles.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{MarkdownDescription: externalSecretFiles.ID, Computed: true},
						"key":                      schema.StringAttribute{MarkdownDescription: externalSecretFiles.Key, Computed: true},
						"description":              schema.StringAttribute{MarkdownDescription: externalSecretFiles.Description, Computed: true},
						"mount_path":               schema.StringAttribute{MarkdownDescription: externalSecretFiles.MountPath, Computed: true},
						"reference":                schema.StringAttribute{MarkdownDescription: externalSecretFiles.Reference, Computed: true},
						"secret_manager_access_id": schema.StringAttribute{MarkdownDescription: externalSecretFiles.SecretManagerAccessID, Computed: true},
					},
				},
			},
			"healthchecks": healthchecksSchemaAttributes(false),
			"custom_domains": schema.SetNestedAttribute{
				MarkdownDescription: customDomains.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: customDomains.ID,
							Computed:            true,
						},
						"domain": schema.StringAttribute{
							MarkdownDescription: customDomains.Domain,
							Computed:            true,
						},
						"generate_certificate": schema.BoolAttribute{
							MarkdownDescription: customDomains.GenerateCertificate,
							Optional:            true,
						},
						"use_cdn": schema.BoolAttribute{
							MarkdownDescription: customDomains.UseCDN,
							Optional:            true,
						},
						"validation_domain": schema.StringAttribute{
							MarkdownDescription: customDomains.ValidationDomain,
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: customDomains.Status,
							Computed:            true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				MarkdownDescription: externalHostDescription("application"),
				Computed:            true,
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("application"),
				Computed:            true,
			},
			"deployment_stage_id": schema.StringAttribute{
				MarkdownDescription: deploymentStageIDDescription,
				Optional:            true,
				Computed:            true,
			},
			"is_skipped": schema.BoolAttribute{
				MarkdownDescription: isSkippedDescription,
				Optional:            true,
				Computed:            true,
			},
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: dataSourceAdvancedSettingsJSONDescription("application"),
				Optional:            true,
				Computed:            true,
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: applicationAutoDeployDescription,
				Optional:            true,
				Computed:            true,
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				MarkdownDescription: restrictions.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: restrictions.ID,
							Computed:            true,
						},
						"mode": schema.StringAttribute{
							MarkdownDescription: restrictions.Mode,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: restrictions.Type,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: restrictions.Value,
							Computed:            true,
						},
					},
				},
			},
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("annotations", "application"),
				Computed:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("labels", "application"),
				Computed:            true,
				ElementType:         types.StringType,
			},
			"docker_target_build_stage": schema.StringAttribute{
				MarkdownDescription: dockerTargetBuildStageDescription,
				Optional:            true,
			},
		},
	}
}

// Read qovery application data source
func (d applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Application
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get application from API
	application, apiErr := d.client.GetApplication(ctx, data.Id.ValueString(), data.AdvancedSettingsJson.ValueString(), true)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Group ids and arguments report the API value; a data source has no plan to match, so none
	// reads as [].
	data.AnnotationsGroupIds = emptyStringSet()
	data.LabelsGroupIds = emptyStringSet()
	data.Arguments = emptyStringList()
	state := convertResponseToApplication(ctx, data, application)
	state.BuildSettings = buildSettingsFromQovery(application.ApplicationBuildSettings)
	tflog.Trace(ctx, "read application", map[string]any{"application_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
