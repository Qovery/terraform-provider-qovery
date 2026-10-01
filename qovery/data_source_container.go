package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &containerDataSource{}

type containerDataSource struct {
	containerService container.Service
}

func newContainerDataSource() datasource.DataSource {
	return &containerDataSource{}
}

func (d containerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

func (d *containerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.containerService = provider.containerService
}

func (r containerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "container")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "container")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "container")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "container")
	secrets := variableListDescriptions("secrets", "container")
	secretAliases := variableListDescriptions("secret_aliases", "container")
	secretOverrides := variableListDescriptions("secret_overrides", "container")
	envVarFiles := variableListDescriptions("environment_variable_files", "container")
	secretFiles := variableListDescriptions("secret_files", "container")
	externalSecrets := variableListDescriptions("external_secrets", "container")
	externalSecretFiles := variableListDescriptions("external_secret_files", "container")
	ports := portDescriptions("container")
	storages := storageDescriptions("container")
	customDomains := customDomainDescriptions("container")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery container.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("container"),
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription,
				Computed:            true,
			},
			"registry_id": schema.StringAttribute{
				MarkdownDescription: containerImageRegistryIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("container"),
				Computed:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: iconURIDescription("container"),
				Optional:            true,
				Computed:            true,
			},
			"image_name": schema.StringAttribute{
				MarkdownDescription: containerImageNameDescription,
				Computed:            true,
			},
			"tag": schema.StringAttribute{
				MarkdownDescription: containerImageTagDescription,
				Computed:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: cpuDescription("container"),
				Optional:            true,
				Computed:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: memoryDescription("container"),
				Optional:            true,
				Computed:            true,
			},
			"ephemeral_storage": schema.Int64Attribute{
				MarkdownDescription: ephemeralStorageDescription("container"),
				Computed:            true,
			},
			"min_running_instances": schema.Int64Attribute{
				MarkdownDescription: minRunningInstancesDescription("container"),
				Optional:            true,
				Computed:            true,
			},
			"max_running_instances": schema.Int64Attribute{
				MarkdownDescription: maxRunningInstancesDescription("container"),
				Optional:            true,
				Computed:            true,
			},
			"autoscaling": autoscalingDataSourceSchema(),
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: autoPreviewDescription("container"),
				Optional:            true,
				Computed:            true,
			},
			"entrypoint": schema.StringAttribute{
				MarkdownDescription: entrypointDescription,
				Optional:            true,
				Computed:            true,
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
			"ports": schema.ListNestedAttribute{
				MarkdownDescription: ports.List,
				Optional:            true,
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
							Optional:            true,
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
							MarkdownDescription: dataSourcePortIsDefaultDescription("container"),
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
						"id":                       schema.StringAttribute{MarkdownDescription: externalSecrets.ID, Computed: true},
						"key":                      schema.StringAttribute{MarkdownDescription: externalSecrets.Key, Computed: true},
						"description":              schema.StringAttribute{MarkdownDescription: externalSecrets.Description, Computed: true},
						"reference":                schema.StringAttribute{MarkdownDescription: externalSecrets.Reference, Computed: true},
						"secret_manager_access_id": schema.StringAttribute{MarkdownDescription: externalSecrets.SecretManagerAccessID, Computed: true},
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
			"arguments": schema.ListAttribute{
				MarkdownDescription: argumentsDescription,
				Optional:            true,
				ElementType:         types.StringType,
			},
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
						"validation_domain": schema.StringAttribute{
							MarkdownDescription: customDomains.ValidationDomain,
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
						"status": schema.StringAttribute{
							MarkdownDescription: customDomains.Status,
							Computed:            true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				MarkdownDescription: externalHostDescription("container"),
				Computed:            true,
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("container"),
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
				MarkdownDescription: dataSourceAdvancedSettingsJSONDescription("container"),
				Optional:            true,
				Computed:            true,
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: containerAutoDeployDescription,
				Optional:            true,
				Computed:            true,
			},
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("annotations", "container"),
				Computed:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("labels", "container"),
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Read qovery container data source
func (d containerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Container
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get container from API
	cont, err := d.containerService.Get(ctx, data.ID.ValueString(), data.AdvancedSettingsJson.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Error on container read", err.Error())
		return
	}

	// Group ids, arguments and entrypoint report the API value; a data source has no plan to
	// match, so none reads as [] and an empty entrypoint as "".
	data.AnnotationsGroupIds = emptyStringSet()
	data.LabelsGroupIds = emptyStringSet()
	data.Arguments = emptyStringList()
	data.Entrypoint = types.StringValue("")
	state := convertDomainContainerToContainer(ctx, data, cont)
	tflog.Trace(ctx, "read container", map[string]any{"container_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
