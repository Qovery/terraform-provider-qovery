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

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &containerResource{}
var _ resource.ResourceWithImportState = containerResource{}

type containerResource struct {
	containerService container.Service
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
}

func (r containerResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	advSettings := map[string]tfsdk.Attribute{}
	for k, v := range GetContainerSettingsDefault() {
		advSettings[k] = tfsdk.Attribute{
			Description: v.Description,
			Required:    true,
			Type:        v.Type,
		}
	}

	return tfsdk.Schema{
		Description: "Provides a Qovery container resource. This can be used to create and manage Qovery container registry.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the container.",
				Type:        types.StringType,
				Computed:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"registry_id": {
				Description: "Id of the registry.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the container.",
				Type:        types.StringType,
				Required:    true,
			},
			"image_name": {
				Description: "Name of the container image.",
				Type:        types.StringType,
				Required:    true,
			},
			"tag": {
				Description: "Tag of the container image.",
				Type:        types.StringType,
				Required:    true,
			},
			"cpu": {
				Description: descriptions.NewInt64MinDescription(
					"CPU of the container in millicores (m) [1000m = 1 CPU].",
					container.MinCPU,
					pointer.ToInt64(container.DefaultCPU),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(container.DefaultCPU),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: container.MinCPU},
				},
			},
			"memory": {
				Description: descriptions.NewInt64MinDescription(
					"RAM of the container in MB [1024MB = 1GB].",
					container.MinMemory,
					pointer.ToInt64(container.DefaultMemory),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(container.DefaultMemory),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: container.MinMemory},
				},
			},
			"min_running_instances": {
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of instances running for the container.",
					container.MinMinRunningInstances,
					pointer.ToInt64(container.DefaultMinRunningInstances),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(container.DefaultMinRunningInstances),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: container.MinMinRunningInstances},
				},
			},
			"max_running_instances": {
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of instances running for the container.",
					container.MinMaxRunningInstances,
					pointer.ToInt64(container.DefaultMaxRunningInstances),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(container.DefaultMaxRunningInstances),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: container.MinMaxRunningInstances},
				},
			},
			"auto_preview": {
				Description: "Specify if the environment preview option is activated or not for this container.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
			},
			"entrypoint": {
				Description: "Entrypoint of the container.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"storage": {
				Description: "List of storages linked to this container.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the storage.",
						Type:        types.StringType,
						Computed:    true,
					},
					"type": {
						Description: descriptions.NewStringEnumDescription(
							"Type of the storage for the container.",
							clientEnumToStringArray(storage.AllowedTypeValues),
							nil,
						),
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							validators.NewStringEnumValidator(clientEnumToStringArray(storage.AllowedTypeValues)),
						},
					},
					"size": {
						Description: descriptions.NewInt64MinDescription(
							"Size of the storage for the container in GB [1024MB = 1GB].",
							container.MinStorageSize,
							nil,
						),
						Type:     types.Int64Type,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							validators.Int64MinValidator{Min: applicationStorageSizeMin},
						},
					},
					"mount_point": {
						Description: "Mount point of the storage for the container.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"ports": {
				Description: "List of storages linked to this container.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the port.",
						Type:        types.StringType,
						Computed:    true,
					},
					"name": {
						Description: "Name of the port.",
						Type:        types.StringType,
						Optional:    true,
					},
					"internal_port": {
						Description: descriptions.NewInt64MinMaxDescription(
							"Internal port of the container.",
							port.MinPort,
							port.MaxPort,
							nil,
						),
						Type:     types.Int64Type,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
						},
					},
					"external_port": {
						Description: descriptions.NewInt64MinMaxDescription(
							"External port of the container.\n\t- Required if: `ports.publicly_accessible=true`.",
							port.MinPort,
							port.MaxPort,
							nil,
						),
						Type:     types.Int64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
						},
					},
					"publicly_accessible": {
						Description: "Specify if the port is exposed to the world or not for this container.",
						Type:        types.BoolType,
						Required:    true,
					},
					"protocol": {
						Description: descriptions.NewStringEnumDescription(
							"Protocol used for the port of the container.",
							clientEnumToStringArray(port.AllowedProtocolValues),
							pointer.ToString(port.DefaultProtocol.String()),
						),
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							modifiers.NewStringDefaultModifier(port.DefaultProtocol.String()),
						},
					},
				}),
			},
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this container.",
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
				Description: "List of environment variables linked to this container.",
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
			"secrets": {
				Description: "List of secrets linked to this container.",
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
						Description: "Value of the secret [NOTE: will always be empty].",
						Type:        types.StringType,
						Required:    true,
						Sensitive:   true,
					},
				}),
			},
			"arguments": {
				Description: "List of arguments of this container.",
				Optional:    true,
				Computed:    true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringSliceDefaultModifier([]string{}),
				},
			},
			//"custom_domains": {
			//	Description: "List of custom domains linked to this container.",
			//	Computed:    true,
			//	Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//		"id": {
			//			Description: "Id of the custom domain.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//		"domain": {
			//			Description: "Your custom domain.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//		"validation_domain": {
			//			Description: "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//		"status": {
			//			Description: "Status of the custom domain.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//	}),
			//},
			"external_host": {
				Description: "The container external FQDN host [NOTE: only if your container is using a publicly accessible port].",
				Type:        types.StringType,
				Computed:    true,
			},
			"internal_host": {
				Description: "The container internal host.",
				Type:        types.StringType,
				Computed:    true,
			},
			"deployment_stage_id": {
				Description: "Id of the deployment stage.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"advanced_settings": {
				Description: "Advanced settings of the container.",
				Optional:    true,
				Computed:    true,
				Attributes:  tfsdk.SingleNestedAttributes(advSettings),
			},
		},
	}, nil
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
	request, err := plan.toUpsertServiceRequest(nil)
	if err != nil {
		resp.Diagnostics.AddError("Error on container create", err.Error())
		return
	}
	cont, err := r.containerService.Create(ctx, plan.EnvironmentID.Value, *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on container create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainContainerToContainer(plan, cont)
	tflog.Trace(ctx, "created container", map[string]interface{}{"container_id": state.ID.Value})

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

	// Get container from the API
	cont, err := r.containerService.Get(ctx, state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on container read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainContainerToContainer(state, cont)
	tflog.Trace(ctx, "read container", map[string]interface{}{"container_id": state.ID.Value})

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
	request, err := plan.toUpsertServiceRequest(&state)
	if err != nil {
		resp.Diagnostics.AddError("Error on container create", err.Error())
		return
	}
	cont, err := r.containerService.Update(ctx, state.ID.Value, *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on container update", err.Error())
		return
	}

	// Update state values
	state = convertDomainContainerToContainer(plan, cont)
	tflog.Trace(ctx, "updated container", map[string]interface{}{"container_id": state.ID.Value})

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
	err := r.containerService.Delete(ctx, state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on container delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted container", map[string]interface{}{"container_id": state.ID.Value})

	// Remove container from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery container resource using its id
func (r containerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
