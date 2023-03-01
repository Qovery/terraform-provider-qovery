package services

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
)

var _ newdeployment.Service = newdeploymentService{}

type newdeploymentService struct {
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

	return &newdeploymentService{
		newDeploymentEnvironmentRepository: newDeploymentEnvironmentRepository,
		deploymentStatusRepository:         deploymentStatusRepository,
	}, nil
}

func (s newdeploymentService) Create(ctx context.Context, params newdeployment.NewDeploymentParams) (*newdeployment.Deployment, error) {
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
		break
	case newdeployment.STOPPED:
		// Do nothing: no need to stop environment as it has just been created
		break
	}

	// Set new deployment id
	//_, err := s.newDeploymentEnvironmentRepository.GetLastDeploymentId(ctx, *deployment.EnvironmentId)
	//if err != nil {
	//	return nil, errors.Wrap(err, newdeployment.ErrFailedToGetDeployment.Error())
	//}
	return deployment, nil
}

func (s newdeploymentService) Get(ctx context.Context, params newdeployment.NewDeploymentParams) (*newdeployment.Deployment, error) {
	deployment, err := newdeployment.NewDeployment(params)
	if err != nil {
		return nil, errors.Wrap(err, newdeployment.ErrFailedToGetDeployment.Error())
	}

	// Return current deployment if ForceTrigger is disabled
	//if params.ForceTrigger == false {
	//	return deployment, nil
	//}

	// Look for next deployment id to trigger the deployment
	//nextDeploymentId, err := s.newDeploymentEnvironmentRepository.GetNextDeploymentId(ctx, *deployment.EnvironmentId)
	//if err != nil {
	//	return nil, err
	//}
	//
	//deployment.Id = *nextDeploymentId
	return deployment, nil
}

func (s newdeploymentService) Update(ctx context.Context, params newdeployment.NewDeploymentParams) (*newdeployment.Deployment, error) {
	deployment, err := newdeployment.NewDeployment(params)
	if err != nil {
		return nil, err
	}

	err = s.deploymentStatusRepository.WaitForTerminatedState(ctx, *deployment.EnvironmentId)
	if err != nil {
		return nil, errors.Wrap(err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
	}

	switch deployment.DesiredState {
	case newdeployment.RUNNING:
		_, err = s.newDeploymentEnvironmentRepository.ReDeploy(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCreateDeployment.Error())
		}
		break
	case newdeployment.STOPPED:
		_, err = s.newDeploymentEnvironmentRepository.Stop(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCreateDeployment.Error())
		}
		break
	case newdeployment.RESTARTED:
		_, err = s.newDeploymentEnvironmentRepository.Restart(ctx, *deployment)
		if err != nil {
			return nil, errors.Wrap(err, newdeployment.ErrFailedToCreateDeployment.Error())
		}
		break
	}

	err = s.deploymentStatusRepository.WaitForExpectedDesiredState(ctx, *deployment)
	if err != nil {
		return nil, errors.Wrap(err, newdeployment.ErrFailedToCheckDeploymentStatus.Error())
	}

	//// Set new deployment id
	//newDeploymentId, err := s.newDeploymentEnvironmentRepository.GetLastDeploymentId(ctx, *deployment.EnvironmentId)
	//if err != nil {
	//	return nil, errors.Wrap(err, newdeployment.ErrFailedToGetNextDeploymentId.Error())
	//}
	//deployment.Id = *newDeploymentId
	return deployment, nil
}

func (s newdeploymentService) Delete(ctx context.Context, params newdeployment.NewDeploymentParams) error {
	deployment, err := newdeployment.NewDeployment(params)
	if err != nil {
		return err
	}

	err = s.deploymentStatusRepository.WaitForTerminatedState(ctx, *deployment.EnvironmentId)
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
