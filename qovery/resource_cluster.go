package qovery

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/qovery/apierror"
	"terraform-provider-qovery/qovery/descriptions"
	"terraform-provider-qovery/qovery/validators"
)

const (
	clusterAPIResource       = "cluster"
	cloudProviderAPIResource = "cloud provider"
)

var cloudProviders = []string{"AWS", "DIGITAL_OCEAN", "SCALEWAY"}

type clusterResourceData struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	CredentialsId  types.String `tfsdk:"credentials_id"`
	Name           types.String `tfsdk:"name"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
	Region         types.String `tfsdk:"region"`
}

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
	var plan clusterResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new cluster
	cluster, res, err := r.client.ClustersApi.
		CreateCluster(ctx, plan.OrganizationId.Value).
		ClusterRequest(qovery.ClusterRequest{
			Name:          plan.Name.Value,
			CloudProvider: plan.CloudProvider.Value,
			Region:        plan.Region.Value,
		}).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := clusterCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Specify cluster credentials
	clusterInfo, res, err := r.client.ClustersApi.
		SpecifyClusterCloudProviderInfo(ctx, plan.OrganizationId.Value, cluster.Id).
		ClusterCloudProviderInfoRequest(qovery.ClusterCloudProviderInfoRequest{
			CloudProvider: &plan.CloudProvider.Value,
			Credentials: &qovery.ClusterCloudProviderInfoRequestCredentials{
				Id:   &plan.CredentialsId.Value,
				Name: &plan.Name.Value,
			},
			Region: &plan.Region.Value,
		}).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := cloudProviderCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Deploy cluster
	r.client.GetConfig().AddDefaultHeader("content-type", "application/json")
	_, res, err = r.client.ClustersApi.
		DeployCluster(ctx, plan.OrganizationId.Value, cluster.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := clusterDeployAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := clusterResourceData{
		Id:             types.String{Value: cluster.Id},
		CredentialsId:  types.String{Value: *clusterInfo.Credentials.Id},
		OrganizationId: plan.OrganizationId,
		Name:           types.String{Value: cluster.Name},
		CloudProvider:  types.String{Value: cluster.CloudProvider},
		Region:         types.String{Value: cluster.Region},
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery cluster resource
func (r clusterResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state clusterResourceData
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

	var toRefresh *clusterResourceData
	for _, cluster := range clusters.GetResults() {
		if state.Id.Value == cluster.Id {
			toRefresh = &clusterResourceData{
				CredentialsId: types.String{Value: *cloudProviderInfo.Credentials.Id},
				Name:          types.String{Value: cluster.Name},
				CloudProvider: types.String{Value: cluster.CloudProvider},
				Region:        types.String{Value: cluster.Region},
			}
			break
		}
	}

	// If cluster id is not in list
	// Returning Not Found error
	if toRefresh == nil {
		res.StatusCode = 404
		apiErr := clusterReadAPIError(state.Id.Value, res, nil)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state.CredentialsId = toRefresh.CredentialsId
	state.Name = toRefresh.Name
	state.CloudProvider = toRefresh.CloudProvider
	state.Region = toRefresh.Region

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery cluster resource
func (r clusterResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state clusterResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update cluster in the backend
	cluster, res, err := r.client.ClustersApi.
		EditCluster(ctx, plan.OrganizationId.Value, plan.Id.Value).
		ClusterRequest(qovery.ClusterRequest{
			Name:          plan.Name.Value,
			CloudProvider: plan.CloudProvider.Value,
			Region:        plan.Region.Value,
		}).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := clusterUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	toUpdate := clusterResourceData{
		Name:          types.String{Value: cluster.Name},
		CloudProvider: types.String{Value: cluster.CloudProvider},
		Region:        types.String{Value: cluster.Region},
	}

	// Update state values
	state.Name = toUpdate.Name
	state.CloudProvider = toUpdate.CloudProvider
	state.Region = toUpdate.Region

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery cluster resource
func (r clusterResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state clusterResourceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete cluster
	res, err := r.client.ClustersApi.DeleteCluster(ctx, state.OrganizationId.Value, state.Id.Value).Execute()
	if err != nil || res.StatusCode >= 300 {
		apiErr := clusterDeleteAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

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

func clusterCreateAPIError(clusterID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(clusterAPIResource, clusterID, apierror.Create, res, err)
}

func clusterReadAPIError(clusterID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(clusterAPIResource, clusterID, apierror.Read, res, err)
}

func clusterUpdateAPIError(clusterID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(clusterAPIResource, clusterID, apierror.Update, res, err)
}

func clusterDeleteAPIError(clusterID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(clusterAPIResource, clusterID, apierror.Delete, res, err)
}

func clusterDeployAPIError(clusterID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(clusterAPIResource, clusterID, apierror.Deploy, res, err)
}

func cloudProviderCreateAPIError(clusterID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(cloudProviderAPIResource, clusterID, apierror.Create, res, err)
}

func cloudProviderReadAPIError(clusterID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(cloudProviderAPIResource, clusterID, apierror.Read, res, err)
}
