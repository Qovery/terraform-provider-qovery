package qoveryapi

import (
	"fmt"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
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
	// ErrInvalidHost is returned when the host is invalid.
	ErrInvalidHost = errors.New("invalid-host")
)

// Configuration represents a function that handle the QoveryAPI configuration.
type Configuration func(qoveryAPI *QoveryAPI) error

// QoveryAPI contains the implementations of domain repositories using the qovery api client.
type QoveryAPI struct {
	Client *qovery.APIClient

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
	Job                            job.Repository
	JobDeployment                  deployment.Repository
	JobEnvironmentVariable         variable.Repository
	JobSecret                      secret.Repository
	Environment                    environment.Repository
	EnvironmentDeployment          deployment.Repository
	EnvironmentEnvironmentVariable variable.Repository
	EnvironmentSecret              secret.Repository
	DeploymentStage                deploymentstage.Repository
	DeploymentEnvironment          newdeployment.EnvironmentRepository
	DeploymentStatus               newdeployment.DeploymentStatusRepository
	Helm                           helm.Repository
	HelmDeployment                 deployment.Repository
	HelmEnvironmentVariable        variable.Repository
	HelmSecret                     secret.Repository
	HelmRepository                 helmRepository.Repository
	AnnotationsGroup               annotations_group.Repository
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

	jobAPI, err := newJobQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	jobDeploymentAPI, err := newJobDeploymentQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	jobEnvironmentVariableAPI, err := newJobEnvironmentVariablesQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	jobSecretAPI, err := newJobSecretsQoveryAPI(apiClient)
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

	deploymentEnvironmentAPI, err := newDeploymentEnvironmentQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	deploymentStatusAPI, err := newDeploymentStatusQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	helmAPI, err := newHelmQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	helmDeploymentAPI, err := newHelmDeploymentQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	helmEnvironmentVariableAPI, err := newHelmEnvironmentVariablesQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	helmSecretAPI, err := newHelmSecretsQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	helmRepositoryAPI, err := newHelmRepositoryQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	annotationsGroupAPI, err := newAnnotationsGroupQoveryAPI(apiClient)
	if err != nil {
		return nil, err
	}

	// Create a new QoveryAPI instance.
	qoveryAPI := &QoveryAPI{
		Client:                         apiClient,
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
		Job:                            jobAPI,
		JobDeployment:                  jobDeploymentAPI,
		JobEnvironmentVariable:         jobEnvironmentVariableAPI,
		JobSecret:                      jobSecretAPI,
		Environment:                    environmentAPI,
		EnvironmentDeployment:          environmentDeploymentAPI,
		EnvironmentEnvironmentVariable: environmentEnvironmentVariableAPI,
		EnvironmentSecret:              environmentSecretAPI,
		DeploymentStage:                deploymentStageAPI,
		DeploymentEnvironment:          deploymentEnvironmentAPI,
		DeploymentStatus:               deploymentStatusAPI,
		Helm:                           helmAPI,
		HelmDeployment:                 helmDeploymentAPI,
		HelmEnvironmentVariable:        helmEnvironmentVariableAPI,
		HelmSecret:                     helmSecretAPI,
		HelmRepository:                 helmRepositoryAPI,
		AnnotationsGroup:               annotationsGroupAPI,
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

		qoveryAPI.Client.GetConfig().AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", apiToken))

		return nil
	}
}

// WithUserAgent sets the user-agent of the api client with the provider version.
func WithUserAgent(userAgent string) Configuration {
	return func(qoveryAPI *QoveryAPI) error {
		if userAgent == "" {
			return ErrInvalidUserAgent
		}

		qoveryAPI.Client.GetConfig().UserAgent = userAgent

		return nil
	}
}

// WithServerHost sets the user-agent of the api client with the provider version.
func WithServerHost(host string) Configuration {
	return func(qoveryAPI *QoveryAPI) error {
		if host == "" {
			return ErrInvalidHost
		}

		qoveryAPI.Client.GetConfig().Servers = qovery.ServerConfigurations{
			{
				URL:         host,
				Description: "No description provided",
			},
		}

		return nil
	}
}
