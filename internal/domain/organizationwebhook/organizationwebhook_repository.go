package organizationwebhook

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	ErrInvalidUpsertRequest = errors.New("invalid organization webhook upsert request")
)

type Repository interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*OrganizationWebhook, error)
	Get(ctx context.Context, organizationID string, organizationWebhookID string, id string) (*OrganizationWebhook, error)
	List(ctx context.Context, organizationID string) ([]*OrganizationWebhook, error)
	Update(ctx context.Context, organizationID string, organizationWebhookID string, id string, request UpsertRequest) (*OrganizationWebhook, error)
	Delete(ctx context.Context, organizationID string, organizationWebhookID string, id string) error
}

// UpsertRequest represents the parameters needed to create & update a organization webhook.
type UpsertRequest struct {
	URL    string  `validate:"required"`
	Kind   Kind    `validate:"required"`
	Events []Event `validate:"required"`

	Secret                 string
	Description            string
	Enabled                bool
	ProjectNamesFilter     []string
	EnvironmentTypesFilter []string
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
