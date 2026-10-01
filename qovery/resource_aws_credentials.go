package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &awsCredentialsResource{}
	_ resource.ResourceWithImportState = awsCredentialsResource{}
)

type awsCredentialsResource struct {
	awsCredentialsService credentials.AwsService
}

func newAwsCredentialsResource() resource.Resource {
	return &awsCredentialsResource{}
}

func (r awsCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_credentials"
}

func (r *awsCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.awsCredentialsService = provider.awsCredentialsService
}

func (r awsCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Qovery AWS credentials: the IAM role or access key Qovery uses to create and manage clusters in an AWS account. Set either `role_arn` or `access_key_id` and `secret_access_key`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("AWS credentials"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription + recreatesOnChange("AWS credentials"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("AWS credentials"),
				Required:            true,
			},
			"access_key_id": schema.StringAttribute{
				MarkdownDescription: credentialsAWSAccessKeyIDDescription,
				Optional:            true,
				Sensitive:           false,
			},
			"secret_access_key": schema.StringAttribute{
				MarkdownDescription: credentialsAWSSecretAccessKeyDescription,
				Optional:            true,
				Sensitive:           true,
			},
			"role_arn": schema.StringAttribute{
				MarkdownDescription: credentialsAWSRoleARNDescription + " When set, `access_key_id` and `secret_access_key` are ignored.",
				Optional:            true,
				Sensitive:           false,
			},
		},
	}
}

// Create qovery aws credentials resource
func (r awsCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan AWSCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new credentials
	creds, err := r.awsCredentialsService.Create(ctx, plan.OrganizationId.ValueString(), plan.toUpsertAwsRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainCredentialsToAWSCredentials(creds, plan)
	tflog.Trace(ctx, "created aws credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery aws credentials resource
func (r awsCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state AWSCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API
	creds, err := r.awsCredentialsService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if handleDomainReadNotFound(ctx, resp, err, "Error on aws credentials read") {
		return
	}

	state = convertDomainCredentialsToAWSCredentials(creds, state)
	tflog.Trace(ctx, "read aws credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery aws credentials resource
func (r awsCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state AWSCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update credentials in the backend
	creds, err := r.awsCredentialsService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), plan.toUpsertAwsRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials update", err.Error())
		return
	}

	// Update state values
	state = convertDomainCredentialsToAWSCredentials(creds, plan)
	tflog.Trace(ctx, "updated aws credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery aws credentials resource
func (r awsCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state AWSCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete credentials in the backend
	err := r.awsCredentialsService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted aws credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Remove credentials from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery aws credentials resource using its id
func (r awsCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,aws_credentials_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
