package services

import (
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/core/repositories"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

var (
	// ErrInvalidRepository is the error return if the given repository is nil or invalid.
	ErrInvalidRepository = errors.New("invalid repository")
)

// Services contains the implementations of domain services using.
type Services struct {
	repos *repositories.Repositories

	CredentialsAws      credentials.AwsService
	CredentialsScaleway credentials.ScalewayService
	Organization        organization.Service
}

// Configuration represents a function that handle the QoveryAPI configuration.
type Configuration func(qoveryAPI *Services) error

// New returns a new instance of QoveryAPI and applies the given configs.
func New(configs ...Configuration) (*Services, error) {
	services := &Services{}

	// Apply all the configs to the qoveryAPI instance.
	for _, config := range configs {
		if err := config(services); err != nil {
			return nil, err
		}
	}

	// Initialize services implementations.
	credentialsAwsService, err := NewCredentialsAwsService(services.repos.CredentialsAws)
	if err != nil {
		return nil, err
	}

	credentialsScalewayService, err := NewCredentialsScalewayService(services.repos.CredentialsScaleway)
	if err != nil {
		return nil, err
	}

	organizationService, err := NewOrganizationService(services.repos.Organization)
	if err != nil {
		return nil, err
	}

	services.CredentialsAws = credentialsAwsService
	services.CredentialsScaleway = credentialsScalewayService
	services.Organization = organizationService

	return services, nil
}

func WithQoveryRepository(apiToken string, providerVersion string) Configuration {
	return func(services *Services) error {
		repos, err := repositories.New(repositories.WithQoveryAPI(apiToken, providerVersion))
		if err != nil {
			return err
		}

		services.repos = repos

		return nil
	}
}
