package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.Resource = scalewayCredentialsResource{}
var _ resource.ResourceWithImportState = scalewayCredentialsResource{}

type scalewayCredentialsResource struct {
	scalewayCredentialsService credentials.ScalewayService
}

func NewScalewayCredentialsResource(service credentials.ScalewayService) func() resource.Resource {
	return func() resource.Resource {
		return scalewayCredentialsResource{
			scalewayCredentialsService: service,
		}
	}
}

func (r scalewayCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scaleway_credentials"
}

func (r scalewayCredentialsResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery SCALEWAY credentials resource. This can be used to create and manage Qovery SCALEWAY credentials.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the SCALEWAY credentials.",
				Type:        types.StringType,
				Computed:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the scaleway credentials.",
				Type:        types.StringType,
				Required:    true,
			},
			"scaleway_access_key": {
				Description: "Your SCALEWAY access key id.",
				Type:        types.StringType,
				Required:    true,
				Sensitive:   true,
			},
			"scaleway_secret_key": {
				Description: "Your SCALEWAY secret key.",
				Type:        types.StringType,
				Required:    true,
				Sensitive:   true,
			},
			"scaleway_project_id": {
				Description: "Your SCALEWAY project ID.",
				Type:        types.StringType,
				Required:    true,
				Sensitive:   true,
			},
		},
	}, nil
}

// Create qovery scaleway credentials resource
func (r scalewayCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan ScalewayCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new credentials
	creds, err := r.scalewayCredentialsService.Create(ctx, plan.OrganizationId.Value, plan.toUpsertScalewayRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainCredentialsToScalewayCredentials(creds, plan)
	tflog.Trace(ctx, "created scaleway credentials", map[string]interface{}{"credentials_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery scaleway credentials resource
func (r scalewayCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ScalewayCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API
	creds, err := r.scalewayCredentialsService.Get(ctx, state.OrganizationId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials read", err.Error())
		return
	}

	state = convertDomainCredentialsToScalewayCredentials(creds, state)
	tflog.Trace(ctx, "read scaleway credentials", map[string]interface{}{"credentials_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery scaleway credentials resource
func (r scalewayCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state ScalewayCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update credentials in the backend
	creds, err := r.scalewayCredentialsService.Update(ctx, state.OrganizationId.Value, state.Id.Value, plan.toUpsertScalewayRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials update", err.Error())
		return
	}

	// Update state values
	state = convertDomainCredentialsToScalewayCredentials(creds, plan)
	tflog.Trace(ctx, "updated scaleway credentials", map[string]interface{}{"credentials_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery scaleway credentials resource
func (r scalewayCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state ScalewayCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete credentials in the backend
	err := r.scalewayCredentialsService.Delete(ctx, state.OrganizationId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted scaleway credentials", map[string]interface{}{"credentials_id": state.Id.Value})

	// Remove credentials from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery scaleway credentials resource using its id
func (r scalewayCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: scaleway_credentials_id,organization_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
