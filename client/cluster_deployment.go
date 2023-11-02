package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) deployCluster(ctx context.Context, organizationID string, cluster *qovery.Cluster) (*qovery.ClusterStateEnum, *apierrors.APIError) {
	// deploy cluster even if it is already in the DEPLOYED state to apply any changed done with editCluster
	_, res, err := c.api.ClustersAPI.
		DeployCluster(ctx, organizationID, cluster.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewDeployError(apierrors.APIResourceCluster, cluster.Id, res, err)
	}

	statusChecker := newClusterStatusCheckerWaitFunc(c, organizationID, cluster.Id, qovery.CLUSTERSTATEENUM_DEPLOYED)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	clusterStatus, apiError := c.getClusterStatus(ctx, organizationID, cluster.Id)
	if apiError != nil {
		return nil, apiError
	}
	return clusterStatus.Status, nil
}

func (c *Client) stopCluster(ctx context.Context, organizationID string, cluster *qovery.Cluster) (*qovery.ClusterStateEnum, *apierrors.APIError) {
	status, apiErr := c.getClusterStatus(ctx, organizationID, cluster.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.GetStatus() {
	case qovery.CLUSTERSTATEENUM_STOPPED, qovery.CLUSTERSTATEENUM_READY:
		status := qovery.CLUSTERSTATEENUM_STOPPED
		return &status, nil
	default:
		_, res, err := c.api.ClustersAPI.
			StopCluster(ctx, organizationID, cluster.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewStopError(apierrors.APIResourceCluster, cluster.Id, res, err)
		}
	}

	statusChecker := newClusterStatusCheckerWaitFunc(c, organizationID, cluster.Id, qovery.CLUSTERSTATEENUM_STOPPED)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	clusterStatus, apiError := c.getClusterStatus(ctx, organizationID, cluster.Id)
	if apiError != nil {
		return nil, apiErr
	}
	return clusterStatus.Status, nil
}
