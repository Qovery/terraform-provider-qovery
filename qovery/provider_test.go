package qovery_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/sethvargo/go-envconfig"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

const (
	testResourcePrefix = "q-test-42-terraform-acc"
)

type testEnvironment struct {
	QoveryApiToken  string `env:"QOVERY_API_TOKEN,required"`
	OrganizationID  string `env:"TEST_ORGANIZATION_ID,required"`
	AccessKeyID     string `env:"TEST_ACCESS_KEY_ID,required"`
	SecretAccessKey string `env:"TEST_SECRET_ACCESS_KEY,required"`
}

var apiClient = client.New(os.Getenv(qovery.APITokenEnvName), "test")

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"qovery": func() (tfprotov6.ProviderServer, error) {
		return tfsdk.NewProtocol6Server(qovery.New("test")()), nil
	},
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

func getTestAccessKeyID() string {
	return os.Getenv("TEST_ACCESS_KEY_ID")
}

func getTestSecretAccessKey() string {
	return os.Getenv("TEST_SECRET_ACCESS_KEY")
}
