package apitoken

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidApiToken            = errors.New("invalid api token")
	ErrInvalidOrganizationIdParam = errors.New("invalid organization id param")
	ErrInvalidApiTokenIdParam     = errors.New("invalid api token id param")
	ErrInvalidCreateRequest       = errors.New("invalid api token create request")
)

// ApiToken represents an organization api token. Token holds the secret value and is only
// returned by the API at creation time; it is nil when the token is fetched afterwards.
type ApiToken struct {
	ID             uuid.UUID `validate:"required"`
	OrganizationID uuid.UUID `validate:"required"`
	Name           string    `validate:"required"`
	Description    *string
	RoleID         string `validate:"required"`
	Token          *string
}

func (t ApiToken) Validate() error {
	if err := validator.New().Struct(t); err != nil {
		return errors.Wrap(err, ErrInvalidApiToken.Error())
	}
	return nil
}

type CreateRequest struct {
	Name        string `validate:"required"`
	Description *string
	RoleID      string `validate:"required"`
}

func (r CreateRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidCreateRequest.Error())
	}
	return nil
}
