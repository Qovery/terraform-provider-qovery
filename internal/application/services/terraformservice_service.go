package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

// Ensure terraformServiceService defined types fully satisfy the terraformservice.Service interface.
var _ terraformservice.Service = terraformServiceService{}

// terraformServiceService implements the interface terraformservice.Service.
type terraformServiceService struct {
	terraformServiceRepository terraformservice.Repository
}

// NewTerraformServiceService return a new instance of a terraformservice.Service that uses the given terraformservice.Repository.
func NewTerraformServiceService(terraformServiceRepository terraformservice.Repository) (terraformservice.Service, error) {
	if terraformServiceRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &terraformServiceService{
		terraformServiceRepository: terraformServiceRepository,
	}, nil
}

// Create handles the domain logic to create a terraform service.
func (s terraformServiceService) Create(ctx context.Context, environmentID string, request terraformservice.UpsertServiceRequest) (*terraformservice.TerraformService, error) {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToCreateTerraformService.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToCreateTerraformService.Error())
	}

	newTerraformService, err := s.terraformServiceRepository.Create(ctx, environmentID, request.TerraformServiceUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToCreateTerraformService.Error())
	}

	return newTerraformService, nil
}

// Get handles the domain logic to retrieve a terraform service.
func (s terraformServiceService) Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*terraformservice.TerraformService, error) {
	if err := s.checkID(terraformServiceID); err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToGetTerraformService.Error())
	}

	fetchedTerraformService, err := s.terraformServiceRepository.Get(ctx, terraformServiceID, advancedSettingsJsonFromState, isTriggeredFromImport)
	if err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToGetTerraformService.Error())
	}

	return fetchedTerraformService, nil
}

// Update handles the domain logic to update a terraform service.
func (s terraformServiceService) Update(ctx context.Context, terraformServiceID string, request terraformservice.UpsertServiceRequest) (*terraformservice.TerraformService, error) {
	if err := s.checkID(terraformServiceID); err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToUpdateTerraformService.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToUpdateTerraformService.Error())
	}

	updatedTerraformService, err := s.terraformServiceRepository.Update(ctx, terraformServiceID, request.TerraformServiceUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToUpdateTerraformService.Error())
	}

	return updatedTerraformService, nil
}

// Delete handles the domain logic to delete a terraform service.
func (s terraformServiceService) Delete(ctx context.Context, terraformServiceID string) error {
	if err := s.checkID(terraformServiceID); err != nil {
		return errors.Wrap(err, terraformservice.ErrFailedToDeleteTerraformService.Error())
	}

	if err := s.terraformServiceRepository.Delete(ctx, terraformServiceID); err != nil {
		return errors.Wrap(err, terraformservice.ErrFailedToDeleteTerraformService.Error())
	}

	return nil
}

// List handles the domain logic to list terraform services in an environment.
func (s terraformServiceService) List(ctx context.Context, environmentID string) ([]terraformservice.TerraformService, error) {
	if err := s.checkEnvironmentID(environmentID); err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToListTerraformServices.Error())
	}

	terraformServices, err := s.terraformServiceRepository.List(ctx, environmentID)
	if err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrFailedToListTerraformServices.Error())
	}

	return terraformServices, nil
}

// checkEnvironmentID validates that the given environmentID is valid.
func (s terraformServiceService) checkEnvironmentID(environmentID string) error {
	if environmentID == "" {
		return terraformservice.ErrInvalidTerraformServiceEnvironmentIDParam
	}

	if _, err := uuid.Parse(environmentID); err != nil {
		return errors.Wrap(err, terraformservice.ErrInvalidTerraformServiceEnvironmentIDParam.Error())
	}

	return nil
}

// checkID validates that the given terraformServiceID is valid.
func (s terraformServiceService) checkID(terraformServiceID string) error {
	if terraformServiceID == "" {
		return terraformservice.ErrInvalidTerraformServiceIDParam
	}

	if _, err := uuid.Parse(terraformServiceID); err != nil {
		return errors.Wrap(err, terraformservice.ErrInvalidTerraformServiceIDParam.Error())
	}

	return nil
}
