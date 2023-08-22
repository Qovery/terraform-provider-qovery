package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
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
func (c secretService) Update(
	ctx context.Context,
	resourceID string,
	secretsRequest secret.DiffRequest,
	secretAliasesRequest secret.DiffRequest,
	secretOverridesRequest secret.DiffRequest,
	overrideAuthorizedScopes map[variable.Scope]struct{},
) (secret.Secrets, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
	}

	if err := secretsRequest.Validate(); err != nil {
		return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
	}

	secrets, err := c.updateSecrets(ctx, resourceID, secretsRequest)
	if err != nil {
		return nil, err
	}

	// The purpose is to get every variable for the current scope.
	// We need them to be able to create aliases & overrides from a higher scope
	if err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToListVariables.Error())
	}
	secretsForCurrentScope, err := c.secretRepository.List(ctx, resourceID)
	var secretsByNameForAliases = make(map[string]secret.Secret)
	var secretsByNameForOverrides = make(map[string]secret.Secret)
	for _, secretForCurrentScope := range secretsForCurrentScope {
		if secretForCurrentScope.Type == "VALUE" || secretForCurrentScope.Type == "BUILT_IN" {
			secretsByNameForAliases[secretForCurrentScope.Key] = secretForCurrentScope
		}
		_, authorizedScope := overrideAuthorizedScopes[secretForCurrentScope.Scope]
		if secretForCurrentScope.Type == "VALUE" && authorizedScope {
			secretsByNameForOverrides[secretForCurrentScope.Key] = secretForCurrentScope
		}
	}

	secretAliases, err := c.updateSecretAliases(ctx, resourceID, secretAliasesRequest, secretsByNameForAliases)
	if err != nil {
		return nil, err
	}
	secretOverrides, err := c.updateSecretOverrides(ctx, resourceID, secretOverridesRequest, secretsByNameForOverrides)
	if err != nil {
		return nil, err
	}

	secrets = append(secrets, secretAliases...)
	secrets = append(secrets, secretOverrides...)

	return secrets, nil
}

func (c secretService) updateSecrets(ctx context.Context, resourceID string, secretsRequest secret.DiffRequest) (secret.Secrets, error) {
	secrets := make(secret.Secrets, 0, len(secretsRequest.Create)+len(secretsRequest.Update))
	for _, toDelete := range secretsRequest.Delete {
		err := c.secretRepository.Delete(ctx, resourceID, toDelete.SecretID)
		if err != nil {
			return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
		}
	}

	for _, toUpdate := range secretsRequest.Update {
		v, err := c.secretRepository.Update(ctx, resourceID, toUpdate.SecretID, toUpdate.UpsertRequest)
		if err != nil {
			return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
		}

		secrets = append(secrets, *v)
	}

	for _, toCreate := range secretsRequest.Create {
		v, err := c.secretRepository.Create(ctx, resourceID, toCreate.UpsertRequest)
		if err != nil {
			return nil, errors.Wrap(err, secret.ErrFailedToUpdateSecrets.Error())
		}

		secrets = append(secrets, *v)
	}

	return secrets, nil
}

func (c secretService) updateSecretAliases(ctx context.Context, resourceID string, request secret.DiffRequest, secretsByName map[string]secret.Secret) (secret.Secrets, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
	}

	aliases := make(secret.Secrets, 0, len(request.Create)+len(request.Update))

	for _, toDelete := range request.Delete {
		err := c.secretRepository.Delete(ctx, resourceID, toDelete.SecretID)
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil && err.Resp == nil || (err != nil && err.Resp.StatusCode != 404) {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}
	}

	for _, toUpdate := range request.Update {
		// If the variable alias value has been updated, it means it targets a new aliased variable.
		// So delete it firstly and re-create it
		errDelete := c.secretRepository.Delete(ctx, resourceID, toUpdate.SecretID)

		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if errDelete != nil && errDelete.Resp == nil || (errDelete != nil && errDelete.Resp.StatusCode != 404) {
			return nil, errors.Wrap(errDelete, variable.ErrFailedToUpdateVariables.Error())
		}
		// The alias variable value contains the name of the aliased variable
		aliasedSecretId := secretsByName[toUpdate.Value].ID
		v, err := c.secretRepository.CreateAlias(ctx, resourceID, toUpdate.UpsertRequest, aliasedSecretId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		aliases = append(aliases, *v)
	}

	for _, toCreate := range request.Create {
		// The alias variable value contains the name of the aliased variable
		aliasedSecretId := secretsByName[toCreate.Value].ID
		v, err := c.secretRepository.CreateAlias(ctx, resourceID, toCreate.UpsertRequest, aliasedSecretId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		aliases = append(aliases, *v)
	}

	return aliases, nil
}

func (c secretService) updateSecretOverrides(ctx context.Context, resourceID string, request secret.DiffRequest, secretsByName map[string]secret.Secret) (secret.Secrets, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
	}

	overrides := make(secret.Secrets, 0, len(request.Create)+len(request.Update))
	for _, toDelete := range request.Delete {
		err := c.secretRepository.Delete(ctx, resourceID, toDelete.SecretID)
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil && err.Resp == nil || (err != nil && err.Resp.StatusCode != 404) {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}
	}

	for _, toUpdate := range request.Update {
		// If the variable override value has been updated, it means it targets a new overridden variable.
		// So delete it firstly and re-create it
		errDelete := c.secretRepository.Delete(ctx, resourceID, toUpdate.SecretID)

		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if errDelete != nil && errDelete.Resp == nil || (errDelete != nil && errDelete.Resp.StatusCode != 404) {
			return nil, errors.Wrap(errDelete, variable.ErrFailedToUpdateVariables.Error())
		}
		// The override variable value contains the name of the overridden variable
		overriddenSecretId := secretsByName[toUpdate.Key].ID
		v, err := c.secretRepository.CreateOverride(ctx, resourceID, toUpdate.UpsertRequest, overriddenSecretId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		overrides = append(overrides, *v)
	}

	for _, toCreate := range request.Create {
		// The override variable value contains the name of the overridden variable
		overriddenSecretId := secretsByName[toCreate.Key].ID
		v, err := c.secretRepository.CreateOverride(ctx, resourceID, toCreate.UpsertRequest, overriddenSecretId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		overrides = append(overrides, *v)
	}

	return overrides, nil
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
