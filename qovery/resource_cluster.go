package qovery

import (
	"bytes"
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.ResourceType = clusterResourceType{}
var _ resource.Resource = clusterResource{}
var _ resource.ResourceWithImportState = clusterResource{}

var (
	//go:embed data/cluster_instance_types/*.json
	instanceTypes embed.FS

	// Cluster State
	clusterStates = clientEnumToStringArray([]qovery.StateEnum{
		qovery.STATEENUM_RUNNING,
		qovery.STATEENUM_STOPPED,
	})
	clusterStateDefault = string(qovery.STATEENUM_RUNNING)

	// Cluster Description
	clusterDescriptionDefault = ""

	// Cloud Provider
	cloudProviders = clientEnumToStringArray(qovery.AllowedCloudProviderEnumEnumValues)

	// Cluster Min Running Nodes
	clusterMinRunningNodesMin     int64 = 1
	clusterMinRunningNodesDefault int64 = 3

	// Cluster Max Running Nodes
	clusterMaxRunningNodesMin     int64 = 1
	clusterMaxRunningNodesDefault int64 = 10

	// Cluster Feature VPC_SUBNET
	clusterFeatureVpcSubnetDefault = "10.0.0.0/16"

	// Cluster Kubernetes Mode
	clusterKubernetesModes = clientEnumToStringArray([]qovery.KubernetesEnum{
		qovery.KUBERNETESENUM_MANAGED,
		qovery.KUBERNETESENUM_K3_S,
	})
	clusterKubernetesModeDefault = string(qovery.KUBERNETESENUM_MANAGED)
)

type clusterResourceType struct {
	client *client.Client
}

func (r clusterResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	// Read cluster instance type for documentation from embedded files
	clusterInstanceTypesByProvider, err := readInstanceTypes()
	if err != nil {
		return tfsdk.Schema{}, []diag.Diagnostic{
			diag.NewErrorDiagnostic("Unable to fetch cluster instance types", err.Error()),
		}
	}
	var clusterInstanceTypes []string
	for _, tt := range clusterInstanceTypesByProvider {
		for _, t := range tt {
			clusterInstanceTypes = append(clusterInstanceTypes, t)
		}
	}

	return tfsdk.Schema{
		Description: "Provides a Qovery cluster resource. This can be used to create and manage Qovery cluster.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"credentials_id": {
				Description: "Id of the credentials.",
				Type:        types.StringType,
				Required:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the cluster.",
				Type:        types.StringType,
				Required:    true,
			},
			"cloud_provider": {
				Description: descriptions.NewStringEnumDescription(
					"Cloud provider of the cluster.",
					cloudProviders,
					nil,
				),
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(cloudProviders),
				},
			},
			"region": {
				Description: "Region of the cluster.",
				Type:        types.StringType,
				Required:    true,
			},
			"description": {
				Description: descriptions.NewStringDefaultDescription(
					"Description of the cluster.",
					clusterDescriptionDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(clusterDescriptionDefault),
				},
			},
			"kubernetes_mode": {
				Description: descriptions.NewStringEnumDescription(
					"Kubernetes mode of the cluster.",
					clusterKubernetesModes,
					&clusterKubernetesModeDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(clusterKubernetesModeDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(clusterKubernetesModes),
				},
			},
			"instance_type": {
				Description: descriptions.NewMapStringArrayEnumDescription(
					"Instance type of the cluster.",
					clusterInstanceTypesByProvider,
					nil,
				),
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(clusterInstanceTypes),
				},
			},
			"min_running_nodes": {
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters].",
					clusterMinRunningNodesMin,
					&clusterMinRunningNodesDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(clusterMinRunningNodesDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: clusterMinRunningNodesMin},
				},
			},
			"max_running_nodes": {
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters]",
					clusterMaxRunningNodesMin,
					&clusterMaxRunningNodesDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(clusterMaxRunningNodesDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: clusterMaxRunningNodesMin},
				},
			},
			"features": {
				Description: "Features of the cluster.",
				Optional:    true,
				Computed:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"vpc_subnet": {
						Description: descriptions.NewStringDefaultDescription(
							"Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureVpcSubnetDefault,
						),
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							modifiers.NewStringDefaultModifier(clusterFeatureVpcSubnetDefault),
						},
					},
				}),
			},
			"routing_table": {
				Description: "List of routes of the cluster.",
				Optional:    true,
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"description": {
						Description: "Description of the route.",
						Type:        types.StringType,
						Required:    true,
					},
					"destination": {
						Description: "Destination of the route.",
						Type:        types.StringType,
						Required:    true,
					},
					"target": {
						Description: "Target of the route.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"state": {
				Description: descriptions.NewStringEnumDescription(
					"State of the cluster.",
					clusterStates,
					&clusterStateDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(clusterStateDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(clusterStates),
				},
			},
		},
	}, nil
}

func (r clusterResourceType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return clusterResource{
		client: p.(*qProvider).client,
	}, nil
}

type clusterResource struct {
	client *client.Client
}

// Create qovery cluster resource
func (r clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new cluster
	request, err := plan.toUpsertClusterRequest(nil)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	cluster, apiErr := r.client.CreateCluster(ctx, plan.OrganizationId.Value, request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := convertResponseToCluster(cluster)
	tflog.Trace(ctx, "created cluster", map[string]interface{}{"cluster_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery cluster resource
func (r clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from the API
	cluster, apiErr := r.client.GetCluster(ctx, state.OrganizationId.Value, state.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state = convertResponseToCluster(cluster)
	tflog.Trace(ctx, "read cluster", map[string]interface{}{"cluster_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery cluster resource
func (r clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update cluster in the backend
	request, err := plan.toUpsertClusterRequest(&state)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	cluster, apiErr := r.client.UpdateCluster(ctx, state.OrganizationId.Value, state.Id.Value, request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}
	// Update state values
	state = convertResponseToCluster(cluster)
	tflog.Trace(ctx, "updated cluster", map[string]interface{}{"cluster_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery cluster resource
func (r clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete cluster
	apiErr := r.client.DeleteCluster(ctx, state.OrganizationId.Value, state.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted cluster", map[string]interface{}{"cluster_id": state.Id.Value})

	// Remove cluster from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery cluster resource using its id
func (r clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: cluster_id,organization_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}

func readInstanceTypes() (map[string][]string, error) {
	dir := "data/cluster_instance_types"
	files, err := instanceTypes.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	instanceTypesByProvider := map[string][]string{}
	for _, f := range files {
		byteArray, err := instanceTypes.ReadFile(fmt.Sprintf("%s/%s", dir, f.Name()))
		if err != nil {
			return nil, err
		}

		var data []string
		if err := json.NewDecoder(bytes.NewBuffer(byteArray)).Decode(&data); err != nil {
			return nil, err
		}

		provider := strings.Split(f.Name(), ".")[0]
		instanceTypesByProvider[strings.ToUpper(provider)] = data
	}

	return instanceTypesByProvider, nil
}
