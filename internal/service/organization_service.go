package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

type organizationService struct {
	organizationAPI organization.API
}

func NewOrganizationService(organizationAPI organization.API) organization.API {
	return &organizationService{
		organizationAPI: organizationAPI,
	}
}

func (o organizationService) Get(ctx context.Context, organizationID string) (*organization.Organization, error) {
	return o.organizationAPI.Get(ctx, organizationID)
}

func (o organizationService) Update(ctx context.Context, organizationID string, request organization.UpdateRequest) (*organization.Organization, error) {
	return o.organizationAPI.Update(ctx, organizationID, request)
}
