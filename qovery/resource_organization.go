package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type resourceOrganizationType struct{}

// Organization Resource schema
func (r resourceOrganizationType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				Computed: false,
			},
			"plan": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

// New resource instance
func (r resourceOrganizationType) NewResource(ct context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceOrganization{
		p: *(p.(*provider)),
	}, nil
}

type resourceOrganization struct {
	p provider
}

// Create a new resource
func (r resourceOrganization) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks! ",
		)
		return
	}

	// Retrieve values from plan
	var plan Organization
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new org
	org, res, err := r.p.client.OrganizationMainCallsApi.
		CreateOrganization(ctx).
		OrganizationRequest(qovery.OrganizationRequest{
			Name: plan.Name.Value,
			Plan: plan.Plan.Value,
		}).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating organization",
			"Could not create organization, unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error creating organization",
			"Could not create organization, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	// Generate resource state struct
	var result = Organization{
		Id: types.String{
			Value: org.Id,
		},
		Name: types.String{
			Value: org.Name,
		},
		Plan: types.String{
			Value: org.Plan,
		},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r resourceOrganization) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Organization
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization from API and then update what is in state from what the API returns
	organization, res, err := r.p.client.OrganizationMainCallsApi.
		GetOrganization(ctx, state.Id.Value).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization",
			"Could not read organization "+state.Id.Value+", unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error reading organization",
			"Could not read organization "+state.Id.Value+", unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	state.Name.Value = organization.Name
	state.Plan.Value = organization.Plan
	state.Id.Value = organization.Id

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r resourceOrganization) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	{
		// Get current state
		var plan Organization
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Update organization in backend
		org, res, err := r.p.client.OrganizationMainCallsApi.
			EditOrganization(ctx, plan.Id.Value).
			OrganizationEditRequest(qovery.OrganizationEditRequest{
				Name: plan.Name.Value,
			}).
			Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating organization",
				"Could not update organization, unexpected error: "+err.Error(),
			)
			return
		}
		if res.StatusCode >= 400 {
			resp.Diagnostics.AddError(
				"Error updating organization",
				"Could not update organization, unexpected status code: "+string(rune(res.StatusCode)),
			)
			return
		}

		// Update state
		var state Organization
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		req.State.Get(ctx, state)

		state.Name.Value = org.Name
		state.Plan.Value = org.Plan
		state.Id.Value = org.Id

		// Set state
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

}

// Delete resource
func (r resourceOrganization) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state Organization
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete organization
	res, err := r.p.client.OrganizationMainCallsApi.DeleteOrganization(ctx, state.Id.Value).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting organization",
			"Could not delete organization "+state.Id.Value+", unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error deleting organization",
			"Could not delete organization, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
