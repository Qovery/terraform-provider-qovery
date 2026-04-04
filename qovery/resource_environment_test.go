//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
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
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key1": "",
					},
					map[string]string{},
					map[string]string{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Update environment variable
			{
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
					map[string]string{
						"key1_alias": "key1",
					},
					map[string]string{
						"environment_variable": "override value",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_overrides.*", map[string]string{
						"key":   "environment_variable",
						"value": "override value",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Add environment variable
			{
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					map[string]string{
						"key1_alias": "key1",
					},
					map[string]string{
						"environment_variable": "override value update",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_overrides.*", map[string]string{
						"key":   "environment_variable",
						"value": "override value update",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Delete environment variable
			{
				Config: testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key2": "value2",
					},
					map[string]string{},
					map[string]string{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secrets.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variable_aliases.0"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variable_overrides.0"),
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
				Config: testAccEnvironmentDefaultConfigWithSecretsAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key1": "",
					},
					map[string]string{
						"key1_alias": "key1",
					},
					map[string]string{
						"environment_secret": "override value",
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
						"value": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_overrides.*", map[string]string{
						"key":   "environment_secret",
						"value": "override value",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Update secret
			{
				Config: testAccEnvironmentDefaultConfigWithSecretsAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key1": "value1-updated",
					},
					map[string]string{
						"key1_alias": "key1",
					},
					map[string]string{
						"environment_secret": "override value updated",
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
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_overrides.*", map[string]string{
						"key":   "environment_secret",
						"value": "override value updated",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Add secret
			{
				Config: testAccEnvironmentDefaultConfigWithSecretsAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					map[string]string{
						"key1_alias": "key1",
					},
					map[string]string{
						"environment_secret": "override value updated",
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
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_overrides.*", map[string]string{
						"key":   "environment_secret",
						"value": "override value updated",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			// Delete secret
			{
				Config: testAccEnvironmentDefaultConfigWithSecretsAndAliasesAndOverrides(
					testName,
					map[string]string{
						"key2": "value2",
					},
					map[string]string{
						"key1_alias": "key2",
					},
					map[string]string{
						"environment_secret": "override value updated",
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
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
				),
			},
			{
				ResourceName:            "qovery_environment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secrets", "secret_aliases", "secret_overrides"},
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

func testAccQoveryEnvironmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("environment not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("environment.id not found")
		}

		_, err := qoveryServices.Environment.Get(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
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

func testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndSecrets(testName string, environmentVariables map[string]string, secrets map[string]string) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  environment_variables = %s
  secrets = %s
}
`, testAccProjectDefaultConfig(testName), getTestClusterID(), generateTestName(testName), convertEnvVarsToString(environmentVariables), convertEnvVarsToString(secrets),
	)
}

func testAccEnvironmentDefaultConfigWithEnvironmentVariables(testName string, environmentVariables map[string]string) string {
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

func testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndAliasesAndOverrides(
	testName string,
	environmentVariables map[string]string,
	environmentVariableAliases map[string]string,
	environmentVariableOverrides map[string]string,
) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  environment_variables = %s
  environment_variable_aliases = %s
  environment_variable_overrides = %s
}
`,
		testAccProjectDefaultConfigWithEnvironmentVariables(testName, map[string]string{"environment_variable": "simple value"}),
		getTestClusterID(),
		generateTestName(testName),
		convertEnvVarsToString(environmentVariables),
		convertEnvVarsToString(environmentVariableAliases),
		convertEnvVarsToString(environmentVariableOverrides),
	)
}

func testAccEnvironmentDefaultConfigWithSecrets(testName string, secrets map[string]string) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  secrets = %s
}
`, testAccProjectDefaultConfig(testName), getTestClusterID(), generateTestName(testName), convertEnvVarsToString(secrets),
	)
}

func testAccEnvironmentDefaultConfigWithSecretsAndAliasesAndOverrides(
	testName string,
	secrets map[string]string,
	secretAliases map[string]string,
	secretOverrides map[string]string,
) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  secrets = %s
  secret_aliases = %s
  secret_overrides = %s
}
`,
		testAccProjectDefaultConfigWithSecrets(testName, map[string]string{"environment_secret": "simple value"}),
		getTestClusterID(),
		generateTestName(testName),
		convertEnvVarsToString(secrets),
		convertEnvVarsToString(secretAliases),
		convertEnvVarsToString(secretOverrides),
	)
}

func testAccEnvironmentDefaultConfigWithEnvVarFiles(
	testName string,
	environmentVariableFiles map[string]fileVar,
) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  environment_variable_files = %s
}
`,
		testAccProjectDefaultConfig(testName),
		getTestClusterID(),
		generateTestName(testName),
		convertFileVarsToString(environmentVariableFiles),
	)
}

func testAccEnvironmentDefaultConfigWithSecretFiles(
	testName string,
	secretFiles map[string]fileVar,
) string {
	return fmt.Sprintf(`
%s

resource "qovery_environment" "test" {
  cluster_id = "%s"
  project_id = qovery_project.test.id
  name = "%s"
  secret_files = %s
}
`,
		testAccProjectDefaultConfig(testName),
		getTestClusterID(),
		generateTestName(testName),
		convertFileVarsToString(secretFiles),
	)
}

func TestAcc_EnvironmentWithEnvironmentVariableFiles(t *testing.T) {
	t.Parallel()
	testName := "env-with-env-var-files"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Step 1: Create with one env var file
			{
				Config: testAccEnvironmentDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config": {key: "APP_CONFIG", value: "config-content", mountPath: "/etc/app/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "config-content",
						"mount_path": "/etc/app/config.yaml",
					}),
				),
			},
			// Step 2: Update value only (mount_path stays the same)
			{
				Config: testAccEnvironmentDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config": {key: "APP_CONFIG", value: "updated-content", mountPath: "/etc/app/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "updated-content",
						"mount_path": "/etc/app/config.yaml",
					}),
				),
			},
			// Step 3: Update mount_path (triggers delete+recreate)
			{
				Config: testAccEnvironmentDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config": {key: "APP_CONFIG", value: "updated-content", mountPath: "/new/path/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "updated-content",
						"mount_path": "/new/path/config.yaml",
					}),
				),
			},
			// Step 4: Add a second file variable
			{
				Config: testAccEnvironmentDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{
						"config":  {key: "APP_CONFIG", value: "updated-content", mountPath: "/new/path/config.yaml"},
						"config2": {key: "DB_CONFIG", value: "db-content", mountPath: "/etc/db/config.yaml"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_files.*", map[string]string{
						"key":        "APP_CONFIG",
						"value":      "updated-content",
						"mount_path": "/new/path/config.yaml",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "environment_variable_files.*", map[string]string{
						"key":        "DB_CONFIG",
						"value":      "db-content",
						"mount_path": "/etc/db/config.yaml",
					}),
				),
			},
			// Step 5: Remove all file variables
			{
				Config: testAccEnvironmentDefaultConfigWithEnvVarFiles(
					testName,
					map[string]fileVar{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "environment_variable_files.0"),
				),
			},
			// Step 6: Import
			{
				ResourceName:      "qovery_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_EnvironmentWithSecretFiles(t *testing.T) {
	t.Parallel()
	testName := "env-with-secret-files"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryEnvironmentDestroy("qovery_environment.test"),
		Steps: []resource.TestStep{
			// Step 1: Create with one secret file
			{
				Config: testAccEnvironmentDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{
						"secret": {key: "API_KEY", value: "secret-value", mountPath: "/usr/local/secrets/api-key"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckResourceAttr("qovery_environment.test", "name", generateTestName(testName)),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_files.*", map[string]string{
						"key":        "API_KEY",
						"value":      "secret-value",
						"mount_path": "/usr/local/secrets/api-key",
					}),
				),
			},
			// Step 2: Update value only (mount_path stays the same)
			{
				Config: testAccEnvironmentDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{
						"secret": {key: "API_KEY", value: "new-secret-value", mountPath: "/usr/local/secrets/api-key"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_files.*", map[string]string{
						"key":        "API_KEY",
						"value":      "new-secret-value",
						"mount_path": "/usr/local/secrets/api-key",
					}),
				),
			},
			// Step 3: Update mount_path (triggers delete+recreate)
			{
				Config: testAccEnvironmentDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{
						"secret": {key: "API_KEY", value: "new-secret-value", mountPath: "/new/path/api-key"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_environment.test", "secret_files.*", map[string]string{
						"key":        "API_KEY",
						"value":      "new-secret-value",
						"mount_path": "/new/path/api-key",
					}),
				),
			},
			// Step 4: Remove all secret files
			{
				Config: testAccEnvironmentDefaultConfigWithSecretFiles(
					testName,
					map[string]fileVar{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					resource.TestCheckNoResourceAttr("qovery_environment.test", "secret_files.0"),
				),
			},
			// Step 5: Import
			{
				ResourceName:            "qovery_environment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_files"},
			},
		},
	})
}
