package qovery

import (
	"context"
	"fmt"

	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &jobDataSource{}

type jobDataSource struct {
	jobService job.Service
}

func newJobDataSource() datasource.DataSource {
	return &jobDataSource{}
}

func (d jobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (d *jobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.jobService = provider.jobService
}

func (d jobDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing job.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the job.",
				Type:        types.StringType,
				Required:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the job.",
				Type:        types.StringType,
				Computed:    true,
			},
			"cpu": {
				Description: "CPU of the job in millicores (m) [1000m = 1 CPU].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"memory": {
				Description: "RAM of the job in MB [1024MB = 1GB].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"max_duration_seconds": {
				Description: "Job's max duration in seconds.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"max_nb_restart": {
				Description: "Job's max number of restarts",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"auto_preview": {
				Description: "Specify if the environment preview option is activated or not for this job.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"port": {
				Description: "Job's probes port.",
				Type:        types.Int64Type,
				Computed:    true,
				Optional:    true,
			},
			"healthchecks": healthchecksSchemaAttributes(false),
			"schedule": {
				Description: "Job's schedule.",
				Computed:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"on_start": {
						Description: "Job's schedule on start.",
						Computed:    true,
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"entrypoint": {
								Description: "Entrypoint of the job.",
								Type:        types.StringType,
								Computed:    true,
								Optional:    true,
							},
							"arguments": {
								Description: "List of arguments of this job.",
								Optional:    true,
								Computed:    true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
								PlanModifiers: tfsdk.AttributePlanModifiers{
									modifiers.NewStringSliceDefaultModifier([]string{}),
								},
							},
						}),
					},
					"on_stop": {
						Description: "Job's schedule on stop.",
						Computed:    true,
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"entrypoint": {
								Description: "Entrypoint of the job.",
								Type:        types.StringType,
								Computed:    true,
								Optional:    true,
							},
							"arguments": {
								Description: "List of arguments of this job.",
								Optional:    true,
								Computed:    true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
								PlanModifiers: tfsdk.AttributePlanModifiers{
									modifiers.NewStringSliceDefaultModifier([]string{}),
								},
							},
						}),
					},
					"on_delete": {
						Description: "Job's schedule on delete.",
						Computed:    true,
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"entrypoint": {
								Description: "Entrypoint of the job.",
								Type:        types.StringType,
								Computed:    true,
								Optional:    true,
							},
							"arguments": {
								Description: "List of arguments of this job.",
								Optional:    true,
								Computed:    true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
								PlanModifiers: tfsdk.AttributePlanModifiers{
									modifiers.NewStringSliceDefaultModifier([]string{}),
								},
							},
						}),
					},
					"cronjob": {
						Description: "Job's cron.",
						Computed:    true,
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"schedule": {
								Description: "Job's cron string.",
								Type:        types.StringType,
								Computed:    true,
								Optional:    false,
							},
							"command": {
								Description: "Job's cron command.",
								Computed:    true,
								Optional:    false,
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"entrypoint": {
										Description: "Entrypoint of the job.",
										Type:        types.StringType,
										Computed:    true,
										Optional:    true,
									},
									"arguments": {
										Description: "List of arguments of this job.",
										Optional:    true,
										Computed:    true,
										Type: types.ListType{
											ElemType: types.StringType,
										},
										PlanModifiers: tfsdk.AttributePlanModifiers{
											modifiers.NewStringSliceDefaultModifier([]string{}),
										},
									},
								}),
							},
						}),
					},
				}),
			},
			"source": {
				Description: "Job's source.",
				Computed:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"image": {
						Description: "Job's image source.",
						Computed:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"registry_id": {
								Description: "Job's image source registry ID.",
								Type:        types.StringType,
								Computed:    true,
							},
							"name": {
								Description: "Job's image source name.",
								Type:        types.StringType,
								Computed:    true,
							},
							"tag": {
								Description: "Job's image source tag.",
								Type:        types.StringType,
								Computed:    true,
							},
						}),
					},
					"docker": {
						Description: "Job's docker source.",
						Computed:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"dockerfile_path": {
								Description: "Job's docker source dockerfile path.",
								Type:        types.StringType,
								Computed:    true,
							},
							"git_repository": {
								Description: "Job's docker source git repository.",
								Computed:    true,
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"url": {
										Description: "Job's docker source git repository URL.",
										Type:        types.StringType,
										Computed:    true,
									},
									"branch": {
										Description: "Job's docker source git repository branch.",
										Type:        types.StringType,
										Computed:    true,
									},
									"root_path": {
										Description: "Job's docker source git repository root path.",
										Type:        types.StringType,
										Computed:    true,
									},
								}),
							},
						}),
					},
				}),
			},
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this job.",
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
				}),
			},
			"environment_variables": {
				Description: "List of environment variables linked to this job.",
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
				}),
			},
			"environment_variable_aliases": {
				Description: "List of environment variable aliases linked to this job.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable alias.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the environment variable alias.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Name of the variable to alias.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"environment_variable_overrides": {
				Description: "List of environment variable overrides linked to this job.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable override.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the environment variable override.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the environment variable override.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"secrets": {
				Description: "List of secrets linked to this job.",
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
				}),
			},
			"secret_aliases": {
				Description: "List of secret aliases linked to this job.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret alias.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the secret alias.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Name of the secret to alias.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"secret_overrides": {
				Description: "List of secret overrides linked to this job.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret override.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the secret override.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the secret override.",
						Type:        types.StringType,
						Required:    true,
						Sensitive:   true,
					},
				}),
			},
			"external_host": {
				Description: "The job external FQDN host [NOTE: only if your job is using a publicly accessible port].",
				Type:        types.StringType,
				Computed:    true,
			},
			"internal_host": {
				Description: "The job internal host.",
				Type:        types.StringType,
				Computed:    true,
			},
			"deployment_stage_id": {
				Description: "Id of the deployment stage.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"advanced_settings_json": {
				Description: "Advanced settings.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"auto_deploy": {
				Description: "Specify if the job will be automatically updated after receiving a new image tag or a new commit according to the source type.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
			},
		},
	}, nil
}

// Read qovery job data source
func (d jobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Job
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get job from API
	cont, err := d.jobService.Get(ctx, data.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on job read", err.Error())
		return
	}

	state := convertDomainJobToJob(data, cont)
	tflog.Trace(ctx, "read job", map[string]interface{}{"job_id": state.ID.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
