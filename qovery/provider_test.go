package qovery_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

var apiClient = client.New(os.Getenv(qovery.APITokenEnvName), "test")

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"qovery": func() (tfprotov6.ProviderServer, error) {
		return tfsdk.NewProtocol6Server(qovery.New("test")()), nil
	},
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv(qovery.APITokenEnvName); v == "" {
		t.Fatalf("%s must be set for acceptance tests", qovery.APITokenEnvName)
	}
}
