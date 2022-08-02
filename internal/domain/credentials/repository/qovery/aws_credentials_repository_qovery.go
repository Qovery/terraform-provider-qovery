package qovery

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/common"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsAwsQoveryRepository implements the interface credentials.AwsRepository
type credentialsAwsQoveryRepository struct {
	client *qovery.APIClient
}

// NewCredentialsAwsQoveryRepository return a new instance of a credentials.AwsRepository that uses Qovery's API.
func NewCredentialsAwsQoveryRepository(client *qovery.APIClient) (credentials.AwsRepository, error) {
	if client == nil {
		return nil, common.ErrInvalidQoveryClient
	}

	return &credentialsAwsQoveryRepository{
		client: client,
	}, nil
}
