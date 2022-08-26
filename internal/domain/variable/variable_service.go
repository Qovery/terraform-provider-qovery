package variable

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

//go:generate mockery --testonly --with-expecter --name=Service --structname=VariableService --filename=variable_service_mock.go --output=../../application/services/mocks_test/ --outpkg=mocks_test

var (
	ErrFailedToListVariables   = errors.New("failed to list variables")
	ErrFailedToUpdateVariables = errors.New("failed to update variables")
)

// Service represents the interface to implement to handle the domain logic of a Variable.
type Service interface {
	List(ctx context.Context, scopeResourceID string) (Variables, error)
	Update(ctx context.Context, scopeResourceID string, request DiffRequest) (Variables, error)
}

// DiffRequest represents the parameters needed to create & update a Variable.
type DiffRequest struct {
	Create []DiffCreateRequest
	Update []DiffUpdateRequest
	Delete []DiffDeleteRequest
}

// Validate returns an error to tell whether the UpsertServiceRequest is valid or not.
func (r DiffRequest) Validate() error {
	for _, c := range r.Create {
		if err := c.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidDiffRequest.Error())
		}
	}

	for _, u := range r.Update {
		if err := validator.New().Struct(u); err != nil {
			return errors.Wrap(err, ErrInvalidDiffRequest.Error())
		}

		if err := u.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidDiffRequest.Error())
		}
	}

	for _, d := range r.Delete {
		if err := validator.New().Struct(d); err != nil {
			return errors.Wrap(err, ErrInvalidDiffRequest.Error())
		}
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertServiceRequest is valid or not.
func (r DiffRequest) IsValid() bool {
	return r.Validate() == nil
}

type DiffCreateRequest struct {
	UpsertRequest
}

type DiffUpdateRequest struct {
	UpsertRequest
	VariableID string `validate:"required"`
}

type DiffDeleteRequest struct {
	VariableID string `validate:"required"`
}
