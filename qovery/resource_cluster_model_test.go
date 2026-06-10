//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"encoding/json"
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
			featureKeyNatGateways:    types.ObjectNull(createNatGatewaysFeatureAttrTypes()),
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

			features, err := toQoveryClusterFeatures(featuresObj, "MANAGED", "GCP")
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

func TestToQoveryClusterFeatures_GcpIgnoresDefaultVpcSubnet(t *testing.T) {
	t.Parallel()

	featuresObj := types.ObjectValueMust(
		createFeaturesAttrTypes(),
		map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue(clusterFeatureVpcSubnetDefault),
			featureKeyStaticIP:       types.BoolValue(false),
			featureKeyNatGateways:    types.ObjectNull(createNatGatewaysFeatureAttrTypes()),
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		},
	)

	features, err := toQoveryClusterFeatures(featuresObj, "MANAGED", "GCP")
	require.NoError(t, err)

	for _, feature := range features {
		assert.NotEqual(t, featureIdVpcSubnet, feature.GetId(), "GCP clusters must not send VPC_SUBNET")
	}
}

func TestToQoveryClusterFeatures_GcpRejectsCustomVpcSubnet(t *testing.T) {
	t.Parallel()

	featuresObj := types.ObjectValueMust(
		createFeaturesAttrTypes(),
		map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue("10.42.0.0/16"),
			featureKeyStaticIP:       types.BoolValue(false),
			featureKeyNatGateways:    types.ObjectNull(createNatGatewaysFeatureAttrTypes()),
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		},
	)

	_, err := toQoveryClusterFeatures(featuresObj, "MANAGED", "GCP")
	require.ErrorContains(t, err, "features.vpc_subnet is not supported for GCP clusters")
}

func TestToQoveryClusterFeatures_NonGcpAllowsCustomVpcSubnet(t *testing.T) {
	t.Parallel()

	featuresObj := types.ObjectValueMust(
		createFeaturesAttrTypes(),
		map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue("10.42.0.0/16"),
			featureKeyStaticIP:       types.BoolValue(false),
			featureKeyNatGateways:    types.ObjectNull(createNatGatewaysFeatureAttrTypes()),
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		},
	)

	features, err := toQoveryClusterFeatures(featuresObj, "MANAGED", "SCW")
	require.NoError(t, err)

	found := false
	for _, feature := range features {
		if feature.GetId() == featureIdVpcSubnet && feature.GetValue().String != nil {
			found = true
			assert.Equal(t, "10.42.0.0/16", *feature.GetValue().String)
			break
		}
	}
	assert.True(t, found, "non-GCP providers should still send VPC_SUBNET when provided")
}

func TestToQoveryClusterFeatures_GcpNatGateways(t *testing.T) {
	t.Parallel()

	natGatewayObj := types.ObjectValueMust(createNatGatewaysFeatureAttrTypes(), map[string]attr.Value{
		"static_ips_count": types.Int64Value(3),
	})
	featuresObj := types.ObjectValueMust(
		createFeaturesAttrTypes(),
		map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue(clusterFeatureVpcSubnetDefault),
			featureKeyStaticIP:       types.BoolValue(true),
			featureKeyNatGateways:    natGatewayObj,
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		},
	)

	features, err := toQoveryClusterFeatures(featuresObj, "MANAGED", "GCP")
	require.NoError(t, err)

	var staticIPFeature *qovery.ClusterRequestFeaturesInner
	var natGatewayFeature *qovery.ClusterRequestFeaturesInner
	for i := range features {
		if features[i].GetId() == featureIdStaticIP {
			staticIPFeature = &features[i]
		}
		if features[i].GetId() == featureIdNatGateway {
			natGatewayFeature = &features[i]
		}
	}

	require.NotNil(t, staticIPFeature)
	require.NotNil(t, staticIPFeature.GetValue().Bool)
	assert.True(t, *staticIPFeature.GetValue().Bool)

	require.NotNil(t, natGatewayFeature)
	natGatewayParams := natGatewayFeature.GetValue().ClusterFeatureNatGatewayParameters
	require.NotNil(t, natGatewayParams)
	gcpNatGateway := natGatewayParams.GetNatGatewayType().ClusterFeatureNatGatewayTypeGcp
	require.NotNil(t, gcpNatGateway)
	assert.Equal(t, "gcp", gcpNatGateway.Provider)
	assert.True(t, gcpNatGateway.StaticIpsEnabled)
	assert.Equal(t, int32(3), gcpNatGateway.StaticIpsCount)

	payload, err := json.Marshal(features)
	require.NoError(t, err)
	assert.JSONEq(t, `[
		{
			"id": "STATIC_IP",
			"value": true
		},
		{
			"id": "NAT_GATEWAY",
			"value": {
				"nat_gateway_type": {
					"provider": "gcp",
					"static_ips_enabled": true,
					"static_ips_count": 3
				}
			}
		}
	]`, string(payload))
}

func TestToQoveryClusterFeatures_NatGatewaysRequiresGCP(t *testing.T) {
	t.Parallel()

	// count=3 triggers the apply-time backstop (count > 1 on non-GCP).
	natGatewayObj := types.ObjectValueMust(createNatGatewaysFeatureAttrTypes(), map[string]attr.Value{
		"static_ips_count": types.Int64Value(3),
	})
	featuresObj := types.ObjectValueMust(
		createFeaturesAttrTypes(),
		map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue(clusterFeatureVpcSubnetDefault),
			featureKeyStaticIP:       types.BoolValue(true),
			featureKeyNatGateways:    natGatewayObj,
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		},
	)

	_, err := toQoveryClusterFeatures(featuresObj, "MANAGED", "AWS")
	require.ErrorContains(t, err, "features.nat_gateways with static_ips_count > 1 is only supported for GCP clusters with static_ip enabled")
}

// TestToQoveryClusterFeatures_GcpStaticIPWithDefaultBlock_EmitsEnabledCountOne pins the
// semantic flip from old behavior: with the default block {static_ips_count:1} and
// static_ip=true on GCP, the converter must emit NAT_GATEWAY with static_ips_enabled=true
// and static_ips_count=1. Previously (presence-based semantics) a null/absent block
// with static_ip=true emitted enabled=false — that dead state is now removed.
func TestToQoveryClusterFeatures_GcpStaticIPWithDefaultBlock_EmitsEnabledCountOne(t *testing.T) {
	t.Parallel()

	// The default block: {static_ips_count: 1} — what the schema ObjectDefault provides.
	natGatewayObj := types.ObjectValueMust(createNatGatewaysFeatureAttrTypes(), map[string]attr.Value{
		"static_ips_count": types.Int64Value(1),
	})
	featuresObj := types.ObjectValueMust(
		createFeaturesAttrTypes(),
		map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue(clusterFeatureVpcSubnetDefault),
			featureKeyStaticIP:       types.BoolValue(true),
			featureKeyNatGateways:    natGatewayObj,
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		},
	)

	features, err := toQoveryClusterFeatures(featuresObj, "MANAGED", "GCP")
	require.NoError(t, err)

	var natGatewayFeature *qovery.ClusterRequestFeaturesInner
	for i := range features {
		if features[i].GetId() == featureIdNatGateway {
			natGatewayFeature = &features[i]
			break
		}
	}
	require.NotNil(t, natGatewayFeature, "NAT_GATEWAY feature must be emitted")
	natGatewayParams := natGatewayFeature.GetValue().ClusterFeatureNatGatewayParameters
	require.NotNil(t, natGatewayParams)
	gcpNatGateway := natGatewayParams.GetNatGatewayType().ClusterFeatureNatGatewayTypeGcp
	require.NotNil(t, gcpNatGateway)
	assert.True(t, gcpNatGateway.StaticIpsEnabled, "default block + static_ip=true must emit enabled=true")
	assert.Equal(t, int32(1), gcpNatGateway.StaticIpsCount, "default count=1")

	payload, err := json.Marshal(features)
	require.NoError(t, err)
	assert.JSONEq(t, `[
		{
			"id": "STATIC_IP",
			"value": true
		},
		{
			"id": "NAT_GATEWAY",
			"value": {
				"nat_gateway_type": {
					"provider": "gcp",
					"static_ips_enabled": true,
					"static_ips_count": 1
				}
			}
		}
	]`, string(payload))
}

// TestToQoveryClusterFeatures_NonGcpIgnoresDefaultNatGateways asserts that on a non-GCP
// cluster (AWS, SCW) the default block {static_ips_count:1} is silently ignored:
// no NAT_GATEWAY feature is emitted and no error is returned. Only count > 1 triggers
// the apply-time backstop on non-GCP.
func TestToQoveryClusterFeatures_NonGcpIgnoresDefaultNatGateways(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider string
	}{
		{"AWS"},
		{"SCW"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.provider, func(t *testing.T) {
			t.Parallel()

			natGatewayObj := types.ObjectValueMust(createNatGatewaysFeatureAttrTypes(), map[string]attr.Value{
				"static_ips_count": types.Int64Value(1),
			})
			featuresObj := types.ObjectValueMust(
				createFeaturesAttrTypes(),
				map[string]attr.Value{
					featureKeyVpcSubnet:      types.StringValue(clusterFeatureVpcSubnetDefault),
					featureKeyStaticIP:       types.BoolValue(true),
					featureKeyNatGateways:    natGatewayObj,
					featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
					featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
					featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
				},
			)

			features, err := toQoveryClusterFeatures(featuresObj, "MANAGED", tt.provider)
			require.NoError(t, err, "default nat_gateways block with count=1 must not error on %s", tt.provider)

			for _, f := range features {
				assert.NotEqual(t, featureIdNatGateway, f.GetId(),
					"no NAT_GATEWAY feature must be emitted for non-GCP provider %s", tt.provider)
			}
		})
	}
}

func TestFromQoveryClusterFeatures_GcpNatGateways(t *testing.T) {
	t.Parallel()

	featureID := featureIdNatGateway
	natGatewayType := qovery.ClusterFeatureNatGatewayTypeGcpAsClusterFeatureNatGatewayParametersNatGatewayType(
		qovery.NewClusterFeatureNatGatewayTypeGcp("gcp", true, 2),
	)
	natGatewayParameters := qovery.ClusterFeatureNatGatewayParameters{}
	natGatewayParameters.SetNatGatewayType(natGatewayType)

	result := fromQoveryClusterFeatures([]qovery.ClusterFeatureResponse{
		{
			Id: &featureID,
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
				&qovery.ClusterFeatureResponseValueObject{
					ClusterFeatureNatGatewayParametersResponse: &qovery.ClusterFeatureNatGatewayParametersResponse{
						Value: *qovery.NewNullableClusterFeatureNatGatewayParameters(&natGatewayParameters),
					},
				},
			),
		},
	})

	require.False(t, result.IsNull())
	staticIPAttr := result.Attributes()[featureKeyStaticIP].(types.Bool)
	assert.True(t, staticIPAttr.ValueBool())

	natGatewaysAttr := result.Attributes()[featureKeyNatGateways].(types.Object)
	require.False(t, natGatewaysAttr.IsNull())
	staticIPCountAttr := natGatewaysAttr.Attributes()["static_ips_count"].(types.Int64)
	assert.Equal(t, int64(2), staticIPCountAttr.ValueInt64())

	vpcSubnetAttr := result.Attributes()[featureKeyVpcSubnet].(types.String)
	assert.Equal(t, clusterFeatureVpcSubnetDefault, vpcSubnetAttr.ValueString())
}

// TestFromQoveryClusterFeatures_NoVpcSubnetFeature_DefaultsToDefaultCidr pins the
// second half of PR #588 finding #1: when the API returns a non-empty feature list
// that contains no VPC_SUBNET entry (the normal GCP case), the Read fallback now
// fills vpc_subnet with clusterFeatureVpcSubnetDefault ("10.0.0.0/16"). Before #588
// this fallback produced "". The flip is harmless on its own, but combined with the
// newly added RequiresReplaceIfKnownChange() on vpc_subnet it lets a legacy state
// value of "" differ from this default and force a cluster replacement — see
// TestRequiresReplaceIfKnownChange_VpcSubnetLegacyEmptyState_ForcesReplacement.
func TestFromQoveryClusterFeatures_NoVpcSubnetFeature_DefaultsToDefaultCidr(t *testing.T) {
	t.Parallel()

	staticIPFeatureID := featureIdStaticIP
	result := fromQoveryClusterFeatures([]qovery.ClusterFeatureResponse{
		{
			Id: &staticIPFeatureID,
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
				&qovery.ClusterFeatureResponseValueObject{
					ClusterFeatureBooleanResponse: qovery.NewClusterFeatureBooleanResponse(
						qovery.CLUSTERFEATURERESPONSETYPEENUM_BOOLEAN,
						false,
					),
				},
			),
		},
	})

	require.False(t, result.IsNull())
	vpcSubnetAttr := result.Attributes()[featureKeyVpcSubnet].(types.String)
	assert.Equal(t, clusterFeatureVpcSubnetDefault, vpcSubnetAttr.ValueString(),
		"PR#588 finding #1: Read fallback for an absent VPC_SUBNET feature now yields the default, not \"\"")
}

// TestFromQoveryClusterFeatures_NoNatGatewayFeature_DefaultsToCountOne pins the Read
// fallback invariant: when the API feature list contains no NAT_GATEWAY entry at all
// (e.g. any non-GCP cluster), nat_gateways must still be a non-null default object
// {static_ips_count: 1} rather than null. This ensures state is always consistent
// regardless of cloud provider.
func TestFromQoveryClusterFeatures_NoNatGatewayFeature_DefaultsToCountOne(t *testing.T) {
	t.Parallel()

	staticIPFeatureID := featureIdStaticIP
	result := fromQoveryClusterFeatures([]qovery.ClusterFeatureResponse{
		{
			Id: &staticIPFeatureID,
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
				&qovery.ClusterFeatureResponseValueObject{
					ClusterFeatureBooleanResponse: qovery.NewClusterFeatureBooleanResponse(
						qovery.CLUSTERFEATURERESPONSETYPEENUM_BOOLEAN,
						false,
					),
				},
			),
		},
	})

	require.False(t, result.IsNull())
	natGatewaysAttr := result.Attributes()[featureKeyNatGateways].(types.Object)
	require.False(t, natGatewaysAttr.IsNull(),
		"nat_gateways must be the default object {static_ips_count:1} when no NAT_GATEWAY feature is present")
	staticIPCountAttr := natGatewaysAttr.Attributes()["static_ips_count"].(types.Int64)
	assert.Equal(t, int64(1), staticIPCountAttr.ValueInt64(),
		"default nat_gateways count must be 1 when no NAT feature is returned by the API")
}

func TestFromQoveryClusterFeatures_GcpNatGatewaysDisabled(t *testing.T) {
	t.Parallel()

	featureID := featureIdNatGateway
	natGatewayType := qovery.ClusterFeatureNatGatewayTypeGcpAsClusterFeatureNatGatewayParametersNatGatewayType(
		qovery.NewClusterFeatureNatGatewayTypeGcp("gcp", false, 2),
	)
	natGatewayParameters := qovery.ClusterFeatureNatGatewayParameters{}
	natGatewayParameters.SetNatGatewayType(natGatewayType)

	result := fromQoveryClusterFeatures([]qovery.ClusterFeatureResponse{
		{
			Id: &featureID,
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
				&qovery.ClusterFeatureResponseValueObject{
					ClusterFeatureNatGatewayParametersResponse: &qovery.ClusterFeatureNatGatewayParametersResponse{
						Value: *qovery.NewNullableClusterFeatureNatGatewayParameters(&natGatewayParameters),
					},
				},
			),
		},
	})

	require.False(t, result.IsNull())
	staticIPAttr := result.Attributes()[featureKeyStaticIP].(types.Bool)
	assert.False(t, staticIPAttr.ValueBool())

	// When NAT_GATEWAY is present but disabled, Read returns the default {static_ips_count:1}
	// rather than null, so state is never null (prevents apply-time inconsistency).
	natGatewaysAttr := result.Attributes()[featureKeyNatGateways].(types.Object)
	require.False(t, natGatewaysAttr.IsNull(), "Read always returns the default object when NAT feature is present, even when disabled")
	staticIPCountAttr := natGatewaysAttr.Attributes()["static_ips_count"].(types.Int64)
	assert.Equal(t, int64(1), staticIPCountAttr.ValueInt64(), "disabled NAT feature → default count=1")
}

func TestFromQoveryClusterFeatures_StaticIPFeatureOverridesNatGatewayEnabledFlag(t *testing.T) {
	t.Parallel()

	staticIPFeatureID := featureIdStaticIP
	natGatewayFeatureID := featureIdNatGateway
	natGatewayType := qovery.ClusterFeatureNatGatewayTypeGcpAsClusterFeatureNatGatewayParametersNatGatewayType(
		qovery.NewClusterFeatureNatGatewayTypeGcp("gcp", true, 2),
	)
	natGatewayParameters := qovery.ClusterFeatureNatGatewayParameters{}
	natGatewayParameters.SetNatGatewayType(natGatewayType)

	result := fromQoveryClusterFeatures([]qovery.ClusterFeatureResponse{
		{
			Id: &staticIPFeatureID,
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
				&qovery.ClusterFeatureResponseValueObject{
					ClusterFeatureBooleanResponse: qovery.NewClusterFeatureBooleanResponse(
						qovery.CLUSTERFEATURERESPONSETYPEENUM_BOOLEAN,
						false,
					),
				},
			),
		},
		{
			Id: &natGatewayFeatureID,
			ValueObject: *qovery.NewNullableClusterFeatureResponseValueObject(
				&qovery.ClusterFeatureResponseValueObject{
					ClusterFeatureNatGatewayParametersResponse: &qovery.ClusterFeatureNatGatewayParametersResponse{
						Value: *qovery.NewNullableClusterFeatureNatGatewayParameters(&natGatewayParameters),
					},
				},
			),
		},
	})

	require.False(t, result.IsNull())
	staticIPAttr := result.Attributes()[featureKeyStaticIP].(types.Bool)
	assert.False(t, staticIPAttr.ValueBool(), "STATIC_IP is the source of truth for the toggle")

	// When NAT_GATEWAY has static_ips_enabled=true, fromQoveryClusterFeatures returns the
	// API count regardless of the STATIC_IP boolean. The nat_gateways block is {count:2}
	// (not null) because the API returned an enabled NAT feature.
	natGatewaysAttr := result.Attributes()[featureKeyNatGateways].(types.Object)
	require.False(t, natGatewaysAttr.IsNull(), "NAT_GATEWAY block should be non-null when NAT feature has static_ips_enabled=true")
	staticIPCountAttr := natGatewaysAttr.Attributes()["static_ips_count"].(types.Int64)
	assert.Equal(t, int64(2), staticIPCountAttr.ValueInt64(), "count from API should be reflected even when STATIC_IP is disabled")
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

func clusterOutputObject(values map[string]string) types.Object {
	attrs := allNullClusterOutputValues()
	for k, v := range values {
		attrs[k] = types.StringValue(v)
	}
	return types.ObjectValueMust(clusterInfrastructureOutputsAttrTypes, attrs)
}

func TestFromQoveryClusterOutput(t *testing.T) {
	t.Parallel()

	priorStateWithOidc := clusterOutputObject(map[string]string{
		"cluster_name":        "qovery-z64614976",
		"cluster_arn":         "arn:aws:eks:us-east-1:123456789012:cluster/qovery-z64614976",
		"cluster_oidc_issuer": "https://oidc.eks.us-east-1.amazonaws.com/id/PRIOR",
		"vpc_id":              "vpc-prior",
	})
	emptyPriorState := types.ObjectNull(clusterInfrastructureOutputsAttrTypes)

	tests := []struct {
		name                string
		api                 *qovery.InfrastructureOutputs
		priorState          types.Object
		expectedName        string
		expectedArn         string
		expectedSelfLink    string
		expectedOidcIssuer  string
		expectedVpcId       string
		expectPriorReturned bool
	}{
		{
			name:                "api nil and prior state empty -> all nulls",
			api:                 nil,
			priorState:          emptyPriorState,
			expectPriorReturned: false,
		},
		{
			name:                "api nil and prior state has values -> prior state preserved",
			api:                 nil,
			priorState:          priorStateWithOidc,
			expectPriorReturned: true,
		},
		{
			name: "api EKS payload -> EKS fields populated",
			api: &qovery.InfrastructureOutputs{
				EksInfrastructureOutputs: qovery.NewEksInfrastructureOutputs(
					"EKS",
					"qovery-zABC",
					"arn:aws:eks:us-east-1:123456789012:cluster/qovery-zABC",
					"https://oidc.eks.us-east-1.amazonaws.com/id/NEW",
					"vpc-new",
				),
			},
			priorState:         priorStateWithOidc,
			expectedName:       "qovery-zABC",
			expectedArn:        "arn:aws:eks:us-east-1:123456789012:cluster/qovery-zABC",
			expectedOidcIssuer: "https://oidc.eks.us-east-1.amazonaws.com/id/NEW",
			expectedVpcId:      "vpc-new",
		},
		{
			name: "api AKS payload -> AKS fields populated",
			api: &qovery.InfrastructureOutputs{
				AksInfrastructureOutputs: qovery.NewAksInfrastructureOutputs(
					"AKS",
					"aks-cluster",
					"https://oidc.aks.example.com/issuer",
				),
			},
			priorState:         emptyPriorState,
			expectedName:       "aks-cluster",
			expectedOidcIssuer: "https://oidc.aks.example.com/issuer",
		},
		{
			name: "api GKE payload -> GKE fields populated",
			api: &qovery.InfrastructureOutputs{
				GkeInfrastructureOutputs: qovery.NewGkeInfrastructureOutputs(
					"GKE",
					"gke-cluster",
					"https://container.googleapis.com/projects/p/locations/l/clusters/c",
				),
			},
			priorState:       emptyPriorState,
			expectedName:     "gke-cluster",
			expectedSelfLink: "https://container.googleapis.com/projects/p/locations/l/clusters/c",
		},
		{
			name: "api Kapsule payload -> only cluster_name populated",
			api: &qovery.InfrastructureOutputs{
				KapsuleInfrastructureOutputs: qovery.NewKapsuleInfrastructureOutputs(
					"SCW_KAPSULE",
					"kapsule-cluster",
				),
			},
			priorState:   emptyPriorState,
			expectedName: "kapsule-cluster",
		},
		{
			name:                "api non-nil but unknown discriminator + prior state has values -> prior preserved",
			api:                 &qovery.InfrastructureOutputs{}, // all sub-types nil
			priorState:          priorStateWithOidc,
			expectPriorReturned: true,
		},
		{
			name:                "api non-nil but unknown discriminator + empty prior state -> all nulls",
			api:                 &qovery.InfrastructureOutputs{},
			priorState:          emptyPriorState,
			expectPriorReturned: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out := fromQoveryClusterOutput(tc.api, tc.priorState)

			if tc.expectPriorReturned {
				assert.True(t, out.Equal(tc.priorState), "expected prior state to be returned unchanged")
				return
			}

			require.False(t, out.IsNull())
			require.False(t, out.IsUnknown())
			attrs := out.Attributes()

			getStr := func(key string) string {
				v, ok := attrs[key].(types.String)
				require.Truef(t, ok, "%s is not a String", key)
				if v.IsNull() || v.IsUnknown() {
					return ""
				}
				return v.ValueString()
			}

			assert.Equal(t, tc.expectedName, getStr("cluster_name"))
			assert.Equal(t, tc.expectedArn, getStr("cluster_arn"))
			assert.Equal(t, tc.expectedSelfLink, getStr("cluster_self_link"))
			assert.Equal(t, tc.expectedOidcIssuer, getStr("cluster_oidc_issuer"))
			assert.Equal(t, tc.expectedVpcId, getStr("vpc_id"))
		})
	}
}

func TestCluster_convertResponseToCluster_InfrastructureOutputs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	priorOidc := "https://oidc.eks.us-east-1.amazonaws.com/id/PRIOR"
	priorArn := "arn:aws:eks:us-east-1:123456789012:cluster/qovery-zABC"
	priorState := Cluster{
		InfrastructureOutputs: clusterOutputObject(map[string]string{
			"cluster_name":        "qovery-zABC",
			"cluster_arn":         priorArn,
			"cluster_oidc_issuer": priorOidc,
			"vpc_id":              "vpc-prior",
		}),
	}

	t.Run("api returns nil infra outputs -> prior state preserved end-to-end", func(t *testing.T) {
		t.Parallel()
		// Reproduces the production scenario: cluster in DEPLOYMENT_ERROR, API returns
		// infrastructure_outputs: null. Without the fallback, every downstream resource
		// referencing qovery_cluster.cluster.infrastructure_outputs.* would see null.
		res := &client.ClusterResponse{
			OrganizationID: "org-123",
			ClusterResponse: &qovery.Cluster{
				Id:                    "cluster-123",
				Name:                  "c",
				CloudProvider:         qovery.CLOUDVENDORENUM_AWS,
				Region:                "us-east-1",
				InfrastructureOutputs: nil,
			},
			ClusterInfo:         makeTestClusterInfo("cred-123"),
			ClusterRoutingTable: &client.ClusterRoutingTable{},
		}

		out := convertResponseToCluster(ctx, res, priorState)

		require.False(t, out.InfrastructureOutputs.IsNull())
		attrs := out.InfrastructureOutputs.Attributes()
		assert.Equal(t, priorOidc, attrs["cluster_oidc_issuer"].(types.String).ValueString())
		assert.Equal(t, priorArn, attrs["cluster_arn"].(types.String).ValueString())
		assert.Equal(t, "vpc-prior", attrs["vpc_id"].(types.String).ValueString())
	})

	t.Run("api returns fresh EKS outputs -> api values win over prior state", func(t *testing.T) {
		t.Parallel()
		res := &client.ClusterResponse{
			OrganizationID: "org-123",
			ClusterResponse: &qovery.Cluster{
				Id:            "cluster-123",
				Name:          "c",
				CloudProvider: qovery.CLOUDVENDORENUM_AWS,
				Region:        "us-east-1",
				InfrastructureOutputs: &qovery.InfrastructureOutputs{
					EksInfrastructureOutputs: qovery.NewEksInfrastructureOutputs(
						"EKS",
						"qovery-zNEW",
						"arn:aws:eks:us-east-1:123456789012:cluster/qovery-zNEW",
						"https://oidc.eks.us-east-1.amazonaws.com/id/NEW",
						"vpc-new",
					),
				},
			},
			ClusterInfo:         makeTestClusterInfo("cred-123"),
			ClusterRoutingTable: &client.ClusterRoutingTable{},
		}

		out := convertResponseToCluster(ctx, res, priorState)

		require.False(t, out.InfrastructureOutputs.IsNull())
		attrs := out.InfrastructureOutputs.Attributes()
		assert.Equal(t, "https://oidc.eks.us-east-1.amazonaws.com/id/NEW", attrs["cluster_oidc_issuer"].(types.String).ValueString())
		assert.Equal(t, "vpc-new", attrs["vpc_id"].(types.String).ValueString())
	})

	t.Run("api returns nil + no prior state -> all nulls", func(t *testing.T) {
		t.Parallel()
		// First-deploy or fresh-import case: cluster has never been deployed, no prior values.
		emptyState := Cluster{
			InfrastructureOutputs: types.ObjectNull(clusterInfrastructureOutputsAttrTypes),
		}
		res := &client.ClusterResponse{
			OrganizationID: "org-123",
			ClusterResponse: &qovery.Cluster{
				Id:                    "cluster-123",
				Name:                  "c",
				CloudProvider:         qovery.CLOUDVENDORENUM_AWS,
				Region:                "us-east-1",
				InfrastructureOutputs: nil,
			},
			ClusterInfo:         makeTestClusterInfo("cred-123"),
			ClusterRoutingTable: &client.ClusterRoutingTable{},
		}

		out := convertResponseToCluster(ctx, res, emptyState)

		require.False(t, out.InfrastructureOutputs.IsNull())
		attrs := out.InfrastructureOutputs.Attributes()
		for k, v := range attrs {
			s, ok := v.(types.String)
			require.Truef(t, ok, "%s is not a String", k)
			assert.Truef(t, s.IsNull(), "expected %s to be null", k)
		}
	})
}

func TestToQoveryInfrastructureChartsParameters_EksAnywhereAndClusterBackup(t *testing.T) {
	t.Parallel()

	attrTypes := createInfrastructureChartsParametersAttrTypes()
	eksAnywhereAttrTypes := attrTypes[infraChartsEksAnywhereKey].(types.ObjectType).AttrTypes
	gitRepositoryAttrTypes := eksAnywhereAttrTypes["git_repository"].(types.ObjectType).AttrTypes
	clusterBackupAttrTypes := eksAnywhereAttrTypes[infraChartsClusterBackupKey].(types.ObjectType).AttrTypes
	clusterBackupS3AttrTypes := clusterBackupAttrTypes["s3"].(types.ObjectType).AttrTypes

	gitRepository := types.ObjectValueMust(gitRepositoryAttrTypes, map[string]attr.Value{
		"url":          types.StringValue("https://github.com/org/eks-anywhere.git"),
		"git_token_id": types.StringValue("git-token-id"),
		"commit_id":    types.StringValue("abcdef1234567890"),
		"branch":       types.StringValue("main"),
		"provider":     types.StringValue("GITHUB"),
	})

	clusterBackupS3 := types.ObjectValueMust(clusterBackupS3AttrTypes, map[string]attr.Value{
		"bucket":     types.StringValue("my-backup-bucket"),
		"region":     types.StringValue("eu-west-3"),
		"role_arn":   types.StringValue("arn:aws:iam::123456789012:role/backup-role"),
		"key_prefix": types.StringValue("eks-anywhere/backups"),
	})

	clusterBackup := types.ObjectValueMust(clusterBackupAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(true),
		"s3":      clusterBackupS3,
	})

	eksAnywhere := types.ObjectValueMust(eksAnywhereAttrTypes, map[string]attr.Value{
		"yaml_file_path":            types.StringValue("clusters/prod/cluster.yaml"),
		"git_repository":            gitRepository,
		infraChartsClusterBackupKey: clusterBackup,
	})

	input := types.ObjectValueMust(attrTypes, map[string]attr.Value{
		infraChartsNginxKey:       types.ObjectNull(attrTypes[infraChartsNginxKey].(types.ObjectType).AttrTypes),
		infraChartsCertManagerKey: types.ObjectNull(attrTypes[infraChartsCertManagerKey].(types.ObjectType).AttrTypes),
		infraChartsMetalLbKey:     types.ObjectNull(attrTypes[infraChartsMetalLbKey].(types.ObjectType).AttrTypes),
		infraChartsEksAnywhereKey: eksAnywhere,
	})

	params, err := toQoveryInfrastructureChartsParameters(input)
	require.NoError(t, err)
	require.NotNil(t, params)
	require.NotNil(t, params.EksAnywhereParameters)

	assert.Equal(t, "clusters/prod/cluster.yaml", params.EksAnywhereParameters.YamlFilePath)
	assert.Equal(t, "https://github.com/org/eks-anywhere.git", params.EksAnywhereParameters.GitRepository.Url)
	assert.Equal(t, "git-token-id", params.EksAnywhereParameters.GitRepository.GitTokenId)
	require.NotNil(t, params.EksAnywhereParameters.GitRepository.CommitId)
	assert.Equal(t, "abcdef1234567890", *params.EksAnywhereParameters.GitRepository.CommitId)
	require.NotNil(t, params.EksAnywhereParameters.GitRepository.Branch)
	assert.Equal(t, "main", *params.EksAnywhereParameters.GitRepository.Branch)
	require.NotNil(t, params.EksAnywhereParameters.GitRepository.Provider)
	assert.Equal(t, qovery.GITPROVIDERENUM_GITHUB, *params.EksAnywhereParameters.GitRepository.Provider)

	require.NotNil(t, params.EksAnywhereParameters.ClusterBackup)
	require.NotNil(t, params.EksAnywhereParameters.ClusterBackup.Enabled)
	assert.Equal(t, true, *params.EksAnywhereParameters.ClusterBackup.Enabled)
	assert.Equal(t, "my-backup-bucket", params.EksAnywhereParameters.ClusterBackup.S3.Bucket)
	assert.Equal(t, "eu-west-3", params.EksAnywhereParameters.ClusterBackup.S3.Region)
	assert.Equal(t, "arn:aws:iam::123456789012:role/backup-role", params.EksAnywhereParameters.ClusterBackup.S3.RoleArn)
	require.NotNil(t, params.EksAnywhereParameters.ClusterBackup.S3.KeyPrefix)
	assert.Equal(t, "eks-anywhere/backups", *params.EksAnywhereParameters.ClusterBackup.S3.KeyPrefix)
}

func TestFromQoveryInfrastructureChartsParameters_EksAnywhereAndClusterBackup(t *testing.T) {
	t.Parallel()

	branch := "main"
	commitID := "abcdef1234567890"
	provider := qovery.GITPROVIDERENUM_GITHUB

	eksAnywhere := qovery.NewClusterInfrastructureEksAnywhereParameters(
		qovery.ClusterEksAnywhereGitRepository{
			Url:        "https://github.com/org/eks-anywhere.git",
			GitTokenId: "git-token-id",
			CommitId:   &commitID,
			Branch:     &branch,
			Provider:   &provider,
		},
		"clusters/prod/cluster.yaml",
	)
	enabled := true
	keyPrefix := "eks-anywhere/backups"
	eksAnywhere.ClusterBackup = &qovery.ClusterInfrastructureEksAnywhereBackupParameters{
		Enabled: &enabled,
		S3: qovery.ClusterInfrastructureEksAnywhereBackupS3Parameters{
			Bucket:    "my-backup-bucket",
			Region:    "eu-west-3",
			RoleArn:   "arn:aws:iam::123456789012:role/backup-role",
			KeyPrefix: &keyPrefix,
		},
	}

	params := qovery.NewClusterInfrastructureChartsParameters()
	params.EksAnywhereParameters = eksAnywhere

	result := fromQoveryInfrastructureChartsParameters(params)
	require.False(t, result.IsNull())

	eksAnywhereAttr := result.Attributes()[infraChartsEksAnywhereKey].(types.Object)
	require.False(t, eksAnywhereAttr.IsNull())
	assert.Equal(t, "clusters/prod/cluster.yaml", eksAnywhereAttr.Attributes()["yaml_file_path"].(types.String).ValueString())

	gitRepositoryAttr := eksAnywhereAttr.Attributes()["git_repository"].(types.Object)
	require.False(t, gitRepositoryAttr.IsNull())
	assert.Equal(t, "https://github.com/org/eks-anywhere.git", gitRepositoryAttr.Attributes()["url"].(types.String).ValueString())
	assert.Equal(t, "git-token-id", gitRepositoryAttr.Attributes()["git_token_id"].(types.String).ValueString())
	assert.Equal(t, "abcdef1234567890", gitRepositoryAttr.Attributes()["commit_id"].(types.String).ValueString())
	assert.Equal(t, "main", gitRepositoryAttr.Attributes()["branch"].(types.String).ValueString())
	assert.Equal(t, "GITHUB", gitRepositoryAttr.Attributes()["provider"].(types.String).ValueString())

	clusterBackupAttr := eksAnywhereAttr.Attributes()[infraChartsClusterBackupKey].(types.Object)
	require.False(t, clusterBackupAttr.IsNull())
	assert.Equal(t, true, clusterBackupAttr.Attributes()["enabled"].(types.Bool).ValueBool())

	s3Attr := clusterBackupAttr.Attributes()["s3"].(types.Object)
	require.False(t, s3Attr.IsNull())
	assert.Equal(t, "my-backup-bucket", s3Attr.Attributes()["bucket"].(types.String).ValueString())
	assert.Equal(t, "eu-west-3", s3Attr.Attributes()["region"].(types.String).ValueString())
	assert.Equal(t, "arn:aws:iam::123456789012:role/backup-role", s3Attr.Attributes()["role_arn"].(types.String).ValueString())
	assert.Equal(t, "eks-anywhere/backups", s3Attr.Attributes()["key_prefix"].(types.String).ValueString())
}

// ----------------------------------------------------------------------------
// PR #588 finding #3 — forceUpdate must fire on deploy-affecting spec changes.
//
// Qovery applies cluster spec changes (instance_type, node counts, disk_size,
// kubernetes_mode, labels) only on a (re)deploy; EditCluster alone persists the spec
// to the DB without applying it to the running cluster. The deploy is gated by
// ClusterUpsertParams.ForceUpdate. Changing such an attribute must set
// ForceUpdate=true so the change is actually deployed; metadata-only changes
// (name/description) must NOT force a redeploy.
// ----------------------------------------------------------------------------

func baseScwCluster() Cluster {
	return Cluster{
		OrganizationId:                 types.StringValue("org-123"),
		CredentialsId:                  types.StringValue("cred-123"),
		Name:                           types.StringValue("c"),
		CloudProvider:                  types.StringValue("SCW"),
		Region:                         types.StringValue("fr-par"),
		KubernetesMode:                 types.StringValue("MANAGED"),
		State:                          types.StringValue("DEPLOYED"),
		InstanceType:                   types.StringValue("DEV1-L"),
		DiskSize:                       types.Int64Value(20),
		MinRunningNodes:                types.Int64Value(3),
		MaxRunningNodes:                types.Int64Value(10),
		Production:                     types.BoolValue(false),
		AdvancedSettingsJson:           types.StringValue("{}"),
		Features:                       types.ObjectNull(map[string]attr.Type{}),
		RoutingTables:                  types.SetNull(types.ObjectType{AttrTypes: clusterRouteAttrTypes}),
		LabelsGroupIds:                 types.SetNull(types.StringType),
		InfrastructureChartsParameters: types.ObjectNull(createInfrastructureChartsParametersAttrTypes()),
	}
}

func TestCluster_toUpsertClusterRequest_SpecChangeForcesDeploy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Cluster)
		want   bool
	}{
		{"no change", func(*Cluster) {}, false},
		{"instance_type change", func(c *Cluster) { c.InstanceType = types.StringValue("DEV1-XL") }, true},
		{"min_running_nodes change", func(c *Cluster) { c.MinRunningNodes = types.Int64Value(5) }, true},
		{"max_running_nodes change", func(c *Cluster) { c.MaxRunningNodes = types.Int64Value(20) }, true},
		{"disk_size change", func(c *Cluster) { c.DiskSize = types.Int64Value(40) }, true},
		{"name only (metadata)", func(c *Cluster) { c.Name = types.StringValue("renamed") }, false},
		{"description only (metadata)", func(c *Cluster) { c.Description = types.StringValue("new desc") }, false},
		{"production only (metadata)", func(c *Cluster) { c.Production = types.BoolValue(true) }, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := baseScwCluster()
			plan := baseScwCluster()
			tc.mutate(&plan)

			params, err := plan.toUpsertClusterRequest(&state)
			require.NoError(t, err)
			assert.Equal(t, tc.want, params.ForceUpdate,
				"PR#588 finding #3: ForceUpdate for change %q", tc.name)
		})
	}
}

// TestCluster_hasClusterSpecDiff exercises the spec-diff helper directly, including
// kubernetes_mode and labels_group_ids, and confirms metadata changes are ignored.
func TestCluster_hasClusterSpecDiff(t *testing.T) {
	t.Parallel()

	labels := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("lg-2")})

	tests := []struct {
		name   string
		mutate func(*Cluster)
		want   bool
	}{
		{"identical", func(*Cluster) {}, false},
		{"instance_type", func(c *Cluster) { c.InstanceType = types.StringValue("DEV1-XL") }, true},
		{"disk_size", func(c *Cluster) { c.DiskSize = types.Int64Value(40) }, true},
		{"min_running_nodes", func(c *Cluster) { c.MinRunningNodes = types.Int64Value(5) }, true},
		{"max_running_nodes", func(c *Cluster) { c.MaxRunningNodes = types.Int64Value(20) }, true},
		{"kubernetes_mode", func(c *Cluster) { c.KubernetesMode = types.StringValue("SELF_MANAGED") }, true},
		{"labels_group_ids", func(c *Cluster) { c.LabelsGroupIds = labels }, true},
		{"name (metadata)", func(c *Cluster) { c.Name = types.StringValue("renamed") }, false},
		{"description (metadata)", func(c *Cluster) { c.Description = types.StringValue("d") }, false},
		{"production only (metadata)", func(c *Cluster) { c.Production = types.BoolValue(true) }, false},
	}

	// nil state is the create path — never a spec "diff".
	t.Run("nil state", func(t *testing.T) {
		t.Parallel()
		plan := baseScwCluster()
		assert.False(t, plan.hasClusterSpecDiff(nil))
	})

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := baseScwCluster()
			plan := baseScwCluster()
			tc.mutate(&plan)
			assert.Equal(t, tc.want, plan.hasClusterSpecDiff(&state),
				"PR#588 finding #3: hasClusterSpecDiff for %q", tc.name)
		})
	}
}

// TestCluster_hasFeaturesDiff guards the reflect.DeepEqual comparison (#588): two
// structurally-identical feature sets must NOT report a diff (the pre-#588 pointer
// comparison always did), while a real feature value change must.
func TestCluster_hasFeaturesDiff(t *testing.T) {
	t.Parallel()

	features := func(staticIP bool) types.Object {
		return types.ObjectValueMust(createFeaturesAttrTypes(), map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue("10.0.0.0/16"),
			featureKeyStaticIP:       types.BoolValue(staticIP),
			featureKeyNatGateways:    types.ObjectNull(createNatGatewaysFeatureAttrTypes()),
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		})
	}
	mk := func(f types.Object) Cluster {
		return Cluster{
			CloudProvider:  types.StringValue("SCW"),
			KubernetesMode: types.StringValue("MANAGED"),
			Features:       f,
		}
	}

	t.Run("identical features -> no diff", func(t *testing.T) {
		t.Parallel()
		plan := mk(features(true))
		state := mk(features(true))
		assert.False(t, plan.hasFeaturesDiff(&state),
			"PR#588: identical features must not report a diff (reflect.DeepEqual)")
	})

	t.Run("changed static_ip -> diff", func(t *testing.T) {
		t.Parallel()
		plan := mk(features(true))
		state := mk(features(false))
		assert.True(t, plan.hasFeaturesDiff(&state),
			"a real feature value change must report a diff")
	})

	// Legacy upgrade path: a pre-#588 provider stored vpc_subnet="" in state (the old
	// Read fallback) while the post-#588 schema Default plans "10.0.0.0/16". On a
	// -refresh=false apply the stale "" survives in state; the two values are the same
	// logical config, so they must not produce a features diff (which would set
	// ForceUpdate=true and redeploy the cluster for a change the user never made).
	t.Run("legacy empty vpc_subnet vs default -> no diff", func(t *testing.T) {
		t.Parallel()

		legacyFeatures := types.ObjectValueMust(createFeaturesAttrTypes(), map[string]attr.Value{
			featureKeyVpcSubnet:      types.StringValue(""),
			featureKeyStaticIP:       types.BoolValue(true),
			featureKeyNatGateways:    types.ObjectNull(createNatGatewaysFeatureAttrTypes()),
			featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
			featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
			featureKeyKarpenter:      types.ObjectNull(createKarpenterFeatureAttrTypes()),
		})

		plan := mk(features(true)) // vpc_subnet = "10.0.0.0/16" (schema default)
		state := mk(legacyFeatures)
		assert.False(t, plan.hasFeaturesDiff(&state),
			"legacy vpc_subnet=\"\" in state vs planned default must not force a redeploy")
	})
}
