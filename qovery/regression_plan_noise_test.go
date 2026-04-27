//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// useStateForUnknownDescription is the description string returned by all
// terraform-plugin-framework UseStateForUnknown() plan modifiers (string, bool,
// list, set, etc.). We match on description because the underlying types are
// unexported by Hashicorp.
const useStateForUnknownDescription = "Once set, the value of this attribute in state will not change."

// stateForUnknownAliases captures the descriptions of custom plan modifiers in
// this provider that preserve state in the same way (or as a stricter
// conditional variant) as UseStateForUnknown.
var stateForUnknownAliases = map[string]bool{
	useStateForUnknownDescription: true,
	"Uses state value unless the resource name is changing, in which case the value is recomputed.":            true,
	"Uses state value unless the resource name or ports are changing, in which case the value is recomputed.": true,
	"Uses state value unless ports are changing, in which case the value is recomputed.":                       true,
}

func anyPreservesState(descriptions []string) bool {
	for _, d := range descriptions {
		if stateForUnknownAliases[d] {
			return true
		}
	}
	return false
}

func stringModifierDescriptions(mods []planmodifier.String) []string {
	out := make([]string, len(mods))
	ctx := context.Background()
	for i, m := range mods {
		out[i] = m.Description(ctx)
	}
	return out
}

func boolModifierDescriptions(mods []planmodifier.Bool) []string {
	out := make([]string, len(mods))
	ctx := context.Background()
	for i, m := range mods {
		out[i] = m.Description(ctx)
	}
	return out
}

func listModifierDescriptions(mods []planmodifier.List) []string {
	out := make([]string, len(mods))
	ctx := context.Background()
	for i, m := range mods {
		out[i] = m.Description(ctx)
	}
	return out
}

func int64ModifierDescriptions(mods []planmodifier.Int64) []string {
	out := make([]string, len(mods))
	ctx := context.Background()
	for i, m := range mods {
		out[i] = m.Description(ctx)
	}
	return out
}

// schemaFor builds a fresh resource schema by invoking the resource's Schema()
// method directly. No provider configuration or API access required.
func schemaFor(t *testing.T, r resource.Resource) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to build schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

// checkPreservesState asserts that the named top-level Computed attribute on
// the given schema has a plan modifier that preserves state across plans.
// Failure means a `terraform plan` against unchanged config will mark this
// attribute as "(known after apply)", which propagates "changes pending" to
// any data source / module with depends_on on the resource — the symptom
// reported in the issue (e.g. data.aws_caller_identity.current will be read
// during apply).
func checkPreservesState(t *testing.T, resourceName string, sch schema.Schema, attrName string) {
	t.Helper()

	attr, ok := sch.Attributes[attrName]
	if !ok {
		t.Fatalf("%s: attribute %q not found in schema", resourceName, attrName)
	}

	var descs []string
	switch a := attr.(type) {
	case schema.StringAttribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true (test only applies to computed attrs)", resourceName, attrName)
		}
		descs = stringModifierDescriptions(a.PlanModifiers)
	case schema.BoolAttribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true", resourceName, attrName)
		}
		descs = boolModifierDescriptions(a.PlanModifiers)
	case schema.ListNestedAttribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true", resourceName, attrName)
		}
		descs = listModifierDescriptions(a.PlanModifiers)
	case schema.Int64Attribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true", resourceName, attrName)
		}
		descs = int64ModifierDescriptions(a.PlanModifiers)
	default:
		t.Fatalf("%s.%s: unsupported attribute type %T", resourceName, attrName, attr)
	}

	if !anyPreservesState(descs) {
		t.Errorf(
			"%s.%s is Computed but has no state-preserving plan modifier "+
				"(UseStateForUnknown / UseStateUnlessNameChanges / UseStateUnlessPortsChange).\n"+
				"This causes the attribute to flicker as `(known after apply)` on every plan, "+
				"which propagates `changes pending` to dependent data sources/modules — the "+
				"regression reported in QOV-1605.\n"+
				"Current plan modifiers: %v",
			resourceName, attrName, descs,
		)
	}
}

// TestRegression_PlanNoise_DeploymentStageID covers the deployment_stage_id
// attribute, which is Optional+Computed on every service resource. The
// QOV-1605 fix added UseStateForUnknown to neighboring attributes
// (internal_host, advanced_settings_json, auto_deploy) but skipped this one,
// causing plan noise to resurface for users on v0.68.0.
func TestRegression_PlanNoise_DeploymentStageID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		resourceName string
		resource     resource.Resource
	}{
		{"qovery_application", &applicationResource{}},
		{"qovery_container", &containerResource{}},
		{"qovery_job", &jobResource{}},
		{"qovery_helm", &helmResource{}},
		{"qovery_database", &databaseResource{}},
		{"qovery_terraform_service", &terraformServiceResource{}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.resourceName, func(t *testing.T) {
			t.Parallel()
			sch := schemaFor(t, tc.resource)
			checkPreservesState(t, tc.resourceName, sch, "deployment_stage_id")
		})
	}
}

// TestRegression_PlanNoise_BuiltInEnvironmentVariables covers
// built_in_environment_variables, which is Computed on environment and
// project resources. The QOV-1605 fix added UseStateUnlessNameChanges() on
// application/container/job/helm but missed environment and project.
func TestRegression_PlanNoise_BuiltInEnvironmentVariables(t *testing.T) {
	t.Parallel()

	cases := []struct {
		resourceName string
		resource     resource.Resource
	}{
		{"qovery_environment", &environmentResource{}},
		{"qovery_project", &projectResource{}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.resourceName, func(t *testing.T) {
			t.Parallel()
			sch := schemaFor(t, tc.resource)
			checkPreservesState(t, tc.resourceName, sch, "built_in_environment_variables")
		})
	}
}

// TestRegression_PlanNoise_DatabaseHostsAndCredentials covers attributes on
// qovery_database that are read-only (Computed without Optional) but lack
// UseStateForUnknown. These are the values most often referenced from
// downstream modules (e.g. an application reading internal_host to construct
// a connection string), so plan noise here cascades widely.
func TestRegression_PlanNoise_DatabaseHostsAndCredentials(t *testing.T) {
	t.Parallel()

	sch := schemaFor(t, &databaseResource{})

	// Computed-only attributes that should be stable across plans.
	for _, attr := range []string{
		"external_host",
		"internal_host",
		"port",
		"login",
		"password",
	} {
		attr := attr
		t.Run(attr, func(t *testing.T) {
			t.Parallel()
			checkPreservesState(t, "qovery_database", sch, attr)
		})
	}
}

// allRegisteredResources returns one (typeName, resource) entry per resource
// the provider exposes. Pulled from qProvider.Resources() so any newly-added
// resource is automatically picked up by the audit below.
func allRegisteredResources(t *testing.T) []struct {
	name     string
	resource resource.Resource
} {
	t.Helper()
	ctx := context.Background()
	out := make([]struct {
		name     string
		resource resource.Resource
	}, 0)
	for _, factory := range (&qProvider{}).Resources(ctx) {
		r := factory()
		metaResp := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "qovery"}, metaResp)
		out = append(out, struct {
			name     string
			resource resource.Resource
		}{metaResp.TypeName, r})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

// TestRegression_PlanNoise_Summary prints a consolidated report of all
// top-level Computed string/bool/list/int64 attributes across EVERY
// registered resource that lack a state-preserving plan modifier. This is
// informational — it always passes — but produces an audit-style log when
// -v is used. The resource list is pulled live from qProvider.Resources()
// so new resources are covered automatically.
func TestRegression_PlanNoise_Summary(t *testing.T) {
	type gap struct {
		resource string
		attr     string
		descs    []string
	}
	var gaps []gap

	resources := allRegisteredResources(t)
	for _, r := range resources {
		sch := schemaFor(t, r.resource)
		for name, attr := range sch.Attributes {
			var descs []string
			var isComputed bool

			switch a := attr.(type) {
			case schema.StringAttribute:
				isComputed = a.Computed
				descs = stringModifierDescriptions(a.PlanModifiers)
			case schema.BoolAttribute:
				isComputed = a.Computed
				descs = boolModifierDescriptions(a.PlanModifiers)
			case schema.ListNestedAttribute:
				isComputed = a.Computed
				descs = listModifierDescriptions(a.PlanModifiers)
			case schema.Int64Attribute:
				isComputed = a.Computed
				descs = int64ModifierDescriptions(a.PlanModifiers)
			default:
				continue
			}

			if isComputed && !anyPreservesState(descs) {
				gaps = append(gaps, gap{r.name, name, descs})
			}
		}
	}

	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].resource != gaps[j].resource {
			return gaps[i].resource < gaps[j].resource
		}
		return gaps[i].attr < gaps[j].attr
	})

	t.Logf("audited %d registered resource(s)", len(resources))
	if len(gaps) == 0 {
		t.Log("no plan-noise gaps found across the entire provider")
		return
	}

	t.Logf("found %d Computed top-level attribute(s) without a state-preserving plan modifier:", len(gaps))
	for _, g := range gaps {
		t.Log(fmt.Sprintf("  %-32s %-32s modifiers=%v", g.resource, g.attr, g.descs))
	}
}
