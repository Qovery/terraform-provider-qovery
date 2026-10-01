package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

var (
	_ resource.ResourceWithConfigure      = &clusterDNSProviderResource{}
	_ resource.ResourceWithImportState    = clusterDNSProviderResource{}
	_ resource.ResourceWithValidateConfig = clusterDNSProviderResource{}
)

var clusterDNSProviderTypes = []string{
	clusterDNSProviderTypeQovery,
	clusterDNSProviderTypeCloudflare,
	clusterDNSProviderTypeRoute53,
}

type clusterDNSProviderResource struct {
	client *client.Client
}

func newClusterDNSProviderResource() resource.Resource {
	return &clusterDNSProviderResource{}
}

func (r clusterDNSProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_dns_provider"
}

func (r *clusterDNSProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r clusterDNSProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the DNS provider of a Qovery cluster.\n\n" +
			"~> **Note:** Destroying this resource only removes it from the Terraform state: the cluster keeps its DNS provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the resource, equal to `cluster_id`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the cluster." + recreatesOnChange("cluster DNS provider"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: "DNS provider of the cluster: `" + clusterDNSProviderTypeQovery + "` for the DNS Qovery manages, or `" +
					clusterDNSProviderTypeCloudflare + "` or `" + clusterDNSProviderTypeRoute53 + "` with the block of the same name.",
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterDNSProviderTypes),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "DNS domain of the cluster.",
				Required:            true,
			},
			"cloudflare": schema.SingleNestedAttribute{
				MarkdownDescription: "Cloudflare configuration. Required when `provider_type` is `CLOUDFLARE`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"email": schema.StringAttribute{
						MarkdownDescription: "Email of the Cloudflare account.",
						Required:            true,
					},
					"api_token": schema.StringAttribute{
						MarkdownDescription: "Cloudflare API token. Omitting it fails at plan time.",
						Optional:            true,
						Sensitive:           true,
					},
					"proxied": schema.BoolAttribute{
						MarkdownDescription: descriptions.NewBoolDefaultDescription("Whether Cloudflare proxies the DNS records of the cluster.", false),
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"route53": schema.SingleNestedAttribute{
				MarkdownDescription: "Route53 configuration. Required when `provider_type` is `ROUTE53`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"credentials": schema.SingleNestedAttribute{
						MarkdownDescription: "AWS credentials Qovery uses to manage the Route53 records.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								MarkdownDescription: "Type of the credentials. Only `" + clusterDNSProviderCredentialsStatic + "` is supported.",
								Required:            true,
								Validators: []validator.String{
									validators.NewStringEnumValidator([]string{clusterDNSProviderCredentialsStatic}),
								},
							},
							"aws_access_key_id": schema.StringAttribute{
								MarkdownDescription: credentialsAWSAccessKeyIDDescription,
								Required:            true,
							},
							"aws_secret_access_key": schema.StringAttribute{
								MarkdownDescription: credentialsAWSSecretAccessKeyDescription + " Omitting it fails at plan time.",
								Optional:            true,
								Sensitive:           true,
							},
						},
					},
					"aws_region": schema.StringAttribute{
						MarkdownDescription: "AWS region.",
						Required:            true,
					},
					"hosted_zone_id": schema.StringAttribute{
						MarkdownDescription: "ID of the Route53 hosted zone.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r clusterDNSProviderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config ClusterDNSProvider
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ProviderType.IsUnknown() || config.ProviderType.IsNull() {
		return
	}

	switch config.ProviderType.ValueString() {
	case clusterDNSProviderTypeCloudflare:
		if config.Cloudflare != nil && isMissingSecret(config.Cloudflare.APIToken) {
			resp.Diagnostics.AddAttributeError(
				path.Root("cloudflare").AtName("api_token"),
				"Missing required attribute",
				"cloudflare.api_token must be set when provider_type is CLOUDFLARE.",
			)
		}
	case clusterDNSProviderTypeRoute53:
		if config.Route53 != nil && config.Route53.Credentials != nil {
			if isMissingSecret(config.Route53.Credentials.AWSSecretAccessKey) {
				resp.Diagnostics.AddAttributeError(
					path.Root("route53").AtName("credentials").AtName("aws_secret_access_key"),
					"Missing required attribute",
					"route53.credentials.aws_secret_access_key must be set when provider_type is ROUTE53.",
				)
			}
		}
	}
}

// isMissingSecret reports a secret the configuration leaves null or empty. An unknown value,
// such as a variable at validate time or another resource's output at plan time, is checked once
// it is known.
func isMissingSecret(secret types.String) bool {
	if secret.IsUnknown() {
		return false
	}
	return secret.IsNull() || secret.ValueString() == ""
}

func (r clusterDNSProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClusterDNSProvider
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toQoveryRequest()
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster DNS provider configuration", err.Error())
		return
	}

	response, apiErr := r.client.UpdateClusterDNSProvider(ctx, plan.ClusterID.ValueString(), *request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state, err := convertResponseToClusterDNSProvider(plan.ClusterID.ValueString(), response, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster DNS provider response", err.Error())
		return
	}

	tflog.Trace(ctx, "created cluster DNS provider", map[string]any{"cluster_id": state.ClusterID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r clusterDNSProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterDNSProvider
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, apiErr := r.client.GetClusterDNSProvider(ctx, state.ClusterID.ValueString())
	if handleReadNotFound(ctx, resp, apiErr) {
		return
	}

	newState, err := convertResponseToClusterDNSProvider(state.ClusterID.ValueString(), response, state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster DNS provider response", err.Error())
		return
	}

	tflog.Trace(ctx, "read cluster DNS provider", map[string]any{"cluster_id": newState.ClusterID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r clusterDNSProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClusterDNSProvider
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toQoveryRequest()
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster DNS provider configuration", err.Error())
		return
	}

	response, apiErr := r.client.UpdateClusterDNSProvider(ctx, plan.ClusterID.ValueString(), *request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state, err := convertResponseToClusterDNSProvider(plan.ClusterID.ValueString(), response, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster DNS provider response", err.Error())
		return
	}

	tflog.Trace(ctx, "updated cluster DNS provider", map[string]any{"cluster_id": state.ClusterID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r clusterDNSProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClusterDNSProvider
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "removed cluster DNS provider from Terraform state", map[string]any{"cluster_id": state.ClusterID.ValueString()})
	resp.State.RemoveResource(ctx)
}

func (r clusterDNSProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			"Expected import identifier with format: cluster_id.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), req.ID)...)
}
