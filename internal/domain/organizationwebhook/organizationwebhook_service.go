package organizationwebhook

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateOrganizationWebhook = errors.New("failed to create organization webhook")
	ErrFailedToGetOrganizationWebhook    = errors.New("failed to get organization webhook")
	ErrFailedToListOrganizationWebhooks  = errors.New("failed to list organization webhooks")
	ErrFailedToUpdateOrganizationWebhook = errors.New("failed to update organization webhook")
	ErrFailedToDeleteOrganizationWebhook = errors.New("failed to delete organization webhook")
)

// Service represents the interface to implement to handle the domain logic of an Organization Webhook.
type Service interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*OrganizationWebhook, error)
	Get(ctx context.Context, organizationID string, organizationWebhookID string, id string) (*OrganizationWebhook, error)
	List(ctx context.Context, organizationID string) ([]*OrganizationWebhook, error)
	Update(ctx context.Context, organizationID string, organizationWebhookID string, id string, request UpsertRequest) (*OrganizationWebhook, error)
	Delete(ctx context.Context, organizationID string, organizationWebhookID string, id string) error
}
