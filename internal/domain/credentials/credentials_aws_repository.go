package credentials

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
