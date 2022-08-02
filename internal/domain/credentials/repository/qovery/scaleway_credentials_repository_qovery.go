package qovery

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/common"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsScalewayQoveryRepository implements the interface credentials.ScalewayRepository
type credentialsScalewayQoveryRepository struct {
	client *qovery.APIClient
}

// NewCredentialsScalewayQoveryRepository return a new instance of a credentials.ScalewayRepository that uses Qovery's API.
func NewCredentialsScalewayQoveryRepository(client *qovery.APIClient) (credentials.ScalewayRepository, error) {
	if client == nil {
		return nil, common.ErrInvalidQoveryClient
	}

	return &credentialsScalewayQoveryRepository{
		client: client,
	}, nil
}
