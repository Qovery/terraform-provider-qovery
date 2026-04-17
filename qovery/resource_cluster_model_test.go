//go:build unit || !integration

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
	require.NotNil(t, params.EksAnywhereParameters.GitRepository.Branch)
	assert.Equal(t, "main", *params.EksAnywhereParameters.GitRepository.Branch)
	require.NotNil(t, params.EksAnywhereParameters.GitRepository.Provider)
	assert.Equal(t, qovery.GITPROVIDERENUM_GITHUB, *params.EksAnywhereParameters.GitRepository.Provider)

	rawClusterBackup, ok := params.EksAnywhereParameters.AdditionalProperties[infraChartsClusterBackupKey]
	require.True(t, ok)

	clusterBackupMap, ok := rawClusterBackup.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, clusterBackupMap["enabled"])

	s3Map, ok := clusterBackupMap["s3"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "my-backup-bucket", s3Map["bucket"])
	assert.Equal(t, "eu-west-3", s3Map["region"])
	assert.Equal(t, "arn:aws:iam::123456789012:role/backup-role", s3Map["role_arn"])
	assert.Equal(t, "eks-anywhere/backups", s3Map["key_prefix"])
}

func TestFromQoveryInfrastructureChartsParameters_EksAnywhereAndClusterBackup(t *testing.T) {
	t.Parallel()

	branch := "main"
	provider := qovery.GITPROVIDERENUM_GITHUB

	eksAnywhere := qovery.NewClusterInfrastructureEksAnywhereParameters(
		qovery.ClusterEksAnywhereGitRepository{
			Url:        "https://github.com/org/eks-anywhere.git",
			GitTokenId: "git-token-id",
			Branch:     &branch,
			Provider:   &provider,
		},
		"clusters/prod/cluster.yaml",
	)
	eksAnywhere.AdditionalProperties = map[string]interface{}{
		infraChartsClusterBackupKey: map[string]interface{}{
			"enabled": true,
			"s3": map[string]interface{}{
				"bucket":     "my-backup-bucket",
				"region":     "eu-west-3",
				"role_arn":   "arn:aws:iam::123456789012:role/backup-role",
				"key_prefix": "eks-anywhere/backups",
			},
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
