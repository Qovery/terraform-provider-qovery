package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrNilCredentials is the error return if a Credentials is nil.
	ErrNilCredentials = errors.New("credentials cannot be nil")
	// ErrInvalidCredentials is the error return if a Credentials is invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidCredentialsID is the error return if a Credentials ID is invalid.
	ErrInvalidCredentialsID = errors.New("invalid credentials id")
	// ErrInvalidCredentialsOrganizationID is the error return if a Credentials ID is invalid.
	ErrInvalidCredentialsOrganizationID = errors.New("invalid credentials organization id")
)

// Credentials represents the domain model for a Qovery credential.
type Credentials struct {
	ID             uuid.UUID `validate:"required"`
	OrganizationID uuid.UUID `validate:"required"`
	Name           string    `validate:"required"`
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
func NewCredentials(credentialsID string, organizationID string, name string) (*Credentials, error) {
	credentialsUUID, err := uuid.Parse(credentialsID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidCredentialsID.Error())
	}

	organizationUUID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidCredentialsOrganizationID.Error())
	}

	creds := &Credentials{
		ID:             credentialsUUID,
		OrganizationID: organizationUUID,
		Name:           name,
	}

	if err := creds.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidCredentials.Error())
	}

	return creds, nil
}
