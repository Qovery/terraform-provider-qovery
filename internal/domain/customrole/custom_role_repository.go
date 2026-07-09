package customrole

//go:generate mockery --testonly --with-expecter --name=Repository --structname=CustomRoleRepository --filename=custom_role_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// Repository represents the interface to implement to handle the persistence of an organization custom role.
type Repository interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*CustomRole, error)
	Get(ctx context.Context, organizationID string, customRoleID string) (*CustomRole, error)
	Update(ctx context.Context, organizationID string, customRoleID string, request UpsertRequest) (*CustomRole, error)
	Delete(ctx context.Context, organizationID string, customRoleID string) error
}
