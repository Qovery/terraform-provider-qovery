package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) GetClusterDNSProvider(ctx context.Context, clusterID string) (*qovery.ClusterDnsProviderResponse, *apierrors.APIError) {
	dnsProvider, res, err := c.api.ClustersAPI.
		GetClusterDnsProvider(ctx, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterDNSProvider, clusterID, res, err)
	}

	return dnsProvider, nil
}

func (c *Client) UpdateClusterDNSProvider(ctx context.Context, clusterID string, request qovery.ClusterDnsProviderRequest) (*qovery.ClusterDnsProviderResponse, *apierrors.APIError) {
	dnsProvider, res, err := c.api.ClustersAPI.
		EditClusterDnsProvider(ctx, clusterID).
		ClusterDnsProviderRequest(request).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterDNSProvider, clusterID, res, err)
	}

	return dnsProvider, nil
}
