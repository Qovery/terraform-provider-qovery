package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestAcc_Environment(t *testing.T) {
	t.Parallel()
	testName := "environment"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEnvironmentDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
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

func TestAcc_EnvironmentWithEnvironmentVariables(t *testing.T) {
	t.Parallel()
	testName := "environment-with-environment-variables"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "env_key_1", value: "env_value_1", scope: "ENVIRONMENT"},
						{key: "project_key_1", value: "project_value_1", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "env_key_1",
						"value": "env_value_1",
						"scope": "ENVIRONMENT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1",
						"scope": "PROJECT",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Update environment variable
			{
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "env_key_1", value: "env_value_1_updated", scope: "ENVIRONMENT"},
						{key: "project_key_1", value: "project_value_1_updated", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "env_key_1",
						"value": "env_value_1_updated",
						"scope": "ENVIRONMENT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1_updated",
						"scope": "PROJECT",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Add environment variable
			{
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "env_key_1", value: "env_value_1", scope: "ENVIRONMENT"},
						{key: "env_key_2", value: "env_value_2", scope: "ENVIRONMENT"},
						{key: "project_key_1", value: "project_value_1", scope: "PROJECT"},
						{key: "project_key_2", value: "project_value_2", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "env_key_1",
						"value": "env_value_1",
						"scope": "ENVIRONMENT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "env_key_2",
						"value": "env_value_2",
						"scope": "ENVIRONMENT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1",
						"scope": "PROJECT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "project_key_2",
						"value": "project_value_2",
						"scope": "PROJECT",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Remove environment variable
			{
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "env_key_2", value: "env_value_2", scope: "ENVIRONMENT"},
						{key: "project_key_2", value: "project_value_2", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "env_key_2",
						"value": "env_value_2",
						"scope": "ENVIRONMENT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "project_key_2",
						"value": "project_value_2",
						"scope": "PROJECT",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
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

func TestAcc_EnvironmentWithSecrets(t *testing.T) {
	t.Parallel()
	testName := "environment-with-secrets"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEnvironmentDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "key1", value: "value1"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Update secret
			{
				Config: testAccEnvironmentDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "key1", value: "value1-updated"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Add secret
			{
				Config: testAccEnvironmentDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "key1", value: "value1"},
						{key: "key2", value: "value2"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Delete secret
			{
				Config: testAccEnvironmentDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "key2", value: "value2"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
		},
	})
}

func TestAcc_EnvironmentWithMode(t *testing.T) {
	t.Parallel()
	testName := "environment-with-mode"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Create and Read testing with mode
			{
				Config: testAccEnvironmentDefaultConfigWithMode(
					testName,
					"PRODUCTION",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "PRODUCTION"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
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

func TestAcc_EnvironmentImport(t *testing.T) {
	t.Parallel()
	testName := "environment-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEnvironmentDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
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

func testAccEnvironmentDefaultConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
}
`, testAccProjectDefaultConfig(testName), getTestClusterID(), generateTestName(testName),
	)
}

func testAccEnvironmentDefaultConfigWithMode(testName string, mode string) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  mode = "%s"
}
`, testAccProjectDefaultConfig(testName), getTestClusterID(), generateTestName(testName), mode,
	)
}

func testAccEnvironmentDefaultConfigWithEnvironmentVariables(testName string, environmentVariables []environmentVariable) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  environment_variables = %s
}
`, testAccProjectDefaultConfig(testName), getTestClusterID(), generateTestName(testName), convertEnvVarsToString(environmentVariables),
	)
}

func testAccEnvironmentDefaultConfigWithSecrets(testName string, secrets []environmentSecret) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  secrets = %s
}
`, testAccProjectDefaultConfig(testName), getTestClusterID(), generateTestName(testName), convertEnvSecretsToString(secrets),
	)
}
