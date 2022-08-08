package credentials

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateScalewayCredentials = errors.New("failed to create scaleway credentials")
	ErrFailedToGetScalewayCredentials    = errors.New("failed to get scaleway credentials")
	ErrFailedToUpdateScalewayCredentials = errors.New("failed to update scaleway credentials")
	ErrFailedToDeleteScalewayCredentials = errors.New("failed to delete scaleway credentials")
)

// ScalewayService represents the interface to implement to handle the domain logic of AWS Credentials.
type ScalewayService interface {
	Create(ctx context.Context, organizationID string, request UpsertScalewayRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertScalewayRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
