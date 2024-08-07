package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/qovery/qovery-client-go"
	"github.com/schollz/progressbar/v3"
	"github.com/sethvargo/go-envconfig"
)

type environment struct {
	QoveryAPIToken     string `env:"QOVERY_API_TOKEN" validate:"required"`
	TestOrganizationID string `env:"TEST_ORGANIZATION_ID" validate:"required"`
}

type project struct {
	ID   string
	Name string
}

type credentials struct {
	ID   string
	Name string
}

type registry struct {
	ID   string
	Name string
}

const testPrefix = "testacc"

func main() {
	var env environment

	ctx := context.Background()
	if err := envconfig.Process(ctx, &env); err != nil {
		log.Fatalf("failed to parse environment variables: %s", err)
	}
	if err := validator.New().Struct(env); err != nil {
		log.Fatalf(err.Error())
	}

	var apiClient = newQoveryAPIClient(env.QoveryAPIToken)

	if err := cleanAwsCredentials(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf(err.Error())
	}

	if err := cleanScalewayCredentials(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf(err.Error())
	}

	if err := cleanProjects(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf(err.Error())
	}

	if err := cleanContainerRegistry(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf(err.Error())
	}
}

func newQoveryAPIClient(apiToken string) *qovery.APIClient {
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", apiToken))
	cfg.AddDefaultHeader("content-type", "application/json")

	cfg.UserAgent = fmt.Sprintf("terraform-provider-qovery/%s", "test-acc")

	return qovery.NewAPIClient(cfg)
}

func cleanProjects(ctx context.Context, apiClient *qovery.APIClient, organizationID string) error {
	projects, err := getProjectsToDelete(ctx, apiClient, organizationID)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(projects)))
	fmt.Printf("Deleting %d projects...\n", len(projects))
	for _, project := range projects {
		if strings.Contains(project.Name, testPrefix) {
			bar.Describe(fmt.Sprintf("%s...", project.Name[0:50]))

			_, err := apiClient.ProjectMainCallsAPI.
				DeleteProject(ctx, project.ID).
				Execute()
			if err != nil {
				return err
			}

			bar.Add(1)
		}
	}

	return nil
}

func cleanAwsCredentials(ctx context.Context, apiClient *qovery.APIClient, organizationID string) error {
	awsCreds, err := getAwsCredentialsToDelete(ctx, apiClient, organizationID)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(awsCreds)))
	fmt.Printf("Deleting %d aws credentials...\n", len(awsCreds))
	for _, creds := range awsCreds {
		if strings.Contains(creds.Name, testPrefix) {
			bar.Describe(fmt.Sprintf("%s...", creds.Name[0:50]))

			_, err := apiClient.CloudProviderCredentialsAPI.
				DeleteAWSCredentials(ctx, creds.ID, organizationID).
				Execute()
			if err != nil {
				return err
			}

			bar.Add(1)
		}
	}

	return nil
}

func cleanScalewayCredentials(ctx context.Context, apiClient *qovery.APIClient, organizationID string) error {
	scalewayCreds, err := getScalewayCredentialsToDelete(ctx, apiClient, organizationID)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(scalewayCreds)))
	fmt.Printf("Deleting %d scaleway credentials...\n", len(scalewayCreds))
	for _, creds := range scalewayCreds {
		if strings.Contains(creds.Name, testPrefix) {
			bar.Describe(fmt.Sprintf("%s...", creds.Name[0:50]))

			_, err := apiClient.CloudProviderCredentialsAPI.
				DeleteScalewayCredentials(ctx, creds.ID, organizationID).
				Execute()
			if err != nil {
				return err
			}

			bar.Add(1)
		}
	}

	return nil
}

func cleanContainerRegistry(ctx context.Context, apiClient *qovery.APIClient, organizationID string) error {
	registries, err := getContainerRegitriesToDelete(ctx, apiClient, organizationID)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(registries)))
	fmt.Printf("Deleting %d container registries...\n", len(registries))
	for _, reg := range registries {
		if strings.Contains(reg.Name, testPrefix) {
			maxSize := len(reg.Name)
			if maxSize > 50 {
				maxSize = 50
			}
			bar.Describe(fmt.Sprintf("%s...", reg.Name[0:maxSize]))

			_, err := apiClient.ContainerRegistriesAPI.
				DeleteContainerRegistry(ctx, organizationID, reg.ID).
				Execute()
			if err != nil {
				return err
			}

			bar.Add(1)
		}
	}

	return nil
}

func getProjectsToDelete(ctx context.Context, apiClient *qovery.APIClient, organizationID string) ([]project, error) {
	projects, _, err := apiClient.ProjectsAPI.
		ListProject(ctx, organizationID).
		Execute()
	if err != nil {
		return nil, err
	}

	projectsToDelete := make([]project, 0, len(projects.GetResults()))
	for _, p := range projects.GetResults() {
		if strings.Contains(p.Name, testPrefix) {
			projectsToDelete = append(projectsToDelete, project{
				ID:   p.Id,
				Name: p.Name,
			})
		}
	}

	return projectsToDelete, nil
}

func getAwsCredentialsToDelete(ctx context.Context, apiClient *qovery.APIClient, organizationID string) ([]credentials, error) {
	awsCreds, _, err := apiClient.CloudProviderCredentialsAPI.
		ListAWSCredentials(ctx, organizationID).
		Execute()
	if err != nil {
		return nil, err
	}

	awsCredsToDelete := make([]credentials, 0, len(awsCreds.GetResults()))
	for _, c := range awsCreds.GetResults() {
		credsName := strings.ToLower(c.AwsClusterCredentials.GetName())
		if strings.Contains(credsName, testPrefix) {
			awsCredsToDelete = append(awsCredsToDelete, credentials{
				ID:   c.AwsClusterCredentials.GetId(),
				Name: c.AwsClusterCredentials.GetName(),
			})
		}
	}

	return awsCredsToDelete, nil
}

func getScalewayCredentialsToDelete(ctx context.Context, apiClient *qovery.APIClient, organizationID string) ([]credentials, error) {
	scalewayCreds, _, err := apiClient.CloudProviderCredentialsAPI.
		ListScalewayCredentials(ctx, organizationID).
		Execute()
	if err != nil {
		return nil, err
	}

	scalewayCredsToDelete := make([]credentials, 0, len(scalewayCreds.GetResults()))
	for _, c := range scalewayCreds.GetResults() {
		credsName := strings.ToLower(c.ScalewayClusterCredentials.GetName())
		if strings.Contains(credsName, testPrefix) {
			scalewayCredsToDelete = append(scalewayCredsToDelete, credentials{
				ID:   c.ScalewayClusterCredentials.GetId(),
				Name: c.ScalewayClusterCredentials.GetName(),
			})
		}
	}

	return scalewayCredsToDelete, nil
}

func getContainerRegitriesToDelete(ctx context.Context, apiClient *qovery.APIClient, organizationID string) ([]registry, error) {
	registries, _, err := apiClient.ContainerRegistriesAPI.
		ListContainerRegistry(ctx, organizationID).
		Execute()
	if err != nil {
		return nil, err
	}

	registriesToDelete := make([]registry, 0, len(registries.GetResults()))
	for _, c := range registries.GetResults() {
		regName := strings.ToLower(c.GetName())
		if strings.Contains(regName, testPrefix) {
			registriesToDelete = append(registriesToDelete, registry{
				ID:   c.GetId(),
				Name: c.GetName(),
			})
		}
	}

	return registriesToDelete, nil
}
