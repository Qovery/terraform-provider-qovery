//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_DeploymentDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(),
		Steps: []resource.TestStep{
			// Create a deployment first, then read it as a data source
			{
				Config: testAccDeploymentDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.qovery_deployment.test", "id"),
					resource.TestCheckResourceAttr("data.qovery_deployment.test", "desired_state", "STOPPED"),
				),
			},
		},
	})
}

func testAccDeploymentDataSourceConfig() string {
	return fmt.Sprintf(`
# Environment + Project
%s

# Needed by container resource
resource "qovery_container_registry" "ds_deployment_registry_test" {
  organization_id = "%s"
  name = "%s"
  kind = "ECR"
  url = "%s"
  config = {
    region = "eu-west-3"
    access_key_id = "%s"
    secret_access_key = "%s"
  }
}

resource "qovery_deployment_stage" "ds_deployment_stage" {
  environment_id = qovery_environment.test.id
  name        = "%s"
}

resource "qovery_container" "ds_container_test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.ds_deployment_registry_test.id
  name = "%s"
  image_name = "terraform-provider-tests-container"
  tag = "1.0.0"
  deployment_stage_id = qovery_deployment_stage.ds_deployment_stage.id
  healthchecks = {}
}

resource "qovery_deployment" "source" {
  environment_id = qovery_environment.test.id
  desired_state  = "STOPPED"

  depends_on = [
    qovery_container.ds_container_test
  ]
}

data "qovery_deployment" "test" {
  id = qovery_deployment.source.id
}
`,
		testAccEnvironmentDefaultConfig("ds-deployment-environment"),
		getTestOrganizationID(),
		generateTestName("ds-deployment-container-registry"),
		getTestAwsEcrURL(),
		getTestAWSCredentialsAccessKeyID(),
		getTestAWSCredentialsSecretAccessKey(),
		generateRandomName("ds-deploymentstage"),
		generateTestName("ds-container"),
	)
}
