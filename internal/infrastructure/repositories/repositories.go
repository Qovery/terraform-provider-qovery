package repositories

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/qoveryapi"
)

var (
	ErrFailedToInitializeQoveryAPI      = errors.New("failed to initialize qovery api")
	ErrMissingRepositoriesConfiguration = errors.New("missing repositories configuration")
)

type Configuration func(repos *Repositories) error

type Repositories struct {
	CredentialsAws                 credentials.AwsRepository
	CredentialsScaleway            credentials.ScalewayRepository
	Organization                   organization.Repository
	Project                        project.Repository
	ProjectEnvironmentVariable     variable.Repository
	ProjectSecret                  secret.Repository
	Container                      container.Repository
	ContainerDeployment            deployment.Repository
	ContainerEnvironmentVariable   variable.Repository
	ContainerSecret                secret.Repository
	ContainerRegistry              registry.Repository
	Environment                    environment.Repository
	EnvironmentDeployment          deployment.Repository
	EnvironmentEnvironmentVariable variable.Repository
	EnvironmentSecret              secret.Repository
	DeploymentStage                deploymentstage.Repository
}

func New(configs ...Configuration) (*Repositories, error) {
	if len(configs) == 0 {
		return nil, ErrMissingRepositoriesConfiguration
	}

	repos := &Repositories{}

	// Apply all the configs to the qoveryAPI instance.
	for _, config := range configs {
		if err := config(repos); err != nil {
			return nil, err
		}
	}

	return repos, nil
}

func WithQoveryAPI(apiToken string, providerVersion string) Configuration {
	return func(repos *Repositories) error {
		qoveryAPI, err := qoveryapi.New(
			qoveryapi.WithQoveryAPIToken(apiToken),
			qoveryapi.WithUserAgent(fmt.Sprintf("terraform-provider-qovery/%s", providerVersion)),
		)
		if err != nil {
			return errors.Wrap(err, ErrFailedToInitializeQoveryAPI.Error())
		}

		repos.CredentialsAws = qoveryAPI.CredentialsAws
		repos.CredentialsScaleway = qoveryAPI.CredentialsScaleway
		repos.Organization = qoveryAPI.Organization
		repos.Project = qoveryAPI.Project
		repos.ProjectEnvironmentVariable = qoveryAPI.ProjectEnvironmentVariable
		repos.ProjectSecret = qoveryAPI.ProjectSecret
		repos.Container = qoveryAPI.Container
		repos.ContainerDeployment = qoveryAPI.ContainerDeployment
		repos.ContainerEnvironmentVariable = qoveryAPI.ContainerEnvironmentVariable
		repos.ContainerSecret = qoveryAPI.ContainerSecret
		repos.ContainerRegistry = qoveryAPI.ContainerRegistry
		repos.Environment = qoveryAPI.Environment
		repos.EnvironmentDeployment = qoveryAPI.EnvironmentDeployment
		repos.EnvironmentEnvironmentVariable = qoveryAPI.EnvironmentEnvironmentVariable
		repos.EnvironmentSecret = qoveryAPI.EnvironmentSecret
		repos.DeploymentStage = qoveryAPI.DeploymentStage

		return nil
	}
}
