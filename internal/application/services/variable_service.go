package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure variableService defined type fully satisfy the variable.Service interface.
var _ variable.Service = variableService{}

// variableService implements the interface variable.Service.
type variableService struct {
	variableRepository variable.Repository
}

// NewVariableService return a new instance of a variable.Service that uses the given variable.Repository.
func NewVariableService(variableRepository variable.Repository) (variable.Service, error) {
	if variableRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &variableService{
		variableRepository: variableRepository,
	}, nil
}

// List handles the domain logic to retrieve a list of variables.
func (c variableService) List(ctx context.Context, resourceID string) (variable.Variables, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToListVariables.Error())
	}

	vars, err := c.variableRepository.List(ctx, resourceID)
	if err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToListVariables.Error())
	}

	return vars, nil
}

// Update handles the domain logic to update a variable.
func (c variableService) Update(ctx context.Context, resourceID string, environmentVariablesRequest variable.DiffRequest, environmentVariableAliasesRequest variable.DiffRequest, environmentVariableOverridesRequest variable.DiffRequest) (variable.Variables, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
	}

	environmentVariables, err := c.updateEnvironmentVariables(ctx, resourceID, environmentVariablesRequest)
	if err != nil {
		return nil, err
	}

	// The purpose is to get every variable for the current scope.
	// We need them to be able to create aliases & overrides from a higher scope
	if err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToListVariables.Error())
	}
	environmentVariablesForCurrentScope, err := c.variableRepository.List(ctx, resourceID)
	var environmentVariablesByName = make(map[string]variable.Variable)
	for _, environmentVariable := range environmentVariablesForCurrentScope {
		environmentVariablesByName[environmentVariable.Key] = environmentVariable
	}

	environmentVariableAliases, err := c.updateEnvironmentVariableAliases(ctx, resourceID, environmentVariableAliasesRequest, environmentVariablesByName)
	if err != nil {
		return nil, err
	}
	environmentVariableOverrides, err := c.updateEnvironmentVariableOverrides(ctx, resourceID, environmentVariableOverridesRequest, environmentVariablesByName)
	if err != nil {
		return nil, err
	}

	environmentVariables = append(environmentVariables, environmentVariableAliases...)
	environmentVariables = append(environmentVariables, environmentVariableOverrides...)

	return environmentVariables, nil
}

func (c variableService) updateEnvironmentVariables(ctx context.Context, resourceID string, request variable.DiffRequest) (variable.Variables, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
	}

	variables := make(variable.Variables, 0, len(request.Create)+len(request.Update))
	for _, toDelete := range request.Delete {
		err := c.variableRepository.Delete(ctx, resourceID, toDelete.VariableID)
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}
	}

	for _, toUpdate := range request.Update {
		v, err := c.variableRepository.Update(ctx, resourceID, toUpdate.VariableID, toUpdate.UpsertRequest)
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		variables = append(variables, *v)
	}

	for _, toCreate := range request.Create {
		v, err := c.variableRepository.Create(ctx, resourceID, toCreate.UpsertRequest)
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		variables = append(variables, *v)
	}

	return variables, nil
}

func (c variableService) updateEnvironmentVariableAliases(ctx context.Context, resourceID string, request variable.DiffRequest, environmentVariablesByName map[string]variable.Variable) (variable.Variables, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
	}

	aliases := make(variable.Variables, 0, len(request.Create)+len(request.Update))

	for _, toDelete := range request.Delete {
		err := c.variableRepository.Delete(ctx, resourceID, toDelete.VariableID)
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}
	}

	for _, toUpdate := range request.Update {
		// If the variable alias value has been updated, it means it targets a new aliased variable.
		// So delete it firstly and re-create it
		err := c.variableRepository.Delete(ctx, resourceID, toUpdate.VariableID)
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}
		// The alias variable value contains the name of the aliased variable
		aliasedVariableId := environmentVariablesByName[toUpdate.Value].ID
		v, err := c.variableRepository.CreateAlias(ctx, resourceID, toUpdate.UpsertRequest, aliasedVariableId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		aliases = append(aliases, *v)
	}

	for _, toCreate := range request.Create {
		// The alias variable value contains the name of the aliased variable
		aliasedVariableId := environmentVariablesByName[toCreate.Value].ID
		v, err := c.variableRepository.CreateAlias(ctx, resourceID, toCreate.UpsertRequest, aliasedVariableId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		aliases = append(aliases, *v)
	}

	return aliases, nil
}

func (c variableService) updateEnvironmentVariableOverrides(ctx context.Context, resourceID string, request variable.DiffRequest, environmentVariablesByName map[string]variable.Variable) (variable.Variables, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
	}

	overrides := make(variable.Variables, 0, len(request.Create)+len(request.Update))
	for _, toDelete := range request.Delete {
		err := c.variableRepository.Delete(ctx, resourceID, toDelete.VariableID)
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}
	}

	for _, toUpdate := range request.Update {
		// If the variable override value has been updated, it means it targets a new overridden variable.
		// So delete it firstly and re-create it
		err := c.variableRepository.Delete(ctx, resourceID, toUpdate.VariableID)
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}
		// The override variable value contains the name of the overridden variable
		overriddenVariableId := environmentVariablesByName[toUpdate.Key].ID
		v, err := c.variableRepository.CreateOverride(ctx, resourceID, toUpdate.UpsertRequest, overriddenVariableId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		overrides = append(overrides, *v)
	}

	for _, toCreate := range request.Create {
		// The override variable value contains the name of the overridden variable
		overriddenVariableId := environmentVariablesByName[toCreate.Key].ID
		v, err := c.variableRepository.CreateOverride(ctx, resourceID, toCreate.UpsertRequest, overriddenVariableId.String())
		if err != nil {
			return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
		}

		overrides = append(overrides, *v)
	}

	return overrides, nil
}

// checkResourceID validates that the given resourceID is valid.
func (c variableService) checkResourceID(resourceID string) error {
	if resourceID == "" {
		return variable.ErrInvalidResourceIDParam
	}

	if _, err := uuid.Parse(resourceID); err != nil {
		return errors.Wrap(err, variable.ErrInvalidResourceIDParam.Error())
	}

	return nil
}
