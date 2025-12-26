package organizationwebhook

import (
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidKindOrganizationWebhook     = errors.New("invalid organization webhook kind")
	ErrInvalidEventOrganizationWebhook    = errors.New("invalid organization webhook event")
	ErrInvalidOrganizationWebhookURLParam = errors.New("invalid organization webhook URL parameter")
)

// OrganizationWebhook represents a webhook associated with an organization.
type OrganizationWebhook struct {
	ID                     uuid.UUID `validate:"required"`
	OrganizationID         uuid.UUID `validate:"required"`
	URL                    string    `validate:"required"`
	Kind                   Kind      `validate:"required"`
	Events                 []Event   `validate:"required"`
	Secret                 string
	Description            string
	Enabled                bool
	ProjectNamesFilter     []string
	EnvironmentTypesFilter []string
	CreatedAt              time.Time
	UpdatedAt              *time.Time
}

// Kind represents the kind of organization webhook.
type Kind string

const (
	KindStandard Kind = "STANDARD"
	KindSlack    Kind = "SLACK"
)

// Validate the Kind of webhook
func (k Kind) Validate() error {
	switch k {
	case KindStandard, KindSlack:
		return nil
	default:
		return ErrInvalidKindOrganizationWebhook
	}
}

// Event represents an event that can trigger the webhook.
type Event string

const (
	EventDeploymentStarted    Event = "DEPLOYMENT_STARTED"
	EventDeploymentCancelled  Event = "DEPLOYMENT_CANCELLED"
	EventDeploymentFailed     Event = "DEPLOYMENT_FAILURE"
	EventDeploymentSuccessful Event = "DEPLOYMENT_SUCCESSFUL"
)

// Validate the Event for webhook trigger
func (e Event) Validate() error {
	switch e {
	case EventDeploymentStarted, EventDeploymentCancelled, EventDeploymentFailed, EventDeploymentSuccessful:
		return nil
	default:
		return ErrInvalidEventOrganizationWebhook
	}
}

// Validate checks if the OrganizationWebhook has valid fields.
func (ow *OrganizationWebhook) Validate() error {

	if ow.URL == "" {
		return ErrInvalidOrganizationWebhookURLParam
	}

	_, err := url.ParseRequestURI(ow.URL)
	if err != nil {
		return ErrInvalidOrganizationWebhookURLParam
	}

	if err := ow.Kind.Validate(); err != nil {
		return err
	}

	for _, event := range ow.Events {
		if err := event.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IsValid returns a boolean indicating whether the OrganizationWebhook is valid.
func (ow *OrganizationWebhook) IsValid() bool {
	return ow.Validate() == nil
}
