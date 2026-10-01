package qovery

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

// useStateUnlessNameChangesModifier preserves the prior state value for
// built_in_environment_variables unless an attribute the API embeds into those
// values changes. Reusing state across such a change causes "Provider produced
// inconsistent result after apply" because the API recomputes the env var list
// during apply and the post-apply value diverges from the planned value.
//
// Triggers: name, mode, ports, source, git_repository, values_override,
// schedule, image_name, tag, registry_id. The mapping from each trigger to the
// affected QOVERY_* env vars lives in q-core's
// `core/variable/domain/VariableDomain.kt`. `mode` flows into ENVIRONMENT_TYPE
// for env-scope built-ins (resolved at read time from `env.type` in
// `VariableService.getVariablesReplacementForEnvironment`).
//
// Absent attributes are skipped gracefully so the same modifier can serve
// resources with different schemas (application has top-level git_repository,
// container has top-level image_name/tag/registry_id, helm has values_override
// + source, job has schedule + source, environment has mode).
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

	// `ports` is a List on application/container and a Map on helm — try both shapes.
	if attrChanged[types.List](ctx, req, "ports") || attrChanged[types.Map](ctx, req, "ports") {
		return
	}

	// `source` is on helm and job; `git_repository` is on application;
	// `values_override` is on helm (commit hash of values file flows into
	// QOVERY_HELM_VALUE_COMMIT_ID); `schedule` is on job (its shape flows
	// into QOVERY_JOB_ACTION).
	for _, name := range []string{"source", "git_repository", "values_override", "schedule"} {
		if objectAttrChanged(ctx, req, name) {
			return
		}
	}

	// On container, image source attributes are at the top level (no `source` block).
	// `mode` is on environment and flows into the ENVIRONMENT_TYPE built-in.
	for _, name := range []string{"image_name", "tag", "registry_id", "mode"} {
		if attrChanged[types.String](ctx, req, name) {
			return
		}
	}

	resp.PlanValue = req.StateValue
}

// objectAttrChanged reports whether the named Object attribute has a *user-visible*
// change between state and plan. Sub-attributes that are unknown in the plan are
// ignored, because terraform-plugin-framework does not guarantee plan-modifier
// execution order across sibling attributes: a nested Computed attribute (e.g.
// schedule.lifecycle_type with stringplanmodifier.UseStateForUnknown()) may still
// be unknown when this list modifier runs, even though it will be frozen to its
// state value before the plan is finalized. Comparing the raw Object would
// otherwise produce false positives whenever a sibling attribute changes.
func objectAttrChanged(ctx context.Context, req planmodifier.ListRequest, name string) bool {
	var sv, pv types.Object
	sd := req.State.GetAttribute(ctx, path.Root(name), &sv)
	pd := req.Plan.GetAttribute(ctx, path.Root(name), &pv)
	if sd.HasError() || pd.HasError() {
		return false
	}
	if sv.IsNull() != pv.IsNull() {
		return true
	}
	if sv.IsNull() {
		return false
	}

	stateAttrs := sv.Attributes()
	planAttrs := pv.Attributes()
	for k, stateVal := range stateAttrs {
		planVal, ok := planAttrs[k]
		if !ok {
			continue
		}
		// Ignore unknown plan values: a sibling/nested plan modifier (e.g.
		// UseStateForUnknown) will freeze them to state before the final plan,
		// so they are not user-driven changes.
		if planVal.IsUnknown() {
			continue
		}
		if !stateVal.Equal(planVal) {
			return true
		}
	}
	return false
}

// attrChanged reports whether the named root attribute differs between state and
// plan. Returns false when the attribute is absent from the schema — the graceful
// "not applicable" path that lets one modifier serve resources with different shapes.
func attrChanged[T attr.Value](ctx context.Context, req planmodifier.ListRequest, name string) bool {
	var sv, pv T
	sd := req.State.GetAttribute(ctx, path.Root(name), &sv)
	pd := req.Plan.GetAttribute(ctx, path.Root(name), &pv)
	return !sd.HasError() && !pd.HasError() && !sv.Equal(pv)
}

// UseStateUnlessNameChanges returns a plan modifier for built_in_environment_variables
// that preserves state across updates that don't change any input the API embeds in
// the env var values. See useStateUnlessNameChangesModifier for the trigger list.
func UseStateUnlessNameChanges() planmodifier.List {
	return useStateUnlessNameChangesModifier{}
}

// RequiresReplaceIfKnownChange triggers resource replacement only when the planned
// value is known AND differs from state. When the planned value is unknown,
// replacement is suppressed.
//
// Use this in place of stringplanmodifier.RequiresReplace() on attributes that may
// be sourced from data sources (e.g. environment_id, cluster_id).
func RequiresReplaceIfKnownChange() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		requiresReplaceIfKnownChangeFunc,
		"If the value changes to a different known value, Terraform will destroy and recreate the resource. Replacement is skipped when the planned value is unknown.",
		"If the value changes to a different known value, Terraform will destroy and recreate the resource. Replacement is **skipped when the planned value is unknown** (e.g., a deferred data source during a `-target` apply).",
	)
}

// requiresReplaceIfKnownChangeFunc is the predicate evaluated by
// RequiresReplaceIfKnownChange after the framework's gates (resource create/destroy,
// plan==state) have passed.
func requiresReplaceIfKnownChangeFunc(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.PlanValue.IsUnknown() {
		return
	}
	resp.RequiresReplace = true
}

const rejectExistingVpcChangeDescription = "Rejects any change to the existing VPC configuration after creation: adding or removing the block, or changing any of its attributes."

// rejectExistingVpcChangeModifier rejects changes to an immutable existing-VPC
// block once the cluster exists, without forcing replacement. It guards the
// whole block at the object level:
//
//   - presence: adding or removing the block after creation is rejected. This
//     must be caught on the object because the framework skips the children's
//     plan modifiers entirely when the planned block is null (fwserver "null
//     and unknown values should not have nested schema to modify").
//   - content: any known child value change is rejected, so attributes added
//     to the block later are immutable by default with no per-attribute
//     wiring. Unknown children (deferred data sources, unresolved Computed
//     values) are skipped. Null is normalized per child type — null≡empty for
//     lists and strings, null≡false for bools — because the API does not
//     distinguish those pairs.
type rejectExistingVpcChangeModifier struct{}

func (m rejectExistingVpcChangeModifier) Description(_ context.Context) string {
	return rejectExistingVpcChangeDescription
}

func (m rejectExistingVpcChangeModifier) MarkdownDescription(_ context.Context) string {
	return rejectExistingVpcChangeDescription
}

func (m rejectExistingVpcChangeModifier) PlanModifyObject(_ context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Resource creation: no prior value exists, nothing is immutable yet.
	if req.State.Raw.IsNull() {
		return
	}

	// Resource destroy: values are planned to null, which must not be mistaken
	// for a removal. Unreachable on framework v1.19.0 (destroy plans skip plan
	// modification entirely) but guarded per the framework's own
	// RequiresReplaceIf convention in case that changes.
	if req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() {
		return
	}

	if req.StateValue.IsNull() != req.PlanValue.IsNull() {
		addExistingVpcImmutableError(req.Path, &resp.Diagnostics)
		return
	}
	if req.StateValue.IsNull() {
		return
	}

	stateAttrs := req.StateValue.Attributes()
	planAttrs := req.PlanValue.Attributes()
	names := make([]string, 0, len(planAttrs))
	for name := range planAttrs {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic diagnostic order

	for _, name := range names {
		planValue := planAttrs[name]
		stateValue, ok := stateAttrs[name]
		// Unknown children are not user-driven changes: a deferred data source
		// or a sibling Computed modifier (e.g. UseStateForUnknown) resolves
		// them before the plan is finalized.
		if !ok || planValue.IsUnknown() {
			continue
		}

		changed, reorderOnly := existingVpcAttributeChanged(stateValue, planValue)
		if !changed {
			continue
		}
		if reorderOnly {
			// An order-only difference cannot be let through: the planned value
			// cannot be normalized to the state order (Terraform core rejects
			// planned values that differ from config on non-Computed attributes),
			// and applying would produce an inconsistent result because the API
			// ignores the reorder. It gets a dedicated message so the remediation
			// is obvious.
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(name),
				"Existing VPC list order changed",
				"This list contains the same elements as the Terraform state but in a different order. "+
					"The Qovery API ignores ordering changes, so applying would produce an inconsistent result. "+
					"Reorder the values in your configuration to match the Terraform state.",
			)
			continue
		}
		addExistingVpcImmutableError(req.Path.AtName(name), &resp.Diagnostics)
	}
}

// existingVpcAttributeChanged reports whether a child attribute of an
// existing-VPC block has a user-visible change between state and plan.
// reorderOnly is true when a list holds the same elements in a different
// order. Lists with unknown elements are never a change — deferred data
// sources must resolve before a concrete comparison is possible.
func existingVpcAttributeChanged(stateValue, planValue attr.Value) (changed bool, reorderOnly bool) {
	switch plan := planValue.(type) {
	case types.List:
		state, ok := stateValue.(types.List)
		if !ok {
			break
		}
		if plan.Equal(state) {
			return false, false
		}
		for _, element := range plan.Elements() {
			if element.IsUnknown() {
				return false, false
			}
		}
		if listIsNullOrEmpty(state) && listIsNullOrEmpty(plan) {
			return false, false
		}
		return true, listsEqualIgnoringOrder(state, plan)
	case types.String:
		state, ok := stateValue.(types.String)
		if ok && stringIsNullOrEmpty(state) && stringIsNullOrEmpty(plan) {
			return false, false
		}
	case types.Bool:
		state, ok := stateValue.(types.Bool)
		if ok && boolIsNullOrFalse(state) && boolIsNullOrFalse(plan) {
			return false, false
		}
	}
	return !planValue.Equal(stateValue), false
}

// addExistingVpcImmutableError emits the shared diagnostic for immutable
// existing-VPC attributes (AWS and GCP): the API ignores these changes, so the
// plan is rejected instead of forcing a cluster replacement.
func addExistingVpcImmutableError(p path.Path, diags *diag.Diagnostics) {
	diags.AddAttributeError(
		p,
		"Cannot change existing VPC configuration",
		"The existing VPC configuration is immutable after cluster creation: the Qovery API ignores these changes. "+
			"To keep this cluster, revert this attribute to match the Terraform state. "+
			"To use a different VPC configuration, recreate the cluster explicitly, e.g. `terraform destroy -target=<cluster resource>` followed by `terraform apply`. "+
			"Note that `terraform apply -replace=...` cannot be used here because the plan is rejected before the replacement is applied.",
	)
}

// listIsNullOrEmpty reports whether the list is null or has no elements.
func listIsNullOrEmpty(v types.List) bool {
	return v.IsNull() || len(v.Elements()) == 0
}

// listsEqualIgnoringOrder reports whether two known lists hold the same elements
// with the same multiplicity, regardless of order.
func listsEqualIgnoringOrder(a, b types.List) bool {
	aElements := a.Elements()
	bElements := b.Elements()
	if len(aElements) != len(bElements) {
		return false
	}

	matched := make([]bool, len(bElements))
	for _, aElement := range aElements {
		found := false
		for i, bElement := range bElements {
			if !matched[i] && aElement.Equal(bElement) {
				matched[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// RejectExistingVpcChange rejects any change to an immutable existing-VPC
// block after cluster creation — presence and content — without forcing
// replacement. Attach it to the block's object attribute; children need no
// per-attribute modifiers and new attributes are immutable by default.
func RejectExistingVpcChange() planmodifier.Object {
	return rejectExistingVpcChangeModifier{}
}

// stringIsNullOrEmpty reports whether the string is null or empty.
func stringIsNullOrEmpty(v types.String) bool {
	return v.IsNull() || v.ValueString() == ""
}

// boolIsNullOrFalse reports whether the bool is null or false.
func boolIsNullOrFalse(v types.Bool) bool {
	return v.IsNull() || !v.ValueBool()
}

// RequiresReplaceIfKnownChangeTreatingEmptyAs behaves like RequiresReplaceIfKnownChange
// but treats an empty-string value as equal to defaultValue when comparing state and
// plan. Use it on an Optional+Computed string attribute whose schema Default is
// defaultValue, when a legacy provider version may have persisted "" in state for the
// same logical value.
//
// Without this, upgrading the provider can manufacture a phantom change: state holds
// the legacy "" while the schema Default makes the planned value defaultValue, and on a
// `-refresh=false` plan (where state isn't rewritten first) a plain RequiresReplace
// modifier would destroy and recreate the resource even though nothing the user owns
// changed. A genuine change between two distinct non-empty values still forces
// replacement. See features.vpc_subnet
func RequiresReplaceIfKnownChangeTreatingEmptyAs(defaultValue string) planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		requiresReplaceIfKnownChangeTreatingEmptyAsFunc(defaultValue),
		"If the value changes to a different known value, Terraform will destroy and recreate the resource. An empty value is treated as the default, and replacement is skipped when the planned value is unknown.",
		"If the value changes to a different known value, Terraform will destroy and recreate the resource. An empty value is treated as the default (`"+defaultValue+"`), and replacement is **skipped when the planned value is unknown**.",
	)
}

// requiresReplaceIfKnownChangeTreatingEmptyAsFunc builds the predicate for
// RequiresReplaceIfKnownChangeTreatingEmptyAs. It normalizes null/empty values to
// defaultValue before deciding whether the change is real.
func requiresReplaceIfKnownChangeTreatingEmptyAsFunc(defaultValue string) stringplanmodifier.RequiresReplaceIfFunc {
	normalize := func(v types.String) string {
		if v.IsNull() || v.IsUnknown() {
			return defaultValue
		}
		if s := v.ValueString(); s != "" {
			return s
		}
		return defaultValue
	}
	return func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		if req.PlanValue.IsUnknown() {
			return
		}
		if normalize(req.StateValue) == normalize(req.PlanValue) {
			return
		}
		resp.RequiresReplace = true
	}
}
