package qovery

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/qovery/apierror"
)

const awsCredentialsAPIResource = "aws credentials"

type awsCredentialsData struct {
	Id              types.String `tfsdk:"id"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	Name            types.String `tfsdk:"name"`
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

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

func (r awsCredentialsResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return awsCredentialsResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type awsCredentialsResource struct {
	client *qovery.APIClient
}

// Create qovery aws credentials resource
func (r awsCredentialsResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Retrieve values from plan
	var plan awsCredentialsData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new credentials
	credentials, res, err := r.client.CloudProviderCredentialsApi.
		CreateAWSCredentials(ctx, plan.OrganizationId.Value).
		AwsCredentialsRequest(qovery.AwsCredentialsRequest{
			Name:            plan.Name.Value,
			AccessKeyId:     &plan.AccessKeyId.Value,
			SecretAccessKey: &plan.SecretAccessKey.Value,
		}).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := awsCredentialsCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := awsCredentialsData{
		Id:              types.String{Value: *credentials.Id},
		Name:            types.String{Value: *credentials.Name},
		OrganizationId:  plan.OrganizationId,
		AccessKeyId:     plan.AccessKeyId,
		SecretAccessKey: plan.SecretAccessKey,
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery aws credentials resource
func (r awsCredentialsResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state awsCredentialsData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API
	credentials, res, err := r.client.CloudProviderCredentialsApi.
		ListAWSCredentials(ctx, state.OrganizationId.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := awsCredentialsReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	var toRefresh *awsCredentialsData
	for _, creds := range credentials.GetResults() {
		if state.Id.Value == *creds.Id {
			toRefresh = &awsCredentialsData{
				Name: types.String{Value: *creds.Name},
			}
			break
		}
	}

	// If credential id is not in list
	// Returning Not Found error
	if toRefresh == nil {
		res.StatusCode = 404
		apiErr := awsCredentialsReadAPIError(state.Id.Value, res, nil)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state.Name = toRefresh.Name

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery aws credentials resource
func (r awsCredentialsResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state awsCredentialsData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update credentials in the backend
	credentials, res, err := r.client.CloudProviderCredentialsApi.
		EditAWSCredentials(ctx, state.OrganizationId.Value, state.Id.Value).
		AwsCredentialsRequest(qovery.AwsCredentialsRequest{
			Name:            plan.Name.Value,
			AccessKeyId:     &plan.AccessKeyId.Value,
			SecretAccessKey: &plan.SecretAccessKey.Value,
		}).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := awsCredentialsUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	toUpdate := awsCredentialsData{
		Name:            types.String{Value: *credentials.Name},
		AccessKeyId:     plan.AccessKeyId,
		SecretAccessKey: plan.SecretAccessKey,
	}

	// Update state values
	state.Name = toUpdate.Name
	state.AccessKeyId = toUpdate.AccessKeyId
	state.SecretAccessKey = toUpdate.SecretAccessKey

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery aws credentials resource
func (r awsCredentialsResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state awsCredentialsData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete credentials in the backend
	res, err := r.client.CloudProviderCredentialsApi.
		DeleteAWSCredentials(ctx, state.OrganizationId.Value, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := awsCredentialsDeleteAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Remove credentials from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery aws credentials resource using its id
func (r awsCredentialsResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: aws_credentials_id,organization_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("organization_id"), idParts[1])...)
}

func awsCredentialsCreateAPIError(credentialsName string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(awsCredentialsAPIResource, credentialsName, apierror.Create, res, err)
}

func awsCredentialsReadAPIError(credentialsID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(awsCredentialsAPIResource, credentialsID, apierror.Read, res, err)
}

func awsCredentialsUpdateAPIError(credentialsID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(awsCredentialsAPIResource, credentialsID, apierror.Update, res, err)
}

func awsCredentialsDeleteAPIError(credentialsID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(awsCredentialsAPIResource, credentialsID, apierror.Delete, res, err)
}
