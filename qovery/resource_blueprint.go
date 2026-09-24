package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

var (
	_ resource.ResourceWithConfigure  = &blueprintResource{}
	_ resource.ResourceWithModifyPlan = blueprintResource{}
)

// Private state key set while saved settings may not be deployed yet
const blueprintPendingApplyKey = "pending_apply"

type blueprintResource struct {
	service blueprint.Service
}

func newBlueprintResource() resource.Resource {
	return &blueprintResource{}
}

func (r blueprintResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (r *blueprintResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T.", req.ProviderData),
		)
		return
	}
	r.service = provider.blueprintService
}

func (r blueprintResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery blueprint resource: a service instantiated from the Qovery service catalog (e.g. a managed database). " +
			"Qovery materializes the blueprint as a terraform or helm service, exposed as `service_id`. Every update is saved then applied, which redeploys that service. " +
			"The API does not return `icon_uri`, `spec_overrides` nor secret values, so changes made to them outside Terraform are not detected.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the blueprint.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the blueprint service.",
				Required:    true,
			},
			"blueprint": schema.StringAttribute{
				Description: "Catalog entry to instantiate, as `<provider>/<service_family>/<service_version>`, e.g. `AWS/postgres/17`. " +
					"Changing the service version upgrades the service in place; changing the provider or the service family replaces it.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					requiresReplaceIfOtherBlueprintService(),
				},
			},
			"tag": schema.StringAttribute{
				Description: "Catalog tag deployed, e.g. `AWS/postgres/17/4.1.0`. Always the latest release of `blueprint`: when the catalog publishes a new one, the next apply upgrades the service to it.",
				Computed:    true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI of the blueprint service.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(defaultBlueprintIconURI),
			},
			"variables": schema.MapAttribute{
				Description: "Blueprint variables, keyed by name. Variables left out keep their catalog default.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"secret_variables": schema.MapAttribute{
				Description: "Secret blueprint variables, keyed by name. The API never returns their values.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"spec_overrides": schema.SingleNestedAttribute{
				Description: "Overrides of the engine settings of the blueprint manifest.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"engine_version": schema.StringAttribute{
						Description: "Terraform or OpenTofu version of the apply job. Must be one of the versions the manifest allows.",
						Optional:    true,
					},
					"credentials": schema.StringAttribute{
						Description: "How the apply job authenticates to the cloud provider: `cluster` reuses the cluster credentials, `env` expects credentials as environment variables.",
						Optional:    true,
					},
					"backend": schema.StringAttribute{
						Description: "Where the Terraform state is stored: `qovery` or `user_provided`.",
						Optional:    true,
					},
					"timeout": schema.Int64Attribute{
						Description: "Maximum duration in seconds of an apply job.",
						Optional:    true,
					},
					"cpu": schema.StringAttribute{
						Description: "CPU of the apply job pod, e.g. `500m`.",
						Optional:    true,
					},
					"ram": schema.StringAttribute{
						Description: "Memory of the apply job pod, e.g. `512Mi`.",
						Optional:    true,
					},
					"storage": schema.StringAttribute{
						Description: "Ephemeral storage of the apply job pod, e.g. `1Gi`.",
						Optional:    true,
					},
				},
			},
			"deploy": schema.BoolAttribute{
				Description: "Whether to deploy the service on creation. Updates always deploy it.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"service_id": schema.StringAttribute{
				Description: "Id of the terraform or helm service the blueprint materialized.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_type": schema.StringAttribute{
				Description: "Type of the service the blueprint materialized: `TERRAFORM` or `HELM`.",
				Computed:    true,
			},
			"catalog_url": schema.StringAttribute{
				Description: "URL of the blueprint catalog entry.",
				Computed:    true,
			},
		},
	}
}

func (r blueprintResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.resolveUnknownTag(ctx, &plan)...)
	request, diags := plan.toCreateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bp, err := r.service.Create(ctx, ToString(plan.EnvironmentID), request)
	if err != nil {
		resp.Diagnostics.AddError("Error on blueprint create", err.Error())
		if bp == nil {
			return
		}
		// Blueprint exists, only dispatch or deploy failed: track it so it gets tainted, not orphaned
	}

	state, diags := convertDomainBlueprintToBlueprint(ctx, bp, plan, false)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r blueprintResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bp, err := r.service.Get(ctx, ToString(state.ID))
	if handleDomainReadNotFound(ctx, resp, err, "Error on blueprint read") {
		return
	}

	pending, diags := isPendingApply(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	newState, diags := convertDomainBlueprintToBlueprint(ctx, bp, state, pending)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r blueprintResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pending, diags := isPendingApply(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	// deploy only matters on create: changing it alone must not redeploy the service
	if !pending && plan.configEqualIgnoringDeploy(state) {
		state.Deploy = plan.Deploy
		state.Blueprint = plan.Blueprint
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	resp.Diagnostics.Append(r.resolveUnknownTag(ctx, &plan)...)
	request, diags := plan.toUpdateRequest(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bp, err := r.service.Update(ctx, ToString(state.ID), request)
	if err != nil {
		resp.Diagnostics.AddError("Error on blueprint update", err.Error())
		// Settings may be saved but not deployed: keep last applied values and flag a retry
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, blueprintPendingApplyKey, []byte("true"))...)
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, blueprintPendingApplyKey, nil)...)

	newState, diags := convertDomainBlueprintToBlueprint(ctx, bp, plan, false)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r blueprintResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.service.Delete(ctx, ToString(state.ID)); err != nil {
		resp.Diagnostics.AddError("Error on blueprint delete", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r blueprintResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.service == nil {
		return
	}
	// Only these attributes: a whole-plan decode fails on an unknown spec_overrides object
	var planned blueprintVersionAttributes
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("environment_id"), &planned.EnvironmentID)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("blueprint"), &planned.Blueprint)...)
	if resp.Diagnostics.HasError() || planned.EnvironmentID.IsUnknown() || planned.Blueprint.IsUnknown() {
		return
	}

	var current *blueprintVersionAttributes
	if !req.State.Raw.IsNull() {
		current = &blueprintVersionAttributes{}
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("environment_id"), &current.EnvironmentID)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("blueprint"), &current.Blueprint)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("tag"), &current.Tag)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	pending, diags := isPendingApply(ctx, req.Private)
	resp.Diagnostics.Append(diags...)

	tag, err := r.latestTag(ctx, planned.EnvironmentID, planned.Blueprint)
	if err != nil {
		// Updates always target the latest tag; only a plan that deploys nothing may keep the deployed one
		if !canKeepDeployedTag(current, planned, pending) {
			resp.Diagnostics.AddAttributeError(path.Root("blueprint"), "Error on blueprint latest tag resolution", err.Error())
			return
		}
		changed, diags := blueprintSettingsChanged(ctx, req.Plan, req.State)
		resp.Diagnostics.Append(diags...)
		if changed {
			resp.Diagnostics.AddAttributeError(path.Root("blueprint"), "Error on blueprint latest tag resolution", err.Error())
			return
		}
		resp.Diagnostics.AddAttributeWarning(path.Root("blueprint"), "Cannot resolve blueprint latest tag, keeping the deployed one", err.Error())
		tag = ToString(current.Tag)
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("tag"), tag)...)

	if pending && current != nil {
		// Unknown computed value makes the plan non-empty, so Update retries the failed apply
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("catalog_url"), types.StringUnknown())...)
	}
}

// A pending retry deploys, so it must resolve the latest tag like any update
func canKeepDeployedTag(current *blueprintVersionAttributes, planned blueprintVersionAttributes, pending bool) bool {
	return current != nil && !pending && current.sameVersion(planned)
}

// blueprintSettingsChanged reports a change Update would send; deploy alone sends nothing
func blueprintSettingsChanged(ctx context.Context, plan tfsdk.Plan, state tfsdk.State) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	changes := []func() (bool, diag.Diagnostics){
		func() (bool, diag.Diagnostics) {
			return attributeChanged[types.String](ctx, plan, state, path.Root("name"))
		},
		func() (bool, diag.Diagnostics) {
			return attributeChanged[types.String](ctx, plan, state, path.Root("icon_uri"))
		},
		func() (bool, diag.Diagnostics) {
			return attributeChanged[types.Map](ctx, plan, state, path.Root("variables"))
		},
		func() (bool, diag.Diagnostics) {
			return attributeChanged[types.Map](ctx, plan, state, path.Root("secret_variables"))
		},
		func() (bool, diag.Diagnostics) {
			return attributeChanged[types.Object](ctx, plan, state, path.Root("spec_overrides"))
		},
	}
	for _, change := range changes {
		changed, d := change()
		diags.Append(d...)
		if changed || diags.HasError() {
			return true, diags
		}
	}
	return false, diags
}

func attributeChanged[T attr.Value](ctx context.Context, plan tfsdk.Plan, state tfsdk.State, attributePath path.Path) (bool, diag.Diagnostics) {
	var planned, current T
	diags := plan.GetAttribute(ctx, attributePath, &planned)
	diags.Append(state.GetAttribute(ctx, attributePath, &current)...)
	return !planned.Equal(current), diags
}

type blueprintVersionAttributes struct {
	EnvironmentID types.String
	Blueprint     types.String
	Tag           types.String
}

func (a blueprintVersionAttributes) sameVersion(planned blueprintVersionAttributes) bool {
	return a.EnvironmentID.Equal(planned.EnvironmentID) &&
		strings.EqualFold(a.Blueprint.ValueString(), planned.Blueprint.ValueString()) &&
		!a.Tag.IsNull()
}

type privateStateReader interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

func isPendingApply(ctx context.Context, private privateStateReader) (bool, diag.Diagnostics) {
	value, diags := private.GetKey(ctx, blueprintPendingApplyKey)
	return string(value) == "true", diags
}

// resolveUnknownTag covers plans made while environment_id was still unknown
func (r blueprintResource) resolveUnknownTag(ctx context.Context, plan *Blueprint) diag.Diagnostics {
	var diags diag.Diagnostics
	if !plan.Tag.IsUnknown() && !plan.Tag.IsNull() {
		return diags
	}
	tag, err := r.latestTag(ctx, plan.EnvironmentID, plan.Blueprint)
	if err != nil {
		diags.AddAttributeError(path.Root("blueprint"), "Error on blueprint latest tag resolution", err.Error())
		return diags
	}
	plan.Tag = FromString(tag)
	return diags
}

func (r blueprintResource) latestTag(ctx context.Context, environmentID types.String, catalogVersion types.String) (string, error) {
	version, err := blueprint.ParseCatalogVersion(ToString(catalogVersion))
	if err != nil {
		return "", err
	}
	return r.service.ResolveLatestTag(ctx, ToString(environmentID), version)
}

func requiresReplaceIfOtherBlueprintService() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			if req.PlanValue.IsUnknown() || req.StateValue.IsNull() {
				return
			}
			planned, plannedErr := blueprint.ParseCatalogVersion(req.PlanValue.ValueString())
			current, currentErr := blueprint.ParseCatalogVersion(req.StateValue.ValueString())
			resp.RequiresReplace = plannedErr != nil || currentErr != nil || !planned.SameService(current)
		},
		"Changing the provider or the service family replaces the blueprint; changing the service version upgrades it in place.",
		"Changing the provider or the service family replaces the blueprint; changing the service version upgrades it in place.",
	)
}
