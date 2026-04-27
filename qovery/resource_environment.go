package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
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

func (r environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery environment resource. This can be used to create and manage Qovery environments. " +
			"An environment is an isolated workspace within a project where services (applications, containers, databases, jobs) are deployed. " +
			"Environment variables and secrets defined at the environment level are inherited by all services within the environment, and can override project-level variables.",
		MarkdownDescription: "Provides a Qovery environment resource. This can be used to create and manage Qovery environments.\n\n" +
			"An environment is an isolated workspace within a project where services (applications, containers, databases, jobs) are deployed. " +
			"Environment variables and secrets defined at the environment level are inherited by all services within the environment, and can override project-level variables.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Unique identifier of the environment (UUID format).",
				MarkdownDescription: "Unique identifier of the environment (UUID format).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description:         "Identifier of the project containing this environment (UUID format).",
				MarkdownDescription: "Identifier of the project containing this environment (UUID format).",
				Required:            true,
			},
			"cluster_id": schema.StringAttribute{
				Description:         "Identifier of the cluster where this environment will be deployed (UUID format). Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "Identifier of the cluster where this environment will be deployed (UUID format). **Cannot be changed after creation** (forces resource replacement).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "Name of the environment.",
				MarkdownDescription: "Name of the environment.",
				Required:            true,
			},
			"mode": schema.StringAttribute{
				Description: "Mode of the environment. The mode affects how the environment behaves and is displayed in the Qovery console.",
				MarkdownDescription: descriptions.NewStringEnumDescription(
					"Mode of the environment. The mode affects how the environment behaves and is displayed in the Qovery console.",
					clientEnumToStringArray(environment.AllowedModeValues),
					new(environment.DefaultMode.String()),
				),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(environment.DefaultMode.String()),
				Validators: []validator.String{
					validators.NewStringEnumValidator(clientEnumToStringArray(environment.AllowedModeValues)),
				},
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				Description:         "List of built-in environment variables linked to this environment. Built-in variables are automatically generated by Qovery and provide metadata about the environment (e.g., environment ID, cluster ID).",
				MarkdownDescription: "List of built-in environment variables linked to this environment. Built-in variables are automatically generated by Qovery and provide metadata about the environment (e.g., environment ID, cluster ID).",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					UseStateUnlessNameChanges(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable.",
							MarkdownDescription: "Identifier of the environment variable.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable.",
							MarkdownDescription: "Key of the environment variable.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable.",
							MarkdownDescription: "Value of the environment variable.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable.",
							MarkdownDescription: "Description of the environment variable.",
							Computed:            true,
						},
					},
				},
			},
			"environment_variables": schema.SetNestedAttribute{
				Description:         "Set of environment variables linked to this environment. These variables are inherited by all services within the environment.",
				MarkdownDescription: "Set of environment variables linked to this environment. These variables are inherited by all services within the environment.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable.",
							MarkdownDescription: "Identifier of the environment variable.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable.",
							MarkdownDescription: "Key of the environment variable.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable.",
							MarkdownDescription: "Value of the environment variable.",
							Required:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable.",
							MarkdownDescription: "Description of the environment variable.",
							Optional:            true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description:         "Set of environment variable aliases linked to this environment. An alias creates an alternative name that points to an existing environment variable.",
				MarkdownDescription: "Set of environment variable aliases linked to this environment. An alias creates an alternative name that points to an existing environment variable.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable alias.",
							MarkdownDescription: "Identifier of the environment variable alias.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the alias. This is the new key that will be available as an environment variable.",
							MarkdownDescription: "Name of the alias. This is the new key that will be available as an environment variable.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the variable to alias. Must match the key of an existing environment variable.",
							MarkdownDescription: "Name of the variable to alias. Must match the `key` of an existing environment variable.",
							Required:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable alias.",
							MarkdownDescription: "Description of the environment variable alias.",
							Optional:            true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description:         "Set of environment variable overrides linked to this environment. An override replaces the value of a variable inherited from the project level.",
				MarkdownDescription: "Set of environment variable overrides linked to this environment. An override replaces the value of a variable inherited from the project level.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable override.",
							MarkdownDescription: "Identifier of the environment variable override.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable to override. Must match the key of a variable defined at a higher scope (e.g., project level).",
							MarkdownDescription: "Name of the environment variable to override. Must match the `key` of a variable defined at a higher scope (e.g., project level).",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Override value of the environment variable.",
							MarkdownDescription: "Override value of the environment variable.",
							Required:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable override.",
							MarkdownDescription: "Description of the environment variable override.",
							Optional:            true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description:         "Set of secrets linked to this environment. Secrets are like environment variables but their values are encrypted and not visible after creation. They are inherited by all services within the environment.",
				MarkdownDescription: "Set of secrets linked to this environment. Secrets are like environment variables but their values are encrypted and not visible after creation. They are inherited by all services within the environment.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the secret.",
							MarkdownDescription: "Identifier of the secret.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the secret.",
							MarkdownDescription: "Key of the secret.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret. The value is write-only and will not be displayed in plan output.",
							MarkdownDescription: "Value of the secret. The value is write-only and will not be displayed in plan output.",
							Required:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret.",
							MarkdownDescription: "Description of the secret.",
							Optional:            true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description:         "Set of secret aliases linked to this environment. An alias creates an alternative name that points to an existing secret.",
				MarkdownDescription: "Set of secret aliases linked to this environment. An alias creates an alternative name that points to an existing secret.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the secret alias.",
							MarkdownDescription: "Identifier of the secret alias.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the alias. This is the new key that will be available as a secret.",
							MarkdownDescription: "Name of the alias. This is the new key that will be available as a secret.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the secret to alias. Must match the key of an existing secret.",
							MarkdownDescription: "Name of the secret to alias. Must match the `key` of an existing secret.",
							Required:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret alias.",
							MarkdownDescription: "Description of the secret alias.",
							Optional:            true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description:         "Set of secret overrides linked to this environment. An override replaces the value of a secret inherited from the project level.",
				MarkdownDescription: "Set of secret overrides linked to this environment. An override replaces the value of a secret inherited from the project level.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the secret override.",
							MarkdownDescription: "Identifier of the secret override.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret to override. Must match the key of a secret defined at a higher scope (e.g., project level).",
							MarkdownDescription: "Name of the secret to override. Must match the `key` of a secret defined at a higher scope (e.g., project level).",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Override value of the secret. The value is write-only and will not be displayed in plan output.",
							MarkdownDescription: "Override value of the secret. The value is write-only and will not be displayed in plan output.",
							Required:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret override.",
							MarkdownDescription: "Description of the secret override.",
							Optional:            true,
						},
					},
				},
			},
			"environment_variable_files": environmentVariableFilesSchemaAttribute("environment"),
			"secret_files":              secretFilesSchemaAttribute("environment"),
		},
	}
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

	env, err := r.environmentService.Create(ctx, plan.ProjectId.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on environment create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainEnvironmentToEnvironment(ctx, plan, env)
	tflog.Trace(ctx, "created environment", map[string]any{"environment_id": state.Id.ValueString()})

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
	env, err := r.environmentService.Get(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on environment read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainEnvironmentToEnvironment(ctx, state, env)
	tflog.Trace(ctx, "read environment", map[string]any{"environment_id": state.Id.ValueString()})

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
	env, err := r.environmentService.Update(ctx, state.Id.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on environment update", err.Error())
		return
	}

	// Update state values
	state = convertDomainEnvironmentToEnvironment(ctx, plan, env)
	tflog.Trace(ctx, "updated environment", map[string]any{"environment_id": state.Id.ValueString()})

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
	err := r.environmentService.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on environment delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted environment", map[string]any{"environment_id": state.Id.ValueString()})

	// Remove environment from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery environment resource using its id
func (r environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
