package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure    = &applicationResource{}
	_ resource.ResourceWithImportState  = applicationResource{}
	_ resource.ResourceWithModifyPlan   = applicationResource{}
	_ resource.ResourceWithUpgradeState = applicationResource{}
)

var (

	// Application Build Mode
	applicationBuildModes       = clientEnumToStringArray(qovery.AllowedBuildModeEnumEnumValues)
	applicationBuildModeDefault = string(qovery.BUILDMODEENUM_DOCKER)

	// Application CPU
	applicationCPUMin     int64 = 10  // in MB
	applicationCPUDefault int64 = 500 // in MB

	// Application Memory
	applicationMemoryMin     int64 = 1   // in MB
	applicationMemoryDefault int64 = 512 // in MB

	// Application Min Running Instances
	applicationMinRunningInstancesMin     int64 = 0
	applicationMinRunningInstancesDefault int64 = 1

	// Application Max Running Instances
	applicationMaxRunningInstancesMin     int64 = -1
	applicationMaxRunningInstancesDefault int64 = 1

	// Application Auto Preview
	applicationAutoPreviewDefault = false

	// Application Storage
	applicationStorageSizeMin int64 = 1 // in GB

	// Application Git Repository
	applicationGitRepositoryRootPathDefault = "/"
)

type applicationResource struct {
	client                  *client.Client
	advancedSettingsService *advanced_settings.ServiceAdvancedSettingsService
}

func newApplicationResource() resource.Resource {
	return &applicationResource{}
}

func (r applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *applicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = provider.client
	r.advancedSettingsService = provider.advancedSettingsService
}

func (r applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "application")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "application")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "application")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "application")
	secrets := variableListDescriptions("secrets", "application")
	secretAliases := variableListDescriptions("secret_aliases", "application")
	secretOverrides := variableListDescriptions("secret_overrides", "application")
	ports := portDescriptions("application")
	storages := storageDescriptions("application")
	customDomains := customDomainDescriptions("application")
	restrictions := deploymentRestrictionDescriptions("application")

	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a Qovery application: a service Qovery builds from a git repository, with a Dockerfile or Buildpacks, and deploys to its environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("application"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription + recreatesOnChange("application"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("application"),
				Required:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(iconURIDescription("application"), applicationIconURIDefault),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(applicationIconURIDefault),
			},
			"git_repository": schema.SingleNestedAttribute{
				MarkdownDescription: "Git repository the application is built from.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: gitRepositoryURLDescription,
						Required:            true,
					},
					"branch": schema.StringAttribute{
						MarkdownDescription: "Branch to build." + gitBranchRemovalNote,
						Optional:            true,
						Computed:            true,
						// Documented exception to the config-is-source-of-truth rule: the default
						// branch depends on the repository, so there is no static Default, and
						// planning unknown whenever it is omitted would give a permanent diff.
						PlanModifiers: []planmodifier.String{
							UseStateUnlessRepositoryChanges(),
						},
					},
					"root_path": schema.StringAttribute{
						MarkdownDescription: descriptions.NewStringDefaultDescription(gitRepositoryRootPathDescription, applicationGitRepositoryRootPathDefault),
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(applicationGitRepositoryRootPathDefault),
					},
					"git_token_id": schema.StringAttribute{
						MarkdownDescription: gitRepositoryTokenIDDescription,
						Optional:            true,
						Computed:            false,
					},
				},
			},
			"build_mode": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(
					"How Qovery builds the application: `DOCKER` builds the Dockerfile at `dockerfile_path`, `BUILDPACKS` detects the language with Cloud Native Buildpacks.",
					applicationBuildModeDefault,
				),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(applicationBuildModeDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(applicationBuildModes),
				},
			},
			"dockerfile_path": schema.StringAttribute{
				MarkdownDescription: dockerfilePathDescription + " Required when `build_mode` is `DOCKER`.",
				Optional:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(cpuDescription("application"), applicationCPUMin, &applicationCPUDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(applicationCPUDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationCPUMin},
				},
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(memoryDescription("application"), applicationMemoryMin, &applicationMemoryDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(applicationMemoryDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMemoryMin},
				},
			},
			"ephemeral_storage": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(ephemeralStorageDescription("application"), serviceEphemeralStorageDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(serviceEphemeralStorageDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: 0},
				},
			},
			"min_running_instances": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(minRunningInstancesDescription("application"), applicationMinRunningInstancesDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(applicationMinRunningInstancesDefault),
				Validators: []validator.Int64{
					validators.MinRunningInstancesAutoscalingValidator{AutoscalingAttributePath: "autoscaling"},
				},
			},
			"max_running_instances": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(maxRunningInstancesDescription("application"), applicationMaxRunningInstancesDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(applicationMaxRunningInstancesDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMaxRunningInstancesMin},
				},
			},
			"autoscaling":    autoscalingResourceSchema(),
			"build_settings": buildSettingsResourceSchemaAttributes(),
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(autoPreviewDescription("application"), applicationAutoPreviewDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(applicationAutoPreviewDefault),
			},
			"entrypoint": schema.StringAttribute{
				MarkdownDescription: entrypointDescription,
				Optional:            true,
			},
			"arguments": schema.ListAttribute{
				MarkdownDescription: argumentsDescription + omittedSetsNoneNote,
				Optional:            true,
				ElementType:         types.StringType,
			},
			"storage": schema.SetNestedAttribute{
				MarkdownDescription: storages.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: storages.ID,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: descriptions.NewStringEnumDescription(storages.Type, clientEnumToStringArray(storage.AllowedTypeValues), nil),
							Required:            true,
							Validators: []validator.String{
								validators.NewStringEnumValidator(clientEnumToStringArray(storage.AllowedTypeValues)),
							},
						},
						"size": schema.Int64Attribute{
							MarkdownDescription: descriptions.NewInt64MinDescription(storages.Size, applicationStorageSizeMin, nil),
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinValidator{Min: applicationStorageSizeMin},
							},
						},
						"mount_point": schema.StringAttribute{
							MarkdownDescription: storages.MountPoint,
							Required:            true,
						},
					},
				},
			},
			"ports": schema.ListNestedAttribute{
				MarkdownDescription: ports.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						validators.PortExternalPortValidator{},
					},
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: ports.ID,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
								UseUnknownForNullString(),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: ports.Name + portNameDefaultNote,
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								PortNameDefault(),
							},
						},
						"internal_port": schema.Int64Attribute{
							MarkdownDescription: descriptions.NewInt64MinMaxDescription(ports.InternalPort, port.MinPort, port.MaxPort, nil),
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"external_port": schema.Int64Attribute{
							MarkdownDescription: descriptions.NewInt64MinMaxDescription(ports.ExternalPort, port.MinPort, port.MaxPort, nil),
							Optional:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"publicly_accessible": schema.BoolAttribute{
							MarkdownDescription: ports.PubliclyAccessible,
							Required:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: descriptions.NewStringEnumDescription(ports.Protocol, clientEnumToStringArray(port.AllowedProtocolValues), new(port.DefaultProtocol.String())),
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(port.DefaultProtocol.String()),
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
			"environment_variable_files": environmentVariableFilesSchemaAttribute("application"),
			"secret_files":               secretFilesSchemaAttribute("application"),
			"external_secrets":           externalSecretsSchemaAttribute("application"),
			"external_secret_files":      externalSecretFilesSchemaAttribute("application"),
			"healthchecks":               healthchecksSchemaAttributes(true),
			"custom_domains": schema.SetNestedAttribute{
				MarkdownDescription: customDomains.List,
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					CustomDomainsBoolDefaults("generate_certificate", "use_cdn"),
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
						// generate_certificate and use_cdn default to false through the
						// CustomDomainsBoolDefaults plan modifier on custom_domains: a Default
						// nested in a set breaks the matching of planned and applied elements.
						"generate_certificate": schema.BoolAttribute{
							MarkdownDescription: descriptions.NewBoolDefaultDescription(customDomains.GenerateCertificate, false),
							Optional:            true,
							Computed:            true,
						},
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
				MarkdownDescription: externalHostDescription("application"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					UseStateUnlessPortsChange(),
				},
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("application"),
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
				MarkdownDescription: advancedSettingsJSONDescription("Applications/operation/getDefaultApplicationAdvancedSettings"),
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: the QOV-2028
				// contract described in advancedSettingsJSONDescription.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(applicationAutoDeployDescription, serviceAutoDeployDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(serviceAutoDeployDefault),
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
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("annotations", "the application's pods"),
				Optional:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("labels", "the application's pods"),
				Optional:            true,
				ElementType:         types.StringType,
			},
			"docker_target_build_stage": schema.StringAttribute{
				MarkdownDescription: dockerTargetBuildStageDescription,
				Optional:            true,
			},
		},
	}
}

// Create qovery application resource
func (r applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Application
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new application
	request, err := plan.toCreateApplicationRequest()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	application, apiErr := r.client.CreateApplication(ctx, ToString(plan.EnvironmentId), request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		if application == nil {
			return
		}
		// The application was created in Qovery and only a follow-up call failed. Save it so
		// Terraform tracks it and taints it for replacement, instead of leaving an orphan
		// that makes every later apply fail with "an application named X already exists".
		resp.Diagnostics.Append(resp.State.Set(ctx, convertResponseToApplication(ctx, plan, application))...)
		return
	}

	// Initialize state values
	state := convertResponseToApplication(ctx, plan, application)
	tflog.Trace(ctx, "created application", map[string]any{"application_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery application resource
func (r applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Application
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hack to know if this method is triggered through an import
	// EnvironmentID is always present except when importing the resource
	isTriggeredFromImport := false
	if state.EnvironmentId.IsNull() {
		isTriggeredFromImport = true
	}

	// Get application from the API
	application, apiErr := r.client.GetApplication(ctx, state.Id.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if handleReadNotFound(ctx, resp, apiErr) {
		return
	}

	// Refresh state values
	state = convertResponseToApplication(ctx, state, application)
	tflog.Trace(ctx, "read application", map[string]any{"application_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update qovery application resource
func (r applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Application
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update application in the backend
	request, err := plan.toUpdateApplicationRequest(state)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	application, apiErr := r.client.UpdateApplication(ctx, state.Id.ValueString(), request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Update state values
	state = convertResponseToApplication(ctx, plan, application)
	tflog.Trace(ctx, "updated application", map[string]any{"application_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery application resource
func (r applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Application
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete application
	apiErr := r.client.DeleteApplication(ctx, state.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted application", map[string]any{"application_id": state.Id.ValueString()})

	// Remove application from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery application resource using its id
func (r applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// UpgradeState migrates application states written by 0.x (schema version 0).
func (r applicationResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// Version 0 has the same attribute types as the current schema.
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	priorSchema := schemaResp.Schema

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &priorSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var app Application
				resp.Diagnostics.Append(req.State.Get(ctx, &app)...)
				if resp.Diagnostics.HasError() {
					return
				}
				app.Arguments = upgradeArgumentsFrom0x(app.Arguments)
				resp.Diagnostics.Append(resp.State.Set(ctx, app)...)
			},
		},
	}
}

// ModifyPlan enforces KEDA autoscaling constraints at plan time so the backend
// never rejects them mid-apply (which would leave the service partially mutated),
// warns about advanced_settings_json keys that are unknown for this service type,
// and lets build_settings own the build.* advanced settings when it is set.
func (r applicationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	validateAutoscalingPlan(ctx, req.Plan, req.State, &resp.Diagnostics)
	warnUnknownAdvancedSettings(ctx, r.advancedSettingsService, domain.APPLICATION, req.Config, &resp.Diagnostics)
	modifyBuildSettingsPlan(ctx, req, resp)
}
