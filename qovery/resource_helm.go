package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &helmResource{}
	_ resource.ResourceWithImportState = helmResource{}
)

var helmPortProtocols = clientEnumToStringArray(helm.AllowedProtocols)

type helmResource struct {
	helmService helm.Service
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
}

func (r helmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Provides a Qovery helm resource. This can be used to create and manage Qovery Helm chart deployments.",
		MarkdownDescription: "Provides a Qovery helm resource. This can be used to create and manage Qovery Helm chart deployments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the helm service.",
				MarkdownDescription: "Id of the helm service.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description:         "Id of the environment.",
				MarkdownDescription: "Id of the environment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "Name of the helm service.",
				MarkdownDescription: "Name of the helm service.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				Description:         "Description of the helm service.",
				MarkdownDescription: "Description of the helm service.",
				Required:            true,
			},
			"icon_uri": schema.StringAttribute{
				Description:         "Icon URI representing the helm service.",
				MarkdownDescription: "Icon URI representing the helm service.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeout_sec": schema.Int64Attribute{
				Description:         "Helm timeout in seconds. Maximum time allowed for the Helm operation to complete.",
				MarkdownDescription: "Helm timeout in seconds. Maximum time allowed for the Helm operation to complete.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(helm.DefaultTimeoutSec),
				// Required: true,
			},
			"auto_preview": schema.BoolAttribute{
				Description:         "Specify if the environment preview option is activated or not for this helm.",
				MarkdownDescription: "Specify if the environment preview option is activated or not for this helm.",
				Optional:            true,
				Computed:            true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description:         "Specify if the helm service will be automatically updated on every new commit on the branch.",
				MarkdownDescription: "Specify if the helm service will be automatically updated on every new commit on the branch.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"arguments": schema.ListAttribute{
				Description:         "Helm CLI arguments passed to the helm command (e.g. --wait, --atomic, --debug).",
				MarkdownDescription: "Helm CLI arguments passed to the helm command (e.g. `--wait`, `--atomic`, `--debug`).",
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
				Description:         "Allow this chart to deploy resources outside of this environment namespace (including CRDs or non-namespaced resources)",
				MarkdownDescription: "Allow this chart to deploy resources outside of this environment namespace (including CRDs or non-namespaced resources)",
				Required:            true,
			},
			"source": schema.SingleNestedAttribute{
				Description:         "Helm chart source. Use helm_repository to deploy from a Helm repository, or git_repository to deploy from a git repository.",
				MarkdownDescription: "Helm chart source. Use `helm_repository` to deploy from a Helm repository, or `git_repository` to deploy from a git repository.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"helm_repository": schema.SingleNestedAttribute{
						Description:         "Helm chart from a Helm repository. Repositories can be HTTPS or OCI-based (ECR, Docker Hub, GHCR, etc.).",
						MarkdownDescription: "Helm chart from a Helm repository. Repositories can be HTTPS or OCI-based (ECR, Docker Hub, GHCR, etc.).",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"helm_repository_id": schema.StringAttribute{
								Description:         "Id of the Helm repository (refers to a qovery_helm_repository resource).",
								MarkdownDescription: "Id of the Helm repository (refers to a `qovery_helm_repository` resource).",
								Required:            true,
							},
							"chart_name": schema.StringAttribute{
								Description:         "Name of the Helm chart to deploy.",
								MarkdownDescription: "Name of the Helm chart to deploy.",
								Required:            true,
							},
							"chart_version": schema.StringAttribute{
								Description:         "Version of the Helm chart to deploy (e.g. 1.0.0).",
								MarkdownDescription: "Version of the Helm chart to deploy (e.g. `1.0.0`).",
								Required:            true,
							},
						},
					},
					"git_repository": schema.SingleNestedAttribute{
						Description:         "Helm chart from a git repository. The repository must contain valid Helm chart files.",
						MarkdownDescription: "Helm chart from a git repository. The repository must contain valid Helm chart files.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description:         "Git repository URL containing the Helm chart.",
								MarkdownDescription: "Git repository URL containing the Helm chart.",
								Required:            true,
							},
							"branch": schema.StringAttribute{
								Description:         "Git branch to use for the Helm chart source.",
								MarkdownDescription: "Git branch to use for the Helm chart source.",
								Optional:            true,
								Computed:            true,
							},
							"root_path": schema.StringAttribute{
								Description:         "Root path in the git repository where the Helm chart is located.",
								MarkdownDescription: "Root path in the git repository where the Helm chart is located.",
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString("/"),
							},
							"git_token_id": schema.StringAttribute{
								Description:         "Git token ID for accessing a private repository (refers to a qovery_git_token resource).",
								MarkdownDescription: "Git token ID for accessing a private repository (refers to a `qovery_git_token` resource).",
								Optional:            true,
								Computed:            true,
							},
						},
					},
				},
			},
			"values_override": schema.SingleNestedAttribute{
				Description:         "Define your own overrides to customize the helm chart behaviour.",
				MarkdownDescription: "Define your own overrides to customize the helm chart behaviour.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"set": schema.MapAttribute{
						Description:         "Override Helm values using --set flag syntax. Map of key-value pairs.",
						MarkdownDescription: "Override Helm values using `--set` flag syntax. Map of key-value pairs.",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"set_string": schema.MapAttribute{
						Description:         "Override Helm values using --set-string flag syntax. Values are always treated as strings.",
						MarkdownDescription: "Override Helm values using `--set-string` flag syntax. Values are always treated as strings.",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"set_json": schema.MapAttribute{
						Description:         "Override Helm values using --set-json flag syntax. Values are treated as JSON.",
						MarkdownDescription: "Override Helm values using `--set-json` flag syntax. Values are treated as JSON.",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"file": schema.SingleNestedAttribute{
						Description:         "Define overrides by selecting a YAML file from a git repository (preferred) or by passing raw YAML files.",
						MarkdownDescription: "Define overrides by selecting a YAML file from a git repository (preferred) or by passing raw YAML files.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"raw": schema.MapNestedAttribute{
								Description:         "Raw YAML files",
								MarkdownDescription: "Raw YAML files",
								Optional:            true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"content": schema.StringAttribute{
											Description:         "content of the file",
											MarkdownDescription: "content of the file",
											Required:            true,
										},
									},
								},
							},
							"git_repository": schema.SingleNestedAttribute{
								Description:         "YAML file from a git repository",
								MarkdownDescription: "YAML file from a git repository",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description:         "YAML file git repository URL",
										MarkdownDescription: "YAML file git repository URL",
										Required:            true,
									},
									"branch": schema.StringAttribute{
										Description:         "YAML file git repository branch",
										MarkdownDescription: "YAML file git repository branch",
										Required:            true,
									},
									"paths": schema.SetAttribute{
										Description:         "YAML files git repository paths",
										MarkdownDescription: "YAML files git repository paths",
										Required:            true,
										ElementType:         types.StringType,
									},
									"git_token_id": schema.StringAttribute{
										Description:         "Git token ID for accessing a private repository (refers to a qovery_git_token resource).",
										MarkdownDescription: "Git token ID for accessing a private repository (refers to a `qovery_git_token` resource).",
										Optional:            true,
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
			"ports": schema.MapNestedAttribute{
				Description:         "List of ports linked to this helm.",
				MarkdownDescription: "List of ports linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_name": schema.StringAttribute{
							Description:         "Name of the Kubernetes service to expose.",
							MarkdownDescription: "Name of the Kubernetes service to expose.",
							Required:            true,
						},
						"namespace": schema.StringAttribute{
							Description:         "Kubernetes namespace where the service is deployed.",
							MarkdownDescription: "Kubernetes namespace where the service is deployed.",
							Optional:            true,
						},
						"internal_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"Internal port of the container.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							MarkdownDescription: descriptions.NewInt64MinMaxDescription(
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
								"External port of the container. Required if: ports.publicly_accessible=true.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							MarkdownDescription: descriptions.NewInt64MinMaxDescription(
								"External port of the container.\n\t- Required if: `ports.publicly_accessible=true`.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							Required: true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the container.",
								helmPortProtocols,
								new(helm.DefaultProtocol.String()),
							),
							MarkdownDescription: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the container.",
								helmPortProtocols,
								new(helm.DefaultProtocol.String()),
							),
							Validators: []validator.String{
								validators.NewStringEnumValidator(helmPortProtocols),
							},
							Optional: true,
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							Description:         "If this port will be used for the root domain. Note: the API may override this value based on port configuration (e.g., when only one publicly accessible port exists, it will be set as default).",
							MarkdownDescription: "If this port will be used for the root domain. Note: the API may override this value based on port configuration (e.g., when only one publicly accessible port exists, it will be set as default).",
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
				Description:         "List of built-in environment variables linked to this helm.",
				MarkdownDescription: "List of built-in environment variables linked to this helm.",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					UseStateUnlessNameChanges(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
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
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description:         "List of environment variables linked to this helm.",
				MarkdownDescription: "List of environment variables linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
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
				Description:         "List of environment variable aliases linked to this helm.",
				MarkdownDescription: "List of environment variable aliases linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable alias.",
							MarkdownDescription: "Id of the environment variable alias.",
							Computed:            true,
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
				Description:         "List of environment variable overrides linked to this helm.",
				MarkdownDescription: "List of environment variable overrides linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable override.",
							MarkdownDescription: "Id of the environment variable override.",
							Computed:            true,
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
				Description:         "List of secrets linked to this helm.",
				MarkdownDescription: "List of secrets linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret.",
							MarkdownDescription: "Id of the secret.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the secret.",
							MarkdownDescription: "Key of the secret.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret.",
							MarkdownDescription: "Value of the secret.",
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
				Description:         "List of secret aliases linked to this helm.",
				MarkdownDescription: "List of secret aliases linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret alias.",
							MarkdownDescription: "Id of the secret alias.",
							Computed:            true,
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
				Description:         "List of secret overrides linked to this helm.",
				MarkdownDescription: "List of secret overrides linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret override.",
							MarkdownDescription: "Id of the secret override.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret override.",
							MarkdownDescription: "Name of the secret override.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret override.",
							MarkdownDescription: "Value of the secret override.",
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
			"environment_variable_files": environmentVariableFilesSchemaAttribute("helm"),
			"secret_files":              secretFilesSchemaAttribute("helm"),
			"custom_domains": schema.SetNestedAttribute{
				Description:         "List of custom domains linked to this helm.",
				MarkdownDescription: "List of custom domains linked to this helm.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the custom domain.",
							MarkdownDescription: "Id of the custom domain.",
							Computed:            true,
						},
						"domain": schema.StringAttribute{
							Description:         "Your custom domain.",
							MarkdownDescription: "Your custom domain.",
							Required:            true,
						},
						"generate_certificate": schema.BoolAttribute{
							Description:         "Qovery will generate and manage the certificate for this domain.",
							MarkdownDescription: "Qovery will generate and manage the certificate for this domain.",
							Required:            true,
						},
						"use_cdn": schema.BoolAttribute{
							Description: "Indicates if the custom domain is behind a CDN (i.e Cloudflare). " +
								"This will condition the way we are checking CNAME before & during a deployment: " +
								"If true then we only check the domain points to an IP. " +
								"If false then we check that the domain resolves to the correct service Load Balancer",
							MarkdownDescription: "Indicates if the custom domain is behind a CDN (i.e Cloudflare).\n" +
								"This will condition the way we are checking CNAME before & during a deployment:\n" +
								" * If `true` then we only check the domain points to an IP\n" +
								" * If `false` then we check that the domain resolves to the correct service Load Balancer",
							Optional: true,
						},
						"validation_domain": schema.StringAttribute{
							Description:         "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
							MarkdownDescription: "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							Description:         "Status of the custom domain.",
							MarkdownDescription: "Status of the custom domain.",
							Computed:            true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description:         "The helm external FQDN host [NOTE: only if your helm is using a publicly accessible port].",
				MarkdownDescription: "The helm external FQDN host [NOTE: only if your helm is using a publicly accessible port].",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal_host": schema.StringAttribute{
				Description:         "The helm internal host.",
				MarkdownDescription: "The helm internal host.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_stage_id": schema.StringAttribute{
				Description:         "Id of the deployment stage. Controls the order of service deployment within an environment.",
				MarkdownDescription: "Id of the deployment stage. Controls the order of service deployment within an environment.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_skipped": schema.BoolAttribute{
				Description:         "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				MarkdownDescription: "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"advanced_settings_json": schema.StringAttribute{
				Description:         "Advanced settings in JSON format. See the Qovery API documentation for available settings: https://api-doc.qovery.com/#tag/Helms/operation/getDefaultHelmAdvancedSettings",
				MarkdownDescription: "Advanced settings in JSON format. See the Qovery API documentation for available settings: https://api-doc.qovery.com/#tag/Helms/operation/getDefaultHelmAdvancedSettings",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				Description:         "List of deployment restrictions.",
				MarkdownDescription: "List of deployment restrictions.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the deployment restriction.",
							MarkdownDescription: "Id of the deployment restriction.",
							Computed:            true,
						},
						"mode": schema.StringAttribute{
							Description:         "Deployment restriction mode. Can be: EXCLUDE, MATCH.",
							MarkdownDescription: "Deployment restriction mode.\n\t- Can be: `EXCLUDE`, `MATCH`.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							Description:         "Deployment restriction type. Can be: PATH.",
							MarkdownDescription: "Deployment restriction type.\n\t- Can be: `PATH`.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the deployment restriction (e.g. a file path pattern).",
							MarkdownDescription: "Value of the deployment restriction (e.g. a file path pattern).",
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
	if err != nil {
		resp.Diagnostics.AddError("Error on helm read", err.Error())
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
