package secret

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	// ErrNilSecret is returned if a Secret is nil.
	ErrNilSecret = errors.New("secret cannot be nil")
	// ErrInvalidSecret is the error return if a Secret is invalid.
	ErrInvalidSecret = errors.New("invalid secret")
	// ErrInvalidSecrets is the error return if a Secrets is invalid.
	ErrInvalidSecrets = errors.New("invalid secrets")
	// ErrInvalidResourceIDParam is returned if the resource id param is invalid.
	ErrInvalidResourceIDParam = errors.New("invalid resource id param")
	// ErrInvalidSecretIDParam is returned if the secret id param is invalid.
	ErrInvalidSecretIDParam = errors.New("invalid secrets id param")
	// ErrInvalidKeyParam is returned if the key param is invalid.
	ErrInvalidKeyParam = errors.New("invalid key param")
	// ErrInvalidScopeParam is returned if the scope param is invalid.
	ErrInvalidScopeParam = errors.New("invalid scope param")
	// ErrInvalidUpsertRequest is returned if the upsert request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid secrets upsert request")
	// ErrInvalidDiffRequest is returned if the diff request is invalid.
	ErrInvalidDiffRequest = errors.New("invalid secrets diff request")
)

type Secrets []Secret

// Validate returns an error to tell whether the Secrets' domain model is valid or not.
func (s Secrets) Validate() error {
	for _, it := range s {
		if err := it.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidSecrets.Error())
		}
	}

	return nil
}

// IsValid returns a bool to tell whether the Secrets' domain model is valid or not.
func (s Secrets) IsValid() bool {
	return s.Validate() == nil
}

type Secret struct {
	ID    uuid.UUID      `validate:"required"`
	Scope variable.Scope `validate:"required"`
	Key   string         `validate:"required"`
	Type  string
}

// Validate returns an error to tell whether the Secret domain model is valid or not.
func (s Secret) Validate() error {
	if err := s.Scope.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidSecret.Error())
	}

	if err := validator.New().Struct(s); err != nil {
		return errors.Wrap(err, ErrInvalidSecret.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the Secret domain model is valid or not.
func (s Secret) IsValid() bool {
	return s.Validate() == nil
}

// NewSecretParams represents the arguments needed to create a Secret.
type NewSecretsParams = []NewSecretParams
type NewSecretParams struct {
	SecretID string
	Scope    string
	Key      string
	Type     string
}

// NewSecret returns a new instance of a Secret domain model.
func NewSecret(params NewSecretParams) (*Secret, error) {
	secretsUUID, err := uuid.Parse(params.SecretID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidSecretIDParam.Error())
	}

	scope, err := variable.NewScopeFromString(params.Scope)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidScopeParam.Error())
	}

	if params.Key == "" {
		return nil, ErrInvalidKeyParam
	}

	v := &Secret{
		ID:    secretsUUID,
		Key:   params.Key,
		Scope: *scope,
		Type:  params.Type,
	}

	if err := v.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidSecrets.Error())
	}

	return v, nil
}

// UpsertRequest represents the parameters needed to create & update a Secret.
type UpsertRequest struct {
	Key   string `validate:"required"`
	Value string
}

// Validate returns an error to tell whether the UpsertRequest is valid or not.
func (r UpsertRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertRequest is valid or not.
func (r UpsertRequest) IsValid() bool {
	return r.Validate() == nil
}
