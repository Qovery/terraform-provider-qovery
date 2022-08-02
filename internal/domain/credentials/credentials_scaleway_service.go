package credentials

import (
	"context"
)

// ScalewayService represents the interface to implement to handle the domain logic of AWS Credentials.
type ScalewayService interface {
	Create(ctx context.Context, organizationID string, request UpsertScalewayRequest) (*Credentials, error)
	Get(ctx context.Context, organizationID string, credentialsID string) (*Credentials, error)
	Update(ctx context.Context, organizationID string, credentialsID string, request UpsertScalewayRequest) (*Credentials, error)
	Delete(ctx context.Context, organizationID string, credentialsID string) error
}
