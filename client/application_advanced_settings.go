package client

import (
	"context"
	"encoding/json"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationAdvancedSettings(ctx context.Context, applicationID string) (*map[string]interface{}, *apierrors.APIError) {
	advancedSettings, res, err := c.api.ApplicationConfigurationApi.
		GetAdvancedSettings(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationAdvancedSettings, applicationID, res, err)
	}

	resp, jsonErr := fromApplicationAdvancedSettings(advancedSettings)
	if jsonErr != nil {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationAdvancedSettings, applicationID, res, jsonErr)
	}
	return &resp, nil
}

func (c *Client) editApplicationAdvancedSettings(ctx context.Context, applicationID string, request map[string]interface{}) (*map[string]interface{}, *apierrors.APIError) {
	req, jsonErr := toApplicationAdvancedSettings(request)
	if jsonErr != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplicationAdvancedSettings, applicationID, nil, jsonErr)
	}
	advancedSettings, res, err := c.api.ApplicationConfigurationApi.
		EditAdvancedSettings(ctx, applicationID).
		ApplicationAdvancedSettings(req).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplicationAdvancedSettings, applicationID, res, jsonErr)
	}

	resp, jsonErr := fromApplicationAdvancedSettings(advancedSettings)
	if jsonErr != nil {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplicationAdvancedSettings, applicationID, res, jsonErr)
	}
	return &resp, nil
}

func fromApplicationAdvancedSettings(s *qovery.ApplicationAdvancedSettings) (map[string]interface{}, error) {
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

func toApplicationAdvancedSettings(s map[string]interface{}) (qovery.ApplicationAdvancedSettings, error) {
	resp, marshalErr := json.Marshal(s)
	if marshalErr != nil {
		return qovery.ApplicationAdvancedSettings{}, marshalErr
	}

	var result qovery.ApplicationAdvancedSettings
	if unmarshalErr := json.Unmarshal(resp, &result); unmarshalErr != nil {
		return qovery.ApplicationAdvancedSettings{}, unmarshalErr
	}

	return result, nil
}
