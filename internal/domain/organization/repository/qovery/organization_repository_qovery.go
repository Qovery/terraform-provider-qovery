package qovery

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/common"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// organizationQoveryRepository implements the interface organization.Repository
type organizationQoveryRepository struct {
	client *qovery.APIClient
}

// NewOrganizationQoveryRepository return a new instance of an organization.Repository that uses Qovery's API.
func NewOrganizationQoveryRepository(client *qovery.APIClient) (organization.Repository, error) {
	if client == nil {
		return nil, common.ErrInvalidQoveryClient
	}

	return &organizationQoveryRepository{
		client: client,
	}, nil
}
