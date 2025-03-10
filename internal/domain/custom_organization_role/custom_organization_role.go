package custom_organization_role

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidCustomOrganizationRoleRequest             = errors.New("invalid custom organization role request")
	ErrInvalidCustomOrganizationRoleOrganizationIdParam = errors.New("invalid organization id format")
	ErrFailedToCreateCustomOrganizationRole             = errors.New("failed to create custom organization role")
	ErrFailedToGetCustomOrganizationRole                = errors.New("failed to get custom organization role")
	ErrFailedToUpdateCustomOrganizationRole             = errors.New("failed to update custom organization role")
	ErrFailedToDeleteCustomOrganizationRole             = errors.New("failed to delete custom organization role")
	ErrInvalidCustomOrganizationRoleIdParam             = errors.New("invalid custom organization role id format")
)

type CustomOrganizationRole struct {
	Id uuid.UUID `validate:"required"`
	//Name string
}
