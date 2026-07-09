package customrole

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateCustomRole = errors.New("failed to create custom role")
	ErrFailedToGetCustomRole    = errors.New("failed to get custom role")
	ErrFailedToUpdateCustomRole = errors.New("failed to update custom role")
	ErrFailedToDeleteCustomRole = errors.New("failed to delete custom role")
)

// Service represents the interface to implement to handle the domain logic of an organization custom role.
type Service interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*CustomRole, error)
	Get(ctx context.Context, organizationID string, customRoleID string) (*CustomRole, error)
	Update(ctx context.Context, organizationID string, customRoleID string, request UpsertRequest) (*CustomRole, error)
	Delete(ctx context.Context, organizationID string, customRoleID string) error
}
