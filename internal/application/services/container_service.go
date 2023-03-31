package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure containerService defined types fully satisfy the container.Service interface.
var _ container.Service = containerService{}

// containerService implements the interface container.Service.
type containerService struct {
	containerRepository        container.Repository
	containerDeploymentService deployment.Service
	variableService            variable.Service
	secretService              secret.Service
}

// NewContainerService return a new instance of a container.Service that uses the given container.Repository.
func NewContainerService(containerRepository container.Repository, containerDeploymentService deployment.Service, variableService variable.Service, secretService secret.Service) (container.Service, error) {
	if containerRepository == nil {
		return nil, ErrInvalidRepository
	}

	if containerDeploymentService == nil {
		return nil, ErrInvalidService
	}

	if variableService == nil {
		return nil, ErrInvalidService
	}

	if secretService == nil {
		return nil, ErrInvalidService
	}

	return &containerService{
		containerRepository:        containerRepository,
		variableService:            variableService,
		secretService:              secretService,
		containerDeploymentService: containerDeploymentService,
	}, nil
}

// Create handles the domain logic to create a container.
func (s containerService) Create(ctx context.Context, environmentID string, request container.UpsertServiceRequest) (*container.Container, error) {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToCreateContainer.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToCreateContainer.Error())
	}

	cont, err := s.containerRepository.Create(ctx, environmentID, request.ContainerUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToCreateContainer.Error())
	}

	_, err = s.variableService.Update(ctx, cont.ID.String(), request.EnvironmentVariables)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToCreateContainer.Error())
	}

	_, err = s.secretService.Update(ctx, cont.ID.String(), request.Secrets)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToCreateContainer.Error())
	}

	cont, err = s.refreshContainer(ctx, *cont)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToCreateContainer.Error())
	}

	return cont, nil
}

// Get handles the domain logic to retrieve a container.
func (s containerService) Get(ctx context.Context, containerID string) (*container.Container, error) {
	if err := s.checkID(containerID); err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToGetContainer.Error())
	}

	cont, err := s.containerRepository.Get(ctx, containerID)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToGetContainer.Error())
	}

	cont, err = s.refreshContainer(ctx, *cont)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToGetContainer.Error())
	}

	return cont, nil
}

// Update handles the domain logic to update a container.
func (s containerService) Update(ctx context.Context, containerID string, request container.UpsertServiceRequest) (*container.Container, error) {
	if err := s.checkID(containerID); err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToUpdateContainer.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToUpdateContainer.Error())
	}

	cont, err := s.containerRepository.Update(ctx, containerID, request.ContainerUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToUpdateContainer.Error())
	}

	_, err = s.variableService.Update(ctx, cont.ID.String(), request.EnvironmentVariables)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToUpdateContainer.Error())
	}

	_, err = s.secretService.Update(ctx, cont.ID.String(), request.Secrets)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToUpdateContainer.Error())
	}

	cont, err = s.refreshContainer(ctx, *cont)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrFailedToUpdateContainer.Error())
	}

	return cont, nil
}

// Delete handles the domain logic to delete a container.
func (s containerService) Delete(ctx context.Context, containerID string) error {
	if err := s.checkID(containerID); err != nil {
		return errors.Wrap(err, container.ErrFailedToDeleteContainer.Error())
	}

	if err := s.containerRepository.Delete(ctx, containerID); err != nil {
		return errors.Wrap(err, container.ErrFailedToDeleteContainer.Error())
	}

	if err := wait(ctx, waitNotFoundFunc(s.containerDeploymentService, containerID), nil); err != nil {
		return errors.Wrap(err, container.ErrFailedToDeleteContainer.Error())
	}

	return nil
}

func (s containerService) refreshContainer(ctx context.Context, cont container.Container) (*container.Container, error) {
	envVars, err := s.variableService.List(ctx, cont.ID.String())
	if err != nil {
		return nil, err
	}

	secrets, err := s.secretService.List(ctx, cont.ID.String())
	if err != nil {
		return nil, err
	}

	status, err := s.containerDeploymentService.GetStatus(ctx, cont.ID.String())
	if err != nil {
		return nil, err
	}

	if err := cont.SetEnvironmentVariables(envVars); err != nil {
		return nil, err
	}

	if err := cont.SetSecrets(secrets); err != nil {
		return nil, err
	}

	if err := cont.SetState(status.State); err != nil {
		return nil, err
	}

	return &cont, err
}

// checkEnvironmentID validates that the given environmentID is valid.
func (s containerService) checkEnvironmentID(environmentID string) error {
	if environmentID == "" {
		return container.ErrInvalidEnvironmentIDParam
	}

	if _, err := uuid.Parse(environmentID); err != nil {
		return errors.Wrap(err, container.ErrInvalidEnvironmentIDParam.Error())
	}

	return nil
}

// checkID validates that the given containerID is valid.
func (s containerService) checkID(containerID string) error {
	if containerID == "" {
		return container.ErrInvalidContainerIDParam
	}

	if _, err := uuid.Parse(containerID); err != nil {
		return errors.Wrap(err, container.ErrInvalidContainerIDParam.Error())
	}

	return nil
}
