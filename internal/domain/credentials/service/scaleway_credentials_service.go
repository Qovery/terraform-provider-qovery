package service

import (
	"github.com/qovery/terraform-provider-qovery/internal/domain/common"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsScalewayService implements the interface credentials.ScalewayService.
type credentialsScalewayService struct {
	credentialsScalewayRepository credentials.ScalewayRepository
}

// NewCredentialsScalewayService return a new instance of a credentials.ScalewayService that uses the given credentials.ScalewayRepository.
func NewCredentialsScalewayService(credentialsScalewayRepository credentials.ScalewayRepository) (credentials.ScalewayService, error) {
	if credentialsScalewayRepository == nil {
		return nil, common.ErrInvalidRepository
	}

	return &credentialsScalewayService{
		credentialsScalewayRepository: credentialsScalewayRepository,
	}, nil
}
