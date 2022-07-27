package credentials

import (
	"context"
)

// AwsService represents the interface to implement to handle the domain logic of AWS Credentials.
type AwsService interface {
	Create(ctx context.Context, organizationID string, request UpsertAwsRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertAwsRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
