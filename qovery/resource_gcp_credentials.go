package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure        = &gcpCredentialsResource{}
	_ resource.ResourceWithImportState      = gcpCredentialsResource{}
	_ resource.ResourceWithConfigValidators = &gcpCredentialsResource{}
)

type gcpCredentialsResource struct {
	gcpCredentialsService credentials.GcpService
}

func newGcpCredentialsResource() resource.Resource {
	return &gcpCredentialsResource{}
}

func (r gcpCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_credentials"
}

func (r *gcpCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.gcpCredentialsService = provider.gcpCredentialsService
}

func (r gcpCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery GCP credentials resource. This can be used to create and manage Qovery GCP credentials. Supports both service account key and Workload Identity Federation authentication modes.",
		MarkdownDescription: "Provides a Qovery GCP credentials resource. This is used to create and manage GCP credentials that Qovery uses to provision and manage GKE clusters in your Google Cloud project.\n\n" +
			"Supports two authentication modes:\n" +
			"- **Service account key** (`gcp_credentials`): a GCP service account key in JSON format.\n" +
			"- **Workload Identity Federation** (`service_account_email` + `workload_identity_provider_resource`): keyless authentication via WIF.\n\n" +
			"Exactly one mode must be configured.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the GCP credentials.",
				MarkdownDescription: "Unique identifier of the GCP credentials (UUID format).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization. Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "ID of the Qovery organization in which to create the credentials. **Cannot be changed after creation** (forces resource replacement).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "Name of the GCP credentials.",
				MarkdownDescription: "Name of the GCP credentials. Used for display purposes in the Qovery console.",
				Required:            true,
			},
			"gcp_credentials": schema.StringAttribute{
				Description:         "Your GCP service account credentials JSON. Mutually exclusive with the Workload Identity Federation fields.",
				MarkdownDescription: "GCP service account key in JSON format. Mutually exclusive with `service_account_email`/`workload_identity_provider_resource`. This is a sensitive value and will not be displayed in plan output. Use `file()` to load from a file: `file(\"${path.module}/service-account.json\")`.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("service_account_email"),
						path.MatchRoot("workload_identity_provider_resource"),
					),
				},
			},
			"service_account_email": schema.StringAttribute{
				Description:         "GCP service account email to impersonate via Workload Identity Federation.",
				MarkdownDescription: "GCP service account email to impersonate (e.g. `qovery@my-project.iam.gserviceaccount.com`). Required together with `workload_identity_provider_resource` when using Workload Identity Federation. Mutually exclusive with `gcp_credentials`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("workload_identity_provider_resource")),
				},
			},
			"workload_identity_provider_resource": schema.StringAttribute{
				Description:         "Full GCP Workload Identity Provider resource.",
				MarkdownDescription: "Full Workload Identity Provider resource path (e.g. `projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider`). Required together with `service_account_email`. Mutually exclusive with `gcp_credentials`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("service_account_email")),
				},
			},
		},
	}
}

// ConfigValidators enforces that exactly one authentication mode is configured:
// either gcp_credentials (service account key) or the Workload Identity Federation fields.
func (r gcpCredentialsResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("gcp_credentials"),
			path.MatchRoot("service_account_email"),
		),
	}
}

// Create qovery gcp credentials resource.
func (r gcpCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan GCPCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new credentials
	creds, err := r.gcpCredentialsService.Create(ctx, plan.OrganizationId.ValueString(), plan.toUpsertGcpRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on gcp credentials create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainCredentialsToGCPCredentials(creds, plan)
	tflog.Trace(ctx, "created gcp credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery gcp credentials resource.
func (r gcpCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state GCPCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API
	creds, err := r.gcpCredentialsService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on gcp credentials read", err.Error())
		return
	}

	state = convertDomainCredentialsToGCPCredentials(creds, state)
	tflog.Trace(ctx, "read gcp credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery gcp credentials resource.
func (r gcpCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state GCPCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update credentials in the backend
	creds, err := r.gcpCredentialsService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), plan.toUpsertGcpRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on gcp credentials update", err.Error())
		return
	}

	// Update state values
	state = convertDomainCredentialsToGCPCredentials(creds, plan)
	tflog.Trace(ctx, "updated gcp credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery gcp credentials resource.
func (r gcpCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state GCPCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete credentials in the backend
	err := r.gcpCredentialsService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on gcp credentials delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted gcp credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Remove credentials from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery gcp credentials resource using its id.
func (r gcpCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,gcp_credentials_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
