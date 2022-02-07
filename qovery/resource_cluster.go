package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"
)

type resourceClusterType struct{}

// Cluster Resource schema
func (r resourceClusterType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				Required: false,
				Optional: false,
			},
			"name": {
				Type:     types.StringType,
				Computed: false,
				Required: true,
				Optional: false,
			},
			"region": {
				Type:     types.StringType,
				Computed: false,
				Required: true,
				Optional: false,
			},
			"cloud_provider": {
				Type:     types.StringType,
				Computed: false,
				Required: true,
				Optional: false,
			},
			"credentials_id": {
				Type:     types.StringType,
				Computed: false,
				Required: true,
				Optional: false,
			},
			"organization_id": {
				Type:     types.StringType,
				Computed: false,
				Required: true,
				Optional: false,
			},
		},
	}, nil
}

// New resource instance
func (r resourceClusterType) NewResource(ct context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceCluster{
		p: *(p.(*provider)),
	}, nil
}

type resourceCluster struct {
	p provider
}

// Create a new resource
func (r resourceCluster) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks! ",
		)
		return
	}

	// Retrieve values from plan
	var plan Cluster
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new cluster
	cluster, res, err := r.p.client.ClustersApi.
		CreateCluster(ctx, plan.OrganizationId.Value).
		ClusterRequest(
			qovery.ClusterRequest{
				Name:          plan.Name.Value,
				CloudProvider: plan.CloudProvider.Value,
				Region:        plan.Region.Value,
			}).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			"Could not create cluster, unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			"Could not create cluster, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	// Generate resource state struct
	var state = Cluster{
		Id: types.String{
			Value: cluster.Id,
		},
		OrganizationId: plan.OrganizationId,
		Name: types.String{
			Value: cluster.Name,
		},
		CloudProvider: types.String{
			Value: cluster.CloudProvider,
		},
		Region: types.String{
			Value: cluster.Region,
		},
	}

	// Specify cluster credentials
	clusterInfo, res, err := r.p.client.ClustersApi.
		SpecifyClusterCloudProviderInfo(ctx, plan.OrganizationId.Value, state.Id.Value).
		ClusterCloudProviderInfoRequest(
			qovery.ClusterCloudProviderInfoRequest{
				CloudProvider: &plan.CloudProvider.Value,
				Credentials: &qovery.ClusterCloudProviderInfoRequestCredentials{
					Id:   &plan.CredentialsId.Value,
					Name: &plan.Name.Value,
				},
				Region: &plan.Region.Value,
			}).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error specifying cluster info",
			"Could not specify cluster info, unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error specifying cluster info",
			"Could not specify cluster info, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}
	state.CredentialsId = types.String{
		Value: *clusterInfo.Credentials.Id,
	}

	// Deploy cluster
	r.p.client.GetConfig().AddDefaultHeader("content-type", "application/json")
	_, res, err = r.p.client.ClustersApi.
		DeployCluster(ctx, plan.OrganizationId.Value, state.Id.Value).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deploying cluster",
			"Could not deploy cluster, unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			"Could not deploy cluster, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r resourceCluster) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from the API
	clusters, res, err := r.p.client.ClustersApi.
		ListOrganizationCluster(ctx, state.OrganizationId.Value).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cluster",
			"Could not read cluster "+state.Id.Value+", unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error reading cluster",
			"Could not read cluster, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	var found = Cluster{}
	for _, cluster := range clusters.GetResults() {
		if state.Id.Value == cluster.Id {
			found = Cluster{
				Id: types.String{
					Value: cluster.Id,
				},
				OrganizationId: state.OrganizationId,
				Name: types.String{
					Value: cluster.Name,
				},
				CloudProvider: types.String{
					Value: cluster.CloudProvider,
				},
				Region: types.String{
					Value: cluster.Region,
				},
			}
		}
	}

	cloudCredentials, res, err := r.p.client.ClustersApi.
		GetOrganizationCloudProviderInfo(ctx, state.OrganizationId.Value, state.Id.Value).
		Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cluster credentials",
			"Could not read cluster credentials for cluster "+state.Id.Value+", unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			"Could not read cluster credentials for cluster "+state.Id.Value+", unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}

	state.CredentialsId = types.String{
		Value: *cloudCredentials.Credentials.Id,
	}
	state.Name = found.Name
	state.Id = found.Id
	state.Region = found.Region
	state.CloudProvider = found.CloudProvider

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceCluster) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,cluster_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("organization_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), idParts[1])...)
}

// Update resource
func (r resourceCluster) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	{
		// Get current state
		var plan Cluster
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Update cluster
		cluster, res, err := r.p.client.ClustersApi.
			EditCluster(ctx, plan.OrganizationId.Value, plan.Id.Value).
			ClusterRequest(
				qovery.ClusterRequest{
					Name:          plan.Name.Value,
					CloudProvider: plan.CloudProvider.Value,
					Region:        plan.CloudProvider.Value,
				},
			).
			Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating cluster",
				"Could not update cluster, unexpected error: "+err.Error(),
			)
			return
		}
		if res.StatusCode >= 400 {
			resp.Diagnostics.AddError(
				"Error updating cluster",
				"Could not update cluster, unexpected status code: "+string(rune(res.StatusCode)),
			)
			return
		}

		// Update state
		var state Cluster
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		req.State.Get(ctx, state)

		state.Id.Value = cluster.Id
		state.Name.Value = cluster.Name
		state.CloudProvider.Value = cluster.CloudProvider
		state.Region.Value = cluster.Region

		// Set state
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

// Delete resource
func (r resourceCluster) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state Cluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete cluster
	res, err := r.p.client.ClustersApi.DeleteCluster(ctx, state.OrganizationId.Value, state.Id.Value).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting cluster",
			"Could not delete cluster "+state.Id.Value+", unexpected error: "+err.Error(),
		)
		return
	}
	if res.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			"Could not delete cluster, unexpected status code: "+string(rune(res.StatusCode)),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
