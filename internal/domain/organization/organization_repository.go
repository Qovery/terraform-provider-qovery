package organization

//go:generate mockery --testonly --with-expecter --name=Repository --structname=OrganizationRepository --filename=organization_mock.go --output=../../core/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// Repository represents the interface to implement to handle the persistence of an Organization.
type Repository interface {
	Get(ctx context.Context, organizationID string) (*Organization, error)
	Update(ctx context.Context, organizationID string, request UpdateRequest) (*Organization, error)
}
