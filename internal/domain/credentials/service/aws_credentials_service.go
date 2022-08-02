package service

import (
	"github.com/qovery/terraform-provider-qovery/internal/domain/common"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsAwsService implements the interface credentials.AwsService.
type credentialsAwsService struct {
	credentialsAwsRepository credentials.AwsRepository
}

// NewCredentialsAwsService return a new instance of a credentials.AwsService that uses the given credentials.AwsRepository.
func NewCredentialsAwsService(credentialsAwsRepository credentials.AwsRepository) (credentials.AwsService, error) {
	if credentialsAwsRepository == nil {
		return nil, common.ErrInvalidRepository
	}

	return &credentialsAwsService{
		credentialsAwsRepository: credentialsAwsRepository,
	}, nil
}
