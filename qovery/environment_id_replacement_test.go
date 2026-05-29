//go:build unit && !integration
// +build unit,!integration

// Tests for RequiresReplaceIfKnownChange and its application to RequiresReplace
// ID attributes (environment_id, cluster_id) in resource schemas.
package qovery

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// testSchemaEnvironmentID is a minimal schema with one string attribute used for
// constructing synthetic State/Plan/Config trees in the tests below.
var testSchemaEnvironmentID = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"environment_id": schema.StringAttribute{},
	},
}

// buildState returns a tfsdk.State whose Raw is non-null and carries the given value
// for the `environment_id` attribute. Pass a null types.String to mark the attribute
// itself null; pass nil to mark the whole state null (resource being created).
func buildState(value *types.String) tfsdk.State {
	ctx := context.Background()
	objType := testSchemaEnvironmentID.Type().TerraformType(ctx)

	if value == nil {
		return tfsdk.State{
			Schema: testSchemaEnvironmentID,
			Raw:    tftypes.NewValue(objType, nil),
		}
	}

	tfValue, err := value.ToTerraformValue(ctx)
	if err != nil {
		panic("ToTerraformValue: " + err.Error())
	}
	return tfsdk.State{
		Schema: testSchemaEnvironmentID,
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"environment_id": tfValue,
		}),
	}
}

func buildPlan(value *types.String) tfsdk.Plan {
	ctx := context.Background()
	objType := testSchemaEnvironmentID.Type().TerraformType(ctx)

	if value == nil {
		return tfsdk.Plan{
			Schema: testSchemaEnvironmentID,
			Raw:    tftypes.NewValue(objType, nil),
		}
	}

	tfValue, err := value.ToTerraformValue(ctx)
	if err != nil {
		panic("ToTerraformValue: " + err.Error())
	}
	return tfsdk.Plan{
		Schema: testSchemaEnvironmentID,
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"environment_id": tfValue,
		}),
	}
}

func buildConfig(value *types.String) tfsdk.Config {
	ctx := context.Background()
	objType := testSchemaEnvironmentID.Type().TerraformType(ctx)

	if value == nil {
		return tfsdk.Config{
			Schema: testSchemaEnvironmentID,
			Raw:    tftypes.NewValue(objType, nil),
		}
	}

	tfValue, err := value.ToTerraformValue(ctx)
	if err != nil {
		panic("ToTerraformValue: " + err.Error())
	}
	return tfsdk.Config{
		Schema: testSchemaEnvironmentID,
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"environment_id": tfValue,
		}),
	}
}

// ptrUnknown returns a pointer to an Unknown string value for use with the build* helpers.
func ptrUnknown() *types.String {
	v := types.StringUnknown()
	return &v
}

// ----------------------------------------------------------------------------
// Predicate unit tests — exercise requiresReplaceIfKnownChangeFunc directly,
// bypassing the framework's RequiresReplaceIf guards.
// ----------------------------------------------------------------------------

func TestRequiresReplaceIfKnownChangeFunc_PlanKnown_SetsReplace(t *testing.T) {
	t.Parallel()

	resp := &stringplanmodifier.RequiresReplaceIfFuncResponse{}
	requiresReplaceIfKnownChangeFunc(context.Background(), planmodifier.StringRequest{
		PlanValue: types.StringValue("env-uuid-NEW"),
	}, resp)

	assert.True(t, resp.RequiresReplace,
		"a known plan value differing from state must require replacement")
}

func TestRequiresReplaceIfKnownChangeFunc_PlanUnknown_DoesNotSetReplace(t *testing.T) {
	t.Parallel()

	resp := &stringplanmodifier.RequiresReplaceIfFuncResponse{}
	requiresReplaceIfKnownChangeFunc(context.Background(), planmodifier.StringRequest{
		PlanValue: types.StringUnknown(),
	}, resp)

	assert.False(t, resp.RequiresReplace,
		"unknown plan value must not require replacement")
}

// ----------------------------------------------------------------------------
// Integration tests — exercise the full RequiresReplaceIfKnownChange() modifier
// through the framework's call path.
// ----------------------------------------------------------------------------

// TestRequiresReplaceIfKnownChange_UnknownPlan_Integration verifies the modifier
// when both ConfigValue and PlanValue are unknown (e.g., the config interpolates
// a deferred data source).
func TestRequiresReplaceIfKnownChange_UnknownPlan_Integration(t *testing.T) {
	t.Parallel()

	stateVal := types.StringValue("env-uuid-123")
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	RequiresReplaceIfKnownChange().PlanModifyString(context.Background(), planmodifier.StringRequest{
		Config:      buildConfig(ptrUnknown()),
		ConfigValue: types.StringUnknown(),
		State:       buildState(&stateVal),
		StateValue:  stateVal,
		Plan:        buildPlan(ptrUnknown()),
		PlanValue:   types.StringUnknown(),
	}, resp)

	assert.False(t, resp.RequiresReplace,
		"unknown plan value must not trigger replacement")
}

// TestRequiresReplaceIfKnownChange_RealChange_Integration verifies that a
// known-value change to a different value triggers replacement.
func TestRequiresReplaceIfKnownChange_RealChange_Integration(t *testing.T) {
	t.Parallel()

	stateVal := types.StringValue("env-uuid-OLD")
	newVal := types.StringValue("env-uuid-NEW")
	resp := &planmodifier.StringResponse{PlanValue: newVal}

	RequiresReplaceIfKnownChange().PlanModifyString(context.Background(), planmodifier.StringRequest{
		Config:      buildConfig(&newVal),
		ConfigValue: newVal,
		State:       buildState(&stateVal),
		StateValue:  stateVal,
		Plan:        buildPlan(&newVal),
		PlanValue:   newVal,
	}, resp)

	assert.True(t, resp.RequiresReplace,
		"a known-value change must require replacement")
}

// ----------------------------------------------------------------------------
// Schema regression tests — verify each known RequiresReplace ID attribute uses
// RequiresReplaceIfKnownChange().
// ----------------------------------------------------------------------------

// hasRequiresReplaceIfKnownChange reports whether the modifier slice contains
// RequiresReplaceIfKnownChange. The framework's modifier types are unexported, so
// we compare by Description string.
func hasRequiresReplaceIfKnownChange(mods []planmodifier.String) bool {
	ctx := context.Background()
	wantDesc := RequiresReplaceIfKnownChange().Description(ctx)
	for _, m := range mods {
		if m.Description(ctx) == wantDesc {
			return true
		}
	}
	return false
}

// hasStockRequiresReplace reports whether the modifier slice contains the stock
// stringplanmodifier.RequiresReplace().
func hasStockRequiresReplace(mods []planmodifier.String) bool {
	ctx := context.Background()
	wantDesc := stringplanmodifier.RequiresReplace().Description(ctx)
	for _, m := range mods {
		if m.Description(ctx) == wantDesc {
			return true
		}
	}
	return false
}

// vulnerableAttribute pairs a resource schema with the name of a RequiresReplace
// string attribute that must use RequiresReplaceIfKnownChange().
type vulnerableAttribute struct {
	resourceName string
	attrName     string
	schema       schema.Schema
}

func collectVulnerableAttributes(t *testing.T) []vulnerableAttribute {
	t.Helper()

	schemaOf := func(r interface {
		Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	}) schema.Schema {
		var resp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
		return resp.Schema
	}

	return []vulnerableAttribute{
		{"qovery_container", "environment_id", schemaOf(containerResource{})},
		{"qovery_helm", "environment_id", schemaOf(helmResource{})},
		{"qovery_application", "environment_id", schemaOf(applicationResource{})},
		{"qovery_job", "environment_id", schemaOf(jobResource{})},
		{"qovery_database", "environment_id", schemaOf(databaseResource{})},
		{"qovery_terraform_service", "environment_id", schemaOf(terraformServiceResource{})},
		{"qovery_deployment_stage", "environment_id", schemaOf(deploymentStageResource{})},
		{"qovery_environment", "cluster_id", schemaOf(environmentResource{})},
		{"qovery_environment", "project_id", schemaOf(environmentResource{})},
		{"qovery_project", "organization_id", schemaOf(projectResource{})},
		{"qovery_cluster", "organization_id", schemaOf(clusterResource{})},
		{"qovery_annotations_group", "organization_id", schemaOf(annotationsGroupResource{})},
		{"qovery_labels_group", "organization_id", schemaOf(labelsGroupResource{})},
		{"qovery_container_registry", "organization_id", schemaOf(containerRegistryResource{})},
		{"qovery_helm_repository", "organization_id", schemaOf(helmRepositoryResource{})},
		{"qovery_git_token", "organization_id", schemaOf(gitTokenResource{})},
		{"qovery_aws_credentials", "organization_id", schemaOf(awsCredentialsResource{})},
		{"qovery_gcp_credentials", "organization_id", schemaOf(gcpCredentialsResource{})},
		{"qovery_scaleway_credentials", "organization_id", schemaOf(scalewayCredentialsResource{})},
		{"qovery_eks_anywhere_vsphere_credentials", "organization_id", schemaOf(eksAnywhereVsphereCredentialsResource{})},
	}
}

func getPlanModifiers(t *testing.T, s schema.Schema, attrName string) []planmodifier.String {
	t.Helper()
	attr, ok := s.Attributes[attrName]
	if !ok {
		t.Fatalf("schema missing %q attribute", attrName)
	}
	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("%q is not a StringAttribute", attrName)
	}
	return strAttr.PlanModifiers
}

// TestVulnerableAttributes_UseRequiresReplaceIfKnownChange asserts that every
// listed RequiresReplace ID attribute uses RequiresReplaceIfKnownChange() and not
// the stock stringplanmodifier.RequiresReplace().
func TestVulnerableAttributes_UseRequiresReplaceIfKnownChange(t *testing.T) {
	t.Parallel()

	for _, v := range collectVulnerableAttributes(t) {
		v := v
		t.Run(v.resourceName+"."+v.attrName, func(t *testing.T) {
			t.Parallel()
			mods := getPlanModifiers(t, v.schema, v.attrName)
			assert.True(t, hasRequiresReplaceIfKnownChange(mods),
				"%s.%s must use RequiresReplaceIfKnownChange()", v.resourceName, v.attrName)
			assert.False(t, hasStockRequiresReplace(mods),
				"%s.%s must not use stock stringplanmodifier.RequiresReplace()", v.resourceName, v.attrName)
		})
	}
}

// Asserts qovery_deployment.environment_id does not force replacement, since its
// Delete deletes the target environment.
func TestDeployment_EnvironmentID_DoesNotForceReplacement(t *testing.T) {
	t.Parallel()

	var resp resource.SchemaResponse
	deploymentResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	mods := getPlanModifiers(t, resp.Schema, "environment_id")

	assert.False(t, hasRequiresReplaceIfKnownChange(mods),
		"qovery_deployment.environment_id must NOT force replacement: Delete deletes the previous environment")
	assert.False(t, hasStockRequiresReplace(mods),
		"qovery_deployment.environment_id must NOT force replacement: Delete deletes the previous environment")
}

// ----------------------------------------------------------------------------
// Provider-wide reflective scan — walks every resource's schema (including nested
// attributes) and fails if any StringAttribute uses stock
// stringplanmodifier.RequiresReplace().
// ----------------------------------------------------------------------------

// walkStringAttributes invokes fn for every StringAttribute found in the schema
// attribute tree, recursing into Single/List/Set/Map nested attributes.
func walkStringAttributes(attrs map[string]schema.Attribute, prefix string, fn func(path string, sa schema.StringAttribute)) {
	for name, attr := range attrs {
		full := name
		if prefix != "" {
			full = prefix + "." + name
		}
		switch a := attr.(type) {
		case schema.StringAttribute:
			fn(full, a)
		case schema.SingleNestedAttribute:
			walkStringAttributes(a.Attributes, full, fn)
		case schema.ListNestedAttribute:
			walkStringAttributes(a.NestedObject.Attributes, full+"[*]", fn)
		case schema.SetNestedAttribute:
			walkStringAttributes(a.NestedObject.Attributes, full+"[*]", fn)
		case schema.MapNestedAttribute:
			walkStringAttributes(a.NestedObject.Attributes, full+"[*]", fn)
		}
	}
}

// TestProvider_NoStockRequiresReplaceAnywhere walks every resource in the provider
// and asserts that no StringAttribute uses stock stringplanmodifier.RequiresReplace().
func TestProvider_NoStockRequiresReplaceAnywhere(t *testing.T) {
	t.Parallel()

	p := New("test")()
	ctx := context.Background()

	resourcesProvider, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	if !ok {
		t.Fatalf("provider does not expose Resources(context.Context)")
	}

	var violations []string
	for _, factory := range resourcesProvider.Resources(ctx) {
		res := factory()

		var meta resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "qovery"}, &meta)

		var sresp resource.SchemaResponse
		res.Schema(ctx, resource.SchemaRequest{}, &sresp)

		walkStringAttributes(sresp.Schema.Attributes, "", func(path string, sa schema.StringAttribute) {
			if hasStockRequiresReplace(sa.PlanModifiers) {
				violations = append(violations, meta.TypeName+"."+path)
			}
		})
	}

	if len(violations) > 0 {
		t.Errorf("the following attributes use stringplanmodifier.RequiresReplace(); replace with RequiresReplaceIfKnownChange() from plan_modifiers.go:\n  - %s",
			strings.Join(violations, "\n  - "))
	}
}
