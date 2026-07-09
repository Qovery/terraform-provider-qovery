package member

//go:generate mockery --testonly --with-expecter --name=Repository --structname=OrganizationMemberRepository --filename=organization_member_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import "context"

// Repository handles the persistence of organization members. The API has no get-single
// endpoint and no stable id across the invite lifecycle, so every operation is keyed on
// the member email: Get lists pending invitations first, then active members.
type Repository interface {
	Create(ctx context.Context, organizationID string, request InviteRequest) (*Member, error)
	Get(ctx context.Context, organizationID string, email string) (*Member, error)
	Update(ctx context.Context, organizationID string, email string, request UpdateRoleRequest) (*Member, error)
	Delete(ctx context.Context, organizationID string, email string) error
}
