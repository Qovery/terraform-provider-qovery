package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &containerRegistryResource{}
	_ resource.ResourceWithImportState = containerRegistryResource{}
)

var registryKinds = clientEnumToStringArray(registry.AllowedKindValues)

type containerRegistryResource struct {
	containerRegistryService registry.Service
}

func newContainerRegistryResource() resource.Resource {
	return &containerRegistryResource{}
}

func (r containerRegistryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_registry"
}

func (r *containerRegistryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.containerRegistryService = provider.containerRegistryService
}

func (r containerRegistryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Qovery container registry: an organization-wide connection to a registry that containers pull their images from.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("container registry"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription + recreatesOnChange("container registry"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("container registry"),
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(registryKindDescription, registryKinds, nil),
				Required:            true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(registryKinds),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: registryURLDescription,
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: storedDescriptionDescription("container registry"),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(storedDescriptionDefault),
			},
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: "Credentials of the container registry. The keys to set depend on `kind`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"access_key_id": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsAWSAccessKeyIDDescription, registry.KindECR, registry.KindPublicECR),
						Optional:            true,
					},
					"secret_access_key": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsAWSSecretAccessKeyDescription, registry.KindECR, registry.KindPublicECR),
						Optional:            true,
						Sensitive:           true,
					},
					"region": schema.StringAttribute{
						MarkdownDescription: usedByKinds(registryRegionDescription, registry.KindECR, registry.KindScalewayCR, registry.KindGcpArtifactRegistry),
						Optional:            true,
					},
					"scaleway_access_key": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsScalewayAccessKeyDescription, registry.KindScalewayCR),
						Optional:            true,
					},
					"scaleway_secret_key": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsScalewaySecretKeyDescription, registry.KindScalewayCR),
						Optional:            true,
						Sensitive:           true,
					},
					"scaleway_project_id": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsScalewayProjectIDDescription, registry.KindScalewayCR),
						Optional:            true,
					},
					"json_credentials": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsGCPJSONKeyDescription, registry.KindGcpArtifactRegistry) + " Omit it when `gcp_credentials_type` is set.",
						Optional:            true,
						Sensitive:           true,
					},
					"gcp_credentials_type": schema.StringAttribute{
						MarkdownDescription: "Set to `workload_identity_federation` to authenticate a `GCP_ARTIFACT_REGISTRY` registry through Workload Identity Federation instead of `json_credentials`. It requires `project_id`, `service_account_email` and `workload_identity_provider_resource`.",
						Optional:            true,
						Validators: []validator.String{
							validators.NewStringEnumValidator([]string{"workload_identity_federation"}),
						},
					},
					"project_id": schema.StringAttribute{
						MarkdownDescription: "ID of the GCP project." + registryWorkloadIdentityNote,
						Optional:            true,
					},
					"service_account_email": schema.StringAttribute{
						MarkdownDescription: credentialsGCPServiceAccountEmailDescription + registryWorkloadIdentityNote,
						Optional:            true,
					},
					"workload_identity_provider_resource": schema.StringAttribute{
						MarkdownDescription: credentialsGCPWorkloadIdentityProviderDescription + registryWorkloadIdentityNote,
						Optional:            true,
					},
					"token_lifetime_seconds": schema.Int64Attribute{
						MarkdownDescription: descriptions.NewInt64DefaultDescription("Lifetime of the tokens Workload Identity Federation issues, in seconds."+registryWorkloadIdentityNote, gcpTokenLifetimeSecondsDefault),
						Optional:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: usedByKinds(registryUsernameDescription, registry.KindDockerHub, registry.KindGithubCr, registry.KindGithubEnterpriseCr, registry.KindGitlabCr, registry.KindGenericCR),
						Optional:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: usedByKinds(registryPasswordDescription, registry.KindDockerHub, registry.KindGithubCr, registry.KindGithubEnterpriseCr, registry.KindGitlabCr, registry.KindGenericCR),
						Optional:            true,
						Sensitive:           true,
					},
				},
			},
		},
	}
}

// Create qovery container registry resource
func (r containerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan ContainerRegistry
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new container registry
	reg, err := r.containerRegistryService.Create(ctx, plan.OrganizationId.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on container registry create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainRegistryToContainerRegistry(plan, reg)
	tflog.Trace(ctx, "created container registry", map[string]any{"container_registry_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery container registry resource
func (r containerRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ContainerRegistry
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get container registry from the API
	reg, err := r.containerRegistryService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if handleDomainReadNotFound(ctx, resp, err, "Error on container registry read") {
		return
	}

	// Refresh state values
	state = convertDomainRegistryToContainerRegistry(state, reg)
	tflog.Trace(ctx, "read container registry", map[string]any{"container_registry_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery container registry resource
func (r containerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state ContainerRegistry
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update container registry in the backend
	reg, err := r.containerRegistryService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on container registry update", err.Error())
		return
	}

	// Update state values
	state = convertDomainRegistryToContainerRegistry(plan, reg)
	tflog.Trace(ctx, "updated container registry", map[string]any{"container_registry_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery container registry resource
func (r containerRegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state ContainerRegistry
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete container registry
	err := r.containerRegistryService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on container registry delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted container registry", map[string]any{"container_registry_id": state.Id.ValueString()})

	// Remove containerRegistry from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery container registry resource using its id
func (r containerRegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,container_registry_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
