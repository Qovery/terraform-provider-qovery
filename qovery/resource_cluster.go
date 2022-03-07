package qovery

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
	"terraform-provider-qovery/qovery/descriptions"
	"terraform-provider-qovery/qovery/modifiers"
	"terraform-provider-qovery/qovery/validators"
)

const (
	clusterAPIResource       = "cluster"
	clusterStatusAPIResource = "cluster status"
	cloudProviderAPIResource = "cloud provider"
)

var ()

var (
	// Cluster State
	clusterStateRunning = "RUNNING"
	clusterStateStopped = "STOPPED"
	clusterStates       = []string{clusterStateRunning, clusterStateStopped}
	clusterStateDefault = clusterStateRunning

	// Cluster Description
	clusterDescriptionDefault = ""

	// Cloud Provider
	cloudProviders = []string{"AWS", "DIGITAL_OCEAN", "SCALEWAY"}

	// Cluster CPU
	clusterCPUMin     int64 = 2000 // in MB
	clusterCPUDefault int64 = 2000 // in MB

	// Cluster Memory
	clusterMemoryMin     int64 = 4096 // in MB
	clusterMemoryDefault int64 = 4096 // in MB

	// Cluster Min Running Nodes
	clusterMinRunningNodesMin     int64 = 3
	clusterMinRunningNodesDefault int64 = 3

	// Cluster Max Running Nodes
	clusterMaxRunningNodesMin     int64 = 3
	clusterMaxRunningNodesDefault int64 = 10
)

type clusterResourceType struct{}

func (r clusterResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
					validators.StringEnumValidator{Enum: cloudProviders},
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
			"cpu": {
				Description: descriptions.NewInt64MinDescription(
					"CPU of the cluster in millicores (m) [1000m = 1 CPU].",
					clusterCPUMin,
					&clusterCPUDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: clusterCPUMin},
				},
			},
			"memory": {
				Description: descriptions.NewInt64MinDescription(
					"RAM of the cluster in MB [1024MB = 1GB].",
					clusterMemoryMin,
					&clusterMemoryDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: clusterMemoryMin},
				},
			},
			"min_running_nodes": {
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of nodes running for the cluster.",
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
					"Maximum number of nodes running for the cluster.",
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
					validators.StringEnumValidator{Enum: clusterStates},
				},
			},
			//"timeouts": NewTimeoutSchemaAttribute(TimeoutParams{
			//	ResourceName:  "cluster",
			//	CreateDefault: "40m",
			//	UpdateDefault: "40m",
			//	DeleteDefault: "40m",
			//}),
		},
	}, nil
}

func (r clusterResourceType) NewResource(ct context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return clusterResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type clusterResource struct {
	client *qovery.APIClient
}

// Create qovery cluster resource
func (r clusterResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Retrieve values from plan
	var plan Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new cluster
	cluster, res, err := r.client.ClustersApi.
		CreateCluster(ctx, plan.OrganizationId.Value).
		ClusterRequest(plan.toUpsertClusterRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := clusterCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Specify cluster credentials
	clusterInfo, res, err := r.client.ClustersApi.
		SpecifyClusterCloudProviderInfo(ctx, plan.OrganizationId.Value, cluster.Id).
		ClusterCloudProviderInfoRequest(plan.toUpdateClusterCloudProviderInfoRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := cloudProviderCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	if plan.State.Value == clusterStateRunning {
		apiErr := r.deployCluster(ctx, plan.OrganizationId.Value, cluster.Id)
		if apiErr != nil {
			resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
			return
		}
	}

	// Initialize state values
	state := convertResponseToCluster(cluster, clusterInfo, plan)
	tflog.Trace(ctx, "created cluster", "cluster_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery cluster resource
func (r clusterResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from the API
	clusters, res, err := r.client.ClustersApi.
		ListOrganizationCluster(ctx, state.OrganizationId.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := clusterReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Get cluster credentials from the API
	cloudProviderInfo, res, err := r.client.ClustersApi.
		GetOrganizationCloudProviderInfo(ctx, state.OrganizationId.Value, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := cloudProviderReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	found := false
	for _, cluster := range clusters.GetResults() {
		if state.Id.Value == cluster.Id {
			found = true
			state = convertResponseToCluster(&cluster, cloudProviderInfo, state)
			break
		}
	}

	// If cluster id is not in list
	// Returning Not Found error
	if !found {
		res.StatusCode = 404
		apiErr := clusterReadAPIError(state.Id.Value, res, nil)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "read cluster", "cluster_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery cluster resource
func (r clusterResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update cluster in the backend
	r.client.GetConfig().AddDefaultHeader("content-type", "application/json")
	cluster, res, err := r.client.ClustersApi.
		EditCluster(ctx, state.OrganizationId.Value, state.Id.Value).
		ClusterRequest(plan.toUpsertClusterRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := clusterUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	if plan.CredentialsId != state.CredentialsId {
		// Specify cluster credentials
		_, res, err := r.client.ClustersApi.
			SpecifyClusterCloudProviderInfo(ctx, plan.OrganizationId.Value, cluster.Id).
			ClusterCloudProviderInfoRequest(plan.toUpdateClusterCloudProviderInfoRequest()).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			apiErr := cloudProviderUpdateAPIError(plan.Name.Value, res, err)
			resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
			return
		}
	}

	cloudProviderInfo, res, err := r.client.ClustersApi.
		GetOrganizationCloudProviderInfo(ctx, state.OrganizationId.Value, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := cloudProviderUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	if plan.State.Value == clusterStateRunning && *cluster.Status != clusterStateRunning {
		apiErr := r.deployCluster(ctx, plan.OrganizationId.Value, cluster.Id)
		if apiErr != nil {
			resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
			return
		}
	}

	if plan.State.Value == clusterStateStopped && *cluster.Status != clusterStateStopped {
		apiErr := r.stopCluster(ctx, plan.OrganizationId.Value, cluster.Id)
		if apiErr != nil {
			resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
			return
		}
	}

	// Update state values
	state = convertResponseToCluster(cluster, cloudProviderInfo, plan)
	tflog.Trace(ctx, "updated cluster", "cluster_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery cluster resource
func (r clusterResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete cluster
	res, err := r.client.ClustersApi.
		DeleteCluster(ctx, state.OrganizationId.Value, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		apiErr := clusterDeleteAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted cluster", "cluster_id", state.Id.Value)

	// Remove cluster from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery cluster resource using its id
func (r clusterResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: cluster_id,organization_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("organization_id"), idParts[1])...)
}

func (r clusterResource) deployCluster(ctx context.Context, organizationID string, clusterID string) *apierrors.APIError {
	// Deploy cluster
	r.client.GetConfig().AddDefaultHeader("content-type", "application/json")
	_, res, err := r.client.ClustersApi.
		DeployCluster(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return clusterDeployAPIError(clusterID, res, err)
	}

	ticker := time.NewTicker(10 * time.Second)
	timeout := time.NewTicker(time.Hour)
	for {
		select {
		case <-timeout.C:
			return clusterDeployAPIError(clusterID, res, err)
		case <-ticker.C:
			status, res, err := r.client.ClustersApi.
				GetClusterStatus(ctx, organizationID, clusterID).
				Execute()
			if err != nil || res.StatusCode >= 400 {
				return clusterStatusReadAPIError(clusterID, res, err)
			}
			if *status.IsDeployed || IsStatusError(status.GetStatus()) {
				return nil
			}
		}
	}
}

func (r clusterResource) stopCluster(ctx context.Context, organizationID string, clusterID string) *apierrors.APIError {
	// Stop cluster
	r.client.GetConfig().AddDefaultHeader("content-type", "application/json")
	_, res, err := r.client.ClustersApi.
		StopCluster(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return clusterDeployAPIError(clusterID, res, err)
	}

	ticker := time.NewTicker(10 * time.Second)
	timeout := time.NewTicker(time.Hour)
	for {
		select {
		case <-timeout.C:
			return clusterDeployAPIError(clusterID, res, err)
		case <-ticker.C:
			status, res, err := r.client.ClustersApi.
				GetClusterStatus(ctx, organizationID, clusterID).
				Execute()
			if err != nil || res.StatusCode >= 400 {
				return clusterStatusReadAPIError(clusterID, res, err)
			}
			if *status.Status == clusterStateStopped || IsStatusError(status.GetStatus()) {
				return nil
			}
		}
	}
}

func clusterCreateAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionCreate, clusterAPIResource, clusterID, res, err)
}

func clusterReadAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionRead, clusterAPIResource, clusterID, res, err)
}

func clusterUpdateAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionUpdate, clusterAPIResource, clusterID, res, err)
}

func clusterDeleteAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionDelete, clusterAPIResource, clusterID, res, err)
}

func clusterDeployAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionDeploy, clusterAPIResource, clusterID, res, err)
}

func clusterStatusReadAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionRead, clusterStatusAPIResource, clusterID, res, err)
}

func cloudProviderCreateAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionCreate, cloudProviderAPIResource, clusterID, res, err)
}

func cloudProviderUpdateAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionUpdate, cloudProviderAPIResource, clusterID, res, err)
}

func cloudProviderReadAPIError(clusterID string, res *http.Response, err error) *apierrors.APIError {
	return apierrors.NewError(apierrors.APIActionRead, cloudProviderAPIResource, clusterID, res, err)
}
