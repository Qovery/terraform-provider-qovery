package credentials

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrNilCredentials is returned if a Credentials is nil.
	ErrNilCredentials = errors.New("credentials cannot be nil")
	// ErrInvalidCredentials is returned if a Credentials is invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidCredentialsID is returned if a Credentials ID is invalid.
	ErrInvalidCredentialsID = errors.New("invalid credentials id")
	// ErrInvalidCredentialsOrganizationID is returned if a Credentials ID is invalid.
	ErrInvalidCredentialsOrganizationID = errors.New("invalid credentials organization id")
	// ErrInvalidCredentialsName is returned if a Credentials name is invalid.
	ErrInvalidCredentialsName = errors.New("invalid credentials name")
	// ErrInvalidOrganizationIDParam is returned if the organization id param is invalid.
	ErrInvalidOrganizationIDParam = errors.New("invalid organization id param")
	// ErrInvalidCredentialsIDParam is returned if the credential id param is invalid.
	ErrInvalidCredentialsIDParam = errors.New("invalid credentials id param")
	// ErrAwsCredentialsNotFound is returned if an aws Credentials doesn't exist.
	ErrAwsCredentialsNotFound = errors.New("aws credentials not found")
	// ErrScalewayCredentialsNotFound is returned if a scaleway Credentials doesn't exist.
	ErrScalewayCredentialsNotFound = errors.New("scaleway credentials not found")
)

// Credentials represents the domain model for a Qovery credential.
type Credentials struct {
	ID             uuid.UUID `validate:"required"`
	OrganizationID uuid.UUID `validate:"required"`
	Name           string    `validate:"required"`
}

// Validate returns an error to tell whether the Credentials domain model is valid or not.
func (r Credentials) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidCredentials.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the Credentials domain model is valid or not.
func (r Credentials) IsValid() bool {
	return r.Validate() == nil
}

// NewCredentialsParams represents the arguments needed to create a Credentials.
type NewCredentialsParams struct {
	CredentialsID  string
	OrganizationID string
	Name           string
}

// NewCredentials returns a new instance of a Credentials domain model.
func NewCredentials(params NewCredentialsParams) (*Credentials, error) {
	credentialsUUID, err := uuid.Parse(params.CredentialsID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidCredentialsID.Error())
	}

	organizationUUID, err := uuid.Parse(params.OrganizationID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidCredentialsOrganizationID.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidCredentialsName
	}

	creds := &Credentials{
		ID:             credentialsUUID,
		OrganizationID: organizationUUID,
		Name:           params.Name,
	}

	if err := creds.Validate(); err != nil {
		return nil, err
	}

	return creds, nil
}
