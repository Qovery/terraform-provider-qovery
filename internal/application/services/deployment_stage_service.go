package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

var _ deploymentstage.Service = deploymentStageService{}

type deploymentStageService struct {
	deploymentStageRepository deploymentstage.Repository
}

func NewDeploymentStageService(deploymentStageRepository deploymentstage.Repository) (deploymentstage.Service, error) {
	if deploymentStageRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &deploymentStageService{
		deploymentStageRepository: deploymentStageRepository,
	}, nil
}

func (s deploymentStageService) Create(ctx context.Context, environmentID string, request deploymentstage.UpsertServiceRequest) (*deploymentstage.DeploymentStage, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToCreateDeploymentStage.Error())
	}

	deploymentStageCreated, err := s.deploymentStageRepository.Create(ctx, environmentID, request.DeploymentStageUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToCreateDeploymentStage.Error())
	}

	return deploymentStageCreated, nil
}

func (s deploymentStageService) Get(ctx context.Context, environmentID string, deploymentStageID string) (*deploymentstage.DeploymentStage, error) {
	if err := s.checkDeploymentStageID(deploymentStageID); err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToGetDeploymentStage.Error())
	}

	deploymentStage, err := s.deploymentStageRepository.Get(ctx, environmentID, deploymentStageID)
	if err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToGetDeploymentStage.Error())
	}

	return deploymentStage, nil
}

func (s deploymentStageService) GetAllByEnvironmentID(ctx context.Context, environmentID string, deploymentStageName string) (*deploymentstage.DeploymentStage, error) {
	deploymentStages, err := s.deploymentStageRepository.GetAllByEnvironmentID(ctx, environmentID)
	if err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToGetDeploymentStage.Error())
	}

	for _, deploymentStage := range *deploymentStages {
		if deploymentStage.Name == deploymentStageName {
			return &deploymentStage, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find deployment stage with name %s in environment with id %s", deploymentStageName, environmentID))
}

func (s deploymentStageService) Update(ctx context.Context, deploymentStageID string, request deploymentstage.UpsertServiceRequest) (*deploymentstage.DeploymentStage, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToUpdateDeploymentStage.Error())
	}
	if err := s.checkDeploymentStageID(deploymentStageID); err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToGetDeploymentStage.Error())
	}

	deploymentStageUpdated, err := s.deploymentStageRepository.Update(ctx, deploymentStageID, request.DeploymentStageUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, deploymentstage.ErrFailedToCreateDeploymentStage.Error())
	}

	return deploymentStageUpdated, nil
}

func (s deploymentStageService) Delete(ctx context.Context, deploymentStageID string) error {
	if err := s.checkDeploymentStageID(deploymentStageID); err != nil {
		return errors.Wrap(err, deploymentstage.ErrFailedToDeleteDeploymentStage.Error())
	}

	err := s.deploymentStageRepository.Delete(ctx, deploymentStageID)
	if err != nil {
		return errors.Wrap(err, deploymentstage.ErrFailedToDeleteDeploymentStage.Error())
	}

	return nil
}

func (s deploymentStageService) checkDeploymentStageID(deploymentStageID string) error {
	if deploymentStageID == "" {
		return deploymentstage.ErrInvalidDeploymentStageIDParam
	}

	if _, err := uuid.Parse(deploymentStageID); err != nil {
		return errors.Wrap(err, deploymentstage.ErrInvalidDeploymentStageIDParam.Error())
	}

	return nil
}
