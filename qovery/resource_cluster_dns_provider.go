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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

var (
	_ resource.ResourceWithConfigure   = &clusterDNSProviderResource{}
	_ resource.ResourceWithImportState = clusterDNSProviderResource{}
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
		Description:         "Provides a Qovery cluster DNS provider resource. This optional resource manages the DNS provider associated with a cluster.",
		MarkdownDescription: "Provides a Qovery cluster DNS provider resource. Declare this resource only when Terraform must explicitly manage the cluster DNS provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the cluster DNS provider resource. This is the cluster id.",
				MarkdownDescription: "Id of the cluster DNS provider resource. This is the cluster id.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description:         "Id of the cluster.",
				MarkdownDescription: "Id of the cluster whose DNS provider is managed by this resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_type": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"DNS provider type.",
					clusterDNSProviderTypes,
					nil,
				),
				MarkdownDescription: descriptions.NewStringEnumDescription(
					"DNS provider type.",
					clusterDNSProviderTypes,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterDNSProviderTypes),
				},
			},
			"domain": schema.StringAttribute{
				Description:         "DNS domain associated with the cluster.",
				MarkdownDescription: "DNS domain associated with the cluster.",
				Required:            true,
			},
			"cloudflare": schema.SingleNestedAttribute{
				Description:         "Cloudflare DNS provider configuration. Required when provider_type is CLOUDFLARE.",
				MarkdownDescription: "Cloudflare DNS provider configuration. Required when `provider_type` is `CLOUDFLARE`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"email": schema.StringAttribute{
						Description:         "Cloudflare account email.",
						MarkdownDescription: "Cloudflare account email.",
						Required:            true,
					},
					"api_token": schema.StringAttribute{
						Description:         "Cloudflare API token. This value must be provided on create and update and is not returned by the Qovery API.",
						MarkdownDescription: "Cloudflare API token. This value must be provided on create and update and is not returned by the Qovery API.",
						Optional:            true,
						Sensitive:           true,
					},
					"proxied": schema.BoolAttribute{
						Description:         "Whether Cloudflare proxying is enabled.",
						MarkdownDescription: "Whether Cloudflare proxying is enabled.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"route53": schema.SingleNestedAttribute{
				Description:         "Route53 DNS provider configuration. Required when provider_type is ROUTE53.",
				MarkdownDescription: "Route53 DNS provider configuration. Required when `provider_type` is `ROUTE53`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"credentials": schema.SingleNestedAttribute{
						Description:         "Route53 credentials.",
						MarkdownDescription: "Route53 credentials.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Description:         "Route53 credentials type. Only STATIC is supported.",
								MarkdownDescription: "Route53 credentials type. Only `STATIC` is supported.",
								Required:            true,
								Validators: []validator.String{
									validators.NewStringEnumValidator([]string{clusterDNSProviderCredentialsStatic}),
								},
							},
							"aws_access_key_id": schema.StringAttribute{
								Description:         "AWS access key id.",
								MarkdownDescription: "AWS access key id.",
								Required:            true,
							},
							"aws_secret_access_key": schema.StringAttribute{
								Description:         "AWS secret access key. This value must be provided on create and update and is not returned by the Qovery API.",
								MarkdownDescription: "AWS secret access key. This value must be provided on create and update and is not returned by the Qovery API.",
								Optional:            true,
								Sensitive:           true,
							},
						},
					},
					"aws_region": schema.StringAttribute{
						Description:         "AWS region.",
						MarkdownDescription: "AWS region.",
						Required:            true,
					},
					"hosted_zone_id": schema.StringAttribute{
						Description:         "Route53 hosted zone id.",
						MarkdownDescription: "Route53 hosted zone id.",
						Optional:            true,
					},
				},
			},
		},
	}
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
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
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
