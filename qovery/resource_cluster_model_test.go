//go:build unit || !integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/client"
)

func makeTestClusterInfo(credID string) *qovery.ClusterCloudProviderInfo {
	id := credID
	return &qovery.ClusterCloudProviderInfo{
		Credentials: &qovery.ClusterCloudProviderInfoCredentials{Id: &id},
	}
}

func TestCluster_toUpsertClusterRequest_KarpenterValidation(t *testing.T) {
	tests := []struct {
		name        string
		cluster     Cluster
		state       *Cluster
		expectError bool
		errorMsg    string
	}{
		{
			name: "new AWS EKS cluster without Karpenter should fail",
			cluster: Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("AWS"),
				Region:          types.StringValue("us-east-1"),
				KubernetesMode:  types.StringValue("MANAGED"),
				InstanceType:    types.StringValue("T3A_MEDIUM"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:           types.StringValue("DEPLOYED"),
				Features:        types.ObjectNull(map[string]attr.Type{}),
			},
			state:       nil, // New cluster
			expectError: true,
			errorMsg:    "Karpenter is required for new EKS (AWS MANAGED) clusters",
		},
		{
			name: "new AWS EKS cluster with Karpenter should succeed",
			cluster: Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("AWS"),
				Region:          types.StringValue("us-east-1"),
				KubernetesMode:  types.StringValue("MANAGED"),
				State:           types.StringValue("DEPLOYED"),
				InstanceType:    types.StringUnknown(),
				MinRunningNodes: types.Int64Unknown(),
				MaxRunningNodes: types.Int64Unknown(),
				DiskSize:        types.Int64Unknown(),
				Features: types.ObjectValueMust(
					map[string]attr.Type{
						"karpenter": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"spot_enabled":                 types.BoolType,
								"disk_size_in_gib":             types.Int64Type,
								"default_service_architecture": types.StringType,
								"qovery_node_pools": types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"requirements": types.ListType{
											ElemType: types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"key":      types.StringType,
													"operator": types.StringType,
													"values":   types.ListType{ElemType: types.StringType},
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]attr.Value{
						"karpenter": types.ObjectValueMust(
							map[string]attr.Type{
								"spot_enabled":                 types.BoolType,
								"disk_size_in_gib":             types.Int64Type,
								"default_service_architecture": types.StringType,
								"qovery_node_pools": types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"requirements": types.ListType{
											ElemType: types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"key":      types.StringType,
													"operator": types.StringType,
													"values":   types.ListType{ElemType: types.StringType},
												},
											},
										},
									},
								},
							},
							map[string]attr.Value{
								"spot_enabled":                 types.BoolValue(true),
								"disk_size_in_gib":             types.Int64Value(50),
								"default_service_architecture": types.StringValue("AMD64"),
								"qovery_node_pools": types.ObjectValueMust(
									map[string]attr.Type{
										"requirements": types.ListType{
											ElemType: types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"key":      types.StringType,
													"operator": types.StringType,
													"values":   types.ListType{ElemType: types.StringType},
												},
											},
										},
									},
									map[string]attr.Value{
										"requirements": types.ListValueMust(
											types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"key":      types.StringType,
													"operator": types.StringType,
													"values":   types.ListType{ElemType: types.StringType},
												},
											},
											[]attr.Value{
												types.ObjectValueMust(
													map[string]attr.Type{
														"key":      types.StringType,
														"operator": types.StringType,
														"values":   types.ListType{ElemType: types.StringType},
													},
													map[string]attr.Value{
														"key":      types.StringValue("InstanceFamily"),
														"operator": types.StringValue("In"),
														"values":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("t3a")}),
													},
												),
												types.ObjectValueMust(
													map[string]attr.Type{
														"key":      types.StringType,
														"operator": types.StringType,
														"values":   types.ListType{ElemType: types.StringType},
													},
													map[string]attr.Value{
														"key":      types.StringValue("InstanceSize"),
														"operator": types.StringValue("In"),
														"values":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("medium")}),
													},
												),
												types.ObjectValueMust(
													map[string]attr.Type{
														"key":      types.StringType,
														"operator": types.StringType,
														"values":   types.ListType{ElemType: types.StringType},
													},
													map[string]attr.Value{
														"key":      types.StringValue("Arch"),
														"operator": types.StringValue("In"),
														"values":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("amd64")}),
													},
												),
											},
										),
									},
								),
							},
						),
					},
				),
			},
			state:       nil, // New cluster
			expectError: false,
		},
		{
			name: "existing AWS EKS cluster without Karpenter should succeed (update allowed)",
			cluster: Cluster{
				Id:              types.StringValue("cluster-123"),
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("AWS"),
				Region:          types.StringValue("us-east-1"),
				KubernetesMode:  types.StringValue("MANAGED"),
				InstanceType:    types.StringValue("T3A_LARGE"),
				MinRunningNodes: types.Int64Value(5),
				MaxRunningNodes: types.Int64Value(15),
				State:           types.StringValue("DEPLOYED"),
				Features:        types.ObjectNull(map[string]attr.Type{}),
			},
			state: &Cluster{
				Id:              types.StringValue("cluster-123"),
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("AWS"),
				Region:          types.StringValue("us-east-1"),
				KubernetesMode:  types.StringValue("MANAGED"),
				InstanceType:    types.StringValue("T3A_MEDIUM"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:           types.StringValue("DEPLOYED"),
				Features:        types.ObjectNull(map[string]attr.Type{}),
			},
			expectError: false,
		},
		{
			name: "new GCP cluster without Karpenter should succeed (not AWS)",
			cluster: Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("GCP"),
				Region:          types.StringValue("us-central1"),
				KubernetesMode:  types.StringValue("MANAGED"),
				InstanceType:    types.StringValue("N2_STANDARD_2"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:           types.StringValue("DEPLOYED"),
				Features:        types.ObjectNull(map[string]attr.Type{}),
			},
			state:       nil, // New cluster
			expectError: false,
		},
		{
			name: "new AWS SELF_MANAGED cluster without Karpenter should succeed (not MANAGED)",
			cluster: Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("AWS"),
				Region:          types.StringValue("us-east-1"),
				KubernetesMode:  types.StringValue("SELF_MANAGED"),
				InstanceType:    types.StringValue("T3A_MEDIUM"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:           types.StringValue("DEPLOYED"),
				Features:        types.ObjectNull(map[string]attr.Type{}),
			},
			state:       nil, // New cluster
			expectError: false,
		},
		{
			name: "new AWS PARTIALLY_MANAGED cluster without Karpenter should succeed (not MANAGED)",
			cluster: Cluster{
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("AWS"),
				Region:         types.StringValue("us-east-1"),
				KubernetesMode: types.StringValue("PARTIALLY_MANAGED"),
				Kubeconfig:     types.StringValue("fake-kubeconfig"),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
				InfrastructureChartsParameters: types.ObjectValueMust(
					map[string]attr.Type{
						"nginx_parameters":        types.ObjectType{AttrTypes: map[string]attr.Type{}},
						"cert_manager_parameters": types.ObjectType{AttrTypes: map[string]attr.Type{}},
						"metal_lb_parameters": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"ip_address_pools": types.ListType{ElemType: types.StringType},
							},
						},
					},
					map[string]attr.Value{
						"nginx_parameters":        types.ObjectNull(map[string]attr.Type{}),
						"cert_manager_parameters": types.ObjectNull(map[string]attr.Type{}),
						"metal_lb_parameters": types.ObjectValueMust(
							map[string]attr.Type{
								"ip_address_pools": types.ListType{ElemType: types.StringType},
							},
							map[string]attr.Value{
								"ip_address_pools": types.ListValueMust(
									types.StringType,
									[]attr.Value{types.StringValue("192.168.1.1-192.168.1.10")},
								),
							},
						),
					},
				),
			},
			state:       nil, // New cluster
			expectError: false,
		},
		{
			name: "new Azure cluster without Karpenter should succeed (not AWS)",
			cluster: Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("AZURE"),
				Region:          types.StringValue("eastus"),
				KubernetesMode:  types.StringValue("MANAGED"),
				InstanceType:    types.StringValue("STANDARD_D2S_V3"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:           types.StringValue("DEPLOYED"),
				Features:        types.ObjectNull(map[string]attr.Type{}),
			},
			state:       nil, // New cluster
			expectError: false,
		},
		{
			name: "new Scaleway cluster without Karpenter should succeed (not AWS)",
			cluster: Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("SCW"),
				Region:          types.StringValue("fr-par"),
				KubernetesMode:  types.StringValue("MANAGED"),
				InstanceType:    types.StringValue("DEV1_M"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:           types.StringValue("DEPLOYED"),
				Features:        types.ObjectNull(map[string]attr.Type{}),
			},
			state:       nil, // New cluster
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.cluster.toUpsertClusterRequest(tt.state)

			if tt.expectError {
				require.Error(t, err, "Expected an error but got none")
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				assert.Nil(t, result, "Result should be nil when there's an error")
			} else {
				require.NoError(t, err, "Expected no error but got: %v", err)
				assert.NotNil(t, result, "Result should not be nil when there's no error")
			}
		})
	}
}

func TestCluster_toUpsertClusterRequest_KarpenterValidationWithEmptyFeatures(t *testing.T) {
	// Test that an empty features object (not null) also fails validation for new AWS EKS clusters
	cluster := Cluster{
		OrganizationId:  types.StringValue("org-123"),
		CredentialsId:   types.StringValue("cred-123"),
		Name:            types.StringValue("test-cluster"),
		CloudProvider:   types.StringValue("AWS"),
		Region:          types.StringValue("us-east-1"),
		KubernetesMode:  types.StringValue("MANAGED"),
		InstanceType:    types.StringValue("T3A_MEDIUM"),
		MinRunningNodes: types.Int64Value(3),
		MaxRunningNodes: types.Int64Value(10),
		State:           types.StringValue("DEPLOYED"),
		Features: types.ObjectValueMust(
			map[string]attr.Type{},
			map[string]attr.Value{},
		),
	}

	result, err := cluster.toUpsertClusterRequest(nil)

	require.Error(t, err, "Expected an error for new AWS EKS cluster without Karpenter")
	assert.Contains(t, err.Error(), "Karpenter is required for new EKS (AWS MANAGED) clusters")
	assert.Nil(t, result)
}

// buildGcpExistingVpcFeatureObject creates a Terraform features object with a gcp_existing_vpc block.
func buildGcpExistingVpcFeatureObject(
	vpcName string,
	vpcProjectID types.String,
	subnetworkName types.String,
	ipRangeServicesName types.String,
	ipRangePodsName types.String,
	additionalIpRangePodsNames types.List,
) types.Object {
	gcpVpcAttrTypes := createGcpExistingVpcFeatureAttrTypes()
	gcpVpcObj := types.ObjectValueMust(gcpVpcAttrTypes, map[string]attr.Value{
		"vpc_name":                       types.StringValue(vpcName),
		"vpc_project_id":                 vpcProjectID,
		"subnetwork_name":                subnetworkName,
		"ip_range_services_name":         ipRangeServicesName,
		"ip_range_pods_name":             ipRangePodsName,
		"additional_ip_range_pods_names": additionalIpRangePodsNames,
	})

	return types.ObjectValueMust(
		createFeaturesAttrTypes(),
		map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue(clusterFeatureVpcSubnetDefault),
			featureKeyStaticIP:       types.BoolValue(false),
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: gcpVpcObj,
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		},
	)
}

func TestToQoveryClusterFeatures_GcpExistingVpc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                       string
		vpcName                    string
		vpcProjectID               types.String
		subnetworkName             types.String
		ipRangeServicesName        types.String
		ipRangePodsName            types.String
		additionalIpRangePodsNames types.List
	}{
		{
			name:                       "all fields populated",
			vpcName:                    "my-vpc",
			vpcProjectID:               types.StringValue("my-project"),
			subnetworkName:             types.StringValue("my-subnet"),
			ipRangeServicesName:        types.StringValue("gke-services"),
			ipRangePodsName:            types.StringValue("gke-pods"),
			additionalIpRangePodsNames: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("extra-1"), types.StringValue("extra-2")}),
		},
		{
			name:                       "only required vpc_name",
			vpcName:                    "minimal-vpc",
			vpcProjectID:               types.StringNull(),
			subnetworkName:             types.StringNull(),
			ipRangeServicesName:        types.StringNull(),
			ipRangePodsName:            types.StringNull(),
			additionalIpRangePodsNames: types.ListNull(types.StringType),
		},
		{
			name:                       "empty additional_ip_range_pods_names list",
			vpcName:                    "vpc-empty-list",
			vpcProjectID:               types.StringNull(),
			subnetworkName:             types.StringNull(),
			ipRangeServicesName:        types.StringNull(),
			ipRangePodsName:            types.StringNull(),
			additionalIpRangePodsNames: types.ListValueMust(types.StringType, []attr.Value{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			featuresObj := buildGcpExistingVpcFeatureObject(
				tt.vpcName, tt.vpcProjectID, tt.subnetworkName,
				tt.ipRangeServicesName, tt.ipRangePodsName, tt.additionalIpRangePodsNames,
			)

			features, err := toQoveryClusterFeatures(featuresObj, "MANAGED")
			require.NoError(t, err)

			// Find the EXISTING_VPC feature (GCP VPC uses the shared ID)
			var gcpFeature *qovery.ClusterRequestFeaturesInner
			for i := range features {
				if features[i].GetId() == featureIdExistingVpc {
					gcpFeature = &features[i]
					break
				}
			}
			require.NotNil(t, gcpFeature, "expected EXISTING_VPC feature to be present")

			gcpValue := gcpFeature.GetValue().ClusterFeatureGcpExistingVpc
			require.NotNil(t, gcpValue, "expected GCP VPC value to be present")
			assert.Equal(t, tt.vpcName, gcpValue.VpcName)

			if tt.vpcProjectID.IsNull() {
				assert.Nil(t, gcpValue.VpcProjectId.Get(), "vpc_project_id value should be nil")
			} else {
				require.NotNil(t, gcpValue.VpcProjectId.Get(), "vpc_project_id value should not be nil")
				assert.Equal(t, tt.vpcProjectID.ValueString(), *gcpValue.VpcProjectId.Get())
			}
		})
	}
}

func TestFromQoveryClusterFeatures_GcpExistingVpc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		buildResponse    func() []qovery.ClusterFeatureResponse
		expectGcpVpcNull bool
		expectedVpcName  string
	}{
		{
			name: "GCP VPC feature with all fields",
			buildResponse: func() []qovery.ClusterFeatureResponse {
				featureID := featureIdExistingVpc
				vpcName := "my-existing-vpc"
				projectID := "my-gcp-project"
				subnetName := "my-subnet"
				servicesRange := "gke-services"
				podsRange := "gke-pods"
				return []qovery.ClusterFeatureResponse{
					{
						Id: &featureID,
						ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
							&qovery.ClusterFeatureResponseValueObject{
								ClusterFeatureGcpExistingVpcResponse: &qovery.ClusterFeatureGcpExistingVpcResponse{
									Value: qovery.ClusterFeatureGcpExistingVpc{
										VpcName:                    vpcName,
										VpcProjectId:               *qovery.NewNullableString(&projectID),
										SubnetworkName:             *qovery.NewNullableString(&subnetName),
										IpRangeServicesName:        *qovery.NewNullableString(&servicesRange),
										IpRangePodsName:            *qovery.NewNullableString(&podsRange),
										AdditionalIpRangePodsNames: []string{"extra-1"},
									},
								},
							},
						),
					},
				}
			},
			expectGcpVpcNull: false,
			expectedVpcName:  "my-existing-vpc",
		},
		{
			name: "GCP VPC feature with only vpc_name",
			buildResponse: func() []qovery.ClusterFeatureResponse {
				featureID := featureIdExistingVpc
				return []qovery.ClusterFeatureResponse{
					{
						Id: &featureID,
						ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
							&qovery.ClusterFeatureResponseValueObject{
								ClusterFeatureGcpExistingVpcResponse: &qovery.ClusterFeatureGcpExistingVpcResponse{
									Value: qovery.ClusterFeatureGcpExistingVpc{
										VpcName: "minimal-vpc",
									},
								},
							},
						),
					},
				}
			},
			expectGcpVpcNull: false,
			expectedVpcName:  "minimal-vpc",
		},
		{
			name: "no features returns null GCP VPC",
			buildResponse: func() []qovery.ClusterFeatureResponse {
				return []qovery.ClusterFeatureResponse{}
			},
			expectGcpVpcNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := fromQoveryClusterFeatures(tt.buildResponse())
			require.False(t, result.IsNull(), "features object should not be null")

			gcpVpcAttr, ok := result.Attributes()[featureKeyGcpExistingVpc]
			require.True(t, ok, "gcp_existing_vpc attribute should exist")

			gcpVpcObj, ok := gcpVpcAttr.(types.Object)
			require.True(t, ok, "gcp_existing_vpc should be a types.Object")

			if tt.expectGcpVpcNull {
				assert.True(t, gcpVpcObj.IsNull(), "gcp_existing_vpc should be null")
				return
			}

			assert.False(t, gcpVpcObj.IsNull(), "gcp_existing_vpc should not be null")

			vpcNameAttr := gcpVpcObj.Attributes()["vpc_name"].(types.String)
			assert.Equal(t, tt.expectedVpcName, vpcNameAttr.ValueString())

			// When GCP VPC is set, AWS existing_vpc should be null
			awsVpcAttr := result.Attributes()[featureKeyExistingVpc].(types.Object)
			assert.True(t, awsVpcAttr.IsNull(), "AWS existing_vpc should be null when GCP VPC is set")
		})
	}
}

func TestCluster_toUpsertClusterRequest_LabelsGroupIds(t *testing.T) {
	tests := []struct {
		name        string
		labelsSet   types.Set
		expectedIds []string
	}{
		{
			name:        "null labels_group_ids -> nil LabelsGroups",
			labelsSet:   types.SetNull(types.StringType),
			expectedIds: nil,
		},
		{
			name: "two labels_group_ids -> two ClusterLabelsGroup entries",
			labelsSet: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("11111111-1111-1111-1111-111111111111"),
				types.StringValue("22222222-2222-2222-2222-222222222222"),
			}),
			expectedIds: []string{
				"11111111-1111-1111-1111-111111111111",
				"22222222-2222-2222-2222-222222222222",
			},
		},
		{
			// Empty set (labels_group_ids = []) produces an empty slice, not nil.
			// This differs from null: the API receives [] rather than omitting the field.
			name:        "empty set labels_group_ids -> empty LabelsGroups slice",
			labelsSet:   types.SetValueMust(types.StringType, []attr.Value{}),
			expectedIds: []string{},
		},
		{
			// Unknown set (e.g. referencing a labels group not yet created) -> nil,
			// so the field is omitted from the request during planning.
			name:        "unknown labels_group_ids -> nil LabelsGroups",
			labelsSet:   types.SetUnknown(types.StringType),
			expectedIds: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("SCW"),
				Region:          types.StringValue("fr-par"),
				KubernetesMode:  types.StringValue("MANAGED"),
				State:           types.StringValue("DEPLOYED"),
				InstanceType:    types.StringValue("DEV1-L"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				DiskSize:        types.Int64Value(50),
				Features:        types.ObjectNull(createFeaturesAttrTypes()),
				LabelsGroupIds:  tc.labelsSet,
			}

			params, err := cluster.toUpsertClusterRequest(nil)
			require.NoError(t, err)
			require.NotNil(t, params)

			if tc.expectedIds == nil {
				assert.Nil(t, params.ClusterRequest.LabelsGroups)
				return
			}

			gotIds := make([]string, 0, len(params.ClusterRequest.LabelsGroups))
			for _, lg := range params.ClusterRequest.LabelsGroups {
				require.NotNil(t, lg.Id)
				gotIds = append(gotIds, *lg.Id)
			}
			assert.ElementsMatch(t, tc.expectedIds, gotIds)
		})
	}
}

func TestCluster_convertResponseToCluster_LabelsGroupIds(t *testing.T) {
	ctx := context.Background()

	id1 := "11111111-1111-1111-1111-111111111111"
	id2 := "22222222-2222-2222-2222-222222222222"

	clusterResp := &qovery.Cluster{
		Id:            "cluster-123",
		Name:          "c",
		CloudProvider: qovery.CLOUDVENDORENUM_AWS,
		Region:        "us-east-1",
		LabelsGroups: []qovery.ClusterLabelsGroup{
			{Id: &id1},
			{Id: &id2},
		},
	}
	res := &client.ClusterResponse{
		OrganizationID:      "org-123",
		ClusterResponse:     clusterResp,
		ClusterInfo:         makeTestClusterInfo("cred-123"),
		ClusterRoutingTable: &client.ClusterRoutingTable{},
	}

	t.Run("plan has non-null labels_group_ids -> populated set", func(t *testing.T) {
		t.Parallel()
		initialPlan := Cluster{
			LabelsGroupIds: types.SetValueMust(types.StringType, []attr.Value{types.StringValue(id1)}),
		}

		out := convertResponseToCluster(ctx, res, initialPlan)

		require.False(t, out.LabelsGroupIds.IsNull())
		elems := out.LabelsGroupIds.Elements()
		gotIds := make([]string, 0, len(elems))
		for _, e := range elems {
			gotIds = append(gotIds, e.(types.String).ValueString())
		}
		assert.ElementsMatch(t, []string{id1, id2}, gotIds)
	})

	t.Run("plan has null labels_group_ids -> null set even when response has labels", func(t *testing.T) {
		t.Parallel()
		initialPlan := Cluster{
			LabelsGroupIds: types.SetNull(types.StringType),
		}

		out := convertResponseToCluster(ctx, res, initialPlan)

		assert.True(t, out.LabelsGroupIds.IsNull())
	})

	t.Run("response with nil Id in ClusterLabelsGroup -> entry is skipped", func(t *testing.T) {
		t.Parallel()
		// Malformed API response: one entry has a nil Id. It must be silently skipped.
		nilIdResp := &qovery.Cluster{
			Id:            "cluster-456",
			Name:          "c",
			CloudProvider: qovery.CLOUDVENDORENUM_AWS,
			Region:        "us-east-1",
			LabelsGroups: []qovery.ClusterLabelsGroup{
				{Id: &id1},
				{Id: nil},
			},
		}
		nilIdRes := &client.ClusterResponse{
			OrganizationID:      "org-123",
			ClusterResponse:     nilIdResp,
			ClusterInfo:         makeTestClusterInfo("cred-123"),
			ClusterRoutingTable: &client.ClusterRoutingTable{},
		}
		initialPlan := Cluster{
			LabelsGroupIds: types.SetValueMust(types.StringType, []attr.Value{types.StringValue(id1)}),
		}

		out := convertResponseToCluster(ctx, nilIdRes, initialPlan)

		require.False(t, out.LabelsGroupIds.IsNull())
		elems := out.LabelsGroupIds.Elements()
		gotIds := make([]string, 0, len(elems))
		for _, e := range elems {
			gotIds = append(gotIds, e.(types.String).ValueString())
		}
		assert.ElementsMatch(t, []string{id1}, gotIds)
	})
}

func TestCluster_toUpsertClusterRequest_DesiredState(t *testing.T) {
	tests := []struct {
		name          string
		state         string
		expectedState qovery.ClusterStateEnum
	}{
		{"READY", "READY", qovery.CLUSTERSTATEENUM_READY},
		{"STOPPED", "STOPPED", qovery.CLUSTERSTATEENUM_STOPPED},
		{"DEPLOYED", "DEPLOYED", qovery.CLUSTERSTATEENUM_DEPLOYED},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cluster := Cluster{
				OrganizationId:  types.StringValue("org-123"),
				CredentialsId:   types.StringValue("cred-123"),
				Name:            types.StringValue("test-cluster"),
				CloudProvider:   types.StringValue("SCW"),
				Region:          types.StringValue("fr-par"),
				KubernetesMode:  types.StringValue("MANAGED"),
				State:           types.StringValue(tc.state),
				InstanceType:    types.StringValue("DEV1-L"),
				MinRunningNodes: types.Int64Value(1),
				MaxRunningNodes: types.Int64Value(1),
				DiskSize:        types.Int64Value(20),
				Features:        types.ObjectNull(createFeaturesAttrTypes()),
			}

			params, err := cluster.toUpsertClusterRequest(nil)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedState, params.DesiredState)
		})
	}
}

func TestCluster_convertResponseToCluster_State(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		apiState      qovery.ClusterStateEnum
		expectedState string
	}{
		{"READY", qovery.CLUSTERSTATEENUM_READY, "READY"},
		{"STOPPED", qovery.CLUSTERSTATEENUM_STOPPED, "STOPPED"},
		{"DEPLOYED", qovery.CLUSTERSTATEENUM_DEPLOYED, "DEPLOYED"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := tc.apiState
			res := &client.ClusterResponse{
				OrganizationID: "org-123",
				ClusterResponse: &qovery.Cluster{
					Id:            "cluster-123",
					Name:          "c",
					CloudProvider: qovery.CLOUDVENDORENUM_AWS,
					Region:        "us-east-1",
					Status:        &state,
				},
				ClusterInfo:         makeTestClusterInfo("cred-123"),
				ClusterRoutingTable: &client.ClusterRoutingTable{},
			}

			out := convertResponseToCluster(ctx, res, Cluster{})
			assert.Equal(t, tc.expectedState, out.State.ValueString())
		})
	}
}

