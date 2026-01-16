package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure credentialsAzureService defined types fully satisfy the credentials.AzureService interface.
var _ credentials.AzureService = credentialsAzureService{}

// credentialsAzureService implements the interface credentials.AzureService.
type credentialsAzureService struct {
	credentialsAzureRepository credentials.AzureRepository
}

// NewCredentialsAzureService return a new instance of a credentials.AzureService that uses the given credentials.AzureRepository.
func NewCredentialsAzureService(credentialsAzureRepository credentials.AzureRepository) (credentials.AzureService, error) {
	if credentialsAzureRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &credentialsAzureService{
		credentialsAzureRepository: credentialsAzureRepository,
	}, nil
}

// Create handles the domain logic to create an azure cluster credentials.
func (c credentialsAzureService) Create(ctx context.Context, organizationID string, request credentials.UpsertAzureRequest) (*credentials.AzureCredentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateAzureCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateAzureCredentials.Error())
	}

	creds, err := c.credentialsAzureRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateAzureCredentials.Error())
	}

	return creds, nil
}

// Get handles the domain logic to retrieve an azure cluster credentials.
func (c credentialsAzureService) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.AzureCredentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetAzureCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetAzureCredentials.Error())
	}

	creds, err := c.credentialsAzureRepository.Get(ctx, organizationID, credentialsID)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetAzureCredentials.Error())
	}

	return creds, nil
}

// Update handles the domain logic to update an azure cluster credentials.
func (c credentialsAzureService) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAzureRequest) (*credentials.AzureCredentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAzureCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAzureCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAzureCredentials.Error())
	}

	creds, err := c.credentialsAzureRepository.Update(ctx, organizationID, credentialsID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAzureCredentials.Error())
	}

	return creds, nil
}

// Delete handles the domain logic to delete an azure cluster credentials.
func (c credentialsAzureService) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteAzureCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteAzureCredentials.Error())
	}

	if err := c.credentialsAzureRepository.Delete(ctx, organizationID, credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteAzureCredentials.Error())
	}

	return nil
}

// checkOrganizationID validates that the given organizationID is valid.
func (c credentialsAzureService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return credentials.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}

// checkCredentialsID validates that the given credentialsID is valid.
func (c credentialsAzureService) checkCredentialsID(credentialsID string) error {
	if credentialsID == "" {
		return credentials.ErrInvalidCredentialsIDParam
	}

	if _, err := uuid.Parse(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidCredentialsIDParam.Error())
	}

	return nil
}
