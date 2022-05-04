package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) restartEnvironment(ctx context.Context, environmentID string) (*qovery.Status, *apierrors.APIError) {
	envFinalStateChecker := newEnvironmentFinalStateCheckerWaitFunc(c, environmentID)
	if apiErr := wait(ctx, envFinalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	_, res, err := c.api.EnvironmentActionsApi.
		RestartEnvironment(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewRestartError(apierrors.APIResourceEnvironment, environmentID, res, err)
	}

	statusChecker := newEnvironmentStatusCheckerWaitFunc(c, environmentID, qovery.STATEENUM_RUNNING)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getEnvironmentStatus(ctx, environmentID)
}
