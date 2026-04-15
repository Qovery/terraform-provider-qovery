package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure credentialsEksAnywhereVsphereService defined types fully satisfy the credentials.EksAnywhereVsphereService interface.
var _ credentials.EksAnywhereVsphereService = credentialsEksAnywhereVsphereService{}

// credentialsEksAnywhereVsphereService implements the interface credentials.EksAnywhereVsphereService.
type credentialsEksAnywhereVsphereService struct {
	credentialsEksAnywhereVsphereRepository credentials.EksAnywhereVsphereRepository
}

// NewCredentialsEksAnywhereVsphereService return a new instance of a credentials.EksAnywhereVsphereService that uses the given credentials.EksAnywhereVsphereRepository.
func NewCredentialsEksAnywhereVsphereService(credentialsEksAnywhereVsphereRepository credentials.EksAnywhereVsphereRepository) (credentials.EksAnywhereVsphereService, error) {
	if credentialsEksAnywhereVsphereRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &credentialsEksAnywhereVsphereService{
		credentialsEksAnywhereVsphereRepository: credentialsEksAnywhereVsphereRepository,
	}, nil
}

// Create handles the domain logic to create an eks anywhere vsphere cluster credentials.
func (c credentialsEksAnywhereVsphereService) Create(ctx context.Context, organizationID string, request credentials.UpsertEksAnywhereVsphereRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateEksAnywhereVsphereCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateEksAnywhereVsphereCredentials.Error())
	}

	creds, err := c.credentialsEksAnywhereVsphereRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToCreateEksAnywhereVsphereCredentials.Error())
	}

	return creds, nil
}

// Get handles the domain logic to retrieve an eks anywhere vsphere cluster credentials.
func (c credentialsEksAnywhereVsphereService) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetEksAnywhereVsphereCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetEksAnywhereVsphereCredentials.Error())
	}

	creds, err := c.credentialsEksAnywhereVsphereRepository.Get(ctx, organizationID, credentialsID)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToGetEksAnywhereVsphereCredentials.Error())
	}

	return creds, nil
}

// Update handles the domain logic to update an eks anywhere vsphere cluster credentials.
func (c credentialsEksAnywhereVsphereService) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertEksAnywhereVsphereRequest) (*credentials.Credentials, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateEksAnywhereVsphereCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateEksAnywhereVsphereCredentials.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateEksAnywhereVsphereCredentials.Error())
	}

	creds, err := c.credentialsEksAnywhereVsphereRepository.Update(ctx, organizationID, credentialsID, request)
	if err != nil {
		return nil, errors.Wrap(err, credentials.ErrFailedToUpdateEksAnywhereVsphereCredentials.Error())
	}

	return creds, nil
}

// Delete handles the domain logic to delete an eks anywhere vsphere cluster credentials.
func (c credentialsEksAnywhereVsphereService) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteEksAnywhereVsphereCredentials.Error())
	}

	if err := c.checkCredentialsID(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteEksAnywhereVsphereCredentials.Error())
	}

	if err := c.credentialsEksAnywhereVsphereRepository.Delete(ctx, organizationID, credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrFailedToDeleteEksAnywhereVsphereCredentials.Error())
	}

	return nil
}

func (c credentialsEksAnywhereVsphereService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return credentials.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}

func (c credentialsEksAnywhereVsphereService) checkCredentialsID(credentialsID string) error {
	if credentialsID == "" {
		return credentials.ErrInvalidCredentialsIDParam
	}

	if _, err := uuid.Parse(credentialsID); err != nil {
		return errors.Wrap(err, credentials.ErrInvalidCredentialsIDParam.Error())
	}

	return nil
}
