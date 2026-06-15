package variable

import "context"

// ExternalSecretFileRepository represents the interface to implement to handle the persistence of external secret files.
type ExternalSecretFileRepository interface {
	Create(ctx context.Context, serviceID string, request ExternalSecretFileUpsertRequest) (*ExternalSecretFile, error)
	Update(ctx context.Context, variableID string, request ExternalSecretFileUpsertRequest) (*ExternalSecretFile, error)
	Delete(ctx context.Context, variableID string) error
	List(ctx context.Context, serviceID string) (ExternalSecretFiles, error)
}
