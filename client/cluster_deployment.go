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

	// Wrap final status call with retry logic to handle transient errors (DNS failures, timeouts, etc.)
	var clusterStatus *qovery.ClusterStatus
	apiError := retryAPICall(ctx, func(ctx context.Context) *apierrors.APIError {
		var err *apierrors.APIError
		clusterStatus, err = c.getClusterStatus(ctx, organizationID, cluster.Id)
		return err
	})
	if apiError != nil {
		return nil, apiError
	}
	return clusterStatus.Status, nil
}

func (c *Client) stopCluster(ctx context.Context, organizationID string, cluster *qovery.Cluster) (*qovery.ClusterStateEnum, *apierrors.APIError) {
	// Wrap initial status check with retry logic to handle transient errors (DNS failures, timeouts, etc.)
	var status *qovery.ClusterStatus
	apiErr := retryAPICall(ctx, func(ctx context.Context) *apierrors.APIError {
		var err *apierrors.APIError
		status, err = c.getClusterStatus(ctx, organizationID, cluster.Id)
		return err
	})
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

	// Wrap final status call with retry logic to handle transient errors (DNS failures, timeouts, etc.)
	var clusterStatus *qovery.ClusterStatus
	apiError := retryAPICall(ctx, func(ctx context.Context) *apierrors.APIError {
		var err *apierrors.APIError
		clusterStatus, err = c.getClusterStatus(ctx, organizationID, cluster.Id)
		return err
	})
	if apiError != nil {
		return nil, apiErr
	}
	return clusterStatus.Status, nil
}
