//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

func TestAcc_Project(t *testing.T) {
	t.Parallel()
	testName := "project"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckNoResourceAttr("qovery_project.test", "secrets.0"),
				),
			},
			// Add description
			{
				Config: testAccProjectDefaultConfigWithDescription(
					testName,
					"this is a description",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", "this is a description"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckNoResourceAttr("qovery_project.test", "secrets.0"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ProjectWithEnvironmentVariables(t *testing.T) {
	t.Parallel()
	testName := "project-with-environment-variables"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariablesAndAliases(
					testName,
					map[string]string{
						"key1": "",
					},
					map[string]string{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "",
					}),
				),
			},
			// Update environment variable
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariablesAndAliases(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
					map[string]string{
						"key1_alias": "key1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
				),
			},
			// Add environment variable
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariablesAndAliases(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					map[string]string{
						"key1_alias": "key1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
				),
			},
			// Remove environment variable
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariablesAndAliases(
					testName,
					map[string]string{
						"key2": "value2",
					},
					map[string]string{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variable_aliases.0"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ProjectWithSecrets(t *testing.T) {
	t.Parallel()
	testName := "project-with-secrets"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfigWithSecretsAndAliases(
					testName,
					map[string]string{
						"key1": "",
					},
					map[string]string{
						"key1_alias": "key1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
				),
			},
			// Update secrets
			{
				Config: testAccProjectDefaultConfigWithSecretsAndAliases(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
					map[string]string{
						"key1_alias": "key1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
				),
			},
			// Add secrets
			{
				Config: testAccProjectDefaultConfigWithSecretsAndAliases(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					map[string]string{
						"key1_alias": "key1",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
				),
			},
			// Remove secrets
			{
				Config: testAccProjectDefaultConfigWithSecretsAndAliases(
					testName,
					map[string]string{
						"key2": "value2",
					},
					map[string]string{
						"key1_alias": "key2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key2",
					}),
				),
			},
		},
	})
}

func TestAcc_ProjectImport(t *testing.T) {
	t.Parallel()
	testName := "project-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "built_in_environment_variables.*", map[string]string{
						"key":   "QOVERY_PROJECT_ID",
						"value": "||Q_PRJ_ID||",
					}),
					resource.TestCheckNoResourceAttr("qovery_project.test", "secrets.0"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryProjectExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("project not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("project.id not found")
		}

		_, err := qoveryServices.Project.Get(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryProjectDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("project not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("project.id not found")
		}

		_, err := qoveryServices.Project.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found project but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted project: %s", err.Error())
		}
		return nil
	}
}

func testAccProjectDefaultConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
}
`, getTestOrganizationID(), generateTestName(testName),
	)
}

func testAccProjectDefaultConfigWithDescription(testName string, description string) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  description = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), description,
	)
}

func testAccProjectDefaultConfigWithEnvironmentVariables(testName string, environmentVariables map[string]string) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  environment_variables = %s
}
`, getTestOrganizationID(), generateTestName(testName), convertEnvVarsToString(environmentVariables),
	)
}

func testAccProjectDefaultConfigWithEnvironmentVariablesAndAliases(
	testName string,
	environmentVariables map[string]string,
	environmentVariableAliases map[string]string,
) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  environment_variables = %s
  environment_variable_aliases = %s
}
`,
		getTestOrganizationID(),
		generateTestName(testName),
		convertEnvVarsToString(environmentVariables),
		convertEnvVarsToString(environmentVariableAliases),
	)
}

func testAccProjectDefaultConfigWithSecrets(testName string, secrets map[string]string) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  secrets = %s
}
`, getTestOrganizationID(), generateTestName(testName), convertEnvVarsToString(secrets),
	)
}

func testAccProjectDefaultConfigWithSecretsAndAliases(
	testName string,
	secrets map[string]string,
	secretAliases map[string]string,
) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  secrets = %s
  secret_aliases = %s
}
`,
		getTestOrganizationID(),
		generateTestName(testName),
		convertEnvVarsToString(secrets),
		convertEnvVarsToString(secretAliases),
	)
}

func convertEnvVarsToString(environmentVariables map[string]string) string {
	vars := make([]string, 0, len(environmentVariables))
	for key, value := range environmentVariables {
		vars = append(vars, fmt.Sprintf(`{key: "%s", value: "%s"}`, key, value))
	}
	return fmt.Sprintf("[%s]", strings.Join(vars, ","))
}

func testAccProjectDefaultConfigWithEnvVarFiles(
	testName string,
	environmentVariableFiles map[string]fileVar,
) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  environment_variable_files = %s
}
`,
		getTestOrganizationID(),
		generateTestName(testName),
		convertFileVarsToString(environmentVariableFiles),
	)
}

func testAccProjectDefaultConfigWithSecretFiles(
	testName string,
	secretFiles map[string]fileVar,
) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  secret_files = %s
}
`,
		getTestOrganizationID(),
		generateTestName(testName),
		convertFileVarsToString(secretFiles),
	)
}

func TestAcc_ProjectWithEnvironmentVariableFiles(t *testing.T) {
	t.Parallel()
	testName := "project-with-env-var-files"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Step 1: Create with one env var file
			{
				Config: testAccProjectDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config": {key: "APP_CONFIG", value: "config-content", mountPath: "/etc/app/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "config-content",
						"mount_path": "/etc/app/config.yaml",
					}),
				),
			},
			// Step 2: Update value only (mount_path stays the same)
			{
				Config: testAccProjectDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config": {key: "APP_CONFIG", value: "updated-content", mountPath: "/etc/app/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "updated-content",
						"mount_path": "/etc/app/config.yaml",
					}),
				),
			},
			// Step 3: Update mount_path (triggers delete+recreate)
			{
				Config: testAccProjectDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config": {key: "APP_CONFIG", value: "updated-content", mountPath: "/new/path/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "updated-content",
						"mount_path": "/new/path/config.yaml",
					}),
				),
			},
			// Step 4: Add a second file variable
			{
				Config: testAccProjectDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config":  {key: "APP_CONFIG", value: "updated-content", mountPath: "/new/path/config.yaml"},
						"config2": {key: "DB_CONFIG", value: "db-content", mountPath: "/etc/db/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "updated-content",
						"mount_path": "/new/path/config.yaml",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variable_files.*", map[string]string{
						"key":        "DB_CONFIG",
						"value":      "db-content",
						"mount_path": "/etc/db/config.yaml",
					}),
				),
			},
			// Step 5: Remove all file variables
			{
				Config: testAccProjectDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variable_files.0"),
				),
			},
			// Step 6: Import
			{
				ResourceName:      "qovery_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ProjectWithSecretFiles(t *testing.T) {
	t.Parallel()
	testName := "project-with-secret-files"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Step 1: Create with one secret file
			{
				Config: testAccProjectDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{
						"secret": {key: "API_KEY", value: "secret-value", mountPath: "/usr/local/secrets/api-key"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secret_files.*", map[string]string{
						"key":        "API_KEY",
						"value":      "secret-value",
						"mount_path": "/usr/local/secrets/api-key",
					}),
				),
			},
			// Step 2: Update value only (mount_path stays the same)
			{
				Config: testAccProjectDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{
						"secret": {key: "API_KEY", value: "new-secret-value", mountPath: "/usr/local/secrets/api-key"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secret_files.*", map[string]string{
						"key":        "API_KEY",
						"value":      "new-secret-value",
						"mount_path": "/usr/local/secrets/api-key",
					}),
				),
			},
			// Step 3: Update mount_path (triggers delete+recreate)
			{
				Config: testAccProjectDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{
						"secret": {key: "API_KEY", value: "new-secret-value", mountPath: "/new/path/api-key"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secret_files.*", map[string]string{
						"key":        "API_KEY",
						"value":      "new-secret-value",
						"mount_path": "/new/path/api-key",
					}),
				),
			},
			// Step 4: Remove all secret files
			{
				Config: testAccProjectDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "secret_files.0"),
				),
			},
			// Step 5: Import
			{
				ResourceName:            "qovery_project.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_files"},
			},
		},
	})
}
