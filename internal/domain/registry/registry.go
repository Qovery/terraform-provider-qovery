package registry

import (
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrNilRegistry is returned if a Registry is nil.
	ErrNilRegistry = errors.New("variable cannot be nil")
	// ErrInvalidRegistry is the error return if a Registry is invalid.
	ErrInvalidRegistry = errors.New("invalid registry")
	// ErrInvalidOrganizationIDParam is returned if the organization id param is invalid.
	ErrInvalidOrganizationIDParam = errors.New("invalid organization id param")
	// ErrInvalidRegistryIDParam is returned if the registry id param is invalid.
	ErrInvalidRegistryIDParam = errors.New("invalid registry id param")
	// ErrInvalidKindParam is returned if the registry kind param is invalid.
	ErrInvalidKindParam = errors.New("invalid kind param")
	// ErrInvalidURLParam is returned if the registry url param is invalid.
	ErrInvalidURLParam = errors.New("invalid url param")
	// ErrInvalidRegistryOrganizationIDParam is returned if the organization id param is invalid.
	ErrInvalidRegistryOrganizationIDParam = errors.New("invalid organization id param")
	// ErrInvalidRegistryNameParam is returned if the value param is invalid.
	ErrInvalidRegistryNameParam = errors.New("invalid registry name param")
	// ErrInvalidUpsertRequest is returned if the create request is invalid.
	ErrInvalidUpsertRequest = errors.New("invalid registry upsert request")
)

type Registry struct {
	ID             uuid.UUID `validate:"required"`
	OrganizationID uuid.UUID `validate:"required"`
	Name           string    `validate:"required"`
	Kind           Kind      `validate:"required"`
	URL            url.URL   `validate:"required"`

	Description *string
	Config      map[string]string
}

// Validate returns an error to tell whether the Registry domain model is valid or not.
func (p Registry) Validate() error {
	return validator.New().Struct(p)
}

// IsValid returns a bool to tell whether the Registry domain model is valid or not.
func (p Registry) IsValid() bool {
	return p.Validate() == nil
}

// NewRegistryParams represents the arguments needed to create a Registry.
type NewRegistryParams struct {
	RegistryID     string
	OrganizationID string
	Name           string
	Kind           string
	URL            string
	Description    *string
}

// NewRegistry returns a new instance of a Registry domain model.
func NewRegistry(params NewRegistryParams) (*Registry, error) {
	registryUUID, err := uuid.Parse(params.RegistryID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidRegistryIDParam.Error())
	}

	organizationUUID, err := uuid.Parse(params.OrganizationID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidRegistryOrganizationIDParam.Error())
	}

	registryURL, err := url.Parse(params.URL)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidURLParam.Error())
	}

	kind, err := NewKindFromString(params.Kind)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidKindParam.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidRegistryNameParam
	}

	v := &Registry{
		ID:             registryUUID,
		OrganizationID: organizationUUID,
		Name:           params.Name,
		Kind:           *kind,
		URL:            *registryURL,
		Description:    params.Description,
	}

	if err := v.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidRegistry.Error())
	}

	return v, nil
}

// UpsertRequest represents the parameters needed to create & update a Variable.
type UpsertRequest struct {
	Name string `validate:"required"`
	Kind string `validate:"required"`
	URL  string `validate:"required"`

	Description *string
	Config      UpsertRequestConfig
}

type UpsertRequestConfig struct {
	AccessKeyID       *string
	SecretAccessKey   *string
	Region            *string
	ScalewayAccessKey *string
	ScalewaySecretKey *string
	Username          *string
	Password          *string
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
