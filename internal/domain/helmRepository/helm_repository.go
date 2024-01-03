package helmRepository

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"net/url"
)

var (
	ErrInvalidRepository                    = errors.New("invalid repository")
	ErrInvalidKindParam                     = errors.New("invalid kind param")
	ErrInvalidURLParam                      = errors.New("invalid url param")
	ErrInvalidRepositoryOrganizationIDParam = errors.New("invalid organization id param")
	ErrInvalidRepositoryNameParam           = errors.New("invalid registry name param")
	ErrInvalidUpsertRequest                 = errors.New("invalid helm repository upsert request")
)

type HelmRepository struct {
	ID                 uuid.UUID `validate:"required"`
	OrganizationID     uuid.UUID `validate:"required"`
	Name               string    `validate:"required"`
	Kind               Kind      `validate:"required"`
	URL                url.URL   `validate:"required"`
	Description        *string
	Config             map[string]string
	SkiTlsVerification *bool
}

func (p HelmRepository) Validate() error {
	return validator.New().Struct(p)
}

func (p HelmRepository) IsValid() bool {
	return p.Validate() == nil
}

// UpsertRequest represents the parameters needed to create & update a Variable.
type UpsertRequest struct {
	Name               string `validate:"required"`
	Kind               string `validate:"required"`
	URL                string `validate:"required"`
	Description        *string
	Config             registry.UpsertRequestConfig
	SkiTlsVerification bool
}

func (r UpsertRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

func (r UpsertRequest) IsValid() bool {
	return r.Validate() == nil
}

type NewHelmRepositoryParams struct {
	RepositoryId       string
	OrganizationID     string
	Name               string
	Kind               string
	URL                string
	Description        *string
	SkiTlsVerification *bool
}

func NewHelmRepository(params NewHelmRepositoryParams) (*HelmRepository, error) {
	repositoryUUID, err := uuid.Parse(params.RepositoryId)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidRepositoryIdParam.Error())
	}

	organizationUUID, err := uuid.Parse(params.OrganizationID)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidRepositoryOrganizationIDParam.Error())
	}

	repositoryUrl, err := url.Parse(params.URL)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidURLParam.Error())
	}

	kind, err := NewKindFromString(params.Kind)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidKindParam.Error())
	}

	if params.Name == "" {
		return nil, ErrInvalidRepositoryNameParam
	}

	r := &HelmRepository{
		ID:                 repositoryUUID,
		OrganizationID:     organizationUUID,
		Name:               params.Name,
		Kind:               *kind,
		URL:                *repositoryUrl,
		Description:        params.Description,
		SkiTlsVerification: params.SkiTlsVerification,
	}

	if err := r.Validate(); err != nil {
		return nil, errors.Wrap(err, ErrInvalidRepository.Error())
	}

	return r, nil
}
