package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

// Ensure secretService defined type fully satisfy the secret.Service interface.
var _ secret.Service = secretService{}

// secretService implements the interface secret.Service.
type secretService struct {
	secretRepository secret.Repository
}

// NewSecretService return a new instance of a secret.Service that uses the given secret.Repository.
func NewSecretService(secretRepository secret.Repository) (secret.Service, error) {
	if secretRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &secretService{
		secretRepository: secretRepository,
	}, nil
}

// List handles the domain logic to retrieve a list of secrets.
func (c secretService) List(ctx context.Context, resourceID string) (secret.Secrets, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, secret.ErrFailedToListSecrets.Error())
	}

	vars, err := c.secretRepository.List(ctx, resourceID)
	if err != nil {
		return nil, errors.Wrap(err, secret.ErrFailedToListSecrets.Error())
	}

	return vars, nil
}

// Update handles the domain logic to update a secret.
func (c secretService) Update(ctx context.Context, resourceID string, request secret.DiffRequest) (secret.Secrets, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
	}

	secrets := make(secret.Secrets, 0, len(request.Create)+len(request.Update))
	for _, toDelete := range request.Delete {
		err := c.secretRepository.Delete(ctx, resourceID, toDelete.SecretID)
		if err != nil {
			return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
		}
	}

	for _, toUpdate := range request.Update {
		v, err := c.secretRepository.Update(ctx, resourceID, toUpdate.SecretID, toUpdate.UpsertRequest)
		if err != nil {
			return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
		}

		secrets = append(secrets, *v)
	}

	for _, toCreate := range request.Create {
		v, err := c.secretRepository.Create(ctx, resourceID, toCreate.UpsertRequest)
		if err != nil {
			return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
		}

		secrets = append(secrets, *v)
	}

	return secrets, nil
}

// checkResourceID validates that the given resourceID is valid.
func (c secretService) checkResourceID(resourceID string) error {
	if resourceID == "" {
		return secret.ErrInvalidResourceIDParam
	}

	if _, err := uuid.Parse(resourceID); err != nil {
		return errors.Wrap(err, secret.ErrInvalidResourceIDParam.Error())
	}

	return nil
}
