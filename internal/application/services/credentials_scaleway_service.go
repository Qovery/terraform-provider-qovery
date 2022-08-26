package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsScalewayService implements the interface credentials.ScalewayService.
type credentialsScalewayService struct {
	credentialsScalewayRepository credentials.ScalewayRepository
}

// NOTE: This forces the implementation of the interface credentials.ScalewayService by credentialsScalewayService at compile time.
var _ credentials.ScalewayService = credentialsScalewayService{}

// NewCredentialsScalewayService return a new instance of a credentials.ScalewayService that uses the given credentials.ScalewayRepository.
func NewCredentialsScalewayService(credentialsScalewayRepository credentials.ScalewayRepository) (credentials.ScalewayService, error) {
	if credentialsScalewayRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &credentialsScalewayService{
		credentialsScalewayRepository: credentialsScalewayRepository,
	}, nil
}

// Create handles the domain logic to create a scaleway cluster credentials.
func (c credentialsScalewayService) Create(ctx context.Context, organizationID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateScalewayCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateScalewayCredentials.Error())
	}

	creds, err := c.credentialsScalewayRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateScalewayCredentials.Error())
	}

	return creds, nil
}

// Get handles the domain logic to retrieve a scaleway cluster credentials.
func (c credentialsScalewayService) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetScalewayCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetScalewayCredentials.Error())
	}

	creds, err := c.credentialsScalewayRepository.Get(ctx, organizationID, credentialsID)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetScalewayCredentials.Error())
	}

	return creds, nil
}

// Update handles the domain logic to update a scaleway cluster credentials.
func (c credentialsScalewayService) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
	}

	creds, err := c.credentialsScalewayRepository.Update(ctx, organizationID, credentialsID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
	}

	return creds, nil
}

// Delete handles the domain logic to delete a scaleway cluster credentials.
func (c credentialsScalewayService) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteScalewayCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteScalewayCredentials.Error())
	}

	if err := c.credentialsScalewayRepository.Delete(ctx, organizationID, credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteScalewayCredentials.Error())
	}
	return nil

}

// checkOrganizationID validates that the given organizationID is valid.
func (c credentialsScalewayService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return credentials.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}

// checkCredentialsID validates that the given credentialsID is valid.
func (c credentialsScalewayService) checkCredentialsID(credentialsID string) error {
	if credentialsID == "" {
		return credentials.ErrInvalidCredentialsIDParam
	}

	if _, err := uuid.Parse(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidCredentialsIDParam.Error())
	}

	return nil
}
