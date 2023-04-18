package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getClusterStatus(ctx context.Context, organizationID string, clusterID string) (*qovery.ClusterStatusGet, *apierrors.APIError) {
	status, res, err := c.api.ClustersApi.
		GetClusterStatus(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterStatus, clusterID, res, err)
	}
	return status, nil
}

func (c *Client) updateClusterStatus(ctx context.Context, organizationID string, cluster *qovery.Cluster, desiredState qovery.StateEnum, forceUpdate bool) (*qovery.ClusterStatusGet, *apierrors.APIError) {
	// wait until we can stop the cluster - otherwise it will fail
	checker := newClusterFinalStateCheckerWaitFunc(c, organizationID, cluster.Id)
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return nil, apiErr
	}

	status, apiErr := c.getClusterStatus(ctx, organizationID, cluster.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	if status.GetStatus() != desiredState || (status.GetStatus() == qovery.STATEENUM_DEPLOYED && forceUpdate == true) {
		switch desiredState {
		case qovery.STATEENUM_DEPLOYED:
			return c.deployCluster(ctx, organizationID, cluster)
		case qovery.STATEENUM_STOPPED:
			return c.stopCluster(ctx, organizationID, cluster)
		}
	}

	return status, nil
}
