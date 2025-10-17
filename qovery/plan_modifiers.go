package qovery

import (
	"context"

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
