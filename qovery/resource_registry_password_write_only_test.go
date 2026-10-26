//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// testAccPublicHelmRepositoryURL is an HTTPS helm repository that accepts any basic auth: q-core
// checks a helm repository by fetching its index.yaml with the credentials, so the tests can send
// any password.
const testAccPublicHelmRepositoryURL = "https://gitlab.com/mulesoft-int/helm-repository/-/raw/master/"

// testAccRandomProvider provides ephemeral "random_password", a write-only value that changes on
// every plan and apply and is never stored.
var testAccRandomProvider = map[string]resource.ExternalProvider{
	"random": {Source: "hashicorp/random", VersionConstraint: "~> 3.7"},
}

// The password of an HTTPS helm repository, sent through password_wo, never reaches the state. The
// steps cover the switch from password, the version bump, an unrelated update, import, and the
// adoption of password_wo after import. The ephemeral password changes on every plan, so each
// empty plan after apply shows that a new value alone plans nothing.
func TestAcc_HelmRepositoryPasswordWriteOnly(t *testing.T) {
	t.Parallel()
	const addr = "qovery_helm_repository.test"
	testName := "helm-repository-password-wo"
	plainPassword := "plain-" + uuid.NewString()
	var repositoryID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        testAccRandomProvider,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		CheckDestroy:             testAccQoveryHelmRepositoryDestroy(addr),
		Steps: []resource.TestStep{
			// The plain password is stored in the state...
			{
				Config: testAccHelmRepositoryPasswordConfig(testName, "", fmt.Sprintf(`password = %q`, plainPassword)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryHelmRepositoryExists(addr),
					resource.TestCheckResourceAttr(addr, "config.username", "qovery"),
					resource.TestCheckResourceAttr(addr, "config.password", plainPassword),
					testAccCaptureResourceID(addr, &repositoryID),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// ... and switching to password_wo is an in-place update that removes it.
			{
				Config: testAccHelmRepositoryPasswordConfig(testName, "", testAccHelmRepositoryPasswordWo("1")),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "id", &repositoryID),
					resource.TestCheckNoResourceAttr(addr, "config.password"),
					resource.TestCheckNoResourceAttr(addr, "config.password_wo"),
					resource.TestCheckResourceAttr(addr, "config.password_wo_version", "1"),
					testAccCheckStateHasNoValue(plainPassword),
				),
			},
			// A new value alone plans nothing.
			{
				Config:   testAccHelmRepositoryPasswordConfig(testName, "", testAccHelmRepositoryPasswordWo("1")),
				PlanOnly: true,
			},
			// Bumping the version sends the current value.
			{
				Config: testAccHelmRepositoryPasswordConfig(testName, "", testAccHelmRepositoryPasswordWo("2")),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "id", &repositoryID),
					resource.TestCheckNoResourceAttr(addr, "config.password_wo"),
					resource.TestCheckResourceAttr(addr, "config.password_wo_version", "2"),
				),
			},
			// An unrelated update also sends it, as the API replaces the whole config.
			{
				Config: testAccHelmRepositoryPasswordConfig(testName, "rotated with the version", testAccHelmRepositoryPasswordWo("2")),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "description", "rotated with the version"),
					resource.TestCheckNoResourceAttr(addr, "config.password_wo"),
					resource.TestCheckResourceAttr(addr, "config.password_wo_version", "2"),
				),
			},
			// Import records neither the password nor its version...
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"config.password_wo_version"},
			},
			// ... and a repository imported into a fresh state...
			{
				Config: testAccEphemeralRandomPassword + fmt.Sprintf(`
removed {
  from = %s

  lifecycle {
    destroy = false
  }
}
`, addr),
				Check: testAccCheckAbsentFromState(addr),
			},
			{
				Config:             testAccHelmRepositoryPasswordConfig(testName, "rotated with the version", testAccHelmRepositoryPasswordWo("2")),
				ResourceName:       addr,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return fmt.Sprintf("%s,%s", getTestOrganizationID(), repositoryID), nil
				},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if version, ok := states[0].Attributes["config.password_wo_version"]; ok {
						return fmt.Errorf("import recorded config.password_wo_version = %q, expected none", version)
					}
					return nil
				},
			},
			// ... so the imported repository adopts password_wo in one update.
			{
				Config: testAccHelmRepositoryPasswordConfig(testName, "rotated with the version", testAccHelmRepositoryPasswordWo("2")),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "id", &repositoryID),
					resource.TestCheckResourceAttr(addr, "config.password_wo_version", "2"),
				),
			},
		},
	})
}

// With ignore_changes = [config], the plan keeps password_wo_version while the config no longer
// sets password_wo. The API replaces the whole config on update, so the update fails instead of
// clearing the password.
func TestAcc_HelmRepositoryPasswordWriteOnlyIgnoredConfig(t *testing.T) {
	t.Parallel()
	const addr = "qovery_helm_repository.test"
	testName := "helm-repository-password-wo-ignored-config"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        testAccRandomProvider,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		CheckDestroy:             testAccQoveryHelmRepositoryDestroy(addr),
		Steps: []resource.TestStep{
			{
				Config: testAccHelmRepositoryPasswordConfig(testName, "", testAccHelmRepositoryPasswordWo("1")),
			},
			{
				Config: testAccEphemeralRandomPassword + fmt.Sprintf(`
resource "qovery_helm_repository" "test" {
  organization_id = "%s"
  name = "%s"
  kind = "HTTPS"
  url = "%s"
  description = "changed"
  skip_tls_verification = false

  lifecycle {
    ignore_changes = [config]
  }
}
`, getTestOrganizationID(), generateTestName(testName), testAccPublicHelmRepositoryURL),
				ExpectError: regexp.MustCompile(`The\s+plan\s+keeps\s+config.password_wo_version`),
			},
		},
	})
}

// testAccProvider1_2FromRegistry is the last release before the write-only passwords.
var testAccProvider1_2FromRegistry = map[string]resource.ExternalProvider{
	"qovery": {Source: "qovery/qovery", VersionConstraint: "1.2.2"},
}

// A helm repository with a plain password, created by 1.2.2, plans nothing after the upgrade: the
// state has no password_wo_version, which reads as null.
func TestAcc_HelmRepositoryPasswordUpgradeFrom1_2(t *testing.T) {
	t.Parallel()
	const addr = "qovery_helm_repository.test"
	config := testAccHelmRepositoryPasswordResource("helm-repository-password-upgrade", "", `password = "plain"`)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccQoveryHelmRepositoryDestroy(addr),
		Steps: []resource.TestStep{
			{
				ExternalProviders: testAccProvider1_2FromRegistry,
				Config:            config,
				Check:             testAccQoveryHelmRepositoryExists(addr),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.TestCheckNoResourceAttr(addr, "config.password_wo_version"),
			},
		},
	})
}

// The write-only password is checked at plan time, on both resources.
func TestAcc_RegistryPasswordWriteOnlyValidation(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        testAccRandomProvider,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		Steps: []resource.TestStep{
			{
				Config:             testAccContainerRegistryPasswordConfig("container-registry-password-wo", testAccHelmRepositoryPasswordWo("1")),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:      testAccContainerRegistryPasswordConfig("container-registry-password-wo-no-version", `password_wo = ephemeral.random_password.secret.result`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`config.password_wo_version\s+is\s+required\s+when\s+config.password_wo\s+is\s+set`),
			},
			{
				Config: testAccHelmRepositoryPasswordConfig("helm-repository-password-and-wo", "", `
    password = "plain"
    password_wo = ephemeral.random_password.secret.result
    password_wo_version = "1"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`config.password_wo\s+cannot\s+be\s+set\s+with\s+config.password`),
			},
			{
				Config:      testAccHelmRepositoryPasswordConfig("helm-repository-password-wo-version-only", "", `password_wo_version = "1"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`config.password_wo_version\s+can\s+only\s+be\s+set\s+with\s+config.password_wo`),
			},
		},
	})
}

// Terraform 1.10 and earlier reject a config that sets a write-only attribute. Run it once with
// TF_ACC_TERRAFORM_VERSION=1.10.5: it skips on the Terraform versions the CI uses.
func TestAcc_RegistryPasswordWriteOnlyNeedsTerraform1_11(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{skipFromTerraform1_11{}},
		Steps: []resource.TestStep{
			{
				Config: testAccHelmRepositoryPasswordResource("helm-repository-password-wo-tf-1-10", "", `
    password_wo = "secret"
    password_wo_version = "1"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Write-only\s+attributes\s+are\s+only\s+supported\s+in\s+Terraform\s+1.11\s+and\s+later`),
			},
		},
	})
}

// skipFromTerraform1_11 skips a test from Terraform 1.11, the first version with write-only
// attributes.
type skipFromTerraform1_11 struct{}

func (skipFromTerraform1_11) CheckTerraformVersion(_ context.Context, req tfversion.CheckTerraformVersionRequest, resp *tfversion.CheckTerraformVersionResponse) {
	if req.TerraformVersion.Core().GreaterThanOrEqual(tfversion.Version1_11_0) {
		resp.Skip = fmt.Sprintf("Terraform CLI version %s supports write-only attributes: skipping test", req.TerraformVersion)
	}
}

// testAccHelmRepositoryPasswordWo sets password_wo from the ephemeral random password, with the
// given version.
func testAccHelmRepositoryPasswordWo(version string) string {
	return fmt.Sprintf(`
    password_wo = ephemeral.random_password.secret.result
    password_wo_version = %q`, version)
}

// testAccHelmRepositoryPasswordConfig declares an HTTPS helm repository authenticated as qovery
// with the given password attributes, next to an ephemeral random password.
func testAccHelmRepositoryPasswordConfig(testName, description, password string) string {
	return testAccEphemeralRandomPassword + testAccHelmRepositoryPasswordResource(testName, description, password)
}

func testAccHelmRepositoryPasswordResource(testName, description, password string) string {
	return fmt.Sprintf(`
resource "qovery_helm_repository" "test" {
  organization_id = "%s"
  name = "%s"
  kind = "HTTPS"
  url = "%s"
  description = %q
  skip_tls_verification = false
  config = {
    username = "qovery"
    %s
  }
}
`, getTestOrganizationID(), generateTestName(testName), testAccPublicHelmRepositoryURL, description, password)
}

// testAccContainerRegistryPasswordConfig declares a GENERIC_CR registry authenticated as qovery with
// the given password attributes, next to an ephemeral random password. q-core checks the password
// on every write and the sandbox has no registry account, so the tests only plan it.
func testAccContainerRegistryPasswordConfig(testName, password string) string {
	return testAccEphemeralRandomPassword + fmt.Sprintf(`
resource "qovery_container_registry" "test" {
  organization_id = "%s"
  name = "%s"
  kind = "GENERIC_CR"
  url = "https://registry.example.com"
  config = {
    username = "qovery"
    %s
  }
}
`, getTestOrganizationID(), generateTestName(testName), password)
}

const testAccEphemeralRandomPassword = `
ephemeral "random_password" "secret" {
  length = 24
}
`

// testAccCheckStateHasNoValue fails when any attribute of any resource in the state contains
// secret.
func testAccCheckStateHasNoValue(secret string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for addr, rs := range s.RootModule().Resources {
			for name, value := range rs.Primary.Attributes {
				if strings.Contains(value, secret) {
					return fmt.Errorf("%s.%s holds the secret", addr, name)
				}
			}
		}
		return nil
	}
}
