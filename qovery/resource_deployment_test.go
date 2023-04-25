//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

func TestAcc_Deployment(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(),
		Steps: []resource.TestStep{
			// Create services and deployment
			{
				Config: testAccDeploymentDefaultConfigWithDesiredState("STOPPED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationExists("qovery_application.application_test"),
					testAccQoveryContainerExists("qovery_container.container_test"),
					testAccQoveryDatabaseExists("qovery_database.database_test"),
					resource.TestCheckResourceAttr("qovery_deployment.deployment_test", "desired_state", "STOPPED"),
				),
			},
			// Apply deployment with RUNNING state
			{
				Config: testAccDeploymentDefaultConfigWithDesiredState("RUNNING"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryApplicationHasState("DEPLOYED"),
					testAccQoveryContainerHasState("DEPLOYED"),
					testAccQoveryDatabaseHasState("DEPLOYED"),
				),
			},
		},
	})
}

func testAccDeploymentDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceName := "qovery_environment.test"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("environment resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("environment.id not found")
		}

		_, err := qoveryServices.Environment.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found environment but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted environment: %s", err.Error())
		}
		return nil
	}
}

func testAccQoveryApplicationHasState(expectedState qovery.StateEnum) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceName := "qovery_application.application_test"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("application resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("application.id not found")
		}

		applicationStatus, response, err := qoveryApiClient.ApplicationMainCallsApi.GetApplicationStatus(context.TODO(), rs.Primary.ID).Execute()
		if err != nil || response.StatusCode >= 400 {
			return fmt.Errorf("Cannot find application status")
		}
		if applicationStatus.State != expectedState {
			return fmt.Errorf("Expected application status %s, got %s", expectedState, applicationStatus.State)
		}

		return nil
	}
}

func testAccQoveryContainerHasState(expectedState qovery.StateEnum) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceName := "qovery_container.container_test"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("container resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("container.id not found")
		}

		containerStatus, response, err := qoveryApiClient.ContainerMainCallsApi.GetContainerStatus(context.TODO(), rs.Primary.ID).Execute()
		if err != nil || response.StatusCode >= 400 {
			return fmt.Errorf("Cannot find container status")
		}
		if containerStatus.State != expectedState {
			return fmt.Errorf("Expected container status %s, got %s", expectedState, containerStatus.State)
		}

		return nil
	}
}

func testAccQoveryDatabaseHasState(expectedState qovery.StateEnum) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceName := "qovery_database.database_test"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("database resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("database.id not found")
		}

		databaseStatus, response, err := qoveryApiClient.DatabaseMainCallsApi.GetDatabaseStatus(context.TODO(), rs.Primary.ID).Execute()
		if err != nil || response.StatusCode >= 400 {
			return fmt.Errorf("Cannot find database status")
		}
		if databaseStatus.State != expectedState {
			return fmt.Errorf("Expected database status %s, got %s", expectedState, databaseStatus.State)
		}

		return nil
	}
}

func testAccDeploymentDefaultConfigWithDesiredState(desiredState string) string {

	return fmt.Sprintf(`
# Environment + Project
%s

# Needed by container resource
resource "qovery_container_registry" "deployment_registry_test" {
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

resource "qovery_deployment_stage" "deployment_stage_1" {
  environment_id = qovery_environment.test.id
  name        = "%s"
}

resource "qovery_deployment_stage" "deployment_stage_2" {
  environment_id = qovery_environment.test.id
  name        = "%s"
  move_after = qovery_deployment_stage.deployment_stage_1.id
}

resource "qovery_application" "application_test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "https://github.com/Qovery/test_http_server.git"
  }
  deployment_stage_id = qovery_deployment_stage.deployment_stage_1.id
}

resource "qovery_container" "container_test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.deployment_registry_test.id
  name = "%s"
  image_name = "terraform-provider-tests-container"
  tag = "1.0.0"
  deployment_stage_id = qovery_deployment_stage.deployment_stage_2.id
}

resource "qovery_database" "database_test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  type = "REDIS"
  version = "6"
  mode = "CONTAINER"
  deployment_stage_id = qovery_deployment_stage.deployment_stage_1.id
}

resource "qovery_deployment" "deployment_test" {
  environment_id = qovery_environment.test.id
  desired_state  = "%s"
  
  depends_on = [
    qovery_application.application_test,
    qovery_container.container_test,
    qovery_database.database_test,
  ]
}
`,
		testAccEnvironmentDefaultConfig("whole-deployment-environment"),
		getTestOrganizationID(),
		generateTestName("deployment-container-registry"),
		getTestAwsEcrURL(),
		getTestAWSCredentialsAccessKeyID(),
		getTestAWSCredentialsSecretAccessKey(),
		generateRandomName("deploymentstage"),
		generateRandomName("deploymentstage"),
		generateTestName("application"),
		generateTestName("container"),
		generateTestName("database"),
		desiredState,
	)
}
