package services

import (
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories"
)

var (
	// ErrInvalidRepository is the error return if the given repository is nil or invalid.
	ErrInvalidRepository = errors.New("invalid repository")
	// ErrInvalidService is the error return if the given service is nil or invalid.
	ErrInvalidService = errors.New("invalid service")
	// ErrMissingConfiguration is the error return if no conf gittokiguration has been given.
	ErrMissingConfiguration = errors.New("missing configuration")
)

// Services contains the implementations of domain services using.
type Services struct {
	repos *repositories.Repositories

	CredentialsAws               credentials.AwsService
	CredentialsScaleway          credentials.ScalewayService
	CredentialsGcp               credentials.GcpService
	Organization                 organization.Service
	Project                      project.Service
	Container                    container.Service
	Job                          job.Service
	ContainerRegistry            registry.Service
	Environment                  environment.Service
	DeploymentStage              deploymentstage.Service
	Deployment                   newdeployment.Service
	GitToken                     gittoken.Service
	Helm                         helm.Service
	HelmRepository               helmRepository.Service
	AnnotationsGroup             annotations_group.Service
	LabelsGroup                  labels_group.Service
	DeploymentRestrictionService deploymentrestriction.DeploymentRestrictionService
	TerraformService             terraformservice.Service
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

	credentialsGcpService, err := NewCredentialsGcpService(services.repos.CredentialsGcp)
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

	jobDeploymentService, err := NewDeploymentService(services.repos.JobDeployment)
	if err != nil {
		return nil, err
	}

	jobEnvironmentVariableService, err := NewVariableService(services.repos.JobEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	jobSecretService, err := NewSecretService(services.repos.JobSecret)
	if err != nil {
		return nil, err
	}

	deploymentRestrictionService, err := deploymentrestriction.NewDeploymentRestrictionService(*services.repos.QoveryClient)
	if err != nil {
		return nil, err
	}

	jobService, err := NewJobService(services.repos.Job, jobDeploymentService, jobEnvironmentVariableService, jobSecretService, deploymentRestrictionService)
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

	deploymentService, err := NewNewDeploymentService(services.repos.DeploymentEnvironment, services.repos.DeploymentStatus)
	if err != nil {
		return nil, err
	}

	gitTokenService, err := NewGitTokenService(services.repos.QoveryClient)
	if err != nil {
		return nil, err
	}

	helmDeploymentService, err := NewDeploymentService(services.repos.HelmDeployment)
	if err != nil {
		return nil, err
	}

	helmEnvironmentVariableService, err := NewVariableService(services.repos.HelmEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	helmSecretService, err := NewSecretService(services.repos.HelmSecret)
	if err != nil {
		return nil, err
	}

	helmService, err := NewHelmService(services.repos.Helm, helmDeploymentService, helmEnvironmentVariableService, helmSecretService, deploymentRestrictionService)
	if err != nil {
		return nil, err
	}

	helmRepositoryService, err := NewHelmRepositoryService(services.repos.HelmRepository)
	if err != nil {
		return nil, err
	}

	annotationsGroupService, err := NewAnnotationsGroupService(services.repos.AnnotationsGroupRepository)
	if err != nil {
		return nil, err
	}

	labelsGroupService, err := NewLabelsGroupService(services.repos.LabelsGroupRepository)
	if err != nil {
		return nil, err
	}

	terraformServiceService, err := NewTerraformServiceService(services.repos.TerraformService)
	if err != nil {
		return nil, err
	}

	services.CredentialsAws = credentialsAwsService
	services.CredentialsScaleway = credentialsScalewayService
	services.CredentialsGcp = credentialsGcpService
	services.Organization = organizationService
	services.Project = projectService
	services.Container = containerService
	services.Job = jobService
	services.ContainerRegistry = containerRegistryService
	services.Environment = environmentService
	services.DeploymentStage = deploymentStageService
	services.Deployment = deploymentService
	services.GitToken = gitTokenService
	services.Helm = helmService
	services.HelmRepository = helmRepositoryService
	services.DeploymentRestrictionService = deploymentRestrictionService
	services.AnnotationsGroup = annotationsGroupService
	services.LabelsGroup = labelsGroupService
	services.TerraformService = terraformServiceService

	return services, nil
}

func WithQoveryRepository(apiToken string, providerVersion string, host string) Configuration {
	return func(services *Services) error {
		repos, err := repositories.New(repositories.WithQoveryAPI(apiToken, providerVersion, host))
		if err != nil {
			return err
		}

		services.repos = repos

		return nil
	}
}
