package credentials

//go:generate mockery --testonly --with-expecter --name=AzureRepository --structname=CredentialsAzureRepository --filename=credentials_azure_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// AzureRepository represents the interface to implement to handle the persistence of Azure Credentials.
type AzureRepository interface {
	Create(ctx context.Context, organizationID string, request UpsertAzureRequest) (*AzureCredentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*AzureCredentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertAzureRequest) (*AzureCredentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
