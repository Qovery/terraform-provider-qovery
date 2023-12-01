package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &scalewayCredentialsResource{}
var _ resource.ResourceWithImportState = scalewayCredentialsResource{}

type scalewayCredentialsResource struct {
	scalewayCredentialsService credentials.ScalewayService
}

func newScalewayCredentialsResource() resource.Resource {
	return &scalewayCredentialsResource{}
}

func (r scalewayCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scaleway_credentials"
}

func (r *scalewayCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.scalewayCredentialsService = provider.scalewayCredentialsService
}

func (r scalewayCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery SCALEWAY credentials resource. This can be used to create and manage Qovery SCALEWAY credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the SCALEWAY credentials.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the scaleway credentials.",
				Required:    true,
			},
			"scaleway_access_key": schema.StringAttribute{
				Description: "Your SCALEWAY access key id.",
				Required:    true,
				Sensitive:   false,
			},
			"scaleway_secret_key": schema.StringAttribute{
				Description: "Your SCALEWAY secret key.",
				Required:    true,
				Sensitive:   true,
			},
			"scaleway_project_id": schema.StringAttribute{
				Description: "Your SCALEWAY project ID.",
				Required:    true,
				Sensitive:   false,
			},
			"scaleway_organization_id": schema.StringAttribute{
				Description: "Your SCALEWAY organization ID.",
				Required:    true,
				Sensitive:   false,
			},
		},
	}
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
	creds, err := r.scalewayCredentialsService.Create(ctx, plan.OrganizationId.ValueString(), plan.toUpsertScalewayRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainCredentialsToScalewayCredentials(creds, plan)
	tflog.Trace(ctx, "created scaleway credentials", map[string]interface{}{"credentials_id": state.Id.ValueString()})

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
	creds, err := r.scalewayCredentialsService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials read", err.Error())
		return
	}

	state = convertDomainCredentialsToScalewayCredentials(creds, state)
	tflog.Trace(ctx, "read scaleway credentials", map[string]interface{}{"credentials_id": state.Id.ValueString()})

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
	creds, err := r.scalewayCredentialsService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), plan.toUpsertScalewayRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials update", err.Error())
		return
	}

	// Update state values
	state = convertDomainCredentialsToScalewayCredentials(creds, plan)
	tflog.Trace(ctx, "updated scaleway credentials", map[string]interface{}{"credentials_id": state.Id.ValueString()})

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
	err := r.scalewayCredentialsService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted scaleway credentials", map[string]interface{}{"credentials_id": state.Id.ValueString()})

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
