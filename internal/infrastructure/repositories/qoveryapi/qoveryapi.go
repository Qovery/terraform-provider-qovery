package qoveryapi

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

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
)

var (
	// ErrInvalidQoveryAPIClient is returned when the qovery client is invalid or nil.
	ErrInvalidQoveryAPIClient = errors.New("invalid qovery api client")
	// ErrInvalidQoveryAPIToken is returned when the qovery api token is invalid.
	ErrInvalidQoveryAPIToken = errors.New("invalid qovery api token")
	// ErrInvalidUserAgent is returned when the user-agent is invalid.
	ErrInvalidUserAgent = errors.New("invalid user-agent")
)

// Configuration represents a function that handle the QoveryAPI configuration.
type Configuration func(qoveryAPI *QoveryAPI) error

// QoveryAPI contains the implementations of domain repositories using the qovery api client.
type QoveryAPI struct {
	client *qovery.APIClient

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

// New returns a new instance of QoveryAPI and applies the given configs.
func New(configs ...Configuration) (*QoveryAPI, error) {
	// Initialize the qovery api client.
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("content-type", "application/json")
	apiClient := qovery.NewAPIClient(cfg)

	// Initialize repositories implementations.
	credentialsAwsAPI, err := newCredentialsAwsQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	credentialsScalewayAPI, err := newCredentialsScalewayQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	organizationAPI, err := newOrganizationQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	projectAPI, err := newProjectQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	projectEnvironmentVariableAPI, err := newProjectEnvironmentVariablesQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	projectSecretAPI, err := newProjectSecretsQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	containerAPI, err := newContainerQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	containerDeploymentAPI, err := newContainerDeploymentQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	containerEnvironmentVariableAPI, err := newContainerEnvironmentVariablesQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	containerSecretAPI, err := newContainerSecretsQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	containerRegistryAPI, err := newContainerRegistryQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	environmentAPI, err := newEnvironmentQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	environmentDeploymentAPI, err := newEnvironmentDeploymentQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	environmentEnvironmentVariableAPI, err := newEnvironmentEnvironmentVariablesQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	environmentSecretAPI, err := newEnvironmentSecretsQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	deploymentStageAPI, err := newDeploymentStageQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	// Create a new QoveryAPI instance.
	qoveryAPI := &QoveryAPI{
		client:                         apiClient,
		CredentialsAws:                 credentialsAwsAPI,
		CredentialsScaleway:            credentialsScalewayAPI,
		Organization:                   organizationAPI,
		Project:                        projectAPI,
		ProjectEnvironmentVariable:     projectEnvironmentVariableAPI,
		ProjectSecret:                  projectSecretAPI,
		Container:                      containerAPI,
		ContainerDeployment:            containerDeploymentAPI,
		ContainerEnvironmentVariable:   containerEnvironmentVariableAPI,
		ContainerSecret:                containerSecretAPI,
		ContainerRegistry:              containerRegistryAPI,
		Environment:                    environmentAPI,
		EnvironmentDeployment:          environmentDeploymentAPI,
		EnvironmentEnvironmentVariable: environmentEnvironmentVariableAPI,
		EnvironmentSecret:              environmentSecretAPI,
		DeploymentStage:                deploymentStageAPI,
	}

	// Apply all the configs to the qoveryAPI instance.
	for _, config := range configs {
		if err := config(qoveryAPI); err != nil {
			return nil, err
		}
	}

	return qoveryAPI, nil
}

// WithQoveryAPIToken sets the api token on the qovery api client.
func WithQoveryAPIToken(apiToken string) Configuration {
	return func(qoveryAPI *QoveryAPI) error {
		if apiToken == "" {
			return ErrInvalidQoveryAPIToken
		}

		qoveryAPI.client.GetConfig().AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", apiToken))

		return nil
	}
}

// WithUserAgent sets the user-agent of the api client with the provider version.
func WithUserAgent(userAgent string) Configuration {
	return func(qoveryAPI *QoveryAPI) error {
		if userAgent == "" {
			return ErrInvalidUserAgent
		}

		qoveryAPI.client.GetConfig().UserAgent = userAgent

		return nil
	}
}
