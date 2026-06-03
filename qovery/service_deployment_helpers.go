package qovery

import (
	"context"
	"fmt"
	"time"

	"github.com/qovery/qovery-client-go"
)

// serviceKind identifies which list to scan inside an EnvironmentStatuses payload.
type serviceKind int

const (
	serviceKindTerraform serviceKind = iota
	serviceKindHelm
)

const (
	// Total budget for a per-service deployment to reach a terminal state.
	serviceDeployWaitTimeout = 1 * time.Hour
	// Poll interval against the env-level status endpoint.
	serviceDeployPollInterval = 10 * time.Second
)

// waitForServiceDeployed polls the environment-level status endpoint until the
// targeted service reaches a terminal state, then returns nil on success or an
// error describing the terminal failure.
//
// The env-level status payload is the only source the API exposes today for
// per-terraform-service status; we reuse it for helm as well to keep both
// resource types on the same code path.
func waitForServiceDeployed(
	ctx context.Context,
	api *qovery.APIClient,
	environmentID string,
	serviceID string,
	kind serviceKind,
) error {
	deadline := time.NewTimer(serviceDeployWaitTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(serviceDeployPollInterval)
	defer ticker.Stop()

	for {
		state, err := readServiceState(ctx, api, environmentID, serviceID, kind)
		if err != nil {
			return err
		}

		switch state {
		case qovery.STATEENUM_DEPLOYED:
			return nil
		case qovery.STATEENUM_BUILD_ERROR,
			qovery.STATEENUM_DEPLOYMENT_ERROR,
			qovery.STATEENUM_RESTART_ERROR,
			qovery.STATEENUM_STOP_ERROR,
			qovery.STATEENUM_DELETE_ERROR,
			qovery.STATEENUM_CANCELED:
			return fmt.Errorf("service %s reached terminal failure state %s", serviceID, state)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for service %s to deploy (last state: %s)", serviceID, state)
		case <-ticker.C:
		}
	}
}

func readServiceState(
	ctx context.Context,
	api *qovery.APIClient,
	environmentID string,
	serviceID string,
	kind serviceKind,
) (qovery.StateEnum, error) {
	envStatus, resp, err := api.EnvironmentMainCallsAPI.GetEnvironmentStatuses(ctx, environmentID).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to read environment %s status: %w", environmentID, err)
	}

	switch kind {
	case serviceKindTerraform:
		for _, s := range envStatus.GetTerraforms() {
			if s.GetId() == serviceID {
				return s.GetState(), nil
			}
		}
	case serviceKindHelm:
		for _, s := range envStatus.GetHelms() {
			if s.GetId() == serviceID {
				return s.GetState(), nil
			}
		}
	}

	return "", fmt.Errorf("service %s not found in environment %s status payload", serviceID, environmentID)
}
