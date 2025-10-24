package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getClusterStatus(ctx context.Context, organizationID string, clusterID string) (*qovery.ClusterStatus, *apierrors.APIError) {
	status, res, err := c.api.ClustersAPI.
		GetClusterStatus(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterStatus, clusterID, res, err)
	}
	return status, nil
}

func (c *Client) updateClusterStatus(ctx context.Context, organizationID string, cluster *qovery.Cluster, desiredState qovery.ClusterStateEnum, forceUpdate bool) (*qovery.ClusterStateEnum, *apierrors.APIError) {
	// wait until we can stop the cluster - otherwise it will fail
	checker := newClusterFinalStateCheckerWaitFunc(c, organizationID, cluster.Id)
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return nil, apiErr
	}

	// Wrap status call with retry logic to handle transient errors (DNS failures, timeouts, etc.)
	var status *qovery.ClusterStatus
	apiErr := retryAPICall(ctx, func(ctx context.Context) *apierrors.APIError {
		var err *apierrors.APIError
		status, err = c.getClusterStatus(ctx, organizationID, cluster.Id)
		return err
	})
	if apiErr != nil {
		return nil, apiErr
	}

	if status.GetStatus() != desiredState || (status.GetStatus() == qovery.CLUSTERSTATEENUM_DEPLOYED && forceUpdate == true) {
		switch desiredState {
		case qovery.CLUSTERSTATEENUM_DEPLOYED:
			return c.deployCluster(ctx, organizationID, cluster)
		case qovery.CLUSTERSTATEENUM_STOPPED:
			return c.stopCluster(ctx, organizationID, cluster)
		}
	}

	return status.Status, nil
}
