package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidUpsertGcpRequest is returned if a GCP Credentials upsert request is invalid.
	ErrInvalidUpsertGcpRequest = errors.New("invalid credentials upsert gcp request")
)

// UpsertGcpRequest represents the parameters needed to create & update GCP Credentials.
type UpsertGcpRequest struct {
	Name           string `validate:"required"`
	GcpCredentials string `validate:"required"`
}

// Validate returns an error to tell whether the UpsertGcpRequest is valid or not.
func (r UpsertGcpRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertGcpRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertGcpRequest is valid or not.
func (r UpsertGcpRequest) IsValid() bool {
	return r.Validate() == nil
}
