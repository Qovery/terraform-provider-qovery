package variable

//go:generate mockery --testonly --with-expecter --name=Repository --structname=VariableRepository --filename=variable_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

// Repository represents the interface to implement to handle the persistence of a Variable.
// scopeResourceID can be either a projectID, environmentID, application or containerID
type Repository interface {
	Create(ctx context.Context, scopeResourceID string, request UpsertRequest) (*Variable, error)
	CreateAlias(ctx context.Context, scopeResourceID string, request UpsertRequest, aliasedVariableId string) (*Variable, error)
	CreateOverride(ctx context.Context, scopeResourceID string, request UpsertRequest, overriddenVariableId string) (*Variable, error)
	List(ctx context.Context, scopeResourceID string) (Variables, error)
	Update(ctx context.Context, scopeResourceID string, variableID string, request UpsertRequest) (*Variable, error)
	Delete(ctx context.Context, scopeResourceID string, variableID string) *apierrors.APIError
}
