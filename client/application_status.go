package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationStatus(ctx context.Context, applicationID string) (*qovery.Status, *apierrors.APIError) {
	status, res, err := c.api.ApplicationMainCallsAPI.
		GetApplicationStatus(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationStatus, applicationID, res, err)
	}

	// Handle READY as STOPPED state
	if status.State == qovery.STATEENUM_READY {
		status.State = qovery.STATEENUM_STOPPED
	}
	return status, nil
}
