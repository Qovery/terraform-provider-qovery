package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
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

func (r containerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery container resource. This can be used to create and manage Qovery container registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the container.",
				Computed:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Required:    true,
			},
			"registry_id": schema.StringAttribute{
				Description: "Id of the registry.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the container.",
				Required:    true,
			},
			"image_name": schema.StringAttribute{
				Description: "Name of the container image.",
				Required:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag of the container image.",
				Required:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the container in millicores (m) [1000m = 1 CPU].",
					container.MinCPU,
					pointer.ToInt64(container.DefaultCPU),
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(container.DefaultCPU),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinCPU},
				},
			},
			"memory": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"RAM of the container in MB [1024MB = 1GB].",
					container.MinMemory,
					pointer.ToInt64(container.DefaultMemory),
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(container.DefaultMemory),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinMemory},
				},
			},
			"min_running_instances": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of instances running for the container.",
					container.MinMinRunningInstances,
					pointer.ToInt64(container.DefaultMinRunningInstances),
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(container.MinMinRunningInstances),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinMinRunningInstances},
				},
			},
			"max_running_instances": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of instances running for the container.",
					container.MinMaxRunningInstances,
					pointer.ToInt64(container.DefaultMaxRunningInstances),
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(container.DefaultMaxRunningInstances),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinMaxRunningInstances},
				},
			},
			"auto_preview": schema.BoolAttribute{
				Description: "Specify if the environment preview option is activated or not for this container.",
				Optional:    true,
				Computed:    true,
			},
			"entrypoint": schema.StringAttribute{
				Description: "Entrypoint of the container.",
				Optional:    true,
				Computed:    true,
			},
			"storage": schema.SetNestedAttribute{
				Description: "List of storages linked to this container.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the storage.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Type of the storage for the container.",
								clientEnumToStringArray(storage.AllowedTypeValues),
								nil,
							),
							Required: true,
							Validators: []validator.String{
								validators.NewStringEnumValidator(clientEnumToStringArray(storage.AllowedTypeValues)),
							},
						},
						"size": schema.Int64Attribute{
							Description: descriptions.NewInt64MinDescription(
								"Size of the storage for the container in GB [1024MB = 1GB].",
								container.MinStorageSize,
								nil,
							),
							Required: true,
							Validators: []validator.Int64{
								validators.Int64MinValidator{Min: applicationStorageSizeMin},
							},
						},
						"mount_point": schema.StringAttribute{
							Description: "Mount point of the storage for the container.",
							Required:    true,
						},
					},
				},
			},
			"ports": schema.SetNestedAttribute{
				Description: "List of ports linked to this container.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the port.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the port.",
							Optional:    true,
							Computed:    true,
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
							Optional: true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"publicly_accessible": schema.BoolAttribute{
							Description: "Specify if the port is exposed to the world or not for this container.",
							Required:    true,
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the container.",
								clientEnumToStringArray(port.AllowedProtocolValues),
								pointer.ToString(port.DefaultProtocol.String()),
							),
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
				Description: "List of built-in environment variables linked to this container.",
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
				Description: "List of environment variables linked to this container.",
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
				Description: "List of environment variable aliases linked to this container.",
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
				Description: "List of environment variable overrides linked to this container.",
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
				Description: "List of secrets linked to this container.",
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
				Description: "List of secret aliases linked to this container.",
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
				Description: "List of secret overrides linked to this container.",
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
			"healthchecks": healthchecksSchemaAttributes(false),
			"arguments": schema.ListAttribute{
				Description: "List of arguments of this container.",
				Optional:    true,
				ElementType: types.StringType,
				Computed:    true,
				//Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"custom_domains": schema.SetNestedAttribute{
				Description: "List of custom domains linked to this container.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the custom domain.",
							Computed:    true,
						},
						"domain": schema.StringAttribute{
							Description: "Your custom domain.",
							Required:    true,
						},
						"validation_domain": schema.StringAttribute{
							Description: "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the custom domain.",
							Computed:    true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description: "The container external FQDN host [NOTE: only if your container is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The container internal host.",
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
			"auto_deploy": schema.BoolAttribute{
				Description: " Specify if the container will be automatically updated after receiving a new image tag.",
				Optional:    true,
				Computed:    true,
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
	request, err := plan.toUpsertServiceRequest(nil)
	if err != nil {
		resp.Diagnostics.AddError("Error on container create", err.Error())
		return
	}
	cont, err := r.containerService.Create(ctx, plan.EnvironmentID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on container create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainContainerToContainer(ctx, plan, cont)
	tflog.Trace(ctx, "created container", map[string]interface{}{"container_id": state.ID.ValueString()})

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
	cont, err := r.containerService.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on container read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainContainerToContainer(ctx, state, cont)
	tflog.Trace(ctx, "read container", map[string]interface{}{"container_id": state.ID.ValueString()})

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
	cont, err := r.containerService.Update(ctx, state.ID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on container update", err.Error())
		return
	}

	// Update state values
	state = convertDomainContainerToContainer(ctx, plan, cont)
	tflog.Trace(ctx, "updated container", map[string]interface{}{"container_id": state.ID.ValueString()})

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

	tflog.Trace(ctx, "deleted container", map[string]interface{}{"container_id": state.ID.ValueString()})

	// Remove container from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery container resource using its id
func (r containerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
