package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsAwsService implements the interface credentials.AwsService.
type credentialsAwsService struct {
	credentialsAwsRepository credentials.AwsRepository
}

// NOTE: This forces the implementation of the interface credentials.AwsService  by credentialsAwsQoveryAPI at compile time.
var _ credentials.AwsService = credentialsAwsService{}

// NewCredentialsAwsService return a new instance of a credentials.AwsService that uses the given credentials.AwsRepository.
func NewCredentialsAwsService(credentialsAwsRepository credentials.AwsRepository) (credentials.AwsService, error) {
	if credentialsAwsRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &credentialsAwsService{
		credentialsAwsRepository: credentialsAwsRepository,
	}, nil
}

// Create handles the domain logic to create an aws cluster credentials.
func (c credentialsAwsService) Create(ctx context.Context, organizationID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateAwsCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateAwsCredentials.Error())
	}

	creds, err := c.credentialsAwsRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateAwsCredentials.Error())
	}

	return creds, nil
}

// Get handles the domain logic to retrieve an aws cluster credentials.
func (c credentialsAwsService) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetAwsCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetAwsCredentials.Error())
	}

	creds, err := c.credentialsAwsRepository.Get(ctx, organizationID, credentialsID)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetAwsCredentials.Error())
	}

	return creds, nil
}

// Update handles the domain logic to update an aws cluster credentials.
func (c credentialsAwsService) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAwsCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAwsCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAwsCredentials.Error())
	}

	creds, err := c.credentialsAwsRepository.Update(ctx, organizationID, credentialsID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateAwsCredentials.Error())
	}

	return creds, nil
}

// Delete handles the domain logic to delete an aws cluster credentials.
func (c credentialsAwsService) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteAwsCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteAwsCredentials.Error())
	}

	if err := c.credentialsAwsRepository.Delete(ctx, organizationID, credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteAwsCredentials.Error())
	}

	return nil
}

// checkOrganizationID validates that the given organizationID is valid.
func (c credentialsAwsService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return credentials.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}

// checkCredentialsID validates that the given credentialsID is valid.
func (c credentialsAwsService) checkCredentialsID(credentialsID string) error {
	if credentialsID == "" {
		return credentials.ErrInvalidCredentialsIDParam
	}

	if _, err := uuid.Parse(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidCredentialsIDParam.Error())
	}

	return nil
}
