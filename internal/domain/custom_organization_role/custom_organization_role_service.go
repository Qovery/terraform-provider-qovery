package custom_organization_role

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, organizationId string, request UpsertServiceRequest) (*CustomOrganizationRole, error)
	Get(ctx context.Context, organizationId string, customOrganizationRoleId string) (*CustomOrganizationRole, error)
	Update(ctx context.Context, organizationId string, customOrganizationRoleId string, request UpsertServiceRequest) (*CustomOrganizationRole, error)
	Delete(ctx context.Context, organizationId string, customOrganizationRoleId string) error
}

type UpsertServiceRequest struct {
	CustomOrganizationRoleUpsertRequest UpsertRequest
}

func (r UpsertServiceRequest) Validate() error {

	return nil
}
