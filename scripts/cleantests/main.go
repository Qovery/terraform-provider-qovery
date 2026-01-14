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

type annotationsGroup struct {
	ID   string
	Name string
}

type labelsGroup struct {
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
		log.Fatalf("failed to validate environment variables: %v", err)
	}

	var apiClient = newQoveryAPIClient(env.QoveryAPIToken)

	if err := cleanAwsCredentials(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf("failed to clean AWS credentials: %v", err)
	}

	if err := cleanScalewayCredentials(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf("failed to clean Scaleway credentials: %v", err)
	}

	if err := cleanGcpCredentials(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf("failed to clean GCP credentials: %v", err)
	}

	if err := cleanProjects(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf("failed to clean projects: %v", err)
	}

	if err := cleanContainerRegistry(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf("failed to clean container registries: %v", err)
	}

	if err := cleanAnnotationsGroups(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf("failed to clean annotations groups: %v", err)
	}

	if err := cleanLabelsGroups(ctx, apiClient, env.TestOrganizationID); err != nil {
		log.Fatalf("failed to clean labels groups: %v", err)
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
			maxSize := len(project.Name)
			if maxSize > 50 {
				maxSize = 50
			}
			bar.Describe(fmt.Sprintf("%s...", project.Name[0:maxSize]))

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
			maxSize := len(creds.Name)
			if maxSize > 50 {
				maxSize = 50
			}
			bar.Describe(fmt.Sprintf("%s...", creds.Name[0:maxSize]))

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
			maxSize := len(creds.Name)
			if maxSize > 50 {
				maxSize = 50
			}
			bar.Describe(fmt.Sprintf("%s...", creds.Name[0:maxSize]))

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

func cleanGcpCredentials(ctx context.Context, apiClient *qovery.APIClient, organizationID string) error {
	gcpCreds, err := getGcpCredentialsToDelete(ctx, apiClient, organizationID)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(gcpCreds)))
	fmt.Printf("Deleting %d gcp credentials...\n", len(gcpCreds))
	for _, creds := range gcpCreds {
		if strings.Contains(creds.Name, testPrefix) {
			maxSize := len(creds.Name)
			if maxSize > 50 {
				maxSize = 50
			}
			bar.Describe(fmt.Sprintf("%s...", creds.Name[0:maxSize]))

			_, err := apiClient.CloudProviderCredentialsAPI.
				DeleteGcpCredentials(ctx, creds.ID, organizationID).
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
		var name string
		var id string
		if c.AwsStaticClusterCredentials != nil {
			name = strings.ToLower(c.AwsStaticClusterCredentials.GetName())
			id = c.AwsStaticClusterCredentials.GetId()
		}
		if c.AwsRoleClusterCredentials != nil {
			name = strings.ToLower(c.AwsRoleClusterCredentials.GetName())
			id = c.AwsRoleClusterCredentials.GetId()
		}

		if strings.Contains(name, testPrefix) {
			awsCredsToDelete = append(awsCredsToDelete, credentials{
				ID:   id,
				Name: name,
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

func getGcpCredentialsToDelete(ctx context.Context, apiClient *qovery.APIClient, organizationID string) ([]credentials, error) {
	gcpCreds, _, err := apiClient.CloudProviderCredentialsAPI.
		ListGcpCredentials(ctx, organizationID).
		Execute()
	if err != nil {
		return nil, err
	}

	gcpCredsToDelete := make([]credentials, 0, len(gcpCreds.GetResults()))
	for _, c := range gcpCreds.GetResults() {
		credsName := strings.ToLower(c.GcpStaticClusterCredentials.GetName())
		if strings.Contains(credsName, testPrefix) {
			gcpCredsToDelete = append(gcpCredsToDelete, credentials{
				ID:   c.GcpStaticClusterCredentials.GetId(),
				Name: c.GcpStaticClusterCredentials.GetName(),
			})
		}
	}

	return gcpCredsToDelete, nil
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

func cleanAnnotationsGroups(ctx context.Context, apiClient *qovery.APIClient, organizationID string) error {
	annotationsGroups, err := getAnnotationsGroupsToDelete(ctx, apiClient, organizationID)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(annotationsGroups)))
	fmt.Printf("Deleting %d annotations groups...\n", len(annotationsGroups))
	for _, group := range annotationsGroups {
		if strings.Contains(group.Name, testPrefix) {
			maxSize := len(group.Name)
			if maxSize > 50 {
				maxSize = 50
			}
			bar.Describe(fmt.Sprintf("%s...", group.Name[0:maxSize]))

			_, err := apiClient.OrganizationAnnotationsGroupAPI.
				DeleteOrganizationAnnotationsGroup(ctx, organizationID, group.ID).
				Execute()
			if err != nil {
				return err
			}

			bar.Add(1)
		}
	}

	return nil
}

func cleanLabelsGroups(ctx context.Context, apiClient *qovery.APIClient, organizationID string) error {
	labelsGroups, err := getLabelsGroupsToDelete(ctx, apiClient, organizationID)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(labelsGroups)))
	fmt.Printf("Deleting %d labels groups...\n", len(labelsGroups))
	for _, group := range labelsGroups {
		if strings.Contains(group.Name, testPrefix) {
			maxSize := len(group.Name)
			if maxSize > 50 {
				maxSize = 50
			}
			bar.Describe(fmt.Sprintf("%s...", group.Name[0:maxSize]))

			_, err := apiClient.OrganizationLabelsGroupAPI.
				DeleteOrganizationLabelsGroup(ctx, organizationID, group.ID).
				Execute()
			if err != nil {
				return err
			}

			bar.Add(1)
		}
	}

	return nil
}

func getAnnotationsGroupsToDelete(ctx context.Context, apiClient *qovery.APIClient, organizationID string) ([]annotationsGroup, error) {
	groups, _, err := apiClient.OrganizationAnnotationsGroupAPI.
		ListOrganizationAnnotationsGroup(ctx, organizationID).
		Execute()
	if err != nil {
		return nil, err
	}

	groupsToDelete := make([]annotationsGroup, 0, len(groups.GetResults()))
	for _, g := range groups.GetResults() {
		groupName := strings.ToLower(g.GetName())
		if strings.Contains(groupName, testPrefix) {
			groupsToDelete = append(groupsToDelete, annotationsGroup{
				ID:   g.GetId(),
				Name: g.GetName(),
			})
		}
	}

	return groupsToDelete, nil
}

func getLabelsGroupsToDelete(ctx context.Context, apiClient *qovery.APIClient, organizationID string) ([]labelsGroup, error) {
	groups, _, err := apiClient.OrganizationLabelsGroupAPI.
		ListOrganizationLabelsGroup(ctx, organizationID).
		Execute()
	if err != nil {
		return nil, err
	}

	groupsToDelete := make([]labelsGroup, 0, len(groups.GetResults()))
	for _, g := range groups.GetResults() {
		groupName := strings.ToLower(g.GetName())
		if strings.Contains(groupName, testPrefix) {
			groupsToDelete = append(groupsToDelete, labelsGroup{
				ID:   g.GetId(),
				Name: g.GetName(),
			})
		}
	}

	return groupsToDelete, nil
}
