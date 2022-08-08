package credentials

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateAwsCredentials = errors.New("failed to create aws credentials")
	ErrFailedToGetAwsCredentials    = errors.New("failed to get aws credentials")
	ErrFailedToUpdateAwsCredentials = errors.New("failed to update aws credentials")
	ErrFailedToDeleteAwsCredentials = errors.New("failed to delete aws credentials")
)

// AwsService represents the interface to implement to handle the domain logic of AWS Credentials.
type AwsService interface {
	Create(ctx context.Context, organizationID string, request UpsertAwsRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertAwsRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
