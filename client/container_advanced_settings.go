package client

import (
	"context"
	"encoding/json"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getContainerAdvancedSettings(ctx context.Context, containerID string) (*map[string]interface{}, *apierrors.APIError) {
	advancedSettings, res, err := c.api.ContainerConfigurationApi.
		GetContainerAdvancedSettings(ctx, containerID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceContainerAdvancedSettings, containerID, res, err)
	}

	resp, jsonErr := fromContainerAdvancedSettings(advancedSettings)
	if jsonErr != nil {
		return nil, apierrors.NewReadError(apierrors.APIResourceContainerAdvancedSettings, containerID, res, jsonErr)
	}
	return &resp, nil
}

func (c *Client) editContainerAdvancedSettings(ctx context.Context, containerID string, request map[string]interface{}) (*map[string]interface{}, *apierrors.APIError) {
	req, jsonErr := toContainerAdvancedSettings(request)
	if jsonErr != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceContainerAdvancedSettings, containerID, nil, jsonErr)
	}
	advancedSettings, res, err := c.api.ContainerConfigurationApi.
		EditContainerAdvancedSettings(ctx, containerID).
		ContainerAdvancedSettings(req).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceContainerAdvancedSettings, containerID, res, jsonErr)
	}

	resp, jsonErr := fromContainerAdvancedSettings(advancedSettings)
	if jsonErr != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceContainerAdvancedSettings, containerID, res, jsonErr)
	}
	return &resp, nil
}

func fromContainerAdvancedSettings(s *qovery.ContainerAdvancedSettings) (map[string]interface{}, error) {
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

func toContainerAdvancedSettings(s map[string]interface{}) (qovery.ContainerAdvancedSettings, error) {
	resp, marshalErr := json.Marshal(s)
	if marshalErr != nil {
		return qovery.ContainerAdvancedSettings{}, marshalErr
	}

	var result qovery.ContainerAdvancedSettings
	if unmarshalErr := json.Unmarshal(resp, &result); unmarshalErr != nil {
		return qovery.ContainerAdvancedSettings{}, unmarshalErr
	}

	return result, nil
}
