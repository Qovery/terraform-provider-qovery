//go:build unit && !integration
// +build unit,!integration

package services

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewDeploymentService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  deployment.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.DeploymentRepository{},
			ExpectError: false,
		},
		{
			TestName:    "error_with_nil_repository",
			Repository:  nil,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewDeploymentService(tc.Repository)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				assert.Equal(t, ErrInvalidRepository, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestDeploymentService_GetStatus(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""

	expectedStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateDeployed,
	}

	testCases := []struct {
		TestName      string
		ResourceID    string
		SetupMock     func(*mocks_test.DeploymentRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_resource_id",
			ResourceID:    emptyResourceID,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:      "error_invalid_resource_id",
			ResourceID:    invalidResourceID,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:   "error_repository_failure",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToGetStatus.Error(),
		},
		{
			TestName:   "success",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(expectedStatus, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentService(mockRepo)
			require.NoError(t, err)

			result, err := service.GetStatus(context.Background(), tc.ResourceID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedStatus.ID, result.ID)
				assert.Equal(t, expectedStatus.State, result.State)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentService_UpdateState(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	version := "1.0.0"

	deployedStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateDeployed,
	}

	stoppedStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateStopped,
	}

	testCases := []struct {
		TestName      string
		ResourceID    string
		DesiredState  status.State
		Version       string
		SetupMock     func(*mocks_test.DeploymentRepository)
		ExpectError   bool
		ErrorContains string
		ExpectedState status.State
	}{
		{
			TestName:     "success_deploy_state",
			ResourceID:   validResourceID,
			DesiredState: status.StateDeployed,
			Version:      version,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(deployedStatus, nil)
			},
			ExpectError:   false,
			ExpectedState: status.StateDeployed,
		},
		{
			TestName:     "success_stop_state",
			ResourceID:   validResourceID,
			DesiredState: status.StateStopped,
			Version:      version,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(stoppedStatus, nil)
			},
			ExpectError:   false,
			ExpectedState: status.StateStopped,
		},
		{
			TestName:      "error_invalid_state",
			ResourceID:    validResourceID,
			DesiredState:  status.StateDeleting,
			Version:       version,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToUpdateState.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentService(mockRepo)
			require.NoError(t, err)

			result, err := service.UpdateState(context.Background(), tc.ResourceID, tc.DesiredState, tc.Version)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.ExpectedState, result.State)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentService_Deploy(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""
	version := "1.0.0"

	deployedStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateDeployed,
	}

	deployingStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateDeploying,
	}

	testCases := []struct {
		TestName      string
		ResourceID    string
		Version       string
		SetupMock     func(*mocks_test.DeploymentRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_resource_id",
			ResourceID:    emptyResourceID,
			Version:       version,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:      "error_invalid_resource_id",
			ResourceID:    invalidResourceID,
			Version:       version,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:   "error_get_status_failure",
			ResourceID: validResourceID,
			Version:    version,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(nil, errors.New("get status error"))
			},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToDeploy.Error(),
		},
		{
			TestName:   "success_already_deployed",
			ResourceID: validResourceID,
			Version:    version,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(deployedStatus, nil)
			},
			ExpectError: false,
		},
		{
			TestName:   "error_deploy_call_failure",
			ResourceID: validResourceID,
			Version:    version,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(deployingStatus, nil)
				m.EXPECT().
					Deploy(mock.Anything, validResourceID, version).
					Return(nil, errors.New("deploy error"))
			},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToDeploy.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentService(mockRepo)
			require.NoError(t, err)

			// Use a context with cancel to prevent waiting in tests
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			result, err := service.Deploy(ctx, tc.ResourceID, tc.Version)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentService_Redeploy(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""

	readyStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateReady,
	}

	deployedStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateDeployed,
	}

	testCases := []struct {
		TestName      string
		ResourceID    string
		SetupMock     func(*mocks_test.DeploymentRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_resource_id",
			ResourceID:    emptyResourceID,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:      "error_invalid_resource_id",
			ResourceID:    invalidResourceID,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:   "error_get_status_failure_during_wait",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				// First call to wait for final state
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(nil, errors.New("get status error"))
			},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToRedeploy.Error(),
		},
		{
			TestName:   "success_already_ready",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				// First GetStatus call during wait for final state
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(readyStatus, nil).
					Once()
				// Second GetStatus call after wait completes
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(readyStatus, nil).
					Once()
			},
			ExpectError: false,
		},
		{
			TestName:   "error_redeploy_call_failure",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				// First GetStatus call during wait for final state
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(deployedStatus, nil).
					Once()
				// Second GetStatus call after wait completes
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(deployedStatus, nil).
					Once()
				// Redeploy call that fails
				m.EXPECT().
					Redeploy(mock.Anything, validResourceID).
					Return(nil, errors.New("redeploy error"))
			},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToRedeploy.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentService(mockRepo)
			require.NoError(t, err)

			// Use a context with cancel to prevent waiting in tests
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			result, err := service.Redeploy(ctx, tc.ResourceID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentService_Stop(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""

	stoppedStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateStopped,
	}

	readyStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateReady,
	}

	deployedStatus := &status.Status{
		ID:    uuid.MustParse(validResourceID),
		State: status.StateDeployed,
	}

	testCases := []struct {
		TestName      string
		ResourceID    string
		SetupMock     func(*mocks_test.DeploymentRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_resource_id",
			ResourceID:    emptyResourceID,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:      "error_invalid_resource_id",
			ResourceID:    invalidResourceID,
			SetupMock:     func(m *mocks_test.DeploymentRepository) {},
			ExpectError:   true,
			ErrorContains: deployment.ErrInvalidResourceIDParam.Error(),
		},
		{
			TestName:   "error_get_status_failure",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(nil, errors.New("get status error"))
			},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToStop.Error(),
		},
		{
			TestName:   "success_already_stopped",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(stoppedStatus, nil)
			},
			ExpectError: false,
		},
		{
			TestName:   "success_already_ready",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(readyStatus, nil)
			},
			ExpectError: false,
		},
		{
			TestName:   "error_stop_call_failure",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.DeploymentRepository) {
				m.EXPECT().
					GetStatus(mock.Anything, validResourceID).
					Return(deployedStatus, nil)
				m.EXPECT().
					Stop(mock.Anything, validResourceID).
					Return(nil, errors.New("stop error"))
			},
			ExpectError:   true,
			ErrorContains: deployment.ErrFailedToStop.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentService(mockRepo)
			require.NoError(t, err)

			// Use a context with cancel to prevent waiting in tests
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			result, err := service.Stop(ctx, tc.ResourceID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
