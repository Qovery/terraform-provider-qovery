package qovery

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/qovery/qovery-client-go"
	"os"
	"terraform-provider-qovery/client"
	"testing"
)

func testAccPreCheck(t *testing.T) {
	token, isOk := os.LookupEnv("API_TOKEN")
	if !isOk || token == "" {
		t.Fatal("You must set API_TOKEN environment variable")
	}

	orgId, isOk := os.LookupEnv("ORG_ID")
	if !isOk || orgId == "" {
		t.Fatal("You must set ORG_ID environment variable")
	}
}

var testAccBaseConfig = fmt.Sprintf(
	`terraform {
		required_providers {
			qovery = {
				source  = "qovery.com/api/qovery"
			}
		}
	}

	provider "qovery" {
		token = "%s"
	}
`, os.Getenv("API_TOKEN"))

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"qovery": func() (tfprotov6.ProviderServer, error) {
		return tfsdk.NewProtocol6Server(&provider{version: "dev", client: testClient(os.Getenv("API_TOKEN"), "dev")}), nil
	},
}

func testClient(token, version string) *client.Client {
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", token))
	cfg.AddDefaultHeader("content-type", "application/json")

	cfg.UserAgent = fmt.Sprintf("terraform-provider-qovery/%s", version)

	return &client.Client{
		API: qovery.NewAPIClient(cfg),
	}
}
