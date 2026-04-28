package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getEnvironmentBuiltInEnvironmentVariables(ctx context.Context, environmentID string) ([]*qovery.EnvironmentVariable, *apierrors.APIError) {
	vars, res, err := c.api.EnvironmentVariableAPI.
		ListEnvironmentEnvironmentVariable(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceEnvironmentEnvironmentVariable, environmentID, res, err)
	}
	return environmentVariableResponseListToArray(vars, qovery.APIVARIABLESCOPEENUM_BUILT_IN), nil
}
