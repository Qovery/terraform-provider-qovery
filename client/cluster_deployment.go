package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
)

func (c *Client) deployCluster(ctx context.Context, organizationID string, cluster *qovery.ClusterResponse) (*qovery.ClusterStatusResponse, *apierrors.APIError) {
	status, apiErr := c.getClusterStatus(ctx, organizationID, cluster.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.GetStatus() {
	case clusterStateRunning:
		return status, nil
	default:
		_, res, err := c.API.ClustersApi.
			DeployCluster(ctx, organizationID, cluster.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeployError(apierrors.APIResourceCluster, cluster.Id, res, err)
		}
	}

	statusChecker := newClusterStatusCheckerWaitFunc(c, organizationID, cluster.Id, clusterStateRunning)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getClusterStatus(ctx, organizationID, cluster.Id)
}

func (c *Client) stopCluster(ctx context.Context, organizationID string, cluster *qovery.ClusterResponse) (*qovery.ClusterStatusResponse, *apierrors.APIError) {
	status, apiErr := c.getClusterStatus(ctx, organizationID, cluster.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.GetStatus() {
	case clusterStateStopped:
		return status, nil
	default:
		_, res, err := c.API.ClustersApi.
			StopCluster(ctx, organizationID, cluster.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewStopError(apierrors.APIResourceCluster, cluster.Id, res, err)
		}
	}

	statusChecker := newClusterStatusCheckerWaitFunc(c, organizationID, cluster.Id, clusterStateStopped)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getClusterStatus(ctx, organizationID, cluster.Id)
}
