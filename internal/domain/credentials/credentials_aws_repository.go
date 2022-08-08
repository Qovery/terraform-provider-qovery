package credentials

//go:generate mockery --testonly --with-expecter --name=AwsRepository --structname=CredentialsAwsRepository --filename=credentials_aws_mock.go --output=../../core/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// AwsRepository represents the interface to implement to handle the persistence of AWS Credentials.
type AwsRepository interface {
	Create(ctx context.Context, organizationID string, request UpsertAwsRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertAwsRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
