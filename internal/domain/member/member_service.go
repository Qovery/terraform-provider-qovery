package member

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToInviteMember = errors.New("failed to invite organization member")
	ErrFailedToGetMember    = errors.New("failed to get organization member")
	ErrFailedToUpdateMember = errors.New("failed to update organization member")
	ErrFailedToDeleteMember = errors.New("failed to delete organization member")
)

// Service represents the interface to implement to handle the domain logic of an organization member.
type Service interface {
	Create(ctx context.Context, organizationID string, request InviteRequest) (*Member, error)
	Get(ctx context.Context, organizationID string, email string) (*Member, error)
	Update(ctx context.Context, organizationID string, email string, request UpdateRoleRequest) (*Member, error)
	Delete(ctx context.Context, organizationID string, email string) error
}
