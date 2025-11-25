package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
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
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery terraform service data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the terraform service.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the terraform service.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the terraform service.",
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Specify if the terraform service will be automatically updated on every new commit.",
				Computed:    true,
			},
			"git_repository": schema.SingleNestedAttribute{
				Description: "Terraform service git repository configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "Git repository URL.",
						Computed:    true,
					},
					"branch": schema.StringAttribute{
						Description: "Git branch.",
						Computed:    true,
					},
					"root_path": schema.StringAttribute{
						Description: "Git root path.",
						Computed:    true,
					},
					"git_token_id": schema.StringAttribute{
						Description: "Git token ID for private repositories.",
						Computed:    true,
					},
				},
			},
			"tfvar_files": schema.ListAttribute{
				Description: "List of .tfvars file paths relative to the root path.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"variable": schema.SetNestedAttribute{
				Description: "Terraform variables.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Variable key.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Variable value.",
							Computed:    true,
							Sensitive:   true,
						},
						"secret": schema.BoolAttribute{
							Description: "Is this variable a secret.",
							Computed:    true,
						},
					},
				},
			},
			"backend": schema.SingleNestedAttribute{
				Description: "Terraform backend configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"kubernetes": schema.SingleNestedAttribute{
						Description: "Use Kubernetes backend for state management.",
						Computed:    true,
						Attributes:  map[string]schema.Attribute{},
					},
					"user_provided": schema.SingleNestedAttribute{
						Description: "Use user-provided backend configuration.",
						Computed:    true,
						Attributes:  map[string]schema.Attribute{},
					},
				},
			},
			"engine": schema.StringAttribute{
				Description: "Terraform engine (TERRAFORM or OPEN_TOFU).",
				Computed:    true,
			},
			"engine_version": schema.SingleNestedAttribute{
				Description: "Terraform/OpenTofu engine version configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"explicit_version": schema.StringAttribute{
						Description: "Explicit version to use for the Terraform/OpenTofu binary.",
						Computed:    true,
					},
					"read_from_terraform_block": schema.BoolAttribute{
						Description: "Whether to read the version from the terraform block in the code.",
						Computed:    true,
					},
				},
			},
			"job_resources": schema.SingleNestedAttribute{
				Description: "Resource allocation for the Terraform job.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"cpu_milli": schema.Int64Attribute{
						Description: descriptions.NewInt64MinDescription(
							"CPU of the terraform job in millicores (m) [1000m = 1 CPU].",
							int64(terraformservice.MinCPU),
							toInt64Pointer(terraformservice.DefaultCPU),
						),
						Computed: true,
					},
					"ram_mib": schema.Int64Attribute{
						Description: descriptions.NewInt64MinDescription(
							"RAM of the terraform job in MiB [1024 MiB = 1GiB].",
							int64(terraformservice.MinRAM),
							toInt64Pointer(terraformservice.DefaultRAM),
						),
						Computed: true,
					},
					"gpu": schema.Int64Attribute{
						Description: descriptions.NewInt64MinDescription(
							"Number of GPUs for the terraform job.",
							int64(terraformservice.MinGPU),
							toInt64Pointer(terraformservice.DefaultGPU),
						),
						Computed: true,
					},
					"storage_gib": schema.Int64Attribute{
						Description: descriptions.NewInt64MinDescription(
							"Storage of the terraform job in GiB [1 GiB = 1024 MiB].",
							int64(terraformservice.MinStorage),
							toInt64Pointer(terraformservice.DefaultStorage),
						),
						Computed: true,
					},
				},
			},
			"timeout_sec": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Timeout in seconds for Terraform operations.",
					int64(terraformservice.MinTimeoutSec),
					toInt64Pointer(terraformservice.DefaultTimeoutSec),
				),
				Computed: true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the terraform service.",
				Computed:    true,
			},
			"use_cluster_credentials": schema.BoolAttribute{
				Description: "Use cluster credentials for cloud provider authentication.",
				Computed:    true,
			},
			"action_extra_arguments": schema.MapAttribute{
				Description: "Extra CLI arguments for specific Terraform actions.",
				Computed:    true,
				ElementType: types.ListType{ElemType: types.StringType},
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings in JSON format.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation date of the terraform service.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update date of the terraform service.",
				Computed:    true,
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
	tflog.Trace(ctx, "read terraform service", map[string]interface{}{"terraform_service_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
