package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &helmRepositoryResource{}
	_ resource.ResourceWithImportState = helmRepositoryResource{}
)

var helmRepositoryKinds = clientEnumToStringArray(helmRepository.AllowedKindValues)

type helmRepositoryResource struct {
	helmRepositoryService helmRepository.Service
}

func newHelmRepositoryResource() resource.Resource {
	return &helmRepositoryResource{}
}

func (r helmRepositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_helm_repository"
}

func (r *helmRepositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.helmRepositoryService = provider.helmRepositoryService
}

func (r helmRepositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Qovery helm repository: an organization-wide connection to a Helm chart repository, HTTPS or OCI, that `qovery_helm` services install their charts from.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("helm repository"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription + recreatesOnChange("helm repository"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("helm repository"),
				Required:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(
					helmRepositoryKindDescription+" Use `HTTPS` for a classic Helm repository or an `OCI_*` kind for an OCI registry.",
					helmRepositoryKinds,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(helmRepositoryKinds),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: helmRepositoryURLDescription,
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: storedDescriptionDescription("helm repository"),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(storedDescriptionDefault),
			},
			"skip_tls_verification": schema.BoolAttribute{
				MarkdownDescription: helmRepositorySkipTLSVerificationDescription + " Set it to `true` for a self-signed certificate.",
				Required:            true,
			},
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: "Credentials of the helm repository. The keys to set depend on `kind`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"access_key_id": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsAWSAccessKeyIDDescription, helmRepository.KindECR, helmRepository.KindPublicECR),
						Optional:            true,
					},
					"secret_access_key": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsAWSSecretAccessKeyDescription, helmRepository.KindECR, helmRepository.KindPublicECR),
						Optional:            true,
						Sensitive:           true,
					},
					"region": schema.StringAttribute{
						MarkdownDescription: usedByKinds(registryRegionDescription, helmRepository.KindECR, helmRepository.KindScalewayCR),
						Optional:            true,
					},
					"scaleway_access_key": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsScalewayAccessKeyDescription, helmRepository.KindScalewayCR),
						Optional:            true,
					},
					"scaleway_secret_key": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsScalewaySecretKeyDescription, helmRepository.KindScalewayCR),
						Optional:            true,
						Sensitive:           true,
					},
					"scaleway_project_id": schema.StringAttribute{
						MarkdownDescription: usedByKinds(credentialsScalewayProjectIDDescription, helmRepository.KindScalewayCR),
						Optional:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: usedByKinds(registryUsernameDescription, helmRepository.KindHttps, helmRepository.KindDockerHub, helmRepository.KindGithubCr, helmRepository.KindGitlabCr, helmRepository.KindGenericCR),
						Optional:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: usedByKinds(registryPasswordDescription, helmRepository.KindHttps, helmRepository.KindDockerHub, helmRepository.KindGithubCr, helmRepository.KindGitlabCr, helmRepository.KindGenericCR),
						Optional:            true,
						Sensitive:           true,
					},
				},
			},
		},
	}
}

// Create qovery helm repository resource
func (r helmRepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan HelmRepository
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new helm repository
	reg, err := r.helmRepositoryService.Create(ctx, plan.OrganizationId.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on helm repository create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainHelmRepositoryToHelmRepository(plan, reg)
	tflog.Trace(ctx, "created helm repository", map[string]any{"helm_repository_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery helm repository resource
func (r helmRepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state HelmRepository
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get helm repository from the API
	reg, err := r.helmRepositoryService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if handleDomainReadNotFound(ctx, resp, err, "Error on helm repository read") {
		return
	}

	// Refresh state values
	state = convertDomainHelmRepositoryToHelmRepository(state, reg)
	tflog.Trace(ctx, "read helm repository", map[string]any{"helm_repository_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery helm repository resource
func (r helmRepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state HelmRepository
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update helm repository in the backend
	reg, err := r.helmRepositoryService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on helm repository update", err.Error())
		return
	}

	// Update state values
	state = convertDomainHelmRepositoryToHelmRepository(plan, reg)
	tflog.Trace(ctx, "updated helm repository", map[string]any{"helm_repository_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery helm repository resource
func (r helmRepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state HelmRepository
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete helm repository
	err := r.helmRepositoryService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on helm repository delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted helm repository", map[string]any{"helm_repository_id": state.Id.ValueString()})

	// Remove helmRepository from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery helm repository resource using its id
func (r helmRepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,helm_repository_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
