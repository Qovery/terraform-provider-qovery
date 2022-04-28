package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) GetClusterInstanceTypes(ctx context.Context) (map[qovery.CloudProviderEnum]*qovery.ClusterInstanceTypeResponseList, *apierrors.APIError) {
	awsInstanceTypes, res, err := c.api.CloudProviderApi.
		ListAWSInstanceType(ctx).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceClusterInstanceType, "", res, err)
	}

	doInstanceTypes, res, err := c.api.CloudProviderApi.
		ListDOInstanceType(ctx).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceClusterInstanceType, "", res, err)
	}

	scwInstanceTypes, res, err := c.api.CloudProviderApi.
		ListScalewayInstanceType(ctx).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceClusterInstanceType, "", res, err)
	}

	return map[qovery.CloudProviderEnum]*qovery.ClusterInstanceTypeResponseList{
		qovery.CLOUDPROVIDERENUM_AWS:           awsInstanceTypes,
		qovery.CLOUDPROVIDERENUM_DIGITAL_OCEAN: doInstanceTypes,
		qovery.CLOUDPROVIDERENUM_SCALEWAY:      scwInstanceTypes,
	}, nil
}
