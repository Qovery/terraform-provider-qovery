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
func (c variableService) Update(ctx context.Context, resourceID string, request variable.DiffRequest) (variable.Variables, error) {
	if err := c.checkResourceID(resourceID); err != nil {
		return nil, errors.Wrap(err, variable.ErrFailedToUpdateVariables.Error())
	}

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
