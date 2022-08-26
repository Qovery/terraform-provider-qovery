package credentials

//go:generate mockery --testonly --with-expecter --name=ScalewayRepository --structname=CredentialsScalewayRepository --filename=credentials_scaleway_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// ScalewayRepository represents the interface to implement to handle the persistence of AWS Credentials.
type ScalewayRepository interface {
	Create(ctx context.Context, organizationID string, request UpsertScalewayRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertScalewayRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
