package services

import (
	"context"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

const (
	defaultWaitTimeout    = 1 * time.Hour
	defaultWaitMaxRetries = 5
)

type waitFunc func(ctx context.Context) (bool, error)

// Ensure deploymentService defined types fully satisfy the deployment.Service interface.
var _ deployment.Service = deploymentService{}

// deploymentService implements the interface deployment.Service.
type deploymentService struct {
	deploymentRepository deployment.Repository
}

// NewDeploymentService return a new instance of a deployment.Service that uses the given deployment.Repository.
func NewDeploymentService(deploymentRepository deployment.Repository) (deployment.Service, error) {
	if deploymentRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &deploymentService{
		deploymentRepository: deploymentRepository,
	}, nil
}

// GetStatus handles the domain logic to get a resource status.
func (c deploymentService) GetStatus(ctx context.Context, resourceID string) (*status.Status, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToGetStatus.Error())
	}

	deploymentStatus, err := c.deploymentRepository.GetStatus(ctx, resourceID)
	if err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToGetStatus.Error())
	}

	return deploymentStatus, nil
}

// UpdateState handles the domain logic to update the state of a resource.
func (c deploymentService) UpdateState(ctx context.Context, resourceID string, desiredState status.State, version string) (*status.Status, error) {
	switch desiredState {
	case status.StateDeployed:
		return c.Deploy(ctx, resourceID, version)
	case status.StateStopped:
		return c.Stop(ctx, resourceID)
	}

	return nil, deployment.ErrFailedToUpdateState
}

// Deploy handles the domain logic to deploy a resource.
func (c deploymentService) Deploy(ctx context.Context, resourceID string, version string) (*status.Status, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToDeploy.Error())
	}

	currentStatus, err := c.GetStatus(ctx, resourceID)
	if err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToDeploy.Error())
	}

	switch currentStatus.State {
	case status.StateDeployed:
		return currentStatus, nil
	case status.StateDeploymentError:
		return c.Redeploy(ctx, resourceID)
	default:
		_, err := c.deploymentRepository.Deploy(ctx, resourceID, version)
		if err != nil {
			return nil, errors.Wrap(err, deployment.ErrFailedToDeploy.Error())
		}
	}

	if err := c.wait(ctx, c.waitDesiredStateFunc(resourceID, status.StateDeployed)); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToDeploy.Error())
	}

	return c.GetStatus(ctx, resourceID)
}

// Redeploy handles the domain logic to redeploy a resource.
func (c deploymentService) Redeploy(ctx context.Context, resourceID string) (*status.Status, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToRedeploy.Error())
	}

	if err := c.wait(ctx, c.waitFinalStateFunc(resourceID)); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToRedeploy.Error())
	}

	currentStatus, err := c.GetStatus(ctx, resourceID)
	if err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToRedeploy.Error())
	}

	switch currentStatus.State {
	case status.StateReady:
		return currentStatus, nil
	default:
		_, err := c.deploymentRepository.Redeploy(ctx, resourceID)
		if err != nil {
			return nil, errors.Wrap(err, deployment.ErrFailedToRedeploy.Error())
		}
	}

	if err := c.wait(ctx, c.waitDesiredStateFunc(resourceID, status.StateDeployed)); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToRedeploy.Error())
	}

	return c.GetStatus(ctx, resourceID)
}

// Stop handles the domain logic to stop a resource.
func (c deploymentService) Stop(ctx context.Context, resourceID string) (*status.Status, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToStop.Error())
	}

	currentStatus, err := c.GetStatus(ctx, resourceID)
	if err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToStop.Error())
	}

	switch currentStatus.State {
	case status.StateReady, status.StateStopped:
		return currentStatus, nil
	default:
		_, err := c.deploymentRepository.Stop(ctx, resourceID)
		if err != nil {
			return nil, errors.Wrap(err, deployment.ErrFailedToStop.Error())
		}
	}

	if err := c.wait(ctx, c.waitDesiredStateFunc(resourceID, status.StateStopped)); err != nil {
		return nil, errors.Wrap(err, deployment.ErrFailedToStop.Error())
	}

	return c.GetStatus(ctx, resourceID)
}

// checkResourceID validates that the given resourceID is valid.
func (c deploymentService) checkResourceID(resourceID string) error {
	if resourceID == "" {
		return deployment.ErrInvalidResourceIDParam
	}

	if _, err := uuid.Parse(resourceID); err != nil {
		return errors.Wrap(err, deployment.ErrInvalidResourceIDParam.Error())
	}

	return nil
}

func (c deploymentService) wait(ctx context.Context, f waitFunc) error {
	return wait(ctx, f)
}

func (c deploymentService) waitDesiredStateFunc(resourceID string, desiredState status.State) waitFunc {
	return func(ctx context.Context) (bool, error) {
		for tryCount := 0; tryCount < defaultWaitMaxRetries; tryCount++ {
			currentStatus, err := c.deploymentRepository.GetStatus(ctx, resourceID)
			if err != nil {
				return false, err
			}

			isExpectedState := currentStatus.State == desiredState
			if !isExpectedState && currentStatus.IsFinalState() {
				time.Sleep(5 * time.Second)
				continue
			}

			return isExpectedState, nil
		}
		return false, deployment.ErrUnexpectedState
	}
}

func (c deploymentService) waitFinalStateFunc(resourceID string) waitFunc {
	return waitFinalStateFunc(c.deploymentRepository, resourceID)
}

func waitFinalStateFunc(deploymentRepository deployment.Repository, resourceID string) waitFunc {
	return func(ctx context.Context) (bool, error) {
		currentStatus, err := deploymentRepository.GetStatus(ctx, resourceID)
		if err != nil {
			return false, err
		}

		return currentStatus.IsFinalState(), nil
	}
}

func waitNotFoundFunc(deploymentRepository deployment.Repository, resourceID string) waitFunc {
	return func(ctx context.Context) (bool, error) {
		_, err := deploymentRepository.GetStatus(ctx, resourceID)
		if err != nil {
			if apierrors.IsErrNotFound(errors.Cause(err)) {
				return true, nil
			}
			return false, err
		}

		return false, nil
	}
}

func wait(ctx context.Context, f waitFunc) error {
	timeout := pointer.ToDuration(defaultWaitTimeout)

	// Run the function once before waiting
	ok, err := f(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	ticker := time.NewTicker(10 * time.Second)
	timeoutTicker := time.NewTicker(*timeout)

	for {
		select {
		case <-timeoutTicker.C:
			return nil
		case <-ticker.C:
			ok, err := f(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}
