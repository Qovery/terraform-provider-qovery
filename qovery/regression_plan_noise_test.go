//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Regression test for QOV-1605 / "[xyz] will be read during apply" plan noise.
//
// Background:
// A `Computed` attribute without a state-preserving plan modifier (and no
// `Default`) is recomputed as `(known after apply)` whenever the resource is
// planned to update for ANY other reason. That false flicker propagates to
// every data source that has an explicit or implicit dependency on the
// resource — they get marked "will be read during apply" — which is the
// noisy plan output users complain about.
//
// This test walks every resource schema in the provider (recursively into
// nested objects) and asserts that every `Computed` attribute either:
//   - has a state-preserving plan modifier (UseStateForUnknown or one of our
//     custom UseStateUnless* modifiers), or
//   - has a `Default` whose value matches what the API returns when the user
//     omits the attribute, or
//   - is on the `flickerAllowlist` with a documented reason.
//
// When this test fails, the fix is almost always to add `UseStateForUnknown()`
// to the offending attribute's `PlanModifiers`. See `qovery/plan_modifiers.go`
// for our custom variants when the attribute genuinely needs to be recomputed
// under specific conditions (e.g. accessibility/name/ports changes).

// useStateForUnknownDescription is the Description() string returned by the
// terraform-plugin-framework's built-in UseStateForUnknown() modifiers (string,
// bool, int64, list, set, map, object). The concrete types are unexported in
// the framework, so matching on description is the most stable identifier.
const useStateForUnknownDescription = "Once set, the value of this attribute in state will not change."

// preservesState reports whether a single plan modifier preserves the prior
// state value, either unconditionally (UseStateForUnknown) or under
// conditions that don't fire on an unrelated update to the same resource
// (UseStateUnlessNameChanges, UseStateUnlessPortsChange, SmartAllowApiOverride).
//
// This list is intentionally curated: only modifiers whose semantics
// guarantee "no flicker on unrelated updates" qualify. New custom modifiers
// added to qovery/plan_modifiers.go must be added here explicitly if they
// preserve state.
func preservesState(m planmodifier.Describer) bool {
	switch m.(type) {
	case useStateUnlessNameChangesModifier,
		useStateUnlessPortsChangeModifier,
		smartAllowApiOverrideModifier:
		return true
	}
	return m.Description(context.Background()) == useStateForUnknownDescription
}

func anyPreservesState[M planmodifier.Describer](mods []M) bool {
	for _, m := range mods {
		if preservesState(m) {
			return true
		}
	}
	return false
}

// flickerAllowlist is the set of fully-qualified attribute paths
// (`<resource>.<dot.path.to.attr>`) that are deliberately allowed to flicker
// as `(known after apply)` on every plan.
//
// Each entry MUST include a justification. Adding to this list should be
// a deliberate, reviewed decision — almost every Computed attribute should
// preserve state, and reaching for the allowlist usually means we've missed
// a plan modifier on the schema.
//
// Format: "<terraform_resource_type>.<dot.separated.attribute.path>" -> reason.
var flickerAllowlist = map[string]string{
	// (intentionally empty — every Computed attribute should preserve state.
	// If a future API change makes an attribute legitimately volatile,
	// document the reason here and reference the API/PR that introduced it.)
}

// attributeStatus describes the relevant flags of a single schema attribute
// for the purposes of this test.
type attributeStatus struct {
	computed       bool
	preservesState bool
	hasDefault     bool
}

// inspectAttribute extracts (computed, has-state-preserving-modifier,
// has-default) for a single attribute. Returns ok=false if the attribute
// type is not one we understand — in that case the caller should add a
// case to this switch rather than silently passing the attribute.
func inspectAttribute(attr schema.Attribute) (status attributeStatus, ok bool) {
	switch a := attr.(type) {
	case schema.StringAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.BoolAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.Int64Attribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.ListAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.SetAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.MapAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.SingleNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.ListNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.SetNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	case schema.MapNestedAttribute:
		return attributeStatus{a.Computed, anyPreservesState(a.PlanModifiers), a.Default != nil}, true
	}
	return attributeStatus{}, false
}

// nestedAttributes returns the inner attribute map for nested-object types,
// or nil for leaf types.
func nestedAttributes(attr schema.Attribute) map[string]schema.Attribute {
	switch a := attr.(type) {
	case schema.SingleNestedAttribute:
		return a.Attributes
	case schema.ListNestedAttribute:
		return a.NestedObject.Attributes
	case schema.SetNestedAttribute:
		return a.NestedObject.Attributes
	case schema.MapNestedAttribute:
		return a.NestedObject.Attributes
	}
	return nil
}

// walkAttributes invokes visit for every attribute under attrs, recursing
// into nested objects. Paths use dot-separated notation (e.g.
// `schedule.cronjob.command.entrypoint`). Iteration order is deterministic
// (alphabetical) so test output is stable.
func walkAttributes(prefix string, attrs map[string]schema.Attribute, visit func(path string, attr schema.Attribute)) {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		attr := attrs[k]
		visit(path, attr)
		if children := nestedAttributes(attr); children != nil {
			walkAttributes(path, children, visit)
		}
	}
}

// resourceCase pairs a registered terraform resource type name with its
// instance, so that test output references the same identifier users see
// in their HCL and plan output.
type resourceCase struct {
	typeName string
	resource resource.Resource
}

// allResourceCases returns one entry per resource registered on the provider.
// Adding a new resource to qProvider.Resources() automatically extends test
// coverage; no test changes required.
func allResourceCases(t *testing.T) []resourceCase {
	t.Helper()
	var p qProvider
	ctx := context.Background()
	ctors := p.Resources(ctx)

	cases := make([]resourceCase, 0, len(ctors))
	for _, ctor := range ctors {
		r := ctor()
		mdResp := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "qovery"}, mdResp)
		if mdResp.TypeName == "" {
			t.Fatalf("resource %T returned an empty TypeName from Metadata; cannot test", r)
		}
		cases = append(cases, resourceCase{typeName: mdResp.TypeName, resource: r})
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].typeName < cases[j].typeName })
	return cases
}

// schemaFor builds a fresh resource schema from the resource's Schema()
// method. No provider configuration or API access required.
func schemaFor(t *testing.T, r resource.Resource) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to build schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

// TestRegression_PlanNoise_NoComputedFlicker is the comprehensive lint test.
// It walks every resource schema and reports any Computed attribute whose
// planned value will flicker as `(known after apply)` on unrelated updates.
//
// To debug a failure: read the failing attribute path, open the resource's
// schema file, locate the attribute, and add `UseStateForUnknown()` (or a
// state-preserving custom modifier) to its `PlanModifiers`. Re-run the test
// to confirm the gap is closed.
func TestRegression_PlanNoise_NoComputedFlicker(t *testing.T) {
	t.Parallel()

	for _, tc := range allResourceCases(t) {
		tc := tc
		t.Run(tc.typeName, func(t *testing.T) {
			t.Parallel()

			sch := schemaFor(t, tc.resource)

			var problems []string
			walkAttributes("", sch.Attributes, func(path string, attr schema.Attribute) {
				status, ok := inspectAttribute(attr)
				if !ok {
					t.Errorf("unhandled attribute type %T at %s.%s — extend inspectAttribute()", attr, tc.typeName, path)
					return
				}
				if !status.computed {
					return
				}
				if status.preservesState || status.hasDefault {
					return
				}
				if _, allowed := flickerAllowlist[tc.typeName+"."+path]; allowed {
					return
				}
				problems = append(problems, path)
			})

			if len(problems) == 0 {
				return
			}

			sort.Strings(problems)
			t.Errorf(
				"%s has %d Computed attribute(s) with no state-preserving plan modifier and no Default.\n"+
					"On any unrelated update to a resource of this type, each will flicker as\n"+
					"`(known after apply)` and propagate spurious `changes pending` to dependent\n"+
					"data sources/modules — the regression reported in QOV-1605.\n\n"+
					"Fix: add `UseStateForUnknown()` (from terraform-plugin-framework) or a custom\n"+
					"modifier from qovery/plan_modifiers.go to each attribute below. If an attribute\n"+
					"is genuinely volatile, add it to flickerAllowlist with a written reason.\n\n"+
					"Failing attributes:\n  - %s",
				tc.typeName, len(problems), strings.Join(problems, "\n  - "),
			)
		})
	}
}
