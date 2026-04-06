package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
		Description: "Provides a Qovery container resource. This can be used to create and manage Qovery containers.",
		MarkdownDescription: "Provides a Qovery container resource. This can be used to create and manage Qovery containers.\n\n" +
			"A container is a service that runs a Docker image from a container registry within a Qovery environment. " +
			"Unlike applications (which are built from source code), containers use pre-built images.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the container.",
				MarkdownDescription: "Id of the container.",
				Computed:             true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				MarkdownDescription: "Id of the environment. Changing this forces the container to be re-created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registry_id": schema.StringAttribute{
				Description: "Id of the registry.",
				MarkdownDescription: "Id of the container registry (from `qovery_container_registry`) that stores the Docker image for this container.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the container.",
				MarkdownDescription: "Name of the container.",
				Required:            true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the container.",
				MarkdownDescription: "Icon URI representing the container. Used in the Qovery console UI.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_name": schema.StringAttribute{
				Description: "Name of the container image.",
				MarkdownDescription: "Name of the container image (e.g. `nginx`, `my-org/my-app`). Do not include the tag.",
				Required:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag of the container image.",
				MarkdownDescription: "Tag of the container image (e.g. `latest`, `1.0.0`, `sha-abc123`).",
				Required:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the container in millicores (m) [1000m = 1 CPU].",
					container.MinCPU,
					pointer.ToInt64(container.DefaultCPU),
				),
				MarkdownDescription: "CPU of the container in millicores (m) [1000m = 1 CPU].",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "RAM of the container in MB [1024MB = 1GB].",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "Minimum number of instances running for the container.",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "Maximum number of instances running for the container.",
				Optional:            true,
				Computed:            true,
				Default:  int64default.StaticInt64(container.DefaultMaxRunningInstances),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: container.MinMaxRunningInstances},
				},
			},
			"auto_preview": schema.BoolAttribute{
				Description: "Specify if the environment preview option is activated or not for this container.",
				MarkdownDescription: "Specify if the environment preview option is activated or not for this container. " +
					"When enabled, Qovery creates a preview environment for each pull request.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"entrypoint": schema.StringAttribute{
				Description: "Entrypoint of the container.",
				MarkdownDescription: "Entrypoint of the container. Overrides the Docker image's default `ENTRYPOINT`.",
				Optional:    true,
			},
			"storage": schema.SetNestedAttribute{
				Description: "List of storages linked to this container.",
				MarkdownDescription: "List of persistent storage volumes linked to this container. " +
					"Data stored in these volumes persists across container restarts.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the storage.",
							MarkdownDescription: "Id of the storage.",
							Computed:             true,
						},
						"type": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Type of the storage for the container.",
								clientEnumToStringArray(storage.AllowedTypeValues),
								nil,
							),
							MarkdownDescription: "Type of the storage for the container.",
							Required:            true,
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
							MarkdownDescription: "Size of the storage for the container in GB [1024MB = 1GB].",
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinValidator{Min: applicationStorageSizeMin},
							},
						},
						"mount_point": schema.StringAttribute{
							Description:         "Mount point of the storage for the container.",
							MarkdownDescription: "Mount point of the storage for the container.",
							Required:            true,
						},
					},
				},
			},
			"ports": schema.ListNestedAttribute{
				Description: "List of ports linked to this container.",
				MarkdownDescription: "List of ports linked to this container. " +
					"At least one port must be set as `publicly_accessible = true` with an `external_port` for the container to be reachable from the internet.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						validators.PortExternalPortValidator{},
					},
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the port.",
							MarkdownDescription: "Id of the port.",
							Computed:             true,
						},
						"name": schema.StringAttribute{
							Description:         "Name of the port.",
							MarkdownDescription: "Name of the port.",
							Optional:            true,
							Computed:            true,
						},
						"internal_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"Internal port of the container.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							MarkdownDescription: "Internal port of the container. Must be between 1 and 65535.",
							Required:            true,
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
							MarkdownDescription: "External port of the container. Required if `ports.publicly_accessible = true`. Must be between 1 and 65535.",
							Optional:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"publicly_accessible": schema.BoolAttribute{
							Description:         "Specify if the port is exposed to the world or not for this container.",
							MarkdownDescription: "Specify if the port is exposed to the world or not for this container.",
							Required:            true,
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the container.",
								clientEnumToStringArray(port.AllowedProtocolValues),
								new(port.DefaultProtocol.String()),
							),
							MarkdownDescription: "Protocol used for the port of the container.",
							Optional:            true,
							Computed:            true,
						},
						"is_default": schema.BoolAttribute{
							Description:         "If this port will be used for the root domain. Note: the API may override this value based on port configuration (e.g., when only one publicly accessible port exists, it will be set as default).",
							MarkdownDescription: "If this port will be used for the root domain. The API may override this value based on port configuration (e.g., when only one publicly accessible port exists, it will be set as default).",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								SmartAllowApiOverride(),
							},
						},
					},
				},
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				Description: "List of built-in environment variables linked to this container.",
				MarkdownDescription: "List of built-in environment variables linked to this container. " +
					"Built-in variables are automatically generated by Qovery and include host information, port mappings, and other service metadata. " +
					"These are read-only and cannot be modified.",
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					UseStateUnlessNameChanges(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
							Computed:             true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable.",
							MarkdownDescription: "Key of the environment variable.",
							Computed:             true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable.",
							MarkdownDescription: "Value of the environment variable.",
							Computed:             true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable.",
							MarkdownDescription: "Description of the environment variable.",
							Computed:             true,
						},
					},
				},
			},
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description: "List of environment variables linked to this container.",
				MarkdownDescription: "List of environment variables linked to this container. " +
					"Environment variables at the container level have the highest precedence and override variables set at the project or environment level.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
							Computed:             true,
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
				Description: "List of environment variable aliases linked to this container.",
				MarkdownDescription: "List of environment variable aliases linked to this container. " +
					"An alias creates a new environment variable name that references the value of an existing variable. " +
					"The `key` is the alias name and `value` is the name of the variable being aliased.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable alias.",
							MarkdownDescription: "Id of the environment variable alias.",
							Computed:             true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable alias.",
							MarkdownDescription: "Name of the environment variable alias.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the variable to alias.",
							MarkdownDescription: "Name of the variable to alias.",
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
				Description: "List of environment variable overrides linked to this container.",
				MarkdownDescription: "List of environment variable overrides linked to this container. " +
					"An override replaces the value of an existing environment variable defined at a higher scope (project or environment). " +
					"The `key` must match the name of the variable to override.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable override.",
							MarkdownDescription: "Id of the environment variable override.",
							Computed:             true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable override.",
							MarkdownDescription: "Name of the environment variable override.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable override.",
							MarkdownDescription: "Value of the environment variable override.",
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
				Description: "List of secrets linked to this container.",
				MarkdownDescription: "List of secrets linked to this container. " +
					"Secrets behave like environment variables but their values are stored securely and not visible in plan outputs.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret.",
							MarkdownDescription: "Id of the secret.",
							Computed:             true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the secret.",
							MarkdownDescription: "Key of the secret.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret.",
							MarkdownDescription: "Value of the secret. The value is write-only and will not be displayed in plan outputs.",
							Required:            true,
							Sensitive:            true,
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
				Description: "List of secret aliases linked to this container.",
				MarkdownDescription: "List of secret aliases linked to this container. " +
					"An alias creates a new secret name that references the value of an existing secret. " +
					"The `key` is the alias name and `value` is the name of the secret being aliased.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret alias.",
							MarkdownDescription: "Id of the secret alias.",
							Computed:             true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret alias.",
							MarkdownDescription: "Name of the secret alias.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the secret to alias.",
							MarkdownDescription: "Name of the secret to alias.",
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
				Description: "List of secret overrides linked to this container.",
				MarkdownDescription: "List of secret overrides linked to this container. " +
					"An override replaces the value of an existing secret defined at a higher scope (project or environment). " +
					"The `key` must match the name of the secret to override.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret override.",
							MarkdownDescription: "Id of the secret override.",
							Computed:             true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret override.",
							MarkdownDescription: "Name of the secret override.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret override.",
							MarkdownDescription: "Value of the secret override. The value is write-only and will not be displayed in plan outputs.",
							Required:            true,
							Sensitive:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret override.",
							MarkdownDescription: "Description of the secret override.",
							Optional:            true,
						},
					},
				},
			},
			"environment_variable_files": schema.SetNestedAttribute{
				Description: "List of environment variable files linked to this container.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable file.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the environment variable file.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable file.",
							Required:    true,
						},
						"mount_path": schema.StringAttribute{
							Description: "Mount path of the environment variable file.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the environment variable file.",
							Optional:    true,
						},
					},
				},
			},
			"secret_files": schema.SetNestedAttribute{
				Description: "List of secret files linked to this container.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret file.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the secret file.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the secret file.",
							Required:    true,
							Sensitive:   true,
						},
						"mount_path": schema.StringAttribute{
							Description: "Mount path of the secret file.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the secret file.",
							Optional:    true,
						},
					},
				},
			},
			"healthchecks": healthchecksSchemaAttributes(true),
			"arguments": schema.ListAttribute{
				Description: "List of arguments of this container.",
				MarkdownDescription: "List of arguments of this container. Overrides the Docker image's default `CMD`.",
				Optional:    true,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				//Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"custom_domains": schema.SetNestedAttribute{
				Description: "List of custom domains linked to this container.",
				MarkdownDescription: "List of custom domains linked to this container. " +
					"You must configure a CNAME record on your DNS provider pointing to the `validation_domain` value.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the custom domain.",
							MarkdownDescription: "Id of the custom domain.",
							Computed:             true,
						},
						"domain": schema.StringAttribute{
							Description: "Your custom domain.",
							MarkdownDescription: "Your custom domain (e.g. `app.example.com`).",
							Required:    true,
						},
						"generate_certificate": schema.BoolAttribute{
							Description: "Qovery will generate and manage the certificate for this domain.",
							MarkdownDescription: "Qovery will generate and manage a TLS/SSL certificate for this domain using Let's Encrypt.",
							Optional:    true,
						},
						"use_cdn": schema.BoolAttribute{
							Description: "Indicates if the custom domain is behind a CDN (i.e Cloudflare).\n" +
								"This will condition the way we are checking CNAME before & during a deployment:\n" +
								" * If `true` then we only check the domain points to an IP\n" +
								" * If `false` then we check that the domain resolves to the correct service Load Balancer",
							MarkdownDescription: "Indicates if the custom domain is behind a CDN (e.g. Cloudflare). " +
								"This affects how Qovery validates the CNAME during deployment:\n" +
								"  - If `true`: Qovery only checks that the domain points to an IP.\n" +
								"  - If `false`: Qovery checks that the domain resolves to the correct service Load Balancer.",
							Optional: true,
						},
						"validation_domain": schema.StringAttribute{
							Description:         "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
							MarkdownDescription: "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
							Computed:             true,
						},
						"status": schema.StringAttribute{
							Description:         "Status of the custom domain.",
							MarkdownDescription: "Status of the custom domain.",
							Computed:             true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description: "The container external FQDN host [NOTE: only if your container is using a publicly accessible port].",
				MarkdownDescription: "The container external FQDN host. Only available if your container has at least one publicly accessible port.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal_host": schema.StringAttribute{
				Description: "The container internal host.",
				MarkdownDescription: "The container internal host. Use this to communicate between services within the same environment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_stage_id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				MarkdownDescription: "Id of the deployment stage. Deployment stages allow you to control the order in which services are deployed within an environment.",
				Optional:    true,
				Computed:    true,
			},
			"is_skipped": schema.BoolAttribute{
				Description:         "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				MarkdownDescription: "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings.",
				MarkdownDescription: "Advanced settings as JSON. " +
					"Use `jsonencode()` to set values. " +
					"Only include settings you want to override. " +
					"Full list available in [Qovery API documentation](https://api-doc.qovery.com/#tag/Containers/operation/getDefaultContainerAdvancedSettings).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Specify if the container will be automatically updated after receiving a new image tag.",
				MarkdownDescription: "Specify if the container will be automatically redeployed after receiving a new image tag from the container registry.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"annotations_group_ids": schema.SetAttribute{
				Description: "List of annotations group ids",
				MarkdownDescription: "List of annotations group ids. Annotations groups allow you to add Kubernetes annotations to the container's pods.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description: "List of labels group ids",
				MarkdownDescription: "List of labels group ids. Labels groups allow you to add Kubernetes labels to the container's pods.",
				Optional:    true,
				ElementType: types.StringType,
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
	var isTriggeredFromImport = false
	if state.EnvironmentID.IsNull() {
		isTriggeredFromImport = true
	}

	// Get container from the API
	cont, err := r.containerService.Get(ctx, state.ID.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if err != nil {
		resp.Diagnostics.AddError("Error on container read", err.Error())
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
