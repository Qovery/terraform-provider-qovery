package services

import (
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories"
)

var (
	// ErrInvalidRepository is the error return if the given repository is nil or invalid.
	ErrInvalidRepository = errors.New("invalid repository")
	// ErrInvalidService is the error return if the given service is nil or invalid.
	ErrInvalidService = errors.New("invalid service")
	// ErrMissingConfiguration is the error return if no configuration has been given.
	ErrMissingConfiguration = errors.New("missing configuration")
)

// Services contains the implementations of domain services using.
type Services struct {
	repos *repositories.Repositories

	CredentialsAws      credentials.AwsService
	CredentialsScaleway credentials.ScalewayService
	Organization        organization.Service
	Project             project.Service
	Container           container.Service
	ContainerRegistry   registry.Service
	Environment         environment.Service
	DeploymentStage     deploymentstage.Service
}

// Configuration represents a function that handle the QoveryAPI configuration.
type Configuration func(qoveryAPI *Services) error

// New returns a new instance of QoveryAPI and applies the given configs.
func New(configs ...Configuration) (*Services, error) {
	services := &Services{}

	if len(configs) == 0 {
		return nil, ErrMissingConfiguration
	}

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

	projectEnvironmentVariableService, err := NewVariableService(services.repos.ProjectEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	projectSecretService, err := NewSecretService(services.repos.ProjectSecret)
	if err != nil {
		return nil, err
	}

	projectService, err := NewProjectService(services.repos.Project, projectEnvironmentVariableService, projectSecretService)
	if err != nil {
		return nil, err
	}

	containerDeploymentService, err := NewDeploymentService(services.repos.ContainerDeployment)
	if err != nil {
		return nil, err
	}

	containerEnvironmentVariableService, err := NewVariableService(services.repos.ContainerEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	containerSecretService, err := NewSecretService(services.repos.ContainerSecret)
	if err != nil {
		return nil, err
	}

	containerService, err := NewContainerService(services.repos.Container, containerDeploymentService, containerEnvironmentVariableService, containerSecretService)
	if err != nil {
		return nil, err
	}

	containerRegistryService, err := NewContainerRegistryService(services.repos.ContainerRegistry)
	if err != nil {
		return nil, err
	}

	environmentDeploymentService, err := NewDeploymentService(services.repos.EnvironmentDeployment)
	if err != nil {
		return nil, err
	}

	environmentEnvironmentVariableService, err := NewVariableService(services.repos.EnvironmentEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	environmentSecretService, err := NewSecretService(services.repos.EnvironmentSecret)
	if err != nil {
		return nil, err
	}

	environmentService, err := NewEnvironmentService(services.repos.Environment, environmentDeploymentService, environmentEnvironmentVariableService, environmentSecretService)
	if err != nil {
		return nil, err
	}

	deploymentStageService, err := NewDeploymentStageService(services.repos.DeploymentStage)
	if err != nil {
		return nil, err
	}

	services.CredentialsAws = credentialsAwsService
	services.CredentialsScaleway = credentialsScalewayService
	services.Organization = organizationService
	services.Project = projectService
	services.Container = containerService
	services.ContainerRegistry = containerRegistryService
	services.Environment = environmentService
	services.DeploymentStage = deploymentStageService

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
