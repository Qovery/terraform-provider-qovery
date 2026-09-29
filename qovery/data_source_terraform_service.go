package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &terraformServiceDataSource{}

type terraformServiceDataSource struct {
	terraformServiceService terraformservice.Service
}

func newTerraformServiceDataSource() datasource.DataSource {
	return &terraformServiceDataSource{}
}

func (d terraformServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_service"
}

func (d *terraformServiceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.terraformServiceService = provider.terraformServiceService
}

func (d terraformServiceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	externalSecrets := variableListDescriptions("external_secrets", "Terraform service")
	externalSecretFiles := variableListDescriptions("external_secret_files", "Terraform service")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery Terraform service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("Terraform service"),
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription,
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
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("Terraform service"),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("Terraform service"),
				Computed:            true,
			},
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: createdFromBlueprintIDDescription("Terraform service"),
				Computed:            true,
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: terraformServiceAutoDeployDescription,
				Computed:            true,
			},
			"terraform_action": schema.StringAttribute{
				MarkdownDescription: terraformServiceActionDescription,
				Computed:            true,
			},
			"git_repository": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceGitRepositoryDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: gitRepositoryURLDescription,
						Computed:            true,
					},
					"branch": schema.StringAttribute{
						MarkdownDescription: terraformServiceBranchDescription,
						Computed:            true,
					},
					"root_path": schema.StringAttribute{
						MarkdownDescription: terraformServiceRootPathDescription,
						Computed:            true,
					},
					"git_token_id": schema.StringAttribute{
						MarkdownDescription: gitRepositoryTokenIDDescription,
						Computed:            true,
					},
				},
			},
			"tfvars_files": schema.ListAttribute{
				MarkdownDescription: terraformServiceTfvarsFilesDescription,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"variables": schema.SetNestedAttribute{
				MarkdownDescription: terraformServiceVariablesDescription,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: terraformServiceVariableKeyDescription,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: terraformServiceVariableValueDescription + " Null for a secret variable.",
							Computed:            true,
							Sensitive:           true,
						},
						"is_secret": schema.BoolAttribute{
							MarkdownDescription: terraformServiceVariableIsSecretDescription,
							Computed:            true,
						},
					},
				},
			},
			"backend": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceBackendDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"kubernetes": schema.SingleNestedAttribute{
						MarkdownDescription: terraformServiceKubernetesBackendDescription,
						Computed:            true,
						Attributes:          map[string]schema.Attribute{},
					},
					"user_provided": schema.SingleNestedAttribute{
						MarkdownDescription: terraformServiceUserProvidedBackendDescription,
						Computed:            true,
						Attributes:          map[string]schema.Attribute{},
					},
					"blueprint": schema.SingleNestedAttribute{
						MarkdownDescription: terraformServiceBlueprintBackendDescription,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								MarkdownDescription: terraformServiceBlueprintBackendTypeDescription,
								Computed:            true,
							},
							"config": schema.MapAttribute{
								MarkdownDescription: terraformServiceBlueprintBackendConfigDescription,
								Computed:            true,
								ElementType:         types.StringType,
							},
						},
					},
				},
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: terraformServiceEngineDescription,
				Computed:            true,
			},
			"engine_version": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceEngineVersionDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"explicit_version": schema.StringAttribute{
						MarkdownDescription: terraformServiceExplicitVersionDescription,
						Computed:            true,
					},
					"read_from_terraform_block": schema.BoolAttribute{
						MarkdownDescription: terraformServiceReadFromTerraformBlockDescription,
						Computed:            true,
					},
				},
			},
			"job_resources": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceJobResourcesDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cpu_milli": schema.Int64Attribute{
						MarkdownDescription: cpuDescription("Terraform job"),
						Computed:            true,
					},
					"ram_mib": schema.Int64Attribute{
						MarkdownDescription: terraformServiceJobRAMDescription,
						Computed:            true,
					},
					"gpu": schema.Int64Attribute{
						MarkdownDescription: terraformServiceJobGPUDescription,
						Computed:            true,
					},
					"storage_gib": schema.Int64Attribute{
						MarkdownDescription: terraformServiceJobStorageDescription,
						Computed:            true,
					},
				},
			},
			"timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: terraformServiceTimeoutDescription,
				Computed:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: iconURIDescription("Terraform service"),
				Computed:            true,
			},
			"use_cluster_credentials": schema.BoolAttribute{
				MarkdownDescription: terraformServiceUseClusterCredentialsDescription,
				Computed:            true,
			},
			"action_extra_arguments": schema.MapAttribute{
				MarkdownDescription: terraformServiceActionExtraArgumentsDescription,
				Computed:            true,
				ElementType:         types.ListType{ElemType: types.StringType},
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
			"build_settings": buildSettingsDataSourceSchemaAttributes(),
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: dataSourceAdvancedSettingsJSONDescription("Terraform service"),
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: terraformServiceCreatedAtDescription,
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: terraformServiceUpdatedAtDescription,
				Computed:            true,
			},
		},
	}
}

func (d terraformServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current config
	var data TerraformService
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get terraform service from API
	terraformSvc, err := d.terraformServiceService.Get(
		ctx,
		ToString(data.ID),
		ToString(data.AdvancedSettingsJson),
		false,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error on terraform service read", err.Error())
		return
	}

	// Convert domain entity to Terraform state
	state := convertDomainTerraformServiceToTerraformService(ctx, data, terraformSvc)
	state.Variables = fromDataSourceVariableArray(terraformSvc.Variables)
	state.BuildSettings = buildSettingsFromQovery(terraformSvc.BuildSettings)
	tflog.Trace(ctx, "read terraform service", map[string]any{"terraform_service_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
