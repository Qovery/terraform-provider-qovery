package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type ClusterResponse struct {
	OrganizationID  string
	ClusterResponse *qovery.ClusterResponse
	ClusterInfo     *qovery.ClusterCloudProviderInfoResponse
}

type ClusterUpsertParams struct {
	ClusterRequest              qovery.ClusterRequest
	ClusterCloudProviderRequest *qovery.ClusterCloudProviderInfoRequest
	DesiredState                string
}

func (c *Client) CreateCluster(ctx context.Context, organizationID string, params ClusterUpsertParams) (*ClusterResponse, *apierrors.APIError) {
	cluster, res, err := c.api.ClustersApi.
		CreateCluster(ctx, organizationID).
		ClusterRequest(params.ClusterRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceCluster, params.ClusterRequest.Name, res, err)
	}
	return c.updateCluster(ctx, organizationID, cluster, params)
}

func (c *Client) GetCluster(ctx context.Context, organizationID string, clusterID string) (*ClusterResponse, *apierrors.APIError) {
	cluster, apiErr := c.getClusterByID(ctx, organizationID, clusterID)
	if apiErr != nil {
		return nil, apiErr
	}

	clusterInfo, res, err := c.api.ClustersApi.
		GetOrganizationCloudProviderInfo(ctx, organizationID, cluster.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceClusterCloudProvider, cluster.Id, res, err)
	}

	return &ClusterResponse{
		OrganizationID:  organizationID,
		ClusterResponse: cluster,
		ClusterInfo:     clusterInfo,
	}, nil
}
func (c *Client) UpdateCluster(ctx context.Context, organizationID string, clusterID string, params ClusterUpsertParams) (*ClusterResponse, *apierrors.APIError) {
	cluster, res, err := c.api.ClustersApi.
		EditCluster(ctx, organizationID, clusterID).
		ClusterRequest(params.ClusterRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceCluster, clusterID, res, err)
	}

	return c.updateCluster(ctx, organizationID, cluster, params)
}

func (c *Client) DeleteCluster(ctx context.Context, organizationID string, clusterID string) *apierrors.APIError {
	finalStateChecker := newClusterFinalStateCheckerWaitFunc(c, organizationID, clusterID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err := c.api.ClustersApi.
		DeleteCluster(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceCluster, clusterID, res, err)
	}

	checker := newClusterStatusCheckerWaitFunc(c, organizationID, clusterID, "DELETED")
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return apiErr
	}
	return nil
}

func (c *Client) getClusterByID(ctx context.Context, organizationID string, clusterID string) (*qovery.ClusterResponse, *apierrors.APIError) {
	clusters, res, err := c.api.ClustersApi.
		ListOrganizationCluster(ctx, organizationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceCluster, clusterID, res, err)
	}

	for _, cluster := range clusters.GetResults() {
		if cluster.Id == clusterID {
			return &cluster, nil
		}
	}

	return nil, apierrors.NewReadError(apierrors.APIResourceCluster, clusterID, res, err)
}

func (c *Client) updateCluster(ctx context.Context, organizationID string, cluster *qovery.ClusterResponse, params ClusterUpsertParams) (*ClusterResponse, *apierrors.APIError) {
	if params.ClusterCloudProviderRequest != nil {
		_, res, err := c.api.ClustersApi.
			SpecifyClusterCloudProviderInfo(ctx, organizationID, cluster.Id).
			ClusterCloudProviderInfoRequest(*params.ClusterCloudProviderRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterCloudProvider, cluster.Id, res, err)
		}
	}

	clusterInfo, res, err := c.api.ClustersApi.
		GetOrganizationCloudProviderInfo(ctx, organizationID, cluster.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterCloudProvider, cluster.Id, res, err)
	}

	clusterStatus, apiErr := c.updateClusterStatus(ctx, organizationID, cluster, params.DesiredState)
	if apiErr != nil {
		return nil, apiErr
	}
	cluster.Status = clusterStatus.Status

	return &ClusterResponse{
		OrganizationID:  organizationID,
		ClusterResponse: cluster,
		ClusterInfo:     clusterInfo,
	}, nil
}
