package credentials

import (
	"context"

	"github.com/pkg/errors"
)

var (
	// ErrFailedToCreateGcpCredentials is returned if GCP credentials creation fails.
	ErrFailedToCreateGcpCredentials = errors.New("failed to create gcp credentials")
	// ErrFailedToGetGcpCredentials is returned if GCP credentials retrieval fails.
	ErrFailedToGetGcpCredentials = errors.New("failed to get gcp credentials")
	// ErrFailedToUpdateGcpCredentials is returned if GCP credentials update fails.
	ErrFailedToUpdateGcpCredentials = errors.New("failed to update gcp credentials")
	// ErrFailedToDeleteGcpCredentials is returned if GCP credentials deletion fails.
	ErrFailedToDeleteGcpCredentials = errors.New("failed to delete gcp credentials")
	// ErrGcpCredentialsNotFound is returned if GCP credentials don't exist.
	ErrGcpCredentialsNotFound = errors.New("gcp credentials not found")
)

// GcpService represents the interface to implement to handle the domain logic of GCP Credentials.
type GcpService interface {
	Create(ctx context.Context, organizationID string, request UpsertGcpRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertGcpRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
