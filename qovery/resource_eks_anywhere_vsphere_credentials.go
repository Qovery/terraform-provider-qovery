package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &eksAnywhereVsphereCredentialsResource{}
var _ resource.ResourceWithImportState = eksAnywhereVsphereCredentialsResource{}

type eksAnywhereVsphereCredentialsResource struct {
	eksAnywhereVsphereCredentialsService credentials.EksAnywhereVsphereService
}

func newEksAnywhereVsphereCredentialsResource() resource.Resource {
	return &eksAnywhereVsphereCredentialsResource{}
}

func (r eksAnywhereVsphereCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eks_anywhere_vsphere_credentials"
}

func (r *eksAnywhereVsphereCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.eksAnywhereVsphereCredentialsService = provider.eksAnywhereVsphereCredentialsService
}

func (r eksAnywhereVsphereCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery EKS Anywhere vSphere credentials resource. This can be used to create and manage Qovery EKS Anywhere vSphere credentials.",
		MarkdownDescription: "Provides a Qovery EKS Anywhere vSphere credentials resource. This is used to create and manage credentials that Qovery uses to provision and manage EKS Anywhere clusters running on vSphere infrastructure.\n\n" +
			"You can authenticate the AWS side using either **IAM access keys** (`access_key_id` + `secret_access_key`) or an **IAM role** (`role_arn`). These two methods are mutually exclusive.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the EKS Anywhere vSphere credentials.",
				MarkdownDescription: "Unique identifier of the EKS Anywhere vSphere credentials (UUID format).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization.",
				MarkdownDescription: "ID of the Qovery organization in which to create the credentials.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the EKS Anywhere vSphere credentials.",
				MarkdownDescription: "Name of the EKS Anywhere vSphere credentials. Used for display purposes in the Qovery console.",
				Required:            true,
			},
			"vsphere_user": schema.StringAttribute{
				Description:         "Your vSphere username.",
				MarkdownDescription: "Username used to authenticate against the vSphere API.",
				Required:            true,
			},
			"vsphere_password": schema.StringAttribute{
				Description:         "Your vSphere password.",
				MarkdownDescription: "Password used to authenticate against the vSphere API. This is a sensitive value and will not be displayed in plan output.",
				Required:            true,
				Sensitive:           true,
			},
			"access_key_id": schema.StringAttribute{
				Description:         "Your AWS access key id.",
				MarkdownDescription: "AWS IAM access key ID. Required when using access key authentication. Must not be set when `role_arn` is specified.",
				Optional:            true,
			},
			"secret_access_key": schema.StringAttribute{
				Description:         "Your AWS secret access key.",
				MarkdownDescription: "AWS IAM secret access key. Required when using access key authentication. This is a sensitive value and will not be displayed in plan output.",
				Optional:            true,
				Sensitive:           true,
			},
			"role_arn": schema.StringAttribute{
				Description:         "Your AWS role ARN that you want Qovery to assume. You can't specify access/secret_key if you use a role.",
				MarkdownDescription: "ARN of the AWS IAM role that Qovery will assume. Must not be set when `access_key_id`/`secret_access_key` are specified.",
				Optional:            true,
			},
		},
	}
}

// Create qovery eks anywhere vsphere credentials resource
func (r eksAnywhereVsphereCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EksAnywhereVsphereCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toUpsertEksAnywhereVsphereRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error on eks anywhere vsphere credentials create", err.Error())
		return
	}

	creds, err := r.eksAnywhereVsphereCredentialsService.Create(ctx, plan.OrganizationId.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on eks anywhere vsphere credentials create", err.Error())
		return
	}

	state := convertDomainCredentialsToEksAnywhereVsphereCredentials(creds, plan)
	tflog.Trace(ctx, "created eks anywhere vsphere credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery eks anywhere vsphere credentials resource
func (r eksAnywhereVsphereCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EksAnywhereVsphereCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := r.eksAnywhereVsphereCredentialsService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on eks anywhere vsphere credentials read", err.Error())
		return
	}

	state = convertDomainCredentialsToEksAnywhereVsphereCredentials(creds, state)
	tflog.Trace(ctx, "read eks anywhere vsphere credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery eks anywhere vsphere credentials resource
func (r eksAnywhereVsphereCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EksAnywhereVsphereCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toUpsertEksAnywhereVsphereRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error on eks anywhere vsphere credentials update", err.Error())
		return
	}

	creds, err := r.eksAnywhereVsphereCredentialsService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on eks anywhere vsphere credentials update", err.Error())
		return
	}

	state = convertDomainCredentialsToEksAnywhereVsphereCredentials(creds, plan)
	tflog.Trace(ctx, "updated eks anywhere vsphere credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery eks anywhere vsphere credentials resource
func (r eksAnywhereVsphereCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EksAnywhereVsphereCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.eksAnywhereVsphereCredentialsService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on eks anywhere vsphere credentials delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted eks anywhere vsphere credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery eks anywhere vsphere credentials resource using its id
func (r eksAnywhereVsphereCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,eks_anywhere_vsphere_credentials_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
