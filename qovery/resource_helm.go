package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &helmResource{}
	_ resource.ResourceWithImportState = helmResource{}
	_ resource.ResourceWithModifyPlan  = helmResource{}
)

var helmPortProtocols = clientEnumToStringArray(helm.AllowedProtocols)

type helmResource struct {
	helmService             helm.Service
	advancedSettingsService *advanced_settings.ServiceAdvancedSettingsService
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
	r.advancedSettingsService = provider.advancedSettingsService
}

func (r helmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "Helm service")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "Helm service")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "Helm service")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "Helm service")
	secrets := variableListDescriptions("secrets", "Helm service")
	secretAliases := variableListDescriptions("secret_aliases", "Helm service")
	secretOverrides := variableListDescriptions("secret_overrides", "Helm service")
	ports := portDescriptions("Helm service")
	customDomains := customDomainDescriptions("Helm service")
	restrictions := deploymentRestrictionDescriptions("Helm service")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Qovery Helm service: a Helm chart that Qovery deploys to its environment, from a Helm repository or a git repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("Helm service"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription + recreatesOnChange("Helm service"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("Helm service"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("Helm service"),
				Required:            true,
			},
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: createdFromBlueprintIDDescription("Helm service") + blueprintIDRemovalNote,
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: q-core records the
				// blueprint only on create, so removal keeps it and a change is a plan error.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					RejectChangeAfterCreate(blueprintIDChangeReason),
				},
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(iconURIDescription("Helm service"), helmIconURIDefault),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(helmIconURIDefault),
			},
			"timeout_sec": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(helmTimeoutSecDescription, helm.DefaultTimeoutSec),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(helm.DefaultTimeoutSec),
				// Required: true,
			},
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(autoPreviewDescription("Helm service"), serviceAutoPreviewDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(serviceAutoPreviewDefault),
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(helmAutoDeployDescription, helmAutoDeployDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(helmAutoDeployDefault),
			},
			"arguments": schema.ListAttribute{
				MarkdownDescription: helmArgumentsDescription + helmArgumentsDefaultNote,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default: listdefault.StaticValue(
					types.ListValueMust(
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
				MarkdownDescription: helmAllowClusterWideResourcesDescription,
				Required:            true,
			},
			"source": schema.SingleNestedAttribute{
				MarkdownDescription: helmSourceDescription,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"helm_repository": schema.SingleNestedAttribute{
						MarkdownDescription: helmSourceHelmRepositoryDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"helm_repository_id": schema.StringAttribute{
								MarkdownDescription: helmSourceHelmRepositoryIDDesc,
								Required:            true,
							},
							"chart_name": schema.StringAttribute{
								MarkdownDescription: helmSourceChartNameDescription,
								Required:            true,
							},
							"chart_version": schema.StringAttribute{
								MarkdownDescription: helmSourceChartVersionDescription,
								Required:            true,
							},
						},
					},
					"git_repository": schema.SingleNestedAttribute{
						MarkdownDescription: helmSourceGitRepositoryDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								MarkdownDescription: gitRepositoryURLDescription,
								Required:            true,
							},
							"branch": schema.StringAttribute{
								MarkdownDescription: helmSourceBranchDescription + gitBranchRemovalNote,
								Optional:            true,
								Computed:            true,
								// Documented exception to the config-is-source-of-truth rule: the
								// default branch depends on the repository, so there is no static
								// Default, and planning unknown whenever it is omitted would give a
								// permanent diff.
								PlanModifiers: []planmodifier.String{
									UseStateUnlessRepositoryChanges(),
								},
							},
							"root_path": schema.StringAttribute{
								MarkdownDescription: descriptions.NewStringDefaultDescription(helmSourceRootPathDescription, helmSourceRootPathDefault),
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString("/"),
							},
							"git_token_id": schema.StringAttribute{
								MarkdownDescription: gitRepositoryTokenIDDescription,
								Optional:            true,
							},
						},
					},
				},
			},
			"values_override": schema.SingleNestedAttribute{
				MarkdownDescription: helmValuesOverrideDescription,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"set": schema.MapAttribute{
						MarkdownDescription: helmValuesSetDescription,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"set_string": schema.MapAttribute{
						MarkdownDescription: helmValuesSetStringDescription,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"set_json": schema.MapAttribute{
						MarkdownDescription: helmValuesSetJSONDescription,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"file": schema.SingleNestedAttribute{
						MarkdownDescription: helmValuesFileDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"raw": schema.MapNestedAttribute{
								MarkdownDescription: helmValuesRawDescription,
								Optional:            true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"content": schema.StringAttribute{
											MarkdownDescription: helmValuesRawContentDescription,
											Required:            true,
										},
									},
								},
							},
							"git_repository": schema.SingleNestedAttribute{
								MarkdownDescription: helmValuesGitRepositoryDescription,
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										MarkdownDescription: gitRepositoryURLDescription,
										Required:            true,
									},
									"branch": schema.StringAttribute{
										MarkdownDescription: helmValuesGitBranchDescription,
										Required:            true,
									},
									"paths": schema.SetAttribute{
										MarkdownDescription: helmValuesGitPathsDescription,
										Required:            true,
										ElementType:         types.StringType,
									},
									"git_token_id": schema.StringAttribute{
										MarkdownDescription: gitRepositoryTokenIDDescription,
										Optional:            true,
									},
								},
							},
						},
					},
				},
			},
			"ports": schema.MapNestedAttribute{
				MarkdownDescription: helmPortsDescription,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_name": schema.StringAttribute{
							MarkdownDescription: helmPortServiceNameDescription,
							Required:            true,
						},
						"namespace": schema.StringAttribute{
							MarkdownDescription: helmPortNamespaceDescription,
							Optional:            true,
						},
						"internal_port": schema.Int64Attribute{
							MarkdownDescription: descriptions.NewInt64MinMaxDescription(helmPortInternalPortDescription, port.MinPort, port.MaxPort, nil),
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"external_port": schema.Int64Attribute{
							MarkdownDescription: descriptions.NewInt64MinMaxDescription(helmPortExternalPortDescription, port.MinPort, port.MaxPort, nil),
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: descriptions.NewStringEnumDescription(ports.Protocol, helmPortProtocols, new(helm.DefaultProtocol.String())),
							Validators: []validator.String{
								validators.NewStringEnumValidator(helmPortProtocols),
							},
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(helm.DefaultProtocol.String()),
						},
						"is_default": schema.BoolAttribute{
							MarkdownDescription: ports.IsDefault,
							Optional:            true,
							Computed:            true,
							// Documented exception to the config-is-source-of-truth rule: the API
							// forces one default port (see smartAllowApiOverrideModifier).
							PlanModifiers: []planmodifier.Bool{
								SmartAllowApiOverride(),
							},
						},
					},
				},
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: builtInEnvVars.List,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					UseStateUnlessNameChanges(),
				},
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVars.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVars.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secrets.Value,
							Required:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secrets.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretAliases.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretAliases.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Value,
							Required:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Description,
							Optional:            true,
						},
					},
				},
			},
			"environment_variable_files": environmentVariableFilesSchemaAttribute("Helm service"),
			"secret_files":               secretFilesSchemaAttribute("Helm service"),
			"external_secrets":           externalSecretsSchemaAttribute("Helm service"),
			"external_secret_files":      externalSecretFilesSchemaAttribute("Helm service"),
			"custom_domains": schema.SetNestedAttribute{
				MarkdownDescription: customDomains.List,
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					CustomDomainsBoolDefaults("use_cdn"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: customDomains.ID,
							Computed:            true,
						},
						"domain": schema.StringAttribute{
							MarkdownDescription: customDomains.Domain,
							Required:            true,
						},
						"generate_certificate": schema.BoolAttribute{
							MarkdownDescription: customDomains.GenerateCertificate,
							Required:            true,
						},
						// use_cdn defaults to false through the CustomDomainsBoolDefaults plan
						// modifier on custom_domains: a Default nested in a set breaks the matching
						// of planned and applied elements.
						"use_cdn": schema.BoolAttribute{
							MarkdownDescription: descriptions.NewBoolDefaultDescription(customDomains.UseCDN, false),
							Optional:            true,
							Computed:            true,
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
				MarkdownDescription: externalHostDescription("Helm service"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("Helm service"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_stage_id": schema.StringAttribute{
				MarkdownDescription: deploymentStageIDDescription + deploymentStageIDRemovalNote,
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: q-core attaches
				// every service to a stage and has no detach, so removal keeps the current stage.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_skipped": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(isSkippedDescription, false),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: advancedSettingsJSONDescription("Helms/operation/getDefaultHelmAdvancedSettings"),
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: the QOV-2028
				// contract described in advancedSettingsJSONDescription.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: restrictions.Type,
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: restrictions.Value,
							Required:            true,
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
		if newHelm == nil {
			return
		}
		// The helm service was created in Qovery and only a follow-up call failed. Save it
		// so Terraform tracks it and taints it for replacement, instead of leaving an orphan
		// that makes every later apply fail with "a helm named X already exists".
		resp.Diagnostics.Append(resp.State.Set(ctx, convertDomainHelmToHelm(ctx, plan, newHelm))...)
		return
	}

	// Initialize state values
	state := convertDomainHelmToHelm(ctx, plan, newHelm)
	tflog.Trace(ctx, "created helm", map[string]any{"helm_id": state.ID.ValueString()})

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

	// Hack to know if this method is triggered through an import
	// EnvironmentID is always present except when importing the resource
	isTriggeredFromImport := false
	if state.EnvironmentID.IsNull() {
		isTriggeredFromImport = true
	}

	// Get helm from the API
	newHelm, err := r.helmService.Get(ctx, state.ID.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if handleDomainReadNotFound(ctx, resp, err, "Error on helm read") {
		return
	}

	// Refresh state values
	state = convertDomainHelmToHelm(ctx, state, newHelm)
	tflog.Trace(ctx, "read helm", map[string]any{"helm_id": state.ID.ValueString()})

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
	tflog.Trace(ctx, "updated helm", map[string]any{"helm_id": state.ID.ValueString()})

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

	tflog.Trace(ctx, "deleted helm", map[string]any{"helm_id": state.ID.ValueString()})

	// Remove helm from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery helm resource using its id
func (r helmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r helmResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	warnUnknownAdvancedSettings(ctx, r.advancedSettingsService, domain.HELM, req.Config, &resp.Diagnostics)
}
