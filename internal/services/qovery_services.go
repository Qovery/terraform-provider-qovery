package services

import (
	"log"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	credsQoveryRepository "github.com/qovery/terraform-provider-qovery/internal/domain/credentials/repository/qovery"
	credsService "github.com/qovery/terraform-provider-qovery/internal/domain/credentials/service"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	orgaQoveryRepository "github.com/qovery/terraform-provider-qovery/internal/domain/organization/repository/qovery"
	orgaService "github.com/qovery/terraform-provider-qovery/internal/domain/organization/service"
)

type Services struct {
	AwsCredentialsService credentials.AwsService
	OrganizationService   organization.Service
}

func NewQoveryServices(client *qovery.APIClient) (*Services, error) {
	// Initializing organization service
	orgaSvc, err := newOrganizationService(client)
	if err != nil {
		return nil, err
	}

	// Initializing aws cluster credentials service
	awsCredsSvc, err := newAwsCredentialsService(client)
	if err != nil {
		return nil, err
	}

	return &Services{
		AwsCredentialsService: awsCredsSvc,
		OrganizationService:   orgaSvc,
	}, nil
}

func MustNewQoveryServices(client *qovery.APIClient) *Services {
	qoveryServices, err := NewQoveryServices(client)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return qoveryServices
}

func newOrganizationService(client *qovery.APIClient) (organization.Service, error) {
	orgaRepo, err := orgaQoveryRepository.NewOrganizationQoveryRepository(client)
	if err != nil {
		return nil, err
	}

	orgaSvc, err := orgaService.NewOrganizationService(orgaRepo)
	if err != nil {
		return nil, err
	}

	return orgaSvc, nil
}

func newAwsCredentialsService(client *qovery.APIClient) (credentials.AwsService, error) {
	awsCredsRepo, err := credsQoveryRepository.NewCredentialsAwsQoveryRepository(client)
	if err != nil {
		return nil, err
	}

	awsCredsSvc, err := credsService.NewCredentialsAwsService(awsCredsRepo)
	if err != nil {
		return nil, err
	}

	return awsCredsSvc, nil
}
