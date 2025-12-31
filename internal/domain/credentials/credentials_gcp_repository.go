package credentials

//go:generate mockery --testonly --with-expecter --name=GcpRepository --structname=CredentialsGcpRepository --filename=credentials_gcp_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// GcpRepository represents the interface to implement to handle the persistence of GCP Credentials.
type GcpRepository interface {
	Create(ctx context.Context, organizationID string, request UpsertGcpRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertGcpRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
