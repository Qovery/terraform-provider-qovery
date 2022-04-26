package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestAcc_Environment(t *testing.T) {
	t.Parallel()
	nameSuffix := uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),

		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEnvironmentConfig(
					getTestProjectID(),
					generateEnvironmentName(nameSuffix),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "project_id", getTestProjectID()),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateEnvironmentName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
				),
			},
			// Add environment variables
			{
				Config: testAccEnvironmentConfigWithEnvironmentVariables(
					getTestProjectID(),
					generateEnvironmentName(nameSuffix),
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "project_id", getTestProjectID()),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateEnvironmentName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
				),
			},
			// Update environment variables
			{
				Config: testAccEnvironmentConfigWithEnvironmentVariables(
					getTestProjectID(),
					generateEnvironmentName(nameSuffix),
					map[string]string{
						"key1": "value1-updated",
						"key2": "value2-updated",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "project_id", getTestProjectID()),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateEnvironmentName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2-updated",
					}),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_EnvironmentWithMode(t *testing.T) {
	t.Parallel()
	nameSuffix := uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Create and Read testing with mode
			{
				Config: testAccEnvironmentConfigWithMode(
					getTestProjectID(),
					generateEnvironmentName(nameSuffix),
					"PRODUCTION",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "project_id", getTestProjectID()),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateEnvironmentName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "PRODUCTION"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_EnvironmentImport(t *testing.T) {
	t.Parallel()
	nameSuffix := uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEnvironmentConfig(
					getTestProjectID(),
					generateEnvironmentName(nameSuffix),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "project_id", getTestProjectID()),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateEnvironmentName(nameSuffix)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryEnvironmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("environment not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("environment.id not found")
		}

		_, apiErr := apiClient.GetEnvironment(context.TODO(), rs.Primary.ID)
		if apiErr != nil {
			return apiErr
		}
		return nil
	}
}

func testAccQoveryEnvironmentDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("environment not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("environment.id not found")
		}

		_, apiErr := apiClient.GetEnvironment(context.TODO(), rs.Primary.ID)
		if apiErr == nil {
			return fmt.Errorf("found environment but expected it to be deleted")
		}
		if !apierrors.IsNotFound(apiErr) {
			return fmt.Errorf("unexpected error checking for deleted environment: %s", apiErr.Summary())
		}
		return nil
	}
}

func testAccEnvironmentConfig(organizationID string, name string) string {
	return fmt.Sprintf(`
resource "qovery_environment" "test" {
  project_id = "%s" 
  name = "%s"
}
`, organizationID, name,
	)
}

func testAccEnvironmentConfigWithMode(organizationID string, name string, mode string) string {
	return fmt.Sprintf(`
resource "qovery_environment" "test" {
  project_id = "%s"
  name = "%s"
  mode = "%s"
}
`, organizationID, name, mode,
	)
}

func testAccEnvironmentConfigWithEnvironmentVariables(organizationID string, name string, environmentVariables map[string]string) string {
	return fmt.Sprintf(`
resource "qovery_environment" "test" {
  project_id = "%s"
  name = "%s"
  environment_variables = %s
}
`, organizationID, name, convertEnvVarsToString(environmentVariables),
	)
}

func generateEnvironmentName(suffix string) string {
	return fmt.Sprintf("%s-environment-%s", testResourcePrefix, suffix)
}
