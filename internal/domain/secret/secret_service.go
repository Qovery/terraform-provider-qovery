package secret

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

//go:generate mockery --testonly --with-expecter --name=Service --structname=SecretService --filename=secret_service_mock.go --output=../../application/services/mocks_test/ --outpkg=mocks_test

var (
	ErrFailedToListSecrets   = errors.New("failed to list secrets")
	ErrFailedToUpdateSecrets = errors.New("failed to update secrets")
)

// Service represents the interface to implement to handle the domain logic of a Secret.
type Service interface {
	List(ctx context.Context, scopeResourceID string) (Secrets, error)
	Update(ctx context.Context, scopeResourceID string, secretsRequest DiffRequest, secretAliasesRequest DiffRequest, secretOverridesRequest DiffRequest) (Secrets, error)
}

// DiffRequest represents the parameters needed to create & update a Secret.
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

// IsValid returns a bool to tell whether the DiffRequest is valid or not.
func (r DiffRequest) IsValid() bool {
	return r.Validate() == nil
}

// IsEmpty returns a bool to tell whether the DiffRequest is empty or not.
func (r DiffRequest) IsEmpty() bool {
	return len(r.Create) == 0 &&
		len(r.Update) == 0 &&
		len(r.Delete) == 0
}

type DiffCreateRequest struct {
	UpsertRequest
}

type DiffUpdateRequest struct {
	UpsertRequest
	SecretID string `validate:"required"`
}

type DiffDeleteRequest struct {
	SecretID string `validate:"required"`
}
