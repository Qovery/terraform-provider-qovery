package customrole

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidCustomRole          = errors.New("invalid custom role")
	ErrInvalidUpsertRequest       = errors.New("invalid custom role upsert request")
	ErrInvalidName                = errors.New("invalid custom role name: must be non-blank without leading or trailing whitespace")
	ErrReservedName               = errors.New("invalid custom role name: owner, admin, devops, billing and viewer are reserved built-in role names")
	ErrInvalidOrganizationIdParam = errors.New("invalid organization id param")
	ErrInvalidCustomRoleIdParam   = errors.New("invalid custom role id param")
	ErrDuplicateClusterID         = errors.New("duplicate cluster_id in cluster_permissions")
	ErrDuplicateProjectID         = errors.New("duplicate project_id in project_permissions")
	ErrAdminWithPermissions       = errors.New("project permissions must not be set when is_admin is true")
	ErrIncompleteEnvironmentTypes = errors.New("project permissions must contain exactly one entry per environment type (DEVELOPMENT, PREVIEW, STAGING, PRODUCTION) when is_admin is false")
	ErrDuplicateEnvironmentType   = errors.New("duplicate environment_type in project permissions")
	ErrInvalidClusterPermission   = errors.New("invalid cluster permission: must be one of VIEWER, ENV_CREATOR, ADMIN")
	ErrInvalidProjectPermission   = errors.New("invalid project permission: must be one of NO_ACCESS, VIEWER, DEPLOYER, MANAGER")
	ErrInvalidEnvironmentType     = errors.New("invalid environment type: must be one of DEVELOPMENT, PREVIEW, STAGING, PRODUCTION")
)

var reservedRoleNames = []string{"owner", "admin", "devops", "billing", "viewer"}

type ClusterPermission string

const (
	ClusterPermissionViewer     ClusterPermission = "VIEWER"
	ClusterPermissionEnvCreator ClusterPermission = "ENV_CREATOR"
	ClusterPermissionAdmin      ClusterPermission = "ADMIN"
)

var AllowedClusterPermissions = []ClusterPermission{ClusterPermissionViewer, ClusterPermissionEnvCreator, ClusterPermissionAdmin}

func (p ClusterPermission) Validate() error {
	for _, allowed := range AllowedClusterPermissions {
		if p == allowed {
			return nil
		}
	}
	return errors.Wrap(ErrInvalidClusterPermission, string(p))
}

type ProjectPermission string

const (
	ProjectPermissionNoAccess ProjectPermission = "NO_ACCESS"
	ProjectPermissionViewer   ProjectPermission = "VIEWER"
	ProjectPermissionDeployer ProjectPermission = "DEPLOYER"
	ProjectPermissionManager  ProjectPermission = "MANAGER"
)

var AllowedProjectPermissions = []ProjectPermission{ProjectPermissionNoAccess, ProjectPermissionViewer, ProjectPermissionDeployer, ProjectPermissionManager}

func (p ProjectPermission) Validate() error {
	for _, allowed := range AllowedProjectPermissions {
		if p == allowed {
			return nil
		}
	}
	return errors.Wrap(ErrInvalidProjectPermission, string(p))
}

type EnvironmentType string

const (
	EnvironmentTypeDevelopment EnvironmentType = "DEVELOPMENT"
	EnvironmentTypePreview     EnvironmentType = "PREVIEW"
	EnvironmentTypeStaging     EnvironmentType = "STAGING"
	EnvironmentTypeProduction  EnvironmentType = "PRODUCTION"
)

var AllowedEnvironmentTypes = []EnvironmentType{EnvironmentTypeDevelopment, EnvironmentTypePreview, EnvironmentTypeStaging, EnvironmentTypeProduction}

func (t EnvironmentType) Validate() error {
	for _, allowed := range AllowedEnvironmentTypes {
		if t == allowed {
			return nil
		}
	}
	return errors.Wrap(ErrInvalidEnvironmentType, string(t))
}

type ClusterRolePermission struct {
	ClusterID  string `validate:"required,uuid"`
	Permission ClusterPermission
}

type EnvironmentPermission struct {
	EnvironmentType EnvironmentType
	Permission      ProjectPermission
}

type ProjectRolePermission struct {
	ProjectID   string `validate:"required,uuid"`
	IsAdmin     bool
	Permissions []EnvironmentPermission
}

type CustomRole struct {
	ID                 uuid.UUID `validate:"required"`
	OrganizationID     uuid.UUID `validate:"required"`
	Name               string    `validate:"required"`
	Description        *string
	ClusterPermissions []ClusterRolePermission
	ProjectPermissions []ProjectRolePermission
}

func (r CustomRole) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidCustomRole.Error())
	}
	return nil
}

type UpsertRequest struct {
	Name               string `validate:"required"`
	Description        *string
	ClusterPermissions []ClusterRolePermission `validate:"dive"`
	ProjectPermissions []ProjectRolePermission `validate:"dive"`
}

func (r UpsertRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}
	if strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.Name) != r.Name {
		return ErrInvalidName
	}
	for _, reserved := range reservedRoleNames {
		if strings.EqualFold(r.Name, reserved) {
			return ErrReservedName
		}
	}
	seenClusters := make(map[string]bool, len(r.ClusterPermissions))
	for _, cp := range r.ClusterPermissions {
		if seenClusters[cp.ClusterID] {
			return errors.Wrap(ErrDuplicateClusterID, cp.ClusterID)
		}
		seenClusters[cp.ClusterID] = true
		if err := cp.Permission.Validate(); err != nil {
			return err
		}
	}
	seenProjects := make(map[string]bool, len(r.ProjectPermissions))
	for _, pp := range r.ProjectPermissions {
		if seenProjects[pp.ProjectID] {
			return errors.Wrap(ErrDuplicateProjectID, pp.ProjectID)
		}
		seenProjects[pp.ProjectID] = true
		if pp.IsAdmin {
			if len(pp.Permissions) > 0 {
				return errors.Wrap(ErrAdminWithPermissions, pp.ProjectID)
			}
			continue
		}
		if len(pp.Permissions) != len(AllowedEnvironmentTypes) {
			return errors.Wrap(ErrIncompleteEnvironmentTypes, pp.ProjectID)
		}
		seenEnvTypes := make(map[EnvironmentType]bool, len(pp.Permissions))
		for _, ep := range pp.Permissions {
			if err := ep.EnvironmentType.Validate(); err != nil {
				return err
			}
			if err := ep.Permission.Validate(); err != nil {
				return err
			}
			if seenEnvTypes[ep.EnvironmentType] {
				return errors.Wrap(ErrDuplicateEnvironmentType, string(ep.EnvironmentType))
			}
			seenEnvTypes[ep.EnvironmentType] = true
		}
	}
	return nil
}
