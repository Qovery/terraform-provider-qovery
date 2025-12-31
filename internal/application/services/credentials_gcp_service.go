package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure credentialsGcpService defined types fully satisfy the credentials.GcpService interface.
var _ credentials.GcpService = credentialsGcpService{}

// credentialsGcpService implements the interface credentials.GcpService.
type credentialsGcpService struct {
	credentialsGcpRepository credentials.GcpRepository
}

// NewCredentialsGcpService return a new instance of a credentials.GcpService that uses the given credentials.GcpRepository.
func NewCredentialsGcpService(credentialsGcpRepository credentials.GcpRepository) (credentials.GcpService, error) {
	if credentialsGcpRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &credentialsGcpService{
		credentialsGcpRepository: credentialsGcpRepository,
	}, nil
}

// Create handles the domain logic to create a gcp cluster credentials.
func (c credentialsGcpService) Create(ctx context.Context, organizationID string, request credentials.UpsertGcpRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateGcpCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateGcpCredentials.Error())
	}

	creds, err := c.credentialsGcpRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateGcpCredentials.Error())
	}

	return creds, nil
}

// Get handles the domain logic to retrieve a gcp cluster credentials.
func (c credentialsGcpService) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetGcpCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetGcpCredentials.Error())
	}

	creds, err := c.credentialsGcpRepository.Get(ctx, organizationID, credentialsID)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetGcpCredentials.Error())
	}

	return creds, nil
}

// Update handles the domain logic to update a gcp cluster credentials.
func (c credentialsGcpService) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertGcpRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateGcpCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateGcpCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateGcpCredentials.Error())
	}

	creds, err := c.credentialsGcpRepository.Update(ctx, organizationID, credentialsID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateGcpCredentials.Error())
	}

	return creds, nil
}

// Delete handles the domain logic to delete a gcp cluster credentials.
func (c credentialsGcpService) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteGcpCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteGcpCredentials.Error())
	}

	if err := c.credentialsGcpRepository.Delete(ctx, organizationID, credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteGcpCredentials.Error())
	}

	return nil
}

// checkOrganizationID validates that the given organizationID is valid.
func (c credentialsGcpService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return credentials.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}

// checkCredentialsID validates that the given credentialsID is valid.
func (c credentialsGcpService) checkCredentialsID(credentialsID string) error {
	if credentialsID == "" {
		return credentials.ErrInvalidCredentialsIDParam
	}

	if _, err := uuid.Parse(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidCredentialsIDParam.Error())
	}

	return nil
}
