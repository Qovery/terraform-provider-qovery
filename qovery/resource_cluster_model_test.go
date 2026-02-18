//go:build unit || !integration
// +build unit !integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		expectFeatureCount         int
	}{
		{
			name:                       "all fields populated",
			vpcName:                    "my-vpc",
			vpcProjectID:               types.StringValue("my-project"),
			subnetworkName:             types.StringValue("my-subnet"),
			ipRangeServicesName:        types.StringValue("gke-services"),
			ipRangePodsName:            types.StringValue("gke-pods"),
			additionalIpRangePodsNames: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("extra-1"), types.StringValue("extra-2")}),
			expectFeatureCount:         4, // vpc_subnet + static_ip + existing_vpc(gcp) + gcp_existing_vpc serialized as EXISTING_VPC
		},
		{
			name:                       "only required vpc_name",
			vpcName:                    "minimal-vpc",
			vpcProjectID:               types.StringNull(),
			subnetworkName:             types.StringNull(),
			ipRangeServicesName:        types.StringNull(),
			ipRangePodsName:            types.StringNull(),
			additionalIpRangePodsNames: types.ListNull(types.StringType),
			expectFeatureCount:         4,
		},
		{
			name:                       "empty additional_ip_range_pods_names list",
			vpcName:                    "vpc-empty-list",
			vpcProjectID:               types.StringNull(),
			subnetworkName:             types.StringNull(),
			ipRangeServicesName:        types.StringNull(),
			ipRangePodsName:            types.StringNull(),
			additionalIpRangePodsNames: types.ListValueMust(types.StringType, []attr.Value{}),
			expectFeatureCount:         4,
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

	vpcName := "my-existing-vpc"
	projectID := "my-gcp-project"
	subnetName := "my-subnet"
	servicesRange := "gke-services"
	podsRange := "gke-pods"

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
			expectedVpcName:  vpcName,
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

			result := fromQoveryClusterFeatures(tt.buildResponse(), Cluster{})
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
