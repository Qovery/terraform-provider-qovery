package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ provider.ResourceType = organizationResourceType{}
var _ resource.Resource = organizationResource{}
var _ resource.ResourceWithImportState = organizationResource{}

var organizationPlans = clientEnumToStringArray(organization.AllowedPlanValues)

type organizationResourceType struct{}

func (r organizationResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery organization resource. This can be used to create and manage Qovery organizations.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"plan": {
				Description: descriptions.NewStringEnumDescription(
					"Plan of the organization.",
					organizationPlans,
					nil,
				),
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(organizationPlans),
				},
			},
			"description": {
				Description: "Description of the organization.",
				Type:        types.StringType,
				Optional:    true,
			},
		},
	}, nil
}

func (r organizationResourceType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return organizationResource{
		organizationService: p.(*qProvider).organizationService,
	}, nil
}

type organizationResource struct {
	organizationService organization.Service
}

// Create qovery organization resource

func (r organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Error on organization create", "Organization creation is not allowed using terraform.")
}

// Read qovery organization resource
func (r organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Organization
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization from API
	orga, err := r.organizationService.Get(ctx, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on organization read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainOrganizationToTerraform(orga)
	tflog.Trace(ctx, "read organization", map[string]interface{}{"organization_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery organization resource
func (r organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Organization
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update organization in backend
	orga, err := r.organizationService.Update(ctx, state.Id.Value, plan.toOrganizationUpdateRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on organization update", err.Error())
		return
	}

	// Update state values
	state = convertDomainOrganizationToTerraform(orga)
	tflog.Trace(ctx, "updated organization", map[string]interface{}{"organization_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery organization resource
func (r organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Error on organization delete", "Organization deletion is not allowed using terraform.")
}

// ImportState imports a qovery organization resource using its id
func (r organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
