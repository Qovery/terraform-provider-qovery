//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clusterResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	clusterResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	return resp.Schema
}

// testClusterFeatures returns the default features object with the given attributes replaced.
func testClusterFeatures(replace map[string]attr.Value) types.Object {
	attributes := maps.Clone(clusterFeaturesDefault().Attributes())
	maps.Copy(attributes, replace)
	return types.ObjectValueMust(createFeaturesAttrTypes(), attributes)
}

func testClusterKarpenterObject() types.Object {
	return types.ObjectValueMust(createKarpenterFeatureAttrTypes(), map[string]attr.Value{
		"disk_size_in_gib":             types.Int64Value(50),
		"disk_iops":                    types.Int64Null(),
		"disk_throughput":              types.Int64Null(),
		"default_service_architecture": types.StringValue("AMD64"),
		"qovery_node_pools":            types.ObjectNull(karpenterNodePoolsAttrTypes()),
	})
}

// newTestClusterState returns a state of the cluster schema holding the given root attributes,
// or the state of a cluster being created when attributes is nil.
func newTestClusterState(t *testing.T, sch schema.Schema, attributes map[string]attr.Value) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	state := tfsdk.State{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
	for name, value := range attributes {
		require.False(t, state.SetAttribute(ctx, path.Root(name), value).HasError())
	}
	return state
}

func TestClusterNodeSizingFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		CloudProvider string
		Mode          string
		Karpenter     bool
		Expected      clusterNodeSizing
	}{
		{CloudProvider: "AWS", Mode: "MANAGED", Expected: clusterNodeSizingNodeGroup},
		{CloudProvider: "AWS", Mode: "MANAGED", Karpenter: true, Expected: clusterNodeSizingServerSet},
		{CloudProvider: "SCW", Mode: "MANAGED", Expected: clusterNodeSizingNodeGroup},
		{CloudProvider: "AZURE", Mode: "MANAGED", Expected: clusterNodeSizingNodeGroup},
		{CloudProvider: "GCP", Mode: "MANAGED", Expected: clusterNodeSizingServerSet},
		{CloudProvider: "ON_PREMISE", Mode: "MANAGED", Expected: clusterNodeSizingServerSet},
		{CloudProvider: "AWS", Mode: "SELF_MANAGED", Expected: clusterNodeSizingServerSet},
		{CloudProvider: "SCW", Mode: "SELF_MANAGED", Expected: clusterNodeSizingServerSet},
		{CloudProvider: "AWS", Mode: "PARTIALLY_MANAGED", Expected: clusterNodeSizingServerSet},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.Expected, clusterNodeSizingFor(tc.CloudProvider, tc.Mode, tc.Karpenter), "%s %s karpenter=%t", tc.CloudProvider, tc.Mode, tc.Karpenter)
	}
}

func TestClusterNodeSizingDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sch := clusterResourceSchema(t)

	karpenterFeatures := testClusterFeatures(map[string]attr.Value{featureKeyKarpenter: testClusterKarpenterObject()})
	unknownKarpenterFeatures := testClusterFeatures(map[string]attr.Value{featureKeyKarpenter: types.ObjectUnknown(createKarpenterFeatureAttrTypes())})

	testCases := []struct {
		TestName      string
		CloudProvider types.String
		Mode          types.String
		Features      types.Object
		// Existing is the recorded state value, or nil on a create.
		Existing      *types.String
		Config        types.String
		Expected      types.String
		ExpectWarning bool
	}{
		{TestName: "aws_without_karpenter_plans_the_default", CloudProvider: types.StringValue("AWS"), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(), Expected: types.StringValue("t3.xlarge")},
		{TestName: "scaleway_plans_the_default", CloudProvider: types.StringValue("SCW"), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(), Expected: types.StringValue("DEV1-L")},
		{TestName: "azure_plans_the_default", CloudProvider: types.StringValue("AZURE"), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(), Expected: types.StringValue("Standard_DS2_v2")},
		{TestName: "null_mode_is_managed", CloudProvider: types.StringValue("SCW"), Mode: types.StringNull(), Features: clusterFeaturesDefault(), Expected: types.StringValue("DEV1-L")},
		{
			TestName: "removing_the_value_plans_the_default", CloudProvider: types.StringValue("SCW"), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(),
			Existing: new(types.StringValue("DEV1-XL")), Expected: types.StringValue("DEV1-L"),
		},
		{
			TestName: "configured_value_is_kept", CloudProvider: types.StringValue("SCW"), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(),
			Existing: new(types.StringValue("DEV1-L")), Config: types.StringValue("DEV1-XL"), Expected: types.StringValue("DEV1-XL"),
		},
		{
			TestName: "karpenter_keeps_the_recorded_value", CloudProvider: types.StringValue("AWS"), Mode: types.StringValue("MANAGED"), Features: karpenterFeatures,
			Existing: new(types.StringValue("KARPENTER")), Expected: types.StringValue("KARPENTER"),
		},
		{TestName: "karpenter_create_leaves_it_unknown", CloudProvider: types.StringValue("AWS"), Mode: types.StringValue("MANAGED"), Features: karpenterFeatures, Expected: types.StringUnknown()},
		{
			TestName: "karpenter_warns_on_a_configured_value", CloudProvider: types.StringValue("AWS"), Mode: types.StringValue("MANAGED"), Features: karpenterFeatures,
			Config: types.StringValue("t3a.medium"), Expected: types.StringValue("t3a.medium"), ExpectWarning: true,
		},
		{
			TestName: "gcp_keeps_the_recorded_value", CloudProvider: types.StringValue("GCP"), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(),
			Existing: new(types.StringValue("AUTO_PILOT")), Expected: types.StringValue("AUTO_PILOT"),
		},
		{
			TestName: "gcp_warns_on_a_configured_value", CloudProvider: types.StringValue("GCP"), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(),
			Config: types.StringValue("AUTO_PILOT"), Expected: types.StringValue("AUTO_PILOT"), ExpectWarning: true,
		},
		{
			TestName: "self_managed_keeps_the_recorded_value", CloudProvider: types.StringValue("AWS"), Mode: types.StringValue("SELF_MANAGED"), Features: clusterFeaturesDefault(),
			Existing: new(types.StringValue("m5.24xlarge")), Expected: types.StringValue("m5.24xlarge"),
		},
		{
			TestName: "partially_managed_keeps_null", CloudProvider: types.StringValue("AWS"), Mode: types.StringValue("PARTIALLY_MANAGED"), Features: clusterFeaturesDefault(),
			Existing: new(types.StringNull()), Expected: types.StringNull(),
		},
		{TestName: "unknown_cloud_provider_changes_nothing", CloudProvider: types.StringUnknown(), Mode: types.StringValue("MANAGED"), Features: clusterFeaturesDefault(), Expected: types.StringUnknown()},
		{TestName: "unknown_karpenter_changes_nothing", CloudProvider: types.StringValue("AWS"), Mode: types.StringValue("MANAGED"), Features: unknownKarpenterFeatures, Expected: types.StringUnknown()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			config := tc.Config
			planValue := types.StringUnknown()
			if !config.IsNull() {
				planValue = config
			}
			plan := newTestPlan(t, sch, map[string]attr.Value{
				"cloud_provider":  tc.CloudProvider,
				"kubernetes_mode": tc.Mode,
				"features":        tc.Features,
				"instance_type":   planValue,
			})

			state := newTestClusterState(t, sch, nil)
			stateValue := types.StringNull()
			if tc.Existing != nil {
				stateValue = *tc.Existing
				state = newTestClusterState(t, sch, map[string]attr.Value{"id": types.StringValue("cluster-123"), "instance_type": stateValue})
			}

			resp := &planmodifier.StringResponse{PlanValue: planValue}
			ClusterInstanceTypeDefault().PlanModifyString(ctx, planmodifier.StringRequest{
				Path:        path.Root("instance_type"),
				Plan:        plan,
				PlanValue:   planValue,
				State:       state,
				StateValue:  stateValue,
				ConfigValue: config,
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.Expected, resp.PlanValue)
			assert.Equal(t, tc.ExpectWarning, resp.Diagnostics.WarningsCount() == 1, "%v", resp.Diagnostics)
		})
	}
}

func TestClusterNodeSizingInt64Default(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sch := clusterResourceSchema(t)

	for _, tc := range []struct {
		Attribute string
		Modifier  planmodifier.Int64
		Expected  int64
	}{
		{Attribute: "disk_size", Modifier: ClusterNodeSizingInt64Default(clusterDiskSizeDefault), Expected: 40},
		{Attribute: "min_running_nodes", Modifier: ClusterNodeSizingInt64Default(clusterMinRunningNodesDefault), Expected: 3},
		{Attribute: "max_running_nodes", Modifier: ClusterNodeSizingInt64Default(clusterMaxRunningNodesDefault), Expected: 10},
	} {
		t.Run(tc.Attribute, func(t *testing.T) {
			t.Parallel()

			plan := newTestPlan(t, sch, map[string]attr.Value{
				"cloud_provider":  types.StringValue("AZURE"),
				"kubernetes_mode": types.StringValue("MANAGED"),
				"features":        clusterFeaturesDefault(),
			})
			state := newTestClusterState(t, sch, map[string]attr.Value{"id": types.StringValue("cluster-123"), tc.Attribute: types.Int64Value(7)})

			resp := &planmodifier.Int64Response{PlanValue: types.Int64Unknown()}
			tc.Modifier.PlanModifyInt64(ctx, planmodifier.Int64Request{
				Path:        path.Root(tc.Attribute),
				Plan:        plan,
				PlanValue:   types.Int64Unknown(),
				State:       state,
				StateValue:  types.Int64Value(7),
				ConfigValue: types.Int64Null(),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, types.Int64Value(tc.Expected), resp.PlanValue, "removing the value plans the Qovery default")
		})
	}
}

// TestClusterSchema_NodeSizingModifiers guards the wiring: the four node sizing attributes plan
// their default through the mode-aware modifier and nothing else keeps the recorded value.
func TestClusterSchema_NodeSizingModifiers(t *testing.T) {
	t.Parallel()
	sch := clusterResourceSchema(t)

	instanceType := sch.Attributes["instance_type"].(schema.StringAttribute)
	require.Len(t, instanceType.PlanModifiers, 1)
	assert.IsType(t, clusterNodeSizingDefaultModifier{}, instanceType.PlanModifiers[0])

	for _, name := range []string{"disk_size", "min_running_nodes", "max_running_nodes"} {
		attribute := sch.Attributes[name].(schema.Int64Attribute)
		require.Len(t, attribute.PlanModifiers, 1, name)
		assert.IsType(t, clusterNodeSizingDefaultModifier{}, attribute.PlanModifiers[0], name)
	}
}

// TestClusterSchema_FeaturesAndKedaDefaults guards the schema Defaults of features and keda: they
// must equal what the read reports for a cluster without any feature or KEDA, otherwise such a
// cluster would plan a change forever.
func TestClusterSchema_FeaturesAndKedaDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sch := clusterResourceSchema(t)

	features := sch.Attributes["features"].(schema.SingleNestedAttribute)
	require.NotNil(t, features.Default)
	var featuresDefault defaults.ObjectResponse
	features.Default.DefaultObject(ctx, defaults.ObjectRequest{}, &featuresDefault)
	assert.True(t, featuresDefault.PlanValue.Equal(clusterFeaturesDefault()))
	assert.True(t, clusterFeaturesFromResponse(nil, types.ObjectNull(createFeaturesAttrTypes()), clusterReadModeResource).Equal(clusterFeaturesDefault()),
		"a response without features reads as the default")

	keda := sch.Attributes["keda"].(schema.SingleNestedAttribute)
	require.NotNil(t, keda.Default)
	var kedaDefault defaults.ObjectResponse
	keda.Default.DefaultObject(ctx, defaults.ObjectRequest{}, &kedaDefault)
	assert.True(t, kedaDefault.PlanValue.Equal(clusterKedaDefault()))
	assert.True(t, fromQoveryClusterKeda(nil).Equal(clusterKedaDefault()), "a response without keda reads as the default")
}

func TestRejectKarpenterDisable(t *testing.T) {
	t.Parallel()

	withKarpenter := testClusterFeatures(map[string]attr.Value{featureKeyKarpenter: testClusterKarpenterObject()})

	testCases := []struct {
		TestName    string
		State       types.Object
		Plan        types.Object
		ExpectError bool
	}{
		{TestName: "removing_karpenter_is_rejected", State: withKarpenter, Plan: clusterFeaturesDefault(), ExpectError: true},
		{TestName: "omitted_features_is_rejected", State: withKarpenter, Plan: types.ObjectNull(createFeaturesAttrTypes()), ExpectError: true},
		{TestName: "keeping_karpenter_is_accepted", State: withKarpenter, Plan: withKarpenter},
		{TestName: "unknown_karpenter_is_accepted", State: withKarpenter, Plan: testClusterFeatures(map[string]attr.Value{featureKeyKarpenter: types.ObjectUnknown(createKarpenterFeatureAttrTypes())})},
		{TestName: "cluster_without_karpenter_is_accepted", State: clusterFeaturesDefault(), Plan: clusterFeaturesDefault()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			rejectKarpenterDisable(tc.State, tc.Plan, &diags)
			assert.Equal(t, tc.ExpectError, diags.HasError(), "%v", diags)
		})
	}
}

func TestRejectStaticIPDisable(t *testing.T) {
	t.Parallel()

	staticIP := func(value types.Bool) types.Object {
		return testClusterFeatures(map[string]attr.Value{featureKeyStaticIP: value})
	}

	testCases := []struct {
		TestName      string
		State         types.Object
		Plan          types.Object
		CloudProvider string
		ClusterState  string
		ExpectError   bool
	}{
		{TestName: "deployed_aws_cluster_is_rejected", State: staticIP(types.BoolValue(true)), Plan: staticIP(types.BoolValue(false)), CloudProvider: "AWS", ClusterState: "DEPLOYED", ExpectError: true},
		{TestName: "stopped_gcp_cluster_is_rejected", State: staticIP(types.BoolValue(true)), Plan: staticIP(types.BoolValue(false)), CloudProvider: "GCP", ClusterState: "STOPPED", ExpectError: true},
		{TestName: "deployed_azure_cluster_is_rejected", State: staticIP(types.BoolValue(true)), Plan: clusterFeaturesDefault(), CloudProvider: "AZURE", ClusterState: "DEPLOYED", ExpectError: true},
		{TestName: "never_deployed_cluster_is_accepted", State: staticIP(types.BoolValue(true)), Plan: staticIP(types.BoolValue(false)), CloudProvider: "AWS", ClusterState: "READY"},
		{TestName: "scaleway_is_accepted", State: staticIP(types.BoolValue(true)), Plan: staticIP(types.BoolValue(false)), CloudProvider: "SCW", ClusterState: "DEPLOYED"},
		{TestName: "enabling_is_left_to_the_api", State: staticIP(types.BoolValue(false)), Plan: staticIP(types.BoolValue(true)), CloudProvider: "AWS", ClusterState: "DEPLOYED"},
		{TestName: "unchanged_is_accepted", State: staticIP(types.BoolValue(true)), Plan: staticIP(types.BoolValue(true)), CloudProvider: "AWS", ClusterState: "DEPLOYED"},
		{TestName: "unknown_is_accepted", State: staticIP(types.BoolValue(true)), Plan: staticIP(types.BoolUnknown()), CloudProvider: "AWS", ClusterState: "DEPLOYED"},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			rejectStaticIPDisable(tc.State, tc.Plan, types.StringValue(tc.CloudProvider), types.StringValue(tc.ClusterState), &diags)
			assert.Equal(t, tc.ExpectError, diags.HasError(), "%v", diags)
		})
	}
}

func TestRejectGkeKmsKeyChange(t *testing.T) {
	t.Parallel()

	key := func(value types.String) types.Object {
		return testClusterFeatures(map[string]attr.Value{featureKeyGkeKmsKey: value})
	}
	const keyA = "projects/p/locations/l/keyRings/r/cryptoKeys/a"

	testCases := []struct {
		TestName      string
		State         types.Object
		Plan          types.Object
		ExpectError   bool
		ExpectInError string
	}{
		{TestName: "removing_the_key_is_rejected", State: key(types.StringValue(keyA)), Plan: key(types.StringNull()), ExpectError: true, ExpectInError: keyA},
		{TestName: "changing_the_key_is_rejected", State: key(types.StringValue(keyA)), Plan: key(types.StringValue(keyA + "-2")), ExpectError: true, ExpectInError: keyA},
		{TestName: "adding_a_key_is_rejected", State: key(types.StringNull()), Plan: key(types.StringValue(keyA)), ExpectError: true, ExpectInError: "holds none"},
		{TestName: "unchanged_key_is_accepted", State: key(types.StringValue(keyA)), Plan: key(types.StringValue(keyA))},
		{TestName: "empty_and_null_are_the_same", State: key(types.StringValue("")), Plan: key(types.StringNull())},
		{TestName: "unknown_key_is_accepted", State: key(types.StringValue(keyA)), Plan: key(types.StringUnknown())},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			rejectGkeKmsKeyChange(tc.State, tc.Plan, &diags)
			require.Equal(t, tc.ExpectError, diags.HasError(), "%v", diags)
			if tc.ExpectError {
				assert.Contains(t, diags.Errors()[0].Detail(), tc.ExpectInError)
				assert.Contains(t, diags.Errors()[0].Detail(), "terraform destroy -target", "the error names how to recreate the cluster")
			}
		})
	}
}

// TestRejectForbiddenClusterFeatureChanges runs the checks through a real plan and state: they
// only apply to an update.
func TestRejectForbiddenClusterFeatureChanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sch := clusterResourceSchema(t)

	withKarpenter := testClusterFeatures(map[string]attr.Value{featureKeyKarpenter: testClusterKarpenterObject()})
	plan := newTestPlan(t, sch, map[string]attr.Value{"cloud_provider": types.StringValue("AWS"), "features": clusterFeaturesDefault()})

	t.Run("create_is_accepted", func(t *testing.T) {
		t.Parallel()
		var diags diag.Diagnostics
		rejectForbiddenClusterFeatureChanges(ctx, newTestClusterState(t, sch, nil), plan, &diags)
		assert.False(t, diags.HasError(), "%v", diags)
	})

	t.Run("update_removing_karpenter_is_rejected", func(t *testing.T) {
		t.Parallel()
		state := newTestClusterState(t, sch, map[string]attr.Value{
			"id": types.StringValue("cluster-123"), "cloud_provider": types.StringValue("AWS"), "state": types.StringValue("READY"), "features": withKarpenter,
		})
		var diags diag.Diagnostics
		rejectForbiddenClusterFeatureChanges(ctx, state, plan, &diags)
		require.True(t, diags.HasError())
		assert.Equal(t, "Cannot disable Karpenter", diags.Errors()[0].Summary())
	})

	t.Run("update_from_a_state_without_features_is_accepted", func(t *testing.T) {
		t.Parallel()
		state := newTestClusterState(t, sch, map[string]attr.Value{"id": types.StringValue("cluster-123"), "cloud_provider": types.StringValue("AWS")})
		var diags diag.Diagnostics
		rejectForbiddenClusterFeatureChanges(ctx, state, plan, &diags)
		assert.False(t, diags.HasError(), "%v", diags)
	})
}
