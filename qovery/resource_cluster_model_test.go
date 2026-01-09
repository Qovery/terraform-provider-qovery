//go:build unit || !integration
// +build unit !integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("AWS"),
				Region:         types.StringValue("us-east-1"),
				KubernetesMode: types.StringValue("MANAGED"),
				InstanceType:   types.StringValue("T3A_MEDIUM"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
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
								"spot_enabled":                  types.BoolType,
								"disk_size_in_gib":              types.Int64Type,
								"default_service_architecture":  types.StringType,
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
								"spot_enabled":                  types.BoolType,
								"disk_size_in_gib":              types.Int64Type,
								"default_service_architecture":  types.StringType,
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
				Id:             types.StringValue("cluster-123"),
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("AWS"),
				Region:         types.StringValue("us-east-1"),
				KubernetesMode: types.StringValue("MANAGED"),
				InstanceType:   types.StringValue("T3A_LARGE"),
				MinRunningNodes: types.Int64Value(5),
				MaxRunningNodes: types.Int64Value(15),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
			},
			state: &Cluster{
				Id:             types.StringValue("cluster-123"),
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("AWS"),
				Region:         types.StringValue("us-east-1"),
				KubernetesMode: types.StringValue("MANAGED"),
				InstanceType:   types.StringValue("T3A_MEDIUM"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
			},
			expectError: false,
		},
		{
			name: "new GCP cluster without Karpenter should succeed (not AWS)",
			cluster: Cluster{
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("GCP"),
				Region:         types.StringValue("us-central1"),
				KubernetesMode: types.StringValue("MANAGED"),
				InstanceType:   types.StringValue("N2_STANDARD_2"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
			},
			state:       nil, // New cluster
			expectError: false,
		},
		{
			name: "new AWS SELF_MANAGED cluster without Karpenter should succeed (not MANAGED)",
			cluster: Cluster{
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("AWS"),
				Region:         types.StringValue("us-east-1"),
				KubernetesMode: types.StringValue("SELF_MANAGED"),
				InstanceType:   types.StringValue("T3A_MEDIUM"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
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
						"nginx_parameters":       types.ObjectType{AttrTypes: map[string]attr.Type{}},
						"cert_manager_parameters": types.ObjectType{AttrTypes: map[string]attr.Type{}},
						"metal_lb_parameters": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"ip_address_pools": types.ListType{ElemType: types.StringType},
							},
						},
					},
					map[string]attr.Value{
						"nginx_parameters":       types.ObjectNull(map[string]attr.Type{}),
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
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("AZURE"),
				Region:         types.StringValue("eastus"),
				KubernetesMode: types.StringValue("MANAGED"),
				InstanceType:   types.StringValue("STANDARD_D2S_V3"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
			},
			state:       nil, // New cluster
			expectError: false,
		},
		{
			name: "new Scaleway cluster without Karpenter should succeed (not AWS)",
			cluster: Cluster{
				OrganizationId: types.StringValue("org-123"),
				CredentialsId:  types.StringValue("cred-123"),
				Name:           types.StringValue("test-cluster"),
				CloudProvider:  types.StringValue("SCW"),
				Region:         types.StringValue("fr-par"),
				KubernetesMode: types.StringValue("MANAGED"),
				InstanceType:   types.StringValue("DEV1_M"),
				MinRunningNodes: types.Int64Value(3),
				MaxRunningNodes: types.Int64Value(10),
				State:          types.StringValue("DEPLOYED"),
				Features:       types.ObjectNull(map[string]attr.Type{}),
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
		OrganizationId: types.StringValue("org-123"),
		CredentialsId:  types.StringValue("cred-123"),
		Name:           types.StringValue("test-cluster"),
		CloudProvider:  types.StringValue("AWS"),
		Region:         types.StringValue("us-east-1"),
		KubernetesMode: types.StringValue("MANAGED"),
		InstanceType:   types.StringValue("T3A_MEDIUM"),
		MinRunningNodes: types.Int64Value(3),
		MaxRunningNodes: types.Int64Value(10),
		State:          types.StringValue("DEPLOYED"),
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
