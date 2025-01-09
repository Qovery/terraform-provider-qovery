package services

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
)

var _ newdeployment.Service = newDeploymentService{}

type newDeploymentService struct {
	newDeploymentEnvironmentRepository newdeployment.EnvironmentRepository
	deploymentStatusRepository         newdeployment.DeploymentStatusRepository
}

func NewNewDeploymentService(newDeploymentEnvironmentRepository newdeployment.EnvironmentRepository, deploymentStatusRepository newdeployment.DeploymentStatusRepository) (newdeployment.Service, error) {
	if newDeploymentEnvironmentRepository == nil {
		return nil, ErrInvalidRepository
	}

	if deploymentStatusRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &newDeploymentService{
		newDeploymentEnvironmentRepository: newDeploymentEnvironmentRepository,
		deploymentStatusRepository:         deploymentStatusRepository,
	}, nil
}

func (s newDeploymentService) Create(ctx context.Context, params newdeployment.NewDeploymentParams) (*newdeployment.Deployment, error) {
	deployment, err := newdeployment.NewDeployment(params)
	if err != nil {
		return nil, err
	}

	if deployment.DesiredState == newdeployment.DELETED || deployment.DesiredState == newdeployment.RESTARTED {
		return nil, newdeployment.ErrDesiredStateForbiddenAtCreation
	}

	switch deployment.DesiredState {
	case newdeployment.RUNNING:
		_, err = s.newDeploymentEnvironmentRepository.Deploy(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCreateDeployment.Error())
		}
		err = s.deploymentStatusRepository.WaitForExpectedDesiredState(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
		}
	case newdeployment.STOPPED:
		// Do nothing: no need to stop environment as it has just been created
		break
	}

	return deployment, nil
}

func (s newDeploymentService) Get(ctx context.Context, params newdeployment.NewDeploymentParams) (*newdeployment.Deployment, error) {
	deployment, err := newdeployment.NewDeployment(params)
	if err != nil {
		return nil, errors.Wrap(err, newdeployment.ErrFailedToGetDeployment.Error())
	}

	return deployment, nil
}

func (s newDeploymentService) Update(ctx context.Context, params newdeployment.NewDeploymentParams) (*newdeployment.Deployment, error) {
	deployment, err := newdeployment.NewDeployment(params)
	if err != nil {
		return nil, err
	}

	err = s.deploymentStatusRepository.WaitForTerminatedState(ctx, *deployment.EnvironmentID)
	if err != nil {
		return nil, errors.Wrap(err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
	}

	switch deployment.DesiredState {
	case newdeployment.RUNNING:
		_, err = s.newDeploymentEnvironmentRepository.ReDeploy(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCreateDeployment.Error())
		}
	case newdeployment.STOPPED:
		_, err = s.newDeploymentEnvironmentRepository.Stop(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCreateDeployment.Error())
		}
	case newdeployment.RESTARTED:
		_, err = s.newDeploymentEnvironmentRepository.Restart(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCreateDeployment.Error())
		}
	}

	err = s.deploymentStatusRepository.WaitForExpectedDesiredState(ctx, *deployment)
	if err != nil {
		return nil, errors.Wrap(err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
	}

	return deployment, nil
}

func (s newDeploymentService) Delete(ctx context.Context, params newdeployment.NewDeploymentParams) error {
	deployment, err := newdeployment.NewDeployment(params)
	if err != nil {
		return err
	}

	err, statusCode := s.deploymentStatusRepository.CheckEnvironmentExists(ctx, *deployment.EnvironmentID)
	if err != nil || statusCode >= 400 {
		if statusCode == 404 {
			message := fmt.Sprintf("The target environment %s doesn't exist anymore so the DELETE deployment cannot be triggered. You can ignore this warning if you have deleted previously your environment", *deployment.EnvironmentID)
			tflog.Warn(ctx, message, map[string]interface{}{"environment_id": *deployment.EnvironmentID})
			return nil
		}
		return errors.Wrap(err, newdeployment.ErrFailedToDeleteDeployment.Error())
	}

	err = s.deploymentStatusRepository.WaitForTerminatedState(ctx, *deployment.EnvironmentID)
	if err != nil {
		return errors.Wrap(err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
	}

	_, err = s.newDeploymentEnvironmentRepository.Delete(ctx, *deployment)
	if err != nil {
		return errors.Wrap(err, newdeployment.ErrFailedToDeleteDeployment.Error())
	}

	err = s.deploymentStatusRepository.WaitForExpectedDesiredState(ctx, *deployment)
	if err != nil {
		return errors.Wrap(err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
	}

	return nil
}
