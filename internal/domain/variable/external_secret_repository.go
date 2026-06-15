package variable

import "context"

// ExternalSecretRepository represents the interface to implement to handle the persistence of external secrets.
type ExternalSecretRepository interface {
	Create(ctx context.Context, serviceID string, request ExternalSecretUpsertRequest) (*ExternalSecret, error)
	Update(ctx context.Context, variableID string, request ExternalSecretUpsertRequest) (*ExternalSecret, error)
	Delete(ctx context.Context, variableID string) error
	List(ctx context.Context, serviceID string) (ExternalSecrets, error)
}
