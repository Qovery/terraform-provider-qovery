package services

import (
	"log"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/core/repositories"
	"github.com/qovery/terraform-provider-qovery/internal/core/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	orgaQoveryRepository "github.com/qovery/terraform-provider-qovery/internal/domain/organization/repository/qovery"
	orgaService "github.com/qovery/terraform-provider-qovery/internal/domain/organization/service"
)

type Services struct {
	AwsCredentialsService      credentials.AwsService
	ScalewayCredentialsService credentials.ScalewayService
	OrganizationService        organization.Service
}

func NewQoveryServices(client *qovery.APIClient, apiToken string, providerVersion string) (*Services, error) {
	repos, err := repositories.New(
		repositories.WithQoveryAPI(apiToken, providerVersion),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize repositories")
	}

	// Initializing organization service
	orgaSvc, err := newOrganizationService(client)
	if err != nil {
		return nil, err
	}

	// Initializing aws cluster credentials service
	awsCredsSvc, err := services.NewCredentialsAwsService(repos.CredentialsAws)
	if err != nil {
		return nil, err
	}

	// Initializing scaleway cluster credentials service
	scalewayCredsSvc, err := services.NewCredentialsScalewayService(repos.CredentialsScaleway)
	if err != nil {
		return nil, err
	}

	return &Services{
		AwsCredentialsService:      awsCredsSvc,
		ScalewayCredentialsService: scalewayCredsSvc,
		OrganizationService:        orgaSvc,
	}, nil
}

func MustNewQoveryServices(client *qovery.APIClient, apiToken string, providerVersion string) *Services {
	qoveryServices, err := NewQoveryServices(client, apiToken, providerVersion)
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
