package custom_organization_role

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, organizationId string, request UpsertRequest) (*CustomOrganizationRole, error)
	Get(ctx context.Context, organizationId string, customOrganizationRoleId string) (*CustomOrganizationRole, error)
	Update(ctx context.Context, organizationId string, customOrganizationRoleId string, request UpsertRequest) (*CustomOrganizationRole, error)
	Delete(ctx context.Context, organizationId string, customOrganizationRoleId string) error
}

type UpsertRequest struct {
	Name               string `validate:"required"`
	Description        *string
	ClusterPermissions []ClusterPermissionsRequest
}

type ClusterPermissionsRequest struct {
	ClusterId  string `validate:"required"`
	Permission string `validate:"required"`
}
