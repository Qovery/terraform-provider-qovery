//go:build unit && !integration

package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
)

// stubNewDeploymentEnvironmentRepository / stubDeploymentStatusRepository replace the
// newdeployment repositories: there is no generated mock for them, and these tests only need
// the environment calls to succeed and the status waits to fail on demand.
type stubNewDeploymentEnvironmentRepository struct{}

func (stubNewDeploymentEnvironmentRepository) Deploy(_ context.Context, d newdeployment.Deployment) (*newdeployment.Deployment, error) {
	return &d, nil
}

func (stubNewDeploymentEnvironmentRepository) ReDeploy(_ context.Context, d newdeployment.Deployment) (*newdeployment.Deployment, error) {
	return &d, nil
}

func (stubNewDeploymentEnvironmentRepository) Stop(_ context.Context, d newdeployment.Deployment) (*newdeployment.Deployment, error) {
	return &d, nil
}

func (stubNewDeploymentEnvironmentRepository) Restart(_ context.Context, d newdeployment.Deployment) (*newdeployment.Deployment, error) {
	return &d, nil
}

func (stubNewDeploymentEnvironmentRepository) Delete(_ context.Context, d newdeployment.Deployment) (*newdeployment.Deployment, error) {
	return &d, nil
}

type stubDeploymentStatusRepository struct {
	waitForTerminatedStateErr      error
	waitForExpectedDesiredStateErr error
}

func (s stubDeploymentStatusRepository) WaitForTerminatedState(_ context.Context, _ uuid.UUID) error {
	return s.waitForTerminatedStateErr
}

func (s stubDeploymentStatusRepository) WaitForExpectedDesiredState(_ context.Context, _ newdeployment.Deployment) error {
	return s.waitForExpectedDesiredStateErr
}

func (stubDeploymentStatusRepository) CheckEnvironmentExists(_ context.Context, _ uuid.UUID) (error, int) {
	return nil, 200
}

// TestNewDeploymentService_WaitTimeoutFailsTheOperation checks that a timed-out status wait is
// surfaced by every lifecycle operation instead of being swallowed as a converged deployment
// (QOV-2299). The Terraform resource turns the returned error into an apply diagnostic.
func TestNewDeploymentService_WaitTimeoutFailsTheOperation(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	timeoutErr := fmt.Errorf("%w: waited 4h0m0s for environment %s to reach state RUNNING", newdeployment.ErrWaitTimeout, environmentID)

	testCases := []struct {
		TestName   string
		StatusRepo stubDeploymentStatusRepository
		Run        func(*testing.T, newdeployment.Service, newdeployment.NewDeploymentParams) error
	}{
		{
			TestName:   "create_fails_when_desired_state_wait_times_out",
			StatusRepo: stubDeploymentStatusRepository{waitForExpectedDesiredStateErr: timeoutErr},
			Run: func(t *testing.T, svc newdeployment.Service, params newdeployment.NewDeploymentParams) error {
				deployment, err := svc.Create(t.Context(), params)
				assert.Nil(t, deployment, "a timed-out create must not hand back a deployment to record in state")
				return err
			},
		},
		{
			TestName:   "update_fails_when_terminal_state_wait_times_out",
			StatusRepo: stubDeploymentStatusRepository{waitForTerminatedStateErr: timeoutErr},
			Run: func(t *testing.T, svc newdeployment.Service, params newdeployment.NewDeploymentParams) error {
				deployment, err := svc.Update(t.Context(), params)
				assert.Nil(t, deployment)
				return err
			},
		},
		{
			TestName:   "update_fails_when_desired_state_wait_times_out",
			StatusRepo: stubDeploymentStatusRepository{waitForExpectedDesiredStateErr: timeoutErr},
			Run: func(t *testing.T, svc newdeployment.Service, params newdeployment.NewDeploymentParams) error {
				deployment, err := svc.Update(t.Context(), params)
				assert.Nil(t, deployment)
				return err
			},
		},
		{
			TestName:   "delete_fails_when_desired_state_wait_times_out",
			StatusRepo: stubDeploymentStatusRepository{waitForExpectedDesiredStateErr: timeoutErr},
			Run: func(t *testing.T, svc newdeployment.Service, params newdeployment.NewDeploymentParams) error {
				return svc.Delete(t.Context(), params)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			svc, err := services.NewNewDeploymentService(stubNewDeploymentEnvironmentRepository{}, tc.StatusRepo)
			require.NoError(t, err)

			err = tc.Run(t, svc, newdeployment.NewDeploymentParams{
				EnvironmentID: environmentID.String(),
				DesiredState:  string(newdeployment.RUNNING),
			})

			require.Error(t, err)
			assert.ErrorIs(t, err, newdeployment.ErrWaitTimeout, "timeout identity must survive service wrapping")
			assert.ErrorContains(t, err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
			assert.ErrorContains(t, err, environmentID.String(), "the diagnostic must name the environment that never converged")
			assert.ErrorContains(t, err, "4h0m0s", "the diagnostic must state how long the provider waited")
		})
	}
}
