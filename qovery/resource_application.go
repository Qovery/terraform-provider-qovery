package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/AlekSi/pointer"
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
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &applicationResource{}
var _ resource.ResourceWithImportState = applicationResource{}

var (

	// Application Build Mode
	applicationBuildModes       = clientEnumToStringArray(qovery.AllowedBuildModeEnumEnumValues)
	applicationBuildModeDefault = string(qovery.BUILDMODEENUM_BUILDPACKS)

	// Application BuildPack
	applicationBuildPackLanguages = clientEnumToStringArray(qovery.AllowedBuildPackLanguageEnumEnumValues)

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
	applicationStorageTypes         = clientEnumToStringArray(qovery.AllowedStorageTypeEnumEnumValues)
	applicationStorageSizeMin int64 = 1 // in GB

	// Application Port
	applicationPortMin                       int64 = 1
	applicationPortMax                       int64 = 65535
	applicationPortProtocols                       = clientEnumToStringArray(qovery.AllowedPortProtocolEnumEnumValues)
	applicationPortProtocolDefault                 = string(qovery.PORTPROTOCOLENUM_HTTP)
	applicationPortPubliclyAccessibleDefault       = false

	// Application Git Repository
	applicationGitRepositoryRootPathDefault = "/"
	applicationGitRepositoryBranchDefault   = "main or master (depending on repository)"
)

type applicationResource struct {
	client *client.Client
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
}

func (r applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery application resource. This can be used to create and manage Qovery applications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the application.",
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
				Description: "Name of the application.",
				Required:    true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the application.",
				Optional:    true,
				Computed:    true,
			},
			"git_repository": schema.SingleNestedAttribute{
				Description: "Git repository of the application.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL of the git repository.",
						Required:    true,
					},
					"branch": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Branch of the git repository.",
							applicationGitRepositoryBranchDefault,
						),
						Optional: true,
						Computed: true,
					},
					"root_path": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Root path of the application.",
							applicationGitRepositoryRootPathDefault,
						),
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(applicationGitRepositoryRootPathDefault),
					},
					"git_token_id": schema.StringAttribute{
						Description: "The git token ID to be used",
						Optional:    true,
						Computed:    false,
					},
				},
			},
			"build_mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Build Mode of the application.",
					applicationBuildModes,
					&applicationBuildModeDefault,
				),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(applicationBuildModeDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(applicationBuildModes),
				},
			},
			"dockerfile_path": schema.StringAttribute{
				Description: "Dockerfile Path of the application.\n\t- Required if: `build_mode=\"DOCKER\"`.",
				Optional:    true,
			},
			"buildpack_language": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Buildpack Language framework.\n\t- Required if: `build_mode=\"BUILDPACKS\"`.",
					applicationBuildPackLanguages,
					nil,
				),
				Optional: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(applicationBuildPackLanguages),
				},
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the application in millicores (m) [1000m = 1 CPU].",
					applicationCPUMin,
					&applicationCPUDefault,
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationCPUDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationCPUMin},
				},
			},
			"memory": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"RAM of the application in MB [1024MB = 1GB].",
					applicationMemoryMin,
					&applicationMemoryDefault,
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationMemoryDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMemoryMin},
				},
			},
			"min_running_instances": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of instances running for the application.",
					applicationMinRunningInstancesMin,
					&applicationMinRunningInstancesDefault,
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationMinRunningInstancesDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMinRunningInstancesMin},
				},
			},
			"max_running_instances": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of instances running for the application.",
					applicationMaxRunningInstancesMin,
					&applicationMaxRunningInstancesDefault,
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationMaxRunningInstancesDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMaxRunningInstancesMin},
				},
			},
			"auto_preview": schema.BoolAttribute{
				Description: descriptions.NewBoolDefaultDescription(
					"Specify if the environment preview option is activated or not for this application.",
					applicationAutoPreviewDefault,
				),
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(applicationAutoPreviewDefault),
			},
			"entrypoint": schema.StringAttribute{
				Description: "Entrypoint of the application.",
				Optional:    true,
			},
			"arguments": schema.ListAttribute{
				Description: "List of arguments of this application.",
				Optional:    true,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				//Default:     listdefault.StaticValue(ListNull(types.StringType)),
			},
			"storage": schema.SetNestedAttribute{
				Description: "List of storages linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the storage.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Type of the storage for the application.",
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
								"Size of the storage for the application in GB [1024MB = 1GB].",
								applicationStorageSizeMin,
								nil,
							),
							Required: true,
							Validators: []validator.Int64{
								validators.Int64MinValidator{Min: applicationStorageSizeMin},
							},
						},
						"mount_point": schema.StringAttribute{
							Description: "Mount point of the storage for the application.",
							Required:    true,
						},
					},
				},
			},
			"ports": schema.ListNestedAttribute{
				Description: "List of ports linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the port.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Description: "Name of the port.",
							Optional:    true,
							Computed:    true,
						},
						"internal_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"Internal port of the application.",
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
								"External port of the application.\n\t- Required if: `ports.publicly_accessible=true`.",
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
							Description: "Specify if the port is exposed to the world or not for this application.",
							Required:    true,
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the application.",
								clientEnumToStringArray(port.AllowedProtocolValues),
								pointer.ToString(port.DefaultProtocol.String()),
							),
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(port.DefaultProtocol.String()),
						},
						"is_default": schema.BoolAttribute{
							Description: "If this port will be used for the root domain",
							Required:    true,
						},
					},
				},
			},
			"built_in_environment_variables": schema.SetNestedAttribute{
				Description: "List of built-in environment variables linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description: "List of environment variables linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the environment variable.",
							Optional:    true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description: "List of environment variable aliases linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the environment variable alias.",
							Optional:    true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description: "List of environment variable overrides linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the environment variable override.",
							Optional:    true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description: "List of secrets linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the secret.",
							Optional:    true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description: "List of secret aliases linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the secret alias.",
							Optional:    true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description: "List of secret overrides linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the secret override.",
							Optional:    true,
						},
					},
				},
			},
			"healthchecks": healthchecksSchemaAttributes(true),
			"custom_domains": schema.SetNestedAttribute{
				Description: "List of custom domains linked to this application.",
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
						"generate_certificate": schema.BoolAttribute{
							Description: "Qovery will generate and manage the certificate for this domain.",
							Optional:    true,
						},
						"use_cdn": schema.BoolAttribute{
							Description: "Indicates if the custom domain is behind a CDN (i.e Cloudflare).\n" +
								"This will condition the way we are checking CNAME before & during a deployment:\n" +
								" * If `true` then we only check the domain points to an IP\n" +
								" * If `false` then we check that the domain resolves to the correct service Load Balancer",
							Optional: true,
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
				Description: "The application external FQDN host [NOTE: only if your application is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The application internal host.",
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
				Description: " Specify if the application will be automatically updated after receiving a new image tag.",
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
			"annotations_group_ids": schema.SetAttribute{
				Description: "List of annotations group ids",
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description: "List of labels group ids",
				Optional:    true,
				ElementType: types.StringType,
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
		return
	}

	// Initialize state values
	state := convertResponseToApplication(ctx, plan, application)
	tflog.Trace(ctx, "created application", map[string]interface{}{"application_id": state.Id.ValueString()})

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

	// Get application from the API
	application, apiErr := r.client.GetApplication(ctx, state.Id.ValueString(), state.AdvancedSettingsJson.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state = convertResponseToApplication(ctx, state, application)
	tflog.Trace(ctx, "read application", map[string]interface{}{"application_id": state.Id.ValueString()})

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
	tflog.Trace(ctx, "updated application", map[string]interface{}{"application_id": state.Id.ValueString()})

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

	tflog.Trace(ctx, "deleted application", map[string]interface{}{"application_id": state.Id.ValueString()})

	// Remove application from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery application resource using its id
func (r applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
