package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.ResourceType = awsCredentialsResourceType{}
var _ resource.Resource = awsCredentialsResource{}
var _ resource.ResourceWithImportState = awsCredentialsResource{}

type awsCredentialsResourceType struct{}

func (r awsCredentialsResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery AWS credentials resource. This can be used to create and manage Qovery AWS credentials.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the AWS credentials.",
				Type:        types.StringType,
				Computed:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the aws credentials.",
				Type:        types.StringType,
				Required:    true,
			},
			"access_key_id": {
				Description: "Your AWS access key id.",
				Type:        types.StringType,
				Required:    true,
				Sensitive:   true,
			},
			"secret_access_key": {
				Description: "Your AWS secret access key.",
				Type:        types.StringType,
				Required:    true,
				Sensitive:   true,
			},
		},
	}, nil
}

func (r awsCredentialsResourceType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return awsCredentialsResource{
		awsCredentialsService: p.(*qProvider).awsCredentialsService,
	}, nil
}

type awsCredentialsResource struct {
	awsCredentialsService credentials.AwsService
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
	creds, err := r.awsCredentialsService.Create(ctx, plan.OrganizationId.Value, plan.toUpsertAwsRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainCredentialsToAWSCredentials(creds, plan)
	tflog.Trace(ctx, "created aws credentials", map[string]interface{}{"credentials_id": state.Id.Value})

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
	creds, err := r.awsCredentialsService.Get(ctx, state.OrganizationId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials read", err.Error())
		return
	}

	state = convertDomainCredentialsToAWSCredentials(creds, state)
	tflog.Trace(ctx, "read aws credentials", map[string]interface{}{"credentials_id": state.Id.Value})

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
	creds, err := r.awsCredentialsService.Update(ctx, state.OrganizationId.Value, state.Id.Value, plan.toUpsertAwsRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials update", err.Error())
		return
	}

	// Update state values
	state = convertDomainCredentialsToAWSCredentials(creds, plan)
	tflog.Trace(ctx, "updated aws credentials", map[string]interface{}{"credentials_id": state.Id.Value})

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
	err := r.awsCredentialsService.Delete(ctx, state.OrganizationId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted aws credentials", map[string]interface{}{"credentials_id": state.Id.Value})

	// Remove credentials from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery aws credentials resource using its id
func (r awsCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: aws_credentials_id,organization_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
