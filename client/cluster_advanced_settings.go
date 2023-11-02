package client

import (
	"context"
	"encoding/json"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getClusterAdvancedSettings(ctx context.Context, organizationID string, clusterID string) (*map[string]interface{}, *apierrors.APIError) {
	advancedSettings, res, err := c.api.ClustersAPI.
		GetClusterAdvancedSettings(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterAdvancedSettings, clusterID, res, err)
	}

	resp, jsonErr := fromClusterAdvancedSettings(advancedSettings)
	if jsonErr != nil {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterAdvancedSettings, clusterID, res, jsonErr)
	}
	return &resp, nil
}

func (c *Client) editClusterAdvancedSettings(ctx context.Context, organizationID string, clusterID string, request map[string]interface{}) (*map[string]interface{}, *apierrors.APIError) {
	req, jsonErr := toClusterAdvancedSettings(request)
	if jsonErr != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterAdvancedSettings, clusterID, nil, jsonErr)
	}
	advancedSettings, res, err := c.api.ClustersAPI.
		EditClusterAdvancedSettings(ctx, organizationID, clusterID).
		ClusterAdvancedSettings(req).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterAdvancedSettings, clusterID, res, jsonErr)
	}

	resp, jsonErr := fromClusterAdvancedSettings(advancedSettings)
	if jsonErr != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterAdvancedSettings, clusterID, res, jsonErr)
	}
	return &resp, nil
}

func fromClusterAdvancedSettings(s *qovery.ClusterAdvancedSettings) (map[string]interface{}, error) {
	// API return nil when map is empty
	if s.CloudProviderContainerRegistryTags == nil {
		s.CloudProviderContainerRegistryTags = &map[string]string{}
	}

	resp, marshalErr := json.Marshal(s)
	if marshalErr != nil {
		return nil, marshalErr
	}

	var unmarshal map[string]interface{}
	if unmarshalErr := json.Unmarshal(resp, &unmarshal); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return unmarshal, nil
}

func toClusterAdvancedSettings(s map[string]interface{}) (qovery.ClusterAdvancedSettings, error) {
	resp, marshalErr := json.Marshal(s)
	if marshalErr != nil {
		return qovery.ClusterAdvancedSettings{}, marshalErr
	}

	var result qovery.ClusterAdvancedSettings
	if unmarshalErr := json.Unmarshal(resp, &result); unmarshalErr != nil {
		return qovery.ClusterAdvancedSettings{}, unmarshalErr
	}

	return result, nil
}
