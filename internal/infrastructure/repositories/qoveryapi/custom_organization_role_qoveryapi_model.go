package qoveryapi

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/custom_organization_role"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func newDomainCustomOrganizationRoleFromQovery(customOrganizationRoleResponse *qovery.OrganizationCustomRole) (*custom_organization_role.CustomOrganizationRole, error) {
	if customOrganizationRoleResponse == nil {
		return nil, variable.ErrNilVariable
	}

	if customOrganizationRoleResponse.Id == nil {
		return nil, variable.ErrNilVariable
	}

	customOrganizationRoleID, err := uuid.Parse(*customOrganizationRoleResponse.Id)
	if err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleIdParam.Error())
	}

	return &custom_organization_role.CustomOrganizationRole{
		Id: customOrganizationRoleID,
	}, nil
}

func newQoveryCustomOrganizationRoleRequestFromDomain(request custom_organization_role.UpsertRequest) (*qovery.OrganizationCustomRoleCreateRequest, error) {
	return &qovery.OrganizationCustomRoleCreateRequest{
		Name:        request.Name,
		Description: request.Description,
	}, nil
}

func newQoveryCustomOrganizationRoleUpdateRequestFromDomain(request custom_organization_role.UpsertRequest) (*qovery.OrganizationCustomRoleUpdateRequest, error) {
	convertToClusterPermission := func(permission string) *qovery.OrganizationCustomRoleClusterPermission {
		var p qovery.OrganizationCustomRoleClusterPermission
		switch permission {
		case string(qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_VIEWER):
			p = qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_VIEWER
		case string(qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_ENV_CREATOR):
			p = qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_ENV_CREATOR
		case string(qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_ADMIN):
			p = qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_ADMIN
		default:
			return nil // Or handle invalid permission values as needed
		}
		return &p
	}

	clusterPerm := make([]qovery.OrganizationCustomRoleUpdateRequestClusterPermissionsInner, 0, len(request.ClusterPermissions))
	for _, clusterPermissions := range request.ClusterPermissions {
		clusterPerm = append(clusterPerm, qovery.OrganizationCustomRoleUpdateRequestClusterPermissionsInner{
			ClusterId:  &clusterPermissions.ClusterId,
			Permission: convertToClusterPermission(clusterPermissions.Permission),
		})
	}

	return &qovery.OrganizationCustomRoleUpdateRequest{
		Name:               request.Name,
		Description:        request.Description,
		ClusterPermissions: clusterPerm,
	}, nil
}

func NewQoveryServiceCustomOrganizationRoleRequestFromDomain(customOrganizationRoleIds []string) ([]qovery.ServiceAnnotationRequest, error) {
	serviceAnnotationRequest := make([]qovery.ServiceAnnotationRequest, 0, len(customOrganizationRoleIds))
	for _, id := range customOrganizationRoleIds {
		newAnnotationsRequest := qovery.ServiceAnnotationRequest{
			Id: id,
		}

		serviceAnnotationRequest = append(serviceAnnotationRequest, newAnnotationsRequest)
	}

	return serviceAnnotationRequest, nil
}
