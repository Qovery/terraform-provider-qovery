package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"strings"

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
var _ resource.ResourceWithConfigure = &containerRegistryResource{}
var _ resource.ResourceWithImportState = containerRegistryResource{}

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
		Description: "Provides a Qovery container registry resource. This can be used to create and manage Qovery container registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the container registry.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the container registry.",
				Required:    true,
			},
			"kind": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Kind of the container registry.",
					registryKinds,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(registryKinds),
				},
			},
			"url": schema.StringAttribute{
				Description: "URL of the container registry.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the container registry.",
				Optional:    true,
				Computed:    true,
			},
			"config": schema.SingleNestedAttribute{
				Description: "Configuration needed to authenticate the container registry.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"access_key_id": schema.StringAttribute{
						Description: "Required if kind is `ECR` or `PUBLIC_ECR`.",
						Optional:    true,
					},
					"secret_access_key": schema.StringAttribute{
						Description: "Required if kind is `ECR` or `PUBLIC_ECR`.",
						Optional:    true,
					},
					"region": schema.StringAttribute{
						Description: "Required if kind is `ECR` or `SCALEWAY_CR`.",
						Optional:    true,
					},
					"scaleway_access_key": schema.StringAttribute{
						Description: "Required if kind is `SCALEWAY_CR`.",
						Optional:    true,
					},
					"scaleway_secret_key": schema.StringAttribute{
						Description: "Required if kind is `SCALEWAY_CR`.",
						Optional:    true,
					},
					"username": schema.StringAttribute{
						Description: "Required if kinds are `DOCKER_HUB`, `GITHUB_CR`, `GITLAB`CR`, `GENERIC_CR`.",
						Optional:    true,
					},
					"password": schema.StringAttribute{
						Description: "Required if kinds are `DOCKER_HUB`, `GITHUB_CR`, `GITLAB`CR`, `GENERIC_CR`.",
						Optional:    true,
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
	tflog.Trace(ctx, "created container registry", map[string]interface{}{"container_registry_id": state.Id.ValueString()})

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
	if err != nil {
		resp.Diagnostics.AddError("Error on container registry read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainRegistryToContainerRegistry(state, reg)
	tflog.Trace(ctx, "read container registry", map[string]interface{}{"container_registry_id": state.Id.ValueString()})

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
	tflog.Trace(ctx, "updated container registry", map[string]interface{}{"container_registry_id": state.Id.ValueString()})

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

	tflog.Trace(ctx, "deleted container registry", map[string]interface{}{"container_registry_id": state.Id.ValueString()})

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
