package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type resourceAwsCredentialsType struct{}

// AwsCredentials Resource schema
func (r resourceAwsCredentialsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				Required: false,
				Optional: false,
			},
			"name": {
				Type:     types.StringType,
				Computed: false,
				Required: true,
				Optional: false,
			},
			"access_key_id": {
				Type:      types.StringType,
				Computed:  false,
				Required:  true,
				Optional:  false,
				Sensitive: true,
			},
			"secret_access_key": {
				Type:      types.StringType,
				Computed:  false,
				Required:  true,
				Optional:  false,
				Sensitive: true,
			},
			"organization_id": {
				Type:     types.StringType,
				Computed: false,
				Required: true,
				Optional: false,
			},
		},
	}, nil
}

// New resource instance
func (r resourceAwsCredentialsType) NewResource(ct context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceAwsCredentials{
		p: *(p.(*provider)),
	}, nil
}

type resourceAwsCredentials struct {
	p provider
}

// Create a new resource
func (r resourceAwsCredentials) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan AwsCredentials
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new credentials
	credentials, res, err := r.p.client.CloudProviderCredentialsApi.
		CreateAWSCredentials(ctx, plan.OrganizationId.Value).
		AwsCredentialsRequest(
			qovery.AwsCredentialsRequest{
				Name:            plan.Name.Value,
				AccessKeyId:     &plan.AccessKeyId.Value,
				SecretAccessKey: &plan.SecretAccessKey.Value,
			}).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating credentials",
			"Could not create credentials, unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error creating credentials",
			"Could not create credentials, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	// Generate resource state struct
	var result = AwsCredentials{
		Id: types.String{
			Value: *credentials.Id,
		},
		Name: types.String{
			Value: *credentials.Name,
		},
		AccessKeyId:     plan.AccessKeyId,
		SecretAccessKey: plan.SecretAccessKey,
		OrganizationId:  plan.OrganizationId,
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r resourceAwsCredentials) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state AwsCredentials
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API and then update what is in state from what the API returns
	credentials, res, err := r.p.client.CloudProviderCredentialsApi.
		ListAWSCredentials(ctx, state.OrganizationId.Value).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading credentials",
			"Could not read credentials of organization "+state.OrganizationId.Value+", unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error reading credentials",
			"Could not read credentials of organization "+state.OrganizationId.Value+", unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	var found = AwsCredentials{}
	for _, credential := range credentials.GetResults() {
		if state.Id.Value == *credential.Id {
			found = AwsCredentials{
				Id: types.String{
					Value: *credential.Id,
				},
				Name: types.String{
					Value: *credential.Name,
				},
			}
		}
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error reading credentials",
			"Could not find credentials of organization "+state.OrganizationId.Value+" with ID: "+state.Id.Value,
		)
		return
	}

	state.Name = found.Name

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r resourceAwsCredentials) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get current state
	var plan AwsCredentials
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update credentials in the backend
	credential, res, err := r.p.client.CloudProviderCredentialsApi.
		EditAWSCredentials(ctx, plan.OrganizationId.Value, plan.Id.Value).
		AwsCredentialsRequest(
			qovery.AwsCredentialsRequest{
				Name:            plan.Name.Value,
				AccessKeyId:     &plan.AccessKeyId.Value,
				SecretAccessKey: &plan.SecretAccessKey.Value},
		).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating credentials",
			"Could not update credentials, unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error updating credentials",
			"Could not update credentials, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	// Get current state
	var state AwsCredentials
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	req.State.Get(ctx, state)

	// Update
	state.Name.Value = *credential.Name
	state.Id.Value = *credential.Id
	state.AccessKeyId = plan.AccessKeyId
	state.SecretAccessKey = plan.SecretAccessKey

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceAwsCredentials) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state AwsCredentials
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete credentials in the backend
	res, err := r.p.client.CloudProviderCredentialsApi.DeleteAWSCredentials(ctx, state.OrganizationId.Value, state.Id.Value).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting credentials",
			"Could not delete credentials "+state.Id.Value+": "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error deleting credentials",
			"Could not delete credentials, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
