package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
		MarkdownDescription: "Provides a Qovery application resource. This can be used to create and manage Qovery applications.\n\n" +
			"An application is a service built from source code in a git repository. " +
			"Qovery builds the application using either Docker (with a Dockerfile) or Buildpacks, then deploys it to your cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the application.",
				MarkdownDescription: "Id of the application.",
				Computed:             true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				MarkdownDescription: "Id of the environment. Changing this forces the application to be re-created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "Name of the application.",
				MarkdownDescription: "Name of the application.",
				Required:            true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the application.",
				MarkdownDescription: "Icon URI representing the application. Used in the Qovery console UI.",
				Optional:    true,
				Computed:    true,
			},
			"git_repository": schema.SingleNestedAttribute{
				Description: "Git repository of the application.",
				MarkdownDescription: "Git repository configuration for the application source code.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL of the git repository.",
						MarkdownDescription: "URL of the git repository (e.g. `https://github.com/my-org/my-app.git`).",
						Required:    true,
					},
					"branch": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Branch of the git repository.",
							applicationGitRepositoryBranchDefault,
						),
						MarkdownDescription: "Branch of the git repository to use for builds. " +
							"Defaults to `main` or `master` (depending on repository).",
						Optional: true,
						Computed: true,
					},
					"root_path": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Root path of the application.",
							applicationGitRepositoryRootPathDefault,
						),
						MarkdownDescription: "Root path of the application within the repository. " +
							"Useful for monorepos where the application code is in a subdirectory. Defaults to `/`.",
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(applicationGitRepositoryRootPathDefault),
					},
					"git_token_id": schema.StringAttribute{
						Description: "The git token ID to be used",
						MarkdownDescription: "The git token ID to be used for authenticating with the git provider. " +
							"Required for private repositories. Reference a `qovery_git_token` resource.",
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
				MarkdownDescription: "Build mode of the application.\n" +
					"  - `DOCKER`: Build using a Dockerfile in the repository. Requires `dockerfile_path` to be set.\n" +
					"  - `BUILDPACKS`: Build using Cloud Native Buildpacks (auto-detects language and framework).\n\n" +
					"Default: `DOCKER`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(applicationBuildModeDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(applicationBuildModes),
				},
			},
			"dockerfile_path": schema.StringAttribute{
				Description: "Dockerfile Path of the application.\n\t- Required if: `build_mode=\"DOCKER\"`.",
				MarkdownDescription: "Path to the Dockerfile relative to the `git_repository.root_path`. " +
					"Required when `build_mode = \"DOCKER\"`. Example: `Dockerfile` or `docker/Dockerfile.prod`.",
				Optional:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the application in millicores (m) [1000m = 1 CPU].",
					applicationCPUMin,
					&applicationCPUDefault,
				),
				MarkdownDescription: "CPU of the application in millicores (m) [1000m = 1 CPU].",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "RAM of the application in MB [1024MB = 1GB].",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "Minimum number of instances running for the application.",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "Maximum number of instances running for the application.",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "Specify if the environment preview option is activated or not for this application. " +
					"When enabled, Qovery creates a preview environment for each pull request. Default: `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(applicationAutoPreviewDefault),
			},
			"entrypoint": schema.StringAttribute{
				Description: "Entrypoint of the application.",
				MarkdownDescription: "Entrypoint of the application. Overrides the Docker image's default `ENTRYPOINT`.",
				Optional:    true,
			},
			"arguments": schema.ListAttribute{
				Description: "List of arguments of this application.",
				MarkdownDescription: "List of arguments of this application. Overrides the Docker image's default `CMD`.",
				Optional:    true,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				//Default:     listdefault.StaticValue(ListNull(types.StringType)),
			},
			"storage": schema.SetNestedAttribute{
				Description:         "List of storages linked to this application.",
				MarkdownDescription: "List of persistent storage volumes linked to this application. Data stored in these volumes persists across application restarts.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the storage.",
							MarkdownDescription: "Id of the storage.",
							Computed:             true,
						},
						"type": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Type of the storage for the application.",
								clientEnumToStringArray(storage.AllowedTypeValues),
								nil,
							),
							MarkdownDescription: "Type of the storage for the application.",
							Required:            true,
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
							MarkdownDescription: "Size of the storage for the application in GB [1024MB = 1GB].",
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinValidator{Min: applicationStorageSizeMin},
							},
						},
						"mount_point": schema.StringAttribute{
							Description:         "Mount point of the storage for the application.",
							MarkdownDescription: "Mount point of the storage for the application.",
							Required:            true,
						},
					},
				},
			},
			"ports": schema.ListNestedAttribute{
				Description: "List of ports linked to this application.",
				MarkdownDescription: "List of ports linked to this application. " +
					"At least one port must be set as `publicly_accessible = true` with an `external_port` for the application to be reachable from the internet.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						validators.PortExternalPortValidator{},
					},
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the port.",
							MarkdownDescription: "Id of the port.",
							Computed:             true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Description:         "Name of the port.",
							MarkdownDescription: "Name of the port.",
							Optional:            true,
							Computed:            true,
						},
						"internal_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"Internal port of the application.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							MarkdownDescription: "Internal port of the application. Must be between 1 and 65535.",
							Required:            true,
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
							MarkdownDescription: "External port of the application. Required if `ports.publicly_accessible = true`. Must be between 1 and 65535.",
							Optional:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"publicly_accessible": schema.BoolAttribute{
							Description:         "Specify if the port is exposed to the world or not for this application.",
							MarkdownDescription: "Specify if the port is exposed to the world or not for this application.",
							Required:            true,
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the application.",
								clientEnumToStringArray(port.AllowedProtocolValues),
								new(port.DefaultProtocol.String()),
							),
							MarkdownDescription: "Protocol used for the port of the application.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(port.DefaultProtocol.String()),
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
				Description: "List of built-in environment variables linked to this application.",
				MarkdownDescription: "List of built-in environment variables linked to this application. " +
					"Built-in variables are automatically generated by Qovery and include host information, port mappings, and other service metadata. " +
					"These are read-only and cannot be modified.",
				Computed:    true,
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
				Description:         "List of environment variables linked to this application.",
				MarkdownDescription: "List of environment variables linked to this application.",
				Optional:            true,
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
				Description:         "List of environment variable aliases linked to this application.",
				MarkdownDescription: "List of environment variable aliases linked to this application.",
				Optional:            true,
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
				Description:         "List of environment variable overrides linked to this application.",
				MarkdownDescription: "List of environment variable overrides linked to this application.",
				Optional:            true,
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
				Description:         "List of secrets linked to this application.",
				MarkdownDescription: "List of secrets linked to this application.",
				Optional:            true,
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
				Description:         "List of secret aliases linked to this application.",
				MarkdownDescription: "List of secret aliases linked to this application.",
				Optional:            true,
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
				Description:         "List of secret overrides linked to this application.",
				MarkdownDescription: "List of secret overrides linked to this application.",
				Optional:            true,
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
			"healthchecks": healthchecksSchemaAttributes(true),
			"custom_domains": schema.SetNestedAttribute{
				Description:         "List of custom domains linked to this application.",
				MarkdownDescription: "List of custom domains linked to this application. You must configure a CNAME record on your DNS provider pointing to the `validation_domain` value.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the custom domain.",
							MarkdownDescription: "Id of the custom domain.",
							Computed:             true,
						},
						"domain": schema.StringAttribute{
							Description:         "Your custom domain.",
							MarkdownDescription: "Your custom domain (e.g. `app.example.com`).",
							Required:            true,
						},
						"generate_certificate": schema.BoolAttribute{
							Description:         "Qovery will generate and manage the certificate for this domain.",
							MarkdownDescription: "Qovery will generate and manage a TLS/SSL certificate for this domain using Let's Encrypt.",
							Optional:            true,
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
				Description: "The application external FQDN host [NOTE: only if your application is using a publicly accessible port].",
				MarkdownDescription: "The application external FQDN host. Only available if your application has at least one publicly accessible port.",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The application internal host.",
				MarkdownDescription: "The application internal host. Use this to communicate between services within the same environment.",
				Computed:    true,
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
					"Full list available in [Qovery API documentation](https://api-doc.qovery.com/#tag/Applications/operation/getDefaultApplicationAdvancedSettings).",
				Optional:    true,
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Specify if the application will be automatically updated after receiving a new commit.",
				MarkdownDescription: "Specify if the application will be automatically redeployed after receiving a new commit on the configured branch.",
				Optional:    true,
				Computed:    true,
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				Description: "List of deployment restrictions",
				MarkdownDescription: "List of deployment restrictions. Deployment restrictions allow you to control when an application is deployed " +
					"based on file path changes in the git repository.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the deployment restriction.",
							MarkdownDescription: "Id of the deployment restriction.",
							Computed:             true,
						},
						"mode": schema.StringAttribute{
							Description: "Can be EXCLUDE or MATCH",
							MarkdownDescription: "Restriction mode. `MATCH`: deploy only when changes match the value. `EXCLUDE`: deploy only when changes do NOT match the value.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "Currently, only PATH is accepted",
							MarkdownDescription: "Type of deployment restriction. Currently only `PATH` is supported.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the deployment restriction",
							MarkdownDescription: "Value of the deployment restriction (e.g. a file path pattern like `src/` or `services/api/`).",
							Required:    true,
						},
					},
				},
			},
			"annotations_group_ids": schema.SetAttribute{
				Description: "List of annotations group ids",
				MarkdownDescription: "List of annotations group ids. Annotations groups allow you to add Kubernetes annotations to the application's pods.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description: "List of labels group ids",
				MarkdownDescription: "List of labels group ids. Labels groups allow you to add Kubernetes labels to the application's pods.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"docker_target_build_stage": schema.StringAttribute{
				Description: "The target build stage in the Dockerfile to build",
				MarkdownDescription: "The target build stage in a multi-stage Dockerfile to build. " +
					"Only applicable when `build_mode = \"DOCKER\"` and using a multi-stage Dockerfile.",
				Optional:    true,
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
	var isTriggeredFromImport = false
	if state.EnvironmentId.IsNull() {
		isTriggeredFromImport = true
	}

	// Get application from the API
	application, apiErr := r.client.GetApplication(ctx, state.Id.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
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
