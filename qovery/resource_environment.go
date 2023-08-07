package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &environmentResource{}
var _ resource.ResourceWithImportState = environmentResource{}

type environmentResource struct {
	environmentService environment.Service
}

func newEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

func (r environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.environmentService = provider.environmentService
}

func (r environmentResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery environment resource. This can be used to create and manage Qovery environments.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"project_id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Required:    true,
			},
			"cluster_id": {
				Description: "Id of the cluster [NOTE: can't be updated after creation].",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"mode": {
				Description: descriptions.NewStringEnumDescription(
					"Mode of the environment [NOTE: can't be updated after creation].",
					clientEnumToStringArray(environment.AllowedModeValues),
					pointer.ToString(environment.DefaultMode.String()),
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(environment.DefaultMode.String()),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(clientEnumToStringArray(environment.AllowedModeValues)),
				},
			},
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this environment.",
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
				Description: "List of environment variables linked to this environment.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
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
				}),
			},
			"environment_variable_aliases": {
				Description: "List of environment variable aliases linked to this environment.",
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
				Description: "List of environment variable overrides linked to this environment.",
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
				Description: "List of secrets linked to this environment.",
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
						Required:    true,
					},
					"value": {
						Description: "Value of the secret.",
						Type:        types.StringType,
						Required:    true,
						Sensitive:   true,
					},
				}),
			},
			"secret_aliases": {
				Description: "List of secret aliases linked to this environment.",
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
				Description: "List of secret overrides linked to this environment.",
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
		},
	}, nil
}

// Create qovery environment resource
func (r environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Environment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new environment
	request, err := plan.toCreateEnvironmentRequest()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	env, err := r.environmentService.Create(ctx, plan.ProjectId.Value, *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on environment create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainEnvironmentToEnvironment(plan, env)
	tflog.Trace(ctx, "created environment", map[string]interface{}{"environment_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery environment resource
func (r environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Environment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get environment from the API
	env, err := r.environmentService.Get(ctx, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on environment read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainEnvironmentToEnvironment(state, env)
	tflog.Trace(ctx, "read environment", map[string]interface{}{"environment_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery environment resource
func (r environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Environment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toUpdateEnvironmentRequest(state)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}

	// Update environment in the backend
	env, err := r.environmentService.Update(ctx, state.Id.Value, *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on environment update", err.Error())
		return
	}

	// Update state values
	state = convertDomainEnvironmentToEnvironment(plan, env)
	tflog.Trace(ctx, "updated environment", map[string]interface{}{"environment_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery environment resource
func (r environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Environment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete environment
	err := r.environmentService.Delete(ctx, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on environment delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted environment", map[string]interface{}{"environment_id": state.Id.Value})

	// Remove environment from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery environment resource using its id
func (r environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
