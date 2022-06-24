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

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

var (
	testNamePrefix = "testacc"
	testNameSuffix = uuid.NewString()
)

type testEnvironment struct {
	QoveryApiToken                string `env:"QOVERY_API_TOKEN,required"`
	OrganizationID                string `env:"TEST_ORGANIZATION_ID,required"`
	AwsCredentialsID              string `env:"TEST_AWS_CREDENTIALS_ID,required"`
	AwsCredentialsAccessKeyID     string `env:"TEST_AWS_CREDENTIALS_ACCESS_KEY_ID,required"`
	AwsCredentialsSecretAccessKey string `env:"TEST_AWS_CREDENTIALS_SECRET_ACCESS_KEY,required"`
	ClusterID                     string `env:"TEST_CLUSTER_ID,required"`
	ProjectID                     string `env:"TEST_PROJECT_ID,required"`
	EnvironmentID                 string `env:"TEST_ENVIRONMENT_ID,required"`
	ApplicationID                 string `env:"TEST_APPLICATION_ID,required"`
	DatabaseID                    string `env:"TEST_DATABASE_ID,required"`
}

var apiClient = client.New(os.Getenv(qovery.APITokenEnvName), "test")

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

func generateTestName(testName string) string {
	return fmt.Sprintf("%s-%s-%s", testNamePrefix, testName, testNameSuffix)
}

func isCIMainBranch() bool {
	return os.Getenv("CI") == "true" && os.Getenv("GITHUB_REF_NAME") == "main"
}

func skipInCIUnlessMainBranch(t *testing.T) {
	if !isCIMainBranch() {
		t.SkipNow()
	}
}
