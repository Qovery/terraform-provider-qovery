//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/sethvargo/go-envconfig"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

var (
	testNamePrefix = "testacc"
	testNameSuffix = uuid.NewString()
)

type testEnvironment struct {
	QoveryAPIToken                string `env:"QOVERY_API_TOKEN,required"`
	OrganizationID                string `env:"TEST_ORGANIZATION_ID,required"`
	AwsCredentialsID              string `env:"TEST_AWS_CREDENTIALS_ID,required"`
	AwsCredentialsAccessKeyID     string `env:"TEST_AWS_CREDENTIALS_ACCESS_KEY_ID,required"`
	AwsCredentialsSecretAccessKey string `env:"TEST_AWS_CREDENTIALS_SECRET_ACCESS_KEY,required"`
	ScalewayCredentialsID         string `env:"TEST_SCALEWAY_CREDENTIALS_ID,required"`
	ScalewayCredentialsProjectID  string `env:"TEST_SCALEWAY_CREDENTIALS_PROJECT_ID,required"`
	ScalewayCredentialsAccessKey  string `env:"TEST_SCALEWAY_CREDENTIALS_ACCESS_KEY,required"`
	ScalewayCredentialsSecretKey  string `env:"TEST_SCALEWAY_CREDENTIALS_SECRET_KEY,required"`
	ClusterID                     string `env:"TEST_CLUSTER_ID,required"`
	ProjectID                     string `env:"TEST_PROJECT_ID,required"`
	EnvironmentID                 string `env:"TEST_ENVIRONMENT_ID,required"`
	ApplicationID                 string `env:"TEST_APPLICATION_ID,required"`
	DatabaseID                    string `env:"TEST_DATABASE_ID,required"`
	AwsEcrURL                     string `env:"TEST_AWS_ECR_URL"`
	ContainerRegistryID           string `env:"TEST_CONTAINER_REGISTRY_ID,required"`
	ContainerID                   string `env:"TEST_CONTAINER_ID,required"`
	JobID                         string `env:"TEST_JOB_ID,required"`
	QoveryHost                    string `env:"TEST_QOVERY_HOST,required"`
	QoverySandboxGitTokenId       string `env:"TEST_QOVERY_SANDBOX_GIT_TOKEN_ID,required"`
}

var (
	apiClient         = client.New(os.Getenv(qovery.APITokenEnvName), "test", getTestQoveryHost())
	qoveryServices, _ = services.New(services.WithQoveryRepository(os.Getenv(qovery.APITokenEnvName), "test", getTestQoveryHost()))
	qoveryAPIClient   = client.NewQoveryAPIClient(os.Getenv(qovery.APITokenEnvName), "test", getTestQoveryHost())
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"qovery": providerserver.NewProtocol6WithError(qovery.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	var e testEnvironment
	if err := envconfig.Process(context.Background(), &e); err != nil {
		t.Fatalf("failed to parse environment variables: %s", err)
	}
}

func getTestOrganizationID() string {
	return os.Getenv("TEST_ORGANIZATION_ID")
}

func getTestAWSCredentialsID() string {
	return os.Getenv("TEST_AWS_CREDENTIALS_ID")
}

func getTestAWSCredentialsAccessKeyID() string {
	return os.Getenv("TEST_AWS_CREDENTIALS_ACCESS_KEY_ID")
}

func getTestAWSCredentialsSecretAccessKey() string {
	return os.Getenv("TEST_AWS_CREDENTIALS_SECRET_ACCESS_KEY")
}

func getTestScalewayCredentialsID() string {
	return os.Getenv("TEST_SCALEWAY_CREDENTIALS_ID")
}

func getTestScalewayCredentialsProjectID() string {
	return os.Getenv("TEST_SCALEWAY_CREDENTIALS_PROJECT_ID")
}

func getTestScalewayCredentialsAccessKey() string {
	return os.Getenv("TEST_SCALEWAY_CREDENTIALS_ACCESS_KEY")
}

func getTestScalewayCredentialsSecretKey() string {
	return os.Getenv("TEST_SCALEWAY_CREDENTIALS_SECRET_KEY")
}

func getTestClusterID() string {
	return os.Getenv("TEST_CLUSTER_ID")
}

func getTestProjectID() string {
	return os.Getenv("TEST_PROJECT_ID")
}

func getTestEnvironmentID() string {
	return os.Getenv("TEST_ENVIRONMENT_ID")
}

func getTestApplicationID() string {
	return os.Getenv("TEST_APPLICATION_ID")
}

func getTestDatabaseID() string {
	return os.Getenv("TEST_DATABASE_ID")
}

func getTestAwsEcrURL() string {
	return os.Getenv("TEST_AWS_ECR_URL")
}

func getTestContainerRegistryID() string {
	return os.Getenv("TEST_CONTAINER_REGISTRY_ID")
}

func getTestContainerID() string {
	return os.Getenv("TEST_CONTAINER_ID")
}

func getTestJobID() string {
	return os.Getenv("TEST_JOB_ID")
}

func getTestQoveryHost() string {
	return os.Getenv("TEST_QOVERY_HOST")
}

func getTestQoverySandboxGitTokenID() string {
	return os.Getenv("TEST_QOVERY_SANDBOX_GIT_TOKEN_ID")
}

func generateTestName(testName string) string {
	return fmt.Sprintf("%s-%s-%s", testNamePrefix, testName, testNameSuffix)
}

func generateRandomName(testName string) string {
	return fmt.Sprintf("%s-%s-%s", testNamePrefix, testName, uuid.NewString())
}

func isCIMainBranch() bool {
	return os.Getenv("CI") == "true" && os.Getenv("GITHUB_REF_NAME") == "main"
}

func skipInCIUnlessMainBranch(t *testing.T) {
	if !isCIMainBranch() {
		t.SkipNow()
	}
}
