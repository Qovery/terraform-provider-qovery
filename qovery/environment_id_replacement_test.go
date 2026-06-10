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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
// PR #588 finding #1 — vpc_subnet phantom-replacement on provider upgrade.
//
// #588 did two things at once:
//   1. added a replace modifier to features.vpc_subnet (it had none before), and
//   2. changed the Read fallback for vpc_subnet from "" to clusterFeatureVpcSubnetDefault
//      ("10.0.0.0/16") — see fromQoveryClusterFeatures (resource_cluster_model.go).
//
// A cluster created by a pre-#588 provider stored features.vpc_subnet="" in state
// (GCP / any cluster whose API response carries no VPC_SUBNET feature). After the
// upgrade the schema Default makes the planned value "10.0.0.0/16". On a normal
// `terraform apply` refresh rewrites state "" -> default first, so plan==state and
// no replacement happens. But on `terraform plan/apply -refresh=false` (common in
// CI) the stale "" survives, so a naive replace modifier sees a *known* plan
// ("10.0.0.0/16") that differs from the *known* state ("") and forces a
// DESTROY+RECREATE of the cluster — a phantom change driven purely by the
// read-default flip, not by user intent.
//
// Fix: features.vpc_subnet must use a replace modifier that treats "" and the
// schema default as equivalent, so the representation flip never triggers
// replacement while a genuine subnet change still does. These two tests pull the
// ACTUAL modifiers wired on features.vpc_subnet from the cluster schema, so they
// fail if anyone reverts to an unsafe modifier.
// ----------------------------------------------------------------------------

// clusterVpcSubnetPlanModifiers returns the plan modifiers wired on the nested
// features.vpc_subnet attribute of the cluster resource schema.
func clusterVpcSubnetPlanModifiers(t *testing.T) []planmodifier.String {
	t.Helper()
	var resp resource.SchemaResponse
	clusterResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	features, ok := resp.Schema.Attributes["features"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("features is not a SingleNestedAttribute")
	}
	vpcSubnet, ok := features.Attributes["vpc_subnet"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("features.vpc_subnet is not a StringAttribute")
	}
	return vpcSubnet.PlanModifiers
}

// runVpcSubnetModifiers runs every modifier wired on features.vpc_subnet with the
// given state/plan values and reports whether any of them requested replacement.
func runVpcSubnetModifiers(t *testing.T, state, plan types.String) bool {
	t.Helper()
	requiresReplace := false
	for _, mod := range clusterVpcSubnetPlanModifiers(t) {
		resp := &planmodifier.StringResponse{PlanValue: plan}
		mod.PlanModifyString(context.Background(), planmodifier.StringRequest{
			Config:      buildConfig(&plan),
			ConfigValue: plan,
			State:       buildState(&state),
			StateValue:  state,
			Plan:        buildPlan(&plan),
			PlanValue:   plan,
		}, resp)
		if resp.RequiresReplace {
			requiresReplace = true
		}
	}
	return requiresReplace
}

// TestClusterVpcSubnet_LegacyEmptyState_DoesNotForceReplacement asserts the safe
// behavior: a legacy state value of "" against the planned schema default
// "10.0.0.0/16" must NOT force a cluster replacement (the phantom-change case).
func TestClusterVpcSubnet_LegacyEmptyState_DoesNotForceReplacement(t *testing.T) {
	t.Parallel()

	legacyState := types.StringValue("")               // written by a pre-#588 provider
	plannedDefault := types.StringValue("10.0.0.0/16") // clusterFeatureVpcSubnetDefault

	assert.False(t, runVpcSubnetModifiers(t, legacyState, plannedDefault),
		"PR#588 finding #1: legacy vpc_subnet=\"\" -> default \"10.0.0.0/16\" must NOT force cluster replacement")
}

// TestClusterVpcSubnet_RealChange_ForcesReplacement guards the other direction:
// vpc_subnet is immutable, so a genuine change between two distinct known CIDRs
// must still force replacement (the fix must not over-suppress).
func TestClusterVpcSubnet_RealChange_ForcesReplacement(t *testing.T) {
	t.Parallel()

	oldCidr := types.StringValue("10.0.0.0/16")
	newCidr := types.StringValue("10.1.0.0/16")

	assert.True(t, runVpcSubnetModifiers(t, oldCidr, newCidr),
		"a genuine vpc_subnet change must force cluster replacement")
}

// ----------------------------------------------------------------------------
// PR #588 finding #2 — features.nat_gateways must be Optional+Computed.
//
// The Read path (fromQoveryClusterFeatures) populates features.nat_gateways with a
// non-null object whenever the API reports a GCP cluster with NAT static IPs
// enabled. If the attribute is Optional-only (not Computed), the framework pins its
// planned value to config: a user importing / managing a Console-created GCP cluster
// whose config omits the block gets a perpetual plan diff, a silent disable of the
// static egress IPs on the next apply, or a "Provider produced inconsistent result
// after apply" error when the backend keeps NAT enabled. The data source already
// declares it Optional+Computed; the resource must match. ObjectDefault fills the
// omitted block with {static_ips_count: 1}, so omission means "reset to default"
// (value-based semantics) instead of "keep whatever state has".
// ----------------------------------------------------------------------------

// clusterNatGatewaysAttribute returns the nested features.nat_gateways attribute
// of the cluster resource schema.
func clusterNatGatewaysAttribute(t *testing.T) schema.SingleNestedAttribute {
	t.Helper()
	var resp resource.SchemaResponse
	clusterResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	features, ok := resp.Schema.Attributes["features"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("features is not a SingleNestedAttribute")
	}
	nat, ok := features.Attributes["nat_gateways"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("features.nat_gateways is not a SingleNestedAttribute")
	}
	return nat
}

// TestClusterNatGateways_IsOptionalAndComputed asserts the fix for finding #2: the
// resource attribute must be Optional AND Computed so the framework can absorb a
// value the API returns for an unconfigured block.
func TestClusterNatGateways_IsOptionalAndComputed(t *testing.T) {
	t.Parallel()

	nat := clusterNatGatewaysAttribute(t)
	assert.True(t, nat.Optional, "features.nat_gateways must be Optional")
	assert.True(t, nat.Computed,
		"PR#588 finding #2: features.nat_gateways must be Computed so the framework can absorb a server-set value for import/Console GCP clusters")
}

// TestClusterNatGateways_HasObjectDefault asserts that features.nat_gateways carries an
// ObjectDefault (not UseStateForUnknown) so omitting the block in config resets the value
// to the default {static_ips_count: 1} rather than keeping the previous state. This is
// the semantic flip introduced by the value-based nat_gateways design.
func TestClusterNatGateways_HasObjectDefault(t *testing.T) {
	t.Parallel()

	nat := clusterNatGatewaysAttribute(t)
	require.NotNil(t, nat.Default,
		"features.nat_gateways must carry an ObjectDefault so omitting the block resets to {static_ips_count:1}")

	ctx := context.Background()
	req := defaults.ObjectRequest{}
	resp := &defaults.ObjectResponse{}
	nat.Default.DefaultObject(ctx, req, resp)
	require.False(t, resp.Diagnostics.HasError(), "DefaultObject must not produce diagnostics: %v", resp.Diagnostics)

	defaultVal := resp.PlanValue
	require.False(t, defaultVal.IsNull(), "default value must not be null")
	require.False(t, defaultVal.IsUnknown(), "default value must not be unknown")

	attrs := defaultVal.Attributes()
	countAttr, ok := attrs["static_ips_count"].(types.Int64)
	require.True(t, ok, "default value must have static_ips_count as Int64")
	assert.Equal(t, int64(1), countAttr.ValueInt64(),
		"default nat_gateways must be {static_ips_count: 1}")

	// Confirm the type structure matches createNatGatewaysFeatureAttrTypes().
	expectedAttrTypes := createNatGatewaysFeatureAttrTypes()
	assert.Equal(t, types.ObjectType{AttrTypes: expectedAttrTypes}, defaultVal.Type(ctx),
		"default object type must match the nat_gateways attribute type schema")
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
