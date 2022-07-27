package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	// ErrNilCredentials is the error return if an Organization is nil.
	ErrNilCredentials = errors.New("credentials cannot be nil")
	// ErrInvalidCredentials is the error return if an Organization is invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Credentials represents the domain model for a Qovery credential.
type Credentials struct {
	ID             string `validate:"required"`
	OrganizationID string `validate:"required"`
	Name           string `validate:"required"`
}

// Validate returns an error to tell whether the Credentials domain model is valid or not.
func (r Credentials) Validate() error {
	return validator.New().Struct(r)
}

// IsValid returns a bool to tell whether the Credentials domain model is valid or not.
func (r Credentials) IsValid() bool {
	return r.Validate() == nil
}

// NewCredentials returns a new instance of a Credentials domain model.
func NewCredentials(id string, organizationID, name string) (*Credentials, error) {
	creds := &Credentials{
		ID:             id,
		OrganizationID: organizationID,
		Name:           name,
	}

	if err := creds.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidCredentials.Error())
	}

	return creds, nil
}
