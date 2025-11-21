//go:build integration && !unit
// +build integration,!unit

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

func TestAcc_TerraformService(t *testing.T) {
	t.Parallel()
	testName := "terraform-service"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryTerraformServiceDestroy("qovery_terraform_service.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTerraformServiceDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryTerraformServiceExists("qovery_terraform_service.test"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "description", "Terraform service for tests"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "auto_deploy", "true"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "git_repository.url", "https://github.com/Qovery/terraform-examples.git"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "git_repository.branch", "main"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "engine", "TERRAFORM"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "provider_version.explicit_version", "1.5.0"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "provider_version.read_from_terraform_block", "false"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "job_resources.cpu_milli", "1000"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "job_resources.ram_mib", "1024"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "job_resources.gpu", "0"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "job_resources.storage_gib", "20"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "timeout_sec", "1800"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "use_cluster_credentials", "false"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "backend.kubernetes.%", "0"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "tfvar_files.#", "0"),
				),
			},
			// Update with variables
			{
				Config: testAccTerraformServiceWithVariablesConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryTerraformServiceExists("qovery_terraform_service.test"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "name", generateTestName(testName)+"-updated"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_terraform_service.test", "variable.*", map[string]string{
						"key":    "AWS_REGION",
						"value":  "us-east-1",
						"secret": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_terraform_service.test", "variable.*", map[string]string{
						"key":    "DATABASE_PASSWORD",
						"secret": "true",
					}),
				),
			},
			// ImportState testing
			{
				ResourceName:            "qovery_terraform_service.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"variable"},
			},
		},
	})
}

func TestAcc_TerraformServiceUserProvidedBackend(t *testing.T) {
	t.Parallel()
	testName := "terraform-service-user-backend"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryTerraformServiceDestroy("qovery_terraform_service.test"),
		Steps: []resource.TestStep{
			// Create with user-provided backend
			{
				Config: testAccTerraformServiceUserProvidedBackendConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryTerraformServiceExists("qovery_terraform_service.test"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "backend.user_provided.%", "0"),
					resource.TestCheckNoResourceAttr("qovery_terraform_service.test", "backend.kubernetes.%"),
				),
			},
		},
	})
}

func TestAcc_TerraformServiceWithAdvancedSettings(t *testing.T) {
	t.Parallel()
	testName := "terraform-service-advanced"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryTerraformServiceDestroy("qovery_terraform_service.test"),
		Steps: []resource.TestStep{
			// Create with advanced settings
			{
				Config: testAccTerraformServiceWithAdvancedSettingsConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryTerraformServiceExists("qovery_terraform_service.test"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "advanced_settings_json", "{\"deployment.termination_grace_period_seconds\":120}"),
				),
			},
		},
	})
}

func TestAcc_TerraformServiceOpenTofu(t *testing.T) {
	t.Parallel()
	testName := "terraform-service-opentofu"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryTerraformServiceDestroy("qovery_terraform_service.test"),
		Steps: []resource.TestStep{
			// Create with OpenTofu engine
			{
				Config: testAccTerraformServiceOpenTofuConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryTerraformServiceExists("qovery_terraform_service.test"),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "engine", "OPEN_TOFU"),
				),
			},
		},
	})
}

func TestAcc_TerraformServiceStorageImmutability(t *testing.T) {
	t.Parallel()
	testName := "terraform-service-storage"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryTerraformServiceDestroy("qovery_terraform_service.test"),
		Steps: []resource.TestStep{
			// Create with 20 GiB storage
			{
				Config: testAccTerraformServiceDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_terraform_service.test", "job_resources.storage_gib", "20"),
				),
			},
			// Try to reduce storage (should fail)
			{
				Config:      testAccTerraformServiceWithReducedStorageConfig(testName),
				ExpectError: regexp.MustCompile("Storage cannot be reduced"),
			},
		},
	})
}

// Helper functions

func testAccQoveryTerraformServiceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("terraform service not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("terraform_service.id not found")
		}

		_, err := qoveryServices.TerraformService.Get(context.TODO(), rs.Primary.ID, "{}", false)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryTerraformServiceDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("terraform service not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("terraform_service.id not found")
		}

		_, err := qoveryServices.TerraformService.Get(context.TODO(), rs.Primary.ID, "{}", false)
		if err == nil {
			return fmt.Errorf("found terraform service but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted terraform service: %s", err.Error())
		}
		return nil
	}
}

// Configuration generators

func testAccTerraformServiceDefaultConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_terraform_service" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  description    = "Terraform service for tests"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  backend = {
    kubernetes = {}
  }

  engine = "TERRAFORM"

  provider_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 20
  }

  timeout_sec             = 1800
  use_cluster_credentials = false
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}

func testAccTerraformServiceWithVariablesConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_terraform_service" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s-updated"
  description    = "Terraform service for tests"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  variable = [
    {
      key    = "AWS_REGION"
      value  = "us-east-1"
      secret = false
    },
    {
      key    = "DATABASE_PASSWORD"
      value  = "supersecret123"
      secret = true
    }
  ]

  backend = {
    kubernetes = {}
  }

  engine = "TERRAFORM"

  provider_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 20
  }

  timeout_sec             = 1800
  use_cluster_credentials = false
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}

func testAccTerraformServiceUserProvidedBackendConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_terraform_service" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  description    = "Terraform service with user-provided backend"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  backend = {
    user_provided = {}
  }

  engine = "TERRAFORM"

  provider_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 20
  }

  timeout_sec             = 1800
  use_cluster_credentials = false
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}

func testAccTerraformServiceWithAdvancedSettingsConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_terraform_service" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  description    = "Terraform service with advanced settings"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  backend = {
    kubernetes = {}
  }

  engine = "TERRAFORM"

  provider_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 20
  }

  timeout_sec             = 1800
  use_cluster_credentials = false

  advanced_settings_json = jsonencode({
    "deployment.termination_grace_period_seconds" : 120
  })
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}

func testAccTerraformServiceOpenTofuConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_terraform_service" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  description    = "Terraform service with OpenTofu"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  backend = {
    kubernetes = {}
  }

  engine = "OPEN_TOFU"

  provider_version = {
    explicit_version          = "1.6.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 20
  }

  timeout_sec             = 1800
  use_cluster_credentials = false
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}

func testAccTerraformServiceWithReducedStorageConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_terraform_service" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  description    = "Terraform service for tests"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  backend = {
    kubernetes = {}
  }

  engine = "TERRAFORM"

  provider_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 10  # Reduced from 20
  }

  timeout_sec             = 1800
  use_cluster_credentials = false
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}
