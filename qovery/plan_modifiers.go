package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// smartAllowApiOverrideModifier is a context-aware plan modifier that intelligently handles
// API-determined values while maintaining backward compatibility with explicitly configured values.
//
// ⚠️ WORKAROUND FOR API LIMITATION ⚠️
// This implementation exists because the Qovery API enforces a business rule:
// "At least one port must have is_default=true"
//
// When users explicitly set is_default=false, the API overrides it to true, causing
// Terraform validation errors: "Provider produced inconsistent result after apply"
//
// This modifier provides the best available solution until the API is updated.
//
// Behavior:
// - If user explicitly sets value in config → uses that value (backward compatible)
// - If user omits value but state exists → uses state value (prevents drift)
// - If user omits value and no state exists → sets to unknown (API determines value)
//
// This solves the problem where:
// 1. Terraform rejects setting an explicit config value to unknown
// 2. We want to allow API to determine value when user omits it
// 3. We need to prevent perpetual drift on refresh
//
// Known Limitation (API-side, cannot fix in provider):
// - Users CANNOT set is_default=false (API will override to true)
// - Workaround: Set is_default=true explicitly for default port, omit for others
//
// TODO: This workaround should remain until the Qovery API is updated to allow
// users to set is_default=false without the API overriding it. Once the API
// respects the user's false values, this modifier can potentially be simplified
// or removed in favor of standard Terraform plan modifiers.
type smartAllowApiOverrideModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m smartAllowApiOverrideModifier) Description(_ context.Context) string {
	return "If config value is explicitly set, uses that value. If omitted, uses state value if available, otherwise allows API to determine the value."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m smartAllowApiOverrideModifier) MarkdownDescription(_ context.Context) string {
	return "If config value is explicitly set, uses that value. If omitted, uses state value if available, otherwise allows API to determine the value."
}

// PlanModifyBool implements the plan modification logic with context-aware behavior.
func (m smartAllowApiOverrideModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// If config is explicitly set, use it (maintains backward compatibility)
	if !req.ConfigValue.IsNull() {
		resp.PlanValue = req.ConfigValue
		return
	}

	// If config is null but state has a value, use state (prevents perpetual drift)
	if !req.StateValue.IsNull() {
		resp.PlanValue = req.StateValue
		return
	}

	// If both config and state are null (initial create with omitted field),
	// set to unknown so API computes it
	resp.PlanValue = types.BoolUnknown()
}

// SmartAllowApiOverride returns a context-aware plan modifier that intelligently handles
// API-determined values while maintaining backward compatibility.
func SmartAllowApiOverride() planmodifier.Bool {
	return smartAllowApiOverrideModifier{}
}

// useUnknownForNullStringModifier sets the plan value to unknown when both config
// and state are null. This handles the case where a new element is added to a list
// during update — Computed attributes on the new element have no state, so they
// would otherwise be planned as null, causing "inconsistent result after apply"
// when the API assigns a value.
type useUnknownForNullStringModifier struct{}

func (m useUnknownForNullStringModifier) Description(_ context.Context) string {
	return "Sets value to unknown when both config and state are null, allowing the API to compute it."
}

func (m useUnknownForNullStringModifier) MarkdownDescription(_ context.Context) string {
	return "Sets value to unknown when both config and state are null, allowing the API to compute it."
}

func (m useUnknownForNullStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() && req.StateValue.IsNull() {
		resp.PlanValue = types.StringUnknown()
	}
}

// UseUnknownForNullString returns a plan modifier that converts null to unknown
// when a Computed attribute has no config value and no prior state (new list element).
func UseUnknownForNullString() planmodifier.String {
	return useUnknownForNullStringModifier{}
}

// useStateUnlessPortsChangeModifier preserves the prior state value for a Computed
// string attribute unless the resource's "ports" attribute is changing. When ports
// change (e.g. adding a public port), attributes like external_host may be assigned
// or removed by the API, so they must be recomputed.
type useStateUnlessPortsChangeModifier struct{}

func (m useStateUnlessPortsChangeModifier) Description(_ context.Context) string {
	return "Uses state value unless ports are changing, in which case the value is recomputed."
}

func (m useStateUnlessPortsChangeModifier) MarkdownDescription(_ context.Context) string {
	return "Uses state value unless ports are changing, in which case the value is recomputed."
}

func (m useStateUnlessPortsChangeModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// No state means create — leave as unknown so the API computes it.
	if req.State.Raw.IsNull() {
		return
	}

	// If plan value is already known (not unknown), nothing to do.
	if !req.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	// Check if ports changed
	var statePorts, planPorts types.List
	stateDiags := req.State.GetAttribute(ctx, path.Root("ports"), &statePorts)
	planDiags := req.Plan.GetAttribute(ctx, path.Root("ports"), &planPorts)
	if !stateDiags.HasError() && !planDiags.HasError() && !statePorts.Equal(planPorts) {
		// Ports changed — leave as unknown so API recomputes
		resp.PlanValue = types.StringUnknown()
		return
	}

	// Ports unchanged — preserve state value
	resp.PlanValue = req.StateValue
}

// UseStateUnlessPortsChange returns a plan modifier for Computed string attributes
// that preserves state unless ports configuration changes.
func UseStateUnlessPortsChange() planmodifier.String {
	return useStateUnlessPortsChangeModifier{}
}

// useStateUnlessNameChangesModifier preserves the prior state value for a computed
// list attribute (built_in_environment_variables) unless any attribute that the API
// embeds into those values is changing.
//
// Built-in env vars are derived server-side from a fixed set of inputs:
//
//	QOVERY_SERVICE_NAME, QOVERY_PROJECT_NAME, …      ← name
//	QOVERY_KUBERNETES_CLUSTER_VPC_ID, …              ← ports
//	QOVERY_BUILD_ID, QOVERY_APPLICATION_..._GIT_..., ← source / git_repository / image_*
//
// If any of those inputs change, the API will recompute the env var values during
// apply. Reusing the prior state value would then cause Terraform to error with
// "Provider produced inconsistent result after apply" (the state we returned does
// not match the state we promised in the plan).
//
// To avoid that, this modifier invalidates the cached state value whenever any of
// the inputs the API derives from is changing. The list of inputs is enumerated
// explicitly below; absent attributes are skipped gracefully (different service
// resources have different shapes — application has top-level git_repository,
// container has top-level image_name/tag/registry_id, helm/job have a `source`
// nested object).
//
// Adding a new attribute that influences built-in env vars: extend the list in
// PlanModifyList AND add coverage in qovery/resource_*_apply_consistency_test.go.
type useStateUnlessNameChangesModifier struct{}

func (m useStateUnlessNameChangesModifier) Description(_ context.Context) string {
	return "Uses state value unless the resource name or ports are changing, in which case the value is recomputed."
}

func (m useStateUnlessNameChangesModifier) MarkdownDescription(_ context.Context) string {
	return "Uses state value unless the resource name or ports are changing, in which case the value is recomputed."
}

func (m useStateUnlessNameChangesModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// No state means create — leave as unknown so the API computes it.
	if req.State.Raw.IsNull() {
		return
	}

	var stateName, planName types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &stateName)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("name"), &planName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Name is changing — built-in env var values will change, so leave as unknown.
	if !stateName.Equal(planName) {
		return
	}

	// Ports changing — adding/removing public ports creates/removes built-in env vars
	// (e.g. QOVERY_KUBERNETES_CLUSTER_VPC_ID), so leave as unknown to recompute.
	// `ports` is a List on application/container and a Map on helm — try both.
	if listAttrChanged(ctx, req, "ports") || mapAttrChanged(ctx, req, "ports") {
		return
	}

	// `source` (single nested object) on helm and job. Image tag / registry / docker
	// changes are embedded in QOVERY_BUILD_ID and friends — recompute.
	if objectAttrChanged(ctx, req, "source") {
		return
	}

	// `git_repository` (single nested object) on application — branch/url/root_path
	// flow into QOVERY_APPLICATION_*_GIT_* env vars.
	if objectAttrChanged(ctx, req, "git_repository") {
		return
	}

	// On container, the image source attributes are at the top level rather than
	// inside a `source` block. These are the API's inputs to QOVERY_BUILD_ID.
	for _, name := range []string{"image_name", "tag", "registry_id"} {
		if stringAttrChanged(ctx, req, name) {
			return
		}
	}

	// Nothing relevant changed — safe to reuse state value.
	resp.PlanValue = req.StateValue
}

// listAttrChanged reports whether the named root attribute is present on both
// state and plan AND its values differ. Returns false when the attribute is
// absent from the schema (e.g. the modifier is attached to a resource that
// doesn't have this attribute) — that's the graceful "not applicable" path.
func listAttrChanged(ctx context.Context, req planmodifier.ListRequest, name string) bool {
	var sv, pv types.List
	sd := req.State.GetAttribute(ctx, path.Root(name), &sv)
	pd := req.Plan.GetAttribute(ctx, path.Root(name), &pv)
	return !sd.HasError() && !pd.HasError() && !sv.Equal(pv)
}

func mapAttrChanged(ctx context.Context, req planmodifier.ListRequest, name string) bool {
	var sv, pv types.Map
	sd := req.State.GetAttribute(ctx, path.Root(name), &sv)
	pd := req.Plan.GetAttribute(ctx, path.Root(name), &pv)
	return !sd.HasError() && !pd.HasError() && !sv.Equal(pv)
}

func objectAttrChanged(ctx context.Context, req planmodifier.ListRequest, name string) bool {
	var sv, pv types.Object
	sd := req.State.GetAttribute(ctx, path.Root(name), &sv)
	pd := req.Plan.GetAttribute(ctx, path.Root(name), &pv)
	return !sd.HasError() && !pd.HasError() && !sv.Equal(pv)
}

func stringAttrChanged(ctx context.Context, req planmodifier.ListRequest, name string) bool {
	var sv, pv types.String
	sd := req.State.GetAttribute(ctx, path.Root(name), &sv)
	pd := req.Plan.GetAttribute(ctx, path.Root(name), &pv)
	return !sd.HasError() && !pd.HasError() && !sv.Equal(pv)
}

// UseStateUnlessNameChanges returns a plan modifier for list attributes that preserves
// the state value when the resource is being updated for reasons that don't affect
// built-in environment variables. See useStateUnlessNameChangesModifier for the full
// list of inputs that invalidate the cached value.
func UseStateUnlessNameChanges() planmodifier.List {
	return useStateUnlessNameChangesModifier{}
}
