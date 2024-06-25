package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &helmResource{}
var _ resource.ResourceWithImportState = helmResource{}

var helmPortProtocols = clientEnumToStringArray(helm.AllowedProtocols)

type helmResource struct {
	helmService helm.Service
}

func newHelmResource() resource.Resource {
	return &helmResource{}
}

func (r helmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_helm"
}

func (r *helmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.helmService = provider.helmService
}

func (r helmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery helm resource. This can be used to create and manage Qovery helm registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the helm.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the helm.",
				Required:    true,
			},
			"timeout_sec": schema.Int64Attribute{
				Description: "Helm timeout in second",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(helm.DefaultTimeoutSec),
				//Required: true,
			},
			"auto_preview": schema.BoolAttribute{
				Description: "Specify if the environment preview option is activated or not for this helm.",
				Optional:    true,
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: " Specify if the service will be automatically updated on every new commit on the branch.",
				Optional:    true,
				Computed:    true,
			},
			"arguments": schema.SetAttribute{
				Description: "Helm arguments",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default: setdefault.StaticValue(
					types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("--wait"),
							types.StringValue("--atomic"),
							types.StringValue("--debug"),
						},
					),
				),
			},
			"allow_cluster_wide_resources": schema.BoolAttribute{
				Description: "Allow this chart to deploy resources outside of this environment namespace (including CRDs or non-namespaced resources)",
				Required:    true,
			},
			"source": schema.SingleNestedAttribute{
				Description: "Helm chart from a Helm repository or from a git repository",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"helm_repository": schema.SingleNestedAttribute{
						Description: "Helm repositories can be private or public",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"helm_repository_id": schema.StringAttribute{
								Description: "helm repository id",
								Required:    true,
							},
							"chart_name": schema.StringAttribute{
								Description: "Chart name",
								Required:    true,
							},
							"chart_version": schema.StringAttribute{
								Description: "Chart version",
								Required:    true,
							},
						},
					},
					"git_repository": schema.SingleNestedAttribute{
						Description: "Git repository",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Helm's source git repository URL",
								Required:    true,
							},
							"branch": schema.StringAttribute{
								Description: "Helm's source git repository branch",
								Optional:    true,
								Computed:    true,
							},
							"root_path": schema.StringAttribute{
								Description: "Helm's source git repository root path",
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("/"),
							},
							"git_token_id": schema.StringAttribute{
								Description: "The git token ID to be used",
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
			},
			"values_override": schema.SingleNestedAttribute{
				Description: "Define your own overrides to customize the helm chart behaviour.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"set": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"set_string": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"set_json": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"file": schema.SingleNestedAttribute{
						Description: "Define the overrides by selecting a YAML file from a git repository (preferred) or by passing raw YAML files.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"raw": schema.MapNestedAttribute{
								Description: "Raw YAML files",
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"content": schema.StringAttribute{
											Description: "content of the file",
											Required:    true,
										},
									},
								},
							},
							"git_repository": schema.SingleNestedAttribute{
								Description: "YAML file from a git repository",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description: "YAML file git repository URL",
										Required:    true,
									},
									"branch": schema.StringAttribute{
										Description: "YAML file git repository branch",
										Required:    true,
									},
									"paths": schema.SetAttribute{
										Description: "YAML files git repository paths",
										Required:    true,
										ElementType: types.StringType,
									},
									"git_token_id": schema.StringAttribute{
										Description: "The git token ID to be used",
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"ports": schema.MapNestedAttribute{
				Description: "List of ports linked to this helm.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_name": schema.StringAttribute{
							Description: "",
							Required:    true,
						},
						"namespace": schema.StringAttribute{
							Description: "",
							Optional:    true,
						},
						"internal_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"Internal port of the container.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							Required: true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"external_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"External port of the container.\n\t- Required if: `ports.publicly_accessible=true`.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							Required: true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the container.",
								helmPortProtocols,
								pointer.ToString(helm.DefaultProtocol.String()),
							),
							Validators: []validator.String{
								validators.NewStringEnumValidator(helmPortProtocols),
							},
							Optional: true,
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							Description: "If this port will be used for the root domain",
							Required:    true,
						},
					},
				},
			},
			"built_in_environment_variables": schema.SetNestedAttribute{
				Description: "List of built-in environment variables linked to this helm.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the environment variable.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description: "List of environment variables linked to this helm.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the environment variable.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable.",
							Required:    true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description: "List of environment variable aliases linked to this helm.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the environment variable alias.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Name of the variable to alias.",
							Required:    true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description: "List of environment variable overrides linked to this helm.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the environment variable override.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable override.",
							Required:    true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description: "List of secrets linked to this helm.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the secret.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the secret.",
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description: "List of secret aliases linked to this helm.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the secret alias.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Name of the secret to alias.",
							Required:    true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description: "List of secret overrides linked to this helm.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the secret override.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the secret override.",
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description: "The helm external FQDN host [NOTE: only if your helm is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The helm internal host.",
				Computed:    true,
			},
			"deployment_stage_id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				Optional:    true,
				Computed:    true,
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings.",
				Optional:    true,
				Computed:    true,
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				Description: "List of deployment restrictions",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the deployment restriction",
							Computed:    true,
						},
						"mode": schema.StringAttribute{
							Description: "Can be EXCLUDE or MATCH",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "Currently, only PATH is accepted",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the deployment restriction",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Create qovery helm resource
func (r helmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Helm
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new helm
	request, err := plan.toUpsertServiceRequest(nil)
	if err != nil {
		resp.Diagnostics.AddError("Error on helm create", err.Error())
		return
	}
	newHelm, err := r.helmService.Create(ctx, plan.EnvironmentID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on helm create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainHelmToHelm(ctx, plan, newHelm)
	tflog.Trace(ctx, "created helm", map[string]interface{}{"helm_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery helm resource
func (r helmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Helm
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get helm from the API
	newHelm, err := r.helmService.Get(ctx, state.ID.ValueString(), state.AdvancedSettingsJson.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on helm read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainHelmToHelm(ctx, state, newHelm)
	tflog.Trace(ctx, "read helm", map[string]interface{}{"helm_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery helm resource
func (r helmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Helm
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update helm in the backend
	request, err := plan.toUpsertServiceRequest(&state)
	if err != nil {
		resp.Diagnostics.AddError("Error on helm create", err.Error())
		return
	}
	newHelm, err := r.helmService.Update(ctx, state.ID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on helm update", err.Error())
		return
	}

	// Update state values
	state = convertDomainHelmToHelm(ctx, plan, newHelm)
	tflog.Trace(ctx, "updated helm", map[string]interface{}{"helm_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery helm resource
func (r helmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Helm
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete helm
	err := r.helmService.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on helm delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted helm", map[string]interface{}{"helm_id": state.ID.ValueString()})

	// Remove helm from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery helm resource using its id
func (r helmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
