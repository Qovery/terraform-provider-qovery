package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure    = &containerResource{}
	_ resource.ResourceWithImportState  = containerResource{}
	_ resource.ResourceWithModifyPlan   = containerResource{}
	_ resource.ResourceWithUpgradeState = containerResource{}
)

type containerResource struct {
	containerService        container.Service
	advancedSettingsService *advanced_settings.ServiceAdvancedSettingsService
}

func newContainerResource() resource.Resource {
	return &containerResource{}
}

func (r containerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

func (r *containerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.containerService = provider.containerService
	r.advancedSettingsService = provider.advancedSettingsService
}

func (r containerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "container")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "container")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "container")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "container")
	secrets := variableListDescriptions("secrets", "container")
	secretAliases := variableListDescriptions("secret_aliases", "container")
	secretOverrides := variableListDescriptions("secret_overrides", "container")
	ports := portDescriptions("container")
	storages := storageDescriptions("container")
	customDomains := customDomainDescriptions("container")

	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: containerResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("container"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription + recreatesOnChange("container"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"registry_id": schema.StringAttribute{
				MarkdownDescription: containerImageRegistryIDDescription,
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("container"),
				Required:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(iconURIDescription("container"), containerIconURIDefault),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(containerIconURIDefault),
			},
			"image_name": schema.StringAttribute{
				MarkdownDescription: containerImageNameDescription,
				Required:            true,
			},
			"tag": schema.StringAttribute{
				MarkdownDescription: containerImageTagDescription,
				Required:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(cpuDescription("container"), container.MinCPU, new(int64(container.DefaultCPU))),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(container.DefaultCPU),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinCPU},
				},
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(memoryDescription("container"), container.MinMemory, new(int64(container.DefaultMemory))),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(container.DefaultMemory),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinMemory},
				},
			},
			"ephemeral_storage": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(ephemeralStorageDescription("container"), serviceEphemeralStorageDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(serviceEphemeralStorageDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: 0},
				},
			},
			"min_running_instances": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(minRunningInstancesDescription("container"), container.MinMinRunningInstances),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(container.MinMinRunningInstances),
				Validators: []validator.Int64{
					validators.MinRunningInstancesAutoscalingValidator{AutoscalingAttributePath: "autoscaling"},
				},
			},
			"max_running_instances": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(maxRunningInstancesDescription("container"), container.DefaultMaxRunningInstances),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(container.DefaultMaxRunningInstances),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinMaxRunningInstances},
				},
			},
			"autoscaling": autoscalingResourceSchema(),
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(autoPreviewDescription("container"), serviceAutoPreviewDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(serviceAutoPreviewDefault),
			},
			"entrypoint": schema.StringAttribute{
				MarkdownDescription: entrypointDescription,
				Optional:            true,
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
			"environment_variable_files": environmentVariableFilesSchemaAttribute("container"),
			"secret_files":               secretFilesSchemaAttribute("container"),
			"external_secrets":           externalSecretsSchemaAttribute("container"),
			"external_secret_files":      externalSecretFilesSchemaAttribute("container"),
			"healthchecks":               healthchecksSchemaAttributes(true),
			"arguments": schema.ListAttribute{
				MarkdownDescription: argumentsDescription + omittedSetsNoneNote,
				Optional:            true,
				ElementType:         types.StringType,
			},
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
				MarkdownDescription: externalHostDescription("container"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					UseStateUnlessPortsChange(),
				},
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("container"),
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
				MarkdownDescription: advancedSettingsJSONDescription("Containers/operation/getDefaultContainerAdvancedSettings"),
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: the QOV-2028
				// contract described in advancedSettingsJSONDescription.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(containerAutoDeployDescription, serviceAutoDeployDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(serviceAutoDeployDefault),
			},
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("annotations", "the container's pods"),
				Optional:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("labels", "the container's pods"),
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Create qovery container resource
func (r containerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Container
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new container
	request := plan.toUpsertServiceRequest(nil)
	cont, err := r.containerService.Create(ctx, plan.EnvironmentID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on container create", err.Error())
		if cont == nil {
			return
		}
		// The container was created in Qovery and only a follow-up call failed. Save it so
		// Terraform tracks it and taints it for replacement, instead of leaving an orphan
		// that makes every later apply fail with "a container named X already exists".
		resp.Diagnostics.Append(resp.State.Set(ctx, convertDomainContainerToContainer(ctx, plan, cont))...)
		return
	}

	// Initialize state values
	state := convertDomainContainerToContainer(ctx, plan, cont)
	tflog.Trace(ctx, "created container", map[string]any{"container_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery container resource
func (r containerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Container
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

	// Get container from the API
	cont, err := r.containerService.Get(ctx, state.ID.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if handleDomainReadNotFound(ctx, resp, err, "Error on container read") {
		return
	}

	// Refresh state values
	state = convertDomainContainerToContainer(ctx, state, cont)
	tflog.Trace(ctx, "read container", map[string]any{"container_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery container resource
func (r containerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Container
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update container in the backend
	request := plan.toUpsertServiceRequest(&state)
	cont, err := r.containerService.Update(ctx, state.ID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on container update", err.Error())
		return
	}

	// Update state values
	state = convertDomainContainerToContainer(ctx, plan, cont)
	tflog.Trace(ctx, "updated container", map[string]any{"container_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery container resource
func (r containerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Container
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete container
	err := r.containerService.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on container delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted container", map[string]any{"container_id": state.ID.ValueString()})

	// Remove container from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery container resource using its id
func (r containerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// UpgradeState migrates container states written by 0.x (schema version 0).
func (r containerResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// Version 0 has the same attribute types as the current schema.
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	priorSchema := schemaResp.Schema

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &priorSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var cont Container
				resp.Diagnostics.Append(req.State.Get(ctx, &cont)...)
				if resp.Diagnostics.HasError() {
					return
				}
				cont.Arguments = upgradeArgumentsFrom0x(cont.Arguments)
				resp.Diagnostics.Append(resp.State.Set(ctx, cont)...)
			},
		},
	}
}

// ModifyPlan enforces KEDA autoscaling constraints at plan time so the backend
// never rejects them mid-apply (which would leave the service partially mutated),
// and warns about advanced_settings_json keys that are unknown for this service type.
func (r containerResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	validateAutoscalingPlan(ctx, req.Plan, req.State, &resp.Diagnostics)
	warnUnknownAdvancedSettings(ctx, r.advancedSettingsService, domain.CONTAINER, req.Config, &resp.Diagnostics)
}
