package secret

//go:generate mockery --testonly --with-expecter --name=Repository --structname=SecretRepository --filename=secret_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

// Repository represents the interface to implement to handle the persistence of a Secret.
// scopeResourceID can be either a projectID, environmentID, application or containerID
type Repository interface {
	Create(ctx context.Context, scopeResourceID string, request UpsertRequest) (*Secret, error)
	CreateAlias(ctx context.Context, scopeResourceID string, request UpsertRequest, aliasedSecretId string) (*Secret, error)
	CreateOverride(ctx context.Context, scopeResourceID string, request UpsertRequest, overriddenSecretId string) (*Secret, error)
	List(ctx context.Context, scopeResourceID string) (Secrets, error)
	Update(ctx context.Context, scopeResourceID string, secretID string, request UpsertRequest) (*Secret, error)
	Delete(ctx context.Context, scopeResourceID string, secretID string) *apierrors.ApiError
}
