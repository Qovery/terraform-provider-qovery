//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// useStateForUnknownDescription is the Description() string returned by
// terraform-plugin-framework's UseStateForUnknown() modifiers (string, bool,
// list, int64, ...). Their underlying types are unexported, so matching on
// description is the only stable identifier for them.
const useStateForUnknownDescription = "Once set, the value of this attribute in state will not change."

func preservesState(m planmodifier.Describer) bool {
	switch m.(type) {
	case useStateUnlessNameChangesModifier, useStateUnlessPortsChangeModifier:
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

func schemaFor(t *testing.T, r resource.Resource) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to build schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func checkPreservesState(t *testing.T, resourceName string, sch schema.Schema, attrName string) {
	t.Helper()

	attr, ok := sch.Attributes[attrName]
	if !ok {
		t.Fatalf("%s: attribute %q not found in schema", resourceName, attrName)
	}

	var preserves bool
	switch a := attr.(type) {
	case schema.StringAttribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true", resourceName, attrName)
		}
		preserves = anyPreservesState(a.PlanModifiers)
	case schema.BoolAttribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true", resourceName, attrName)
		}
		preserves = anyPreservesState(a.PlanModifiers)
	case schema.ListNestedAttribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true", resourceName, attrName)
		}
		preserves = anyPreservesState(a.PlanModifiers)
	case schema.Int64Attribute:
		if !a.Computed {
			t.Fatalf("%s.%s: expected Computed=true", resourceName, attrName)
		}
		preserves = anyPreservesState(a.PlanModifiers)
	default:
		t.Fatalf("%s.%s: unsupported attribute type %T", resourceName, attrName, attr)
	}

	if !preserves {
		t.Errorf(
			"%s.%s is Computed but has no state-preserving plan modifier — "+
				"the attribute will flicker as `(known after apply)` on every plan, "+
				"propagating spurious diffs to dependent data sources and modules.",
			resourceName, attrName,
		)
	}
}

type resourceCase struct {
	name     string
	resource resource.Resource
}

func TestRegression_PlanNoise_DeploymentStageID(t *testing.T) {
	t.Parallel()

	cases := []resourceCase{
		{"qovery_application", &applicationResource{}},
		{"qovery_container", &containerResource{}},
		{"qovery_job", &jobResource{}},
		{"qovery_helm", &helmResource{}},
		{"qovery_database", &databaseResource{}},
		{"qovery_terraform_service", &terraformServiceResource{}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sch := schemaFor(t, tc.resource)
			checkPreservesState(t, tc.name, sch, "deployment_stage_id")
		})
	}
}

func TestRegression_PlanNoise_BuiltInEnvironmentVariables(t *testing.T) {
	t.Parallel()

	cases := []resourceCase{
		{"qovery_environment", &environmentResource{}},
		{"qovery_project", &projectResource{}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sch := schemaFor(t, tc.resource)
			checkPreservesState(t, tc.name, sch, "built_in_environment_variables")
		})
	}
}

func TestRegression_PlanNoise_DatabaseHostsAndCredentials(t *testing.T) {
	t.Parallel()

	sch := schemaFor(t, &databaseResource{})

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
