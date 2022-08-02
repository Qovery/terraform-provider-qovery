package credentials

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
