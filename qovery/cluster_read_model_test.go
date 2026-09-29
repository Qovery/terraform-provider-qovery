//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/client"
)

func testClusterResponse(cloudProvider qovery.CloudVendorEnum, mode qovery.KubernetesEnum, features []qovery.ClusterFeatureResponse) *qovery.Cluster {
	return &qovery.Cluster{
		Id:              "cluster-123",
		Name:            "c",
		CloudProvider:   cloudProvider,
		Region:          "eu-west-3",
		Kubernetes:      &mode,
		InstanceType:    new("KARPENTER"),
		DiskSize:        new(int32(20)),
		MinRunningNodes: new(int32(3)),
		MaxRunningNodes: new(int32(2147483647)),
		Features:        features,
	}
}

func TestClusterNodeSizingFromResponse(t *testing.T) {
	t.Parallel()

	karpenter := []qovery.ClusterFeatureResponse{{Id: new(featureIdKarpenter)}}
	planned := Cluster{
		InstanceType:    types.StringValue("t3a.medium"),
		DiskSize:        types.Int64Value(50),
		MinRunningNodes: types.Int64Value(4),
		MaxRunningNodes: types.Int64Value(6),
	}
	unknownPlan := Cluster{
		InstanceType:    types.StringUnknown(),
		DiskSize:        types.Int64Unknown(),
		MinRunningNodes: types.Int64Unknown(),
		MaxRunningNodes: types.Int64Unknown(),
	}
	api := [4]attr.Value{types.StringValue("KARPENTER"), types.Int64Value(20), types.Int64Value(3), types.Int64Value(2147483647)}
	plan := [4]attr.Value{types.StringValue("t3a.medium"), types.Int64Value(50), types.Int64Value(4), types.Int64Value(6)}
	none := [4]attr.Value{types.StringNull(), types.Int64Null(), types.Int64Null(), types.Int64Null()}

	testCases := []struct {
		TestName string
		Response *qovery.Cluster
		Prior    Cluster
		Mode     clusterReadMode
		Expected [4]attr.Value
	}{
		{TestName: "node_group_reads_the_api", Response: testClusterResponse(qovery.CLOUDVENDORENUM_SCW, qovery.KUBERNETESENUM_MANAGED, nil), Prior: planned, Expected: api},
		{TestName: "aws_without_karpenter_reads_the_api", Response: testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_MANAGED, nil), Prior: planned, Expected: api},
		{TestName: "karpenter_keeps_the_prior_value", Response: testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_MANAGED, karpenter), Prior: planned, Expected: plan},
		{TestName: "karpenter_without_prior_reads_the_api", Response: testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_MANAGED, karpenter), Prior: unknownPlan, Expected: api},
		{TestName: "gcp_keeps_the_prior_value", Response: testClusterResponse(qovery.CLOUDVENDORENUM_GCP, qovery.KUBERNETESENUM_MANAGED, nil), Prior: planned, Expected: plan},
		{TestName: "self_managed_import_reads_the_api", Response: testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_SELF_MANAGED, nil), Prior: Cluster{}, Expected: api},
		{TestName: "partially_managed_reads_nothing", Response: testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_PARTIALLY_MANAGED, nil), Prior: Cluster{}, Expected: none},
		{TestName: "partially_managed_keeps_a_configured_value", Response: testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_PARTIALLY_MANAGED, nil), Prior: planned, Expected: plan},
		{TestName: "data_source_reports_the_karpenter_sentinels", Response: testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_MANAGED, karpenter), Prior: planned, Mode: clusterReadModeDataSource, Expected: api},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			instanceType, diskSize, minRunningNodes, maxRunningNodes := clusterNodeSizingFromResponse(tc.Response, tc.Prior, tc.Mode)
			assert.Equal(t, tc.Expected, [4]attr.Value{instanceType, diskSize, minRunningNodes, maxRunningNodes})
		})
	}
}

// TestConvertResponseToCluster_PartiallyManagedReadsTheDefaults checks that a partially managed
// cluster, which supports no feature and no KEDA, reads as the schema Defaults of features and
// keda, so that its plan stays empty.
func TestConvertResponseToCluster_PartiallyManagedReadsTheDefaults(t *testing.T) {
	t.Parallel()

	res := &client.ClusterResponse{
		OrganizationID:      "org-123",
		ClusterResponse:     testClusterResponse(qovery.CLOUDVENDORENUM_AWS, qovery.KUBERNETESENUM_PARTIALLY_MANAGED, nil),
		ClusterInfo:         makeTestClusterInfo("cred-123"),
		ClusterRoutingTable: &client.ClusterRoutingTable{},
	}

	cluster := convertResponseToCluster(context.Background(), res, Cluster{})

	assert.True(t, cluster.Features.Equal(clusterFeaturesDefault()), "features: %v", cluster.Features)
	assert.True(t, cluster.Keda.Equal(clusterKedaDefault()), "keda: %v", cluster.Keda)
	assert.True(t, cluster.InstanceType.IsNull())
}

func testExistingVpcResponse(rdsZoneA []string) []qovery.ClusterFeatureResponse {
	return []qovery.ClusterFeatureResponse{{
		Id: new(featureIdExistingVpc),
		ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(&qovery.ClusterFeatureResponseValueObject{
			ClusterFeatureAwsExistingVpcResponse: &qovery.ClusterFeatureAwsExistingVpcResponse{
				Value: qovery.ClusterFeatureAwsExistingVpc{
					AwsVpcEksId:        "vpc-1",
					EksSubnetsZoneAIds: []string{"subnet-a"},
					EksSubnetsZoneBIds: []string{"subnet-b"},
					EksSubnetsZoneCIds: []string{"subnet-c"},
					RdsSubnetsZoneAIds: rdsZoneA,
					RdsSubnetsZoneBIds: []string{},
				},
			},
		}),
	}}
}

// testExistingVpcFeatures returns planned features holding an existing_vpc block whose
// rds_subnets_zone_a_ids is rdsZoneA and whose other optional lists are null.
func testExistingVpcFeatures(rdsZoneA types.List) types.Object {
	attributes := map[string]attr.Value{}
	for name, attrType := range createExistingVpcFeatureAttrTypes() {
		switch attrType.(type) {
		case types.ListType:
			attributes[name] = types.ListNull(types.StringType)
		default:
			attributes[name] = types.BoolValue(false)
		}
	}
	attributes["aws_vpc_eks_id"] = types.StringValue("vpc-1")
	attributes["rds_subnets_zone_a_ids"] = rdsZoneA
	return testClusterFeatures(map[string]attr.Value{
		featureKeyExistingVpc: types.ObjectValueMust(createExistingVpcFeatureAttrTypes(), attributes),
	})
}

func existingVpcList(t *testing.T, features types.Object, name string) types.List {
	t.Helper()
	existingVpc, ok := features.Attributes()[featureKeyExistingVpc].(types.Object)
	require.True(t, ok)
	require.False(t, existingVpc.IsNull())
	return existingVpc.Attributes()[name].(types.List)
}

// TestClusterFeaturesFromResponse_ExistingVpcOptionalLists checks that an empty optional subnet
// list reads back in the shape of the prior value, and that a non-empty one always reads the API.
func TestClusterFeaturesFromResponse_ExistingVpcOptionalLists(t *testing.T) {
	t.Parallel()

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})

	testCases := []struct {
		TestName string
		API      []string
		Prior    types.Object
		Mode     clusterReadMode
		Expected types.List
	}{
		{TestName: "omitted_list_stays_null", API: []string{}, Prior: testExistingVpcFeatures(types.ListNull(types.StringType)), Expected: types.ListNull(types.StringType)},
		{TestName: "empty_list_stays_empty", API: []string{}, Prior: testExistingVpcFeatures(emptyList), Expected: emptyList},
		{TestName: "import_reads_an_empty_list_as_null", API: []string{}, Prior: types.ObjectNull(createFeaturesAttrTypes()), Expected: types.ListNull(types.StringType)},
		{
			TestName: "list_set_outside_terraform_is_read", API: []string{"subnet-rds"}, Prior: testExistingVpcFeatures(types.ListNull(types.StringType)),
			Expected: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-rds")}),
		},
		{TestName: "data_source_reports_an_empty_list", API: []string{}, Prior: types.ObjectNull(createFeaturesAttrTypes()), Mode: clusterReadModeDataSource, Expected: emptyList},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			features := clusterFeaturesFromResponse(testExistingVpcResponse(tc.API), tc.Prior, tc.Mode)
			assert.Equal(t, tc.Expected, existingVpcList(t, features, "rds_subnets_zone_a_ids"))
		})
	}
}

func TestClusterFeaturesFromResponse_GcpAdditionalIpRangePodsNames(t *testing.T) {
	t.Parallel()

	response := []qovery.ClusterFeatureResponse{{
		Id: new(featureIdExistingVpc),
		ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(&qovery.ClusterFeatureResponseValueObject{
			ClusterFeatureGcpExistingVpcResponse: &qovery.ClusterFeatureGcpExistingVpcResponse{
				Value: qovery.ClusterFeatureGcpExistingVpc{VpcName: "vpc", AdditionalIpRangePodsNames: []string{}},
			},
		}),
	}}
	gcpFeatures := func(additional types.List) types.Object {
		return testClusterFeatures(map[string]attr.Value{
			featureKeyGcpExistingVpc: types.ObjectValueMust(createGcpExistingVpcFeatureAttrTypes(), map[string]attr.Value{
				"vpc_name":                       types.StringValue("vpc"),
				"vpc_project_id":                 types.StringNull(),
				"subnetwork_name":                types.StringNull(),
				"ip_range_services_name":         types.StringNull(),
				"ip_range_pods_name":             types.StringNull(),
				"additional_ip_range_pods_names": additional,
				"private_nodes":                  types.BoolValue(false),
			}),
		})
	}
	additionalOf := func(features types.Object) types.List {
		return features.Attributes()[featureKeyGcpExistingVpc].(types.Object).Attributes()["additional_ip_range_pods_names"].(types.List)
	}

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	assert.Equal(t, types.ListNull(types.StringType), additionalOf(clusterFeaturesFromResponse(response, gcpFeatures(types.ListNull(types.StringType)), clusterReadModeResource)))
	assert.Equal(t, emptyList, additionalOf(clusterFeaturesFromResponse(response, gcpFeatures(emptyList), clusterReadModeResource)),
		"an explicit [] no longer reads back as null, which failed the apply as an inconsistent result")

	private := clusterFeaturesFromResponse(response, gcpFeatures(emptyList), clusterReadModeResource).
		Attributes()[featureKeyGcpExistingVpc].(types.Object).Attributes()["private_nodes"]
	assert.Equal(t, types.BoolValue(false), private, "an omitted private_nodes reads as the false default")
}

func TestClusterFeaturesFromResponse_VpcSubnetWithoutValueReadsTheDefault(t *testing.T) {
	t.Parallel()

	for _, value := range []*qovery.ClusterFeatureStringResponse{nil, {Value: ""}} {
		response := []qovery.ClusterFeatureResponse{{
			Id:          new(featureIdVpcSubnet),
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(&qovery.ClusterFeatureResponseValueObject{ClusterFeatureStringResponse: value}),
		}}
		features := clusterFeaturesFromResponse(response, types.ObjectNull(createFeaturesAttrTypes()), clusterReadModeResource)
		assert.Equal(t, types.StringValue(clusterFeatureVpcSubnetDefault), features.Attributes()[featureKeyVpcSubnet])
	}
}

// TestToQoveryClusterFeatures_VpcSubnetOnlyForAwsManaged checks that VPC_SUBNET only goes to an
// EKS cluster: q-core rejects the feature when it creates any other cluster, even with the
// default subnet, so the features schema Default must not send it there.
func TestToQoveryClusterFeatures_VpcSubnetOnlyForAwsManaged(t *testing.T) {
	t.Parallel()

	custom := testClusterFeatures(map[string]attr.Value{featureKeyVpcSubnet: types.StringValue("10.42.0.0/16")})

	testCases := []struct {
		TestName      string
		Features      types.Object
		Mode          string
		CloudProvider string
		ExpectSent    bool
		ExpectError   string
	}{
		{TestName: "aws_managed_sends_the_default", Features: clusterFeaturesDefault(), Mode: "MANAGED", CloudProvider: "AWS", ExpectSent: true},
		{TestName: "aws_managed_sends_a_custom_subnet", Features: custom, Mode: "MANAGED", CloudProvider: "AWS", ExpectSent: true},
		{TestName: "scaleway_does_not_send_the_default", Features: clusterFeaturesDefault(), Mode: "MANAGED", CloudProvider: "SCW"},
		{TestName: "azure_does_not_send_the_default", Features: clusterFeaturesDefault(), Mode: "MANAGED", CloudProvider: "AZURE"},
		{TestName: "aws_self_managed_does_not_send_the_default", Features: clusterFeaturesDefault(), Mode: "SELF_MANAGED", CloudProvider: "AWS"},
		{TestName: "scaleway_rejects_a_custom_subnet", Features: custom, Mode: "MANAGED", CloudProvider: "SCW", ExpectError: "only supported for AWS clusters"},
		{TestName: "gcp_rejects_a_custom_subnet", Features: custom, Mode: "MANAGED", CloudProvider: "GCP", ExpectError: "not supported for GCP clusters"},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			features, err := toQoveryClusterFeatures(tc.Features, tc.Mode, tc.CloudProvider)
			if tc.ExpectError != "" {
				require.ErrorContains(t, err, tc.ExpectError)
				return
			}
			require.NoError(t, err)

			sent := false
			for _, feature := range features {
				sent = sent || feature.GetId() == featureIdVpcSubnet
			}
			assert.Equal(t, tc.ExpectSent, sent)
		})
	}
}

// TestCluster_toUpsertClusterRequest_PartiallyManagedKeda checks that the keda default, disabled,
// is accepted on a partially managed cluster, and that only enabling KEDA is rejected.
func TestCluster_toUpsertClusterRequest_PartiallyManagedKeda(t *testing.T) {
	t.Parallel()

	metalLbAttrTypes := map[string]attr.Type{"ip_address_pools": types.ListType{ElemType: types.StringType}}
	infraChartsAttrTypes := createInfrastructureChartsParametersAttrTypes()
	infraChartsValues := map[string]attr.Value{}
	for name, attrType := range infraChartsAttrTypes {
		infraChartsValues[name] = types.ObjectNull(attrType.(types.ObjectType).AttrTypes)
	}
	infraChartsValues[infraChartsMetalLbKey] = types.ObjectValueMust(metalLbAttrTypes, map[string]attr.Value{
		"ip_address_pools": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.168.1.1-192.168.1.10")}),
	})

	cluster := Cluster{
		OrganizationId:                 types.StringValue("org-123"),
		CredentialsId:                  types.StringValue("cred-123"),
		Name:                           types.StringValue("c"),
		CloudProvider:                  types.StringValue("AWS"),
		Region:                         types.StringValue("on-premise"),
		KubernetesMode:                 types.StringValue("PARTIALLY_MANAGED"),
		Kubeconfig:                     types.StringValue("fake-kubeconfig"),
		State:                          types.StringValue("DEPLOYED"),
		Features:                       clusterFeaturesDefault(),
		Keda:                           clusterKedaDefault(),
		InfrastructureChartsParameters: types.ObjectValueMust(infraChartsAttrTypes, infraChartsValues),
	}

	_, err := cluster.toUpsertClusterRequest(nil)
	require.NoError(t, err, "the keda and features defaults are what a partially managed cluster has")

	cluster.Keda = types.ObjectValueMust(createKedaAttrTypes(), map[string]attr.Value{"enabled": types.BoolValue(true)})
	_, err = cluster.toUpsertClusterRequest(nil)
	require.ErrorContains(t, err, "keda is not supported")
}

func TestClusterResource_UpgradeStateV0ToV1_ExistingVpcSubnetLists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	declared := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-rds")})
	// 0.x stored [] for every optional subnet list the configuration omitted.
	priorFeatures := func(rdsZoneA types.List) types.Object {
		features := testExistingVpcFeatures(rdsZoneA)
		existingVpc := maps.Clone(features.Attributes()[featureKeyExistingVpc].(types.Object).Attributes())
		for _, name := range existingVpcOptionalSubnetLists[1:] {
			existingVpc[name] = emptyList
		}
		return testClusterFeatures(map[string]attr.Value{featureKeyExistingVpc: types.ObjectValueMust(createExistingVpcFeatureAttrTypes(), existingVpc)})
	}

	for _, tc := range []struct {
		TestName string
		Prior    types.List
		Expected types.List
	}{
		{TestName: "empty_list_written_by_0x_becomes_null", Prior: emptyList, Expected: types.ListNull(types.StringType)},
		{TestName: "declared_list_is_kept", Prior: declared, Expected: declared},
	} {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			upgrader := clusterResource{}.UpgradeState(ctx)[0]
			priorState := tfsdk.State{Schema: *upgrader.PriorSchema, Raw: tftypes.NewValue(upgrader.PriorSchema.Type().TerraformType(ctx), nil)}
			require.False(t, priorState.SetAttribute(ctx, path.Root("id"), "cluster-123").HasError())
			require.False(t, priorState.SetAttribute(ctx, path.Root("features"), priorFeatures(tc.Prior)).HasError())

			var schemaResp resource.SchemaResponse
			clusterResource{}.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &priorState}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

			var features types.Object
			require.False(t, resp.State.GetAttribute(ctx, path.Root("features"), &features).HasError())
			assert.Equal(t, tc.Expected, existingVpcList(t, features, "rds_subnets_zone_a_ids"))
			for _, name := range existingVpcOptionalSubnetLists[1:] {
				assert.True(t, existingVpcList(t, features, name).IsNull(), name)
			}
		})
	}
}

func TestClusterFeaturesFromResponse_EmptyGkeKmsKeyReadsAsNull(t *testing.T) {
	t.Parallel()

	for value, expected := range map[string]types.String{
		"":                 types.StringNull(),
		"projects/p/key/k": types.StringValue("projects/p/key/k"),
	} {
		response := []qovery.ClusterFeatureResponse{{
			Id:          new(featureIdGkeKmsKey),
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(&qovery.ClusterFeatureResponseValueObject{ClusterFeatureStringResponse: &qovery.ClusterFeatureStringResponse{Value: value}}),
		}}
		features := clusterFeaturesFromResponse(response, types.ObjectNull(createFeaturesAttrTypes()), clusterReadModeResource)
		assert.Equal(t, expected, features.Attributes()[featureKeyGkeKmsKey], "API value %q", value)
	}
}

func TestUpgradeGkeKmsKeyFrom0x(t *testing.T) {
	t.Parallel()

	withKey := func(value types.String) types.Object {
		return testClusterFeatures(map[string]attr.Value{featureKeyGkeKmsKey: value})
	}
	assert.True(t, upgradeGkeKmsKeyFrom0x(withKey(types.StringValue(""))).Equal(withKey(types.StringNull())), "the empty key 0.x stored becomes null")
	assert.True(t, upgradeGkeKmsKeyFrom0x(withKey(types.StringValue("k"))).Equal(withKey(types.StringValue("k"))), "a key is kept")
	assert.True(t, upgradeGkeKmsKeyFrom0x(types.ObjectNull(createFeaturesAttrTypes())).IsNull(), "a state without features is kept")
}
