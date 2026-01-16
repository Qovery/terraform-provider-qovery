package credentials

import (
	"context"

	"github.com/pkg/errors"
)

var (
	// ErrFailedToCreateAzureCredentials is returned if Azure credentials creation fails.
	ErrFailedToCreateAzureCredentials = errors.New("failed to create azure credentials")
	// ErrFailedToGetAzureCredentials is returned if Azure credentials retrieval fails.
	ErrFailedToGetAzureCredentials = errors.New("failed to get azure credentials")
	// ErrFailedToUpdateAzureCredentials is returned if Azure credentials update fails.
	ErrFailedToUpdateAzureCredentials = errors.New("failed to update azure credentials")
	// ErrFailedToDeleteAzureCredentials is returned if Azure credentials deletion fails.
	ErrFailedToDeleteAzureCredentials = errors.New("failed to delete azure credentials")
	// ErrAzureCredentialsNotFound is returned if Azure credentials don't exist.
	ErrAzureCredentialsNotFound = errors.New("azure credentials not found")
)

// AzureService represents the interface to implement to handle the domain logic of Azure Credentials.
type AzureService interface {
	Create(ctx context.Context, organizationID string, request UpsertAzureRequest) (*AzureCredentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*AzureCredentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertAzureRequest) (*AzureCredentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
