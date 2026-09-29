//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"
)

// The organization-level contract tests cover the attributes QOV-2333 moved to the
// config-is-source-of-truth rule: the description of projects, container registries and helm
// repositories defaults to "", and the non-secret keys of the registry config blocks are read from
// the API. They reuse the service contract steps (testAccServiceContract).

// --- project -------------------------------------------------------------------------------------

func TestAcc_ProjectContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_project.test"
	testName := "project-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address:  address,
		config:   func(declared bool) string { return testAccProjectContractConfig(testName, declared) },
		declared: []testAccContractValue{{"description", "declared"}},
		defaults: []testAccContractValue{{"description", ""}},
		outOfBand: func(id string) error {
			return testAccEditServiceOutOfBand("/project/"+id, nil, func(project map[string]any) {
				project["description"] = "set from the Console"
			})
		},
		changed: []testAccContractValue{{"description", "set from the Console"}},
		exists:  testAccQoveryProjectExists(address),
		destroy: testAccQoveryProjectDestroy(address),
	})
}

func TestAcc_ProjectContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testAccServiceContractUpgradeFrom0x(t, testAccProjectContractConfig("project-contract-upgrade", false), testAccQoveryProjectDestroy("qovery_project.test"))
}

func testAccProjectContractConfig(testName string, declared bool) string {
	description := ""
	if declared {
		description = `
  description     = "declared"`
	}
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name            = "%s"%s
}
`, getTestOrganizationID(), generateTestName(testName), description)
}

// --- container registry --------------------------------------------------------------------------

// TestAcc_ContainerRegistryContract uses a Docker Hub registry with a username and no password:
// q-core checks registry credentials only when both are set, so the registry needs no real account.
func TestAcc_ContainerRegistryContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_container_registry.test"
	testName := "container-registry-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address: address,
		config:  func(declared bool) string { return testAccContainerRegistryContractConfig(testName, declared) },
		declared: []testAccContractValue{
			{"description", "declared"},
			{"config.username", "declared"},
		},
		defaults: []testAccContractValue{
			{"description", ""},
			{"config.%", nil},
		},
		outOfBand: func(id string) error {
			apiPath := fmt.Sprintf("/organization/%s/containerRegistry/%s", getTestOrganizationID(), id)
			return testAccEditServiceOutOfBand(apiPath, nil, func(registry map[string]any) {
				registry["description"] = "set from the Console"
				registry["config"] = map[string]any{"username": "console"}
			})
		},
		changed: []testAccContractValue{
			{"description", "set from the Console"},
			{"config.username", "console"},
		},
		importStateIDPrefix: getTestOrganizationID() + ",",
		exists:              testAccQoveryContainerRegistryExists(address),
		destroy:             testAccQoveryContainerRegistryDestroy(address),
	})
}

func TestAcc_ContainerRegistryContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testAccServiceContractUpgradeFrom0x(t, testAccContainerRegistryDefaultConfig("container-registry-contract-upgrade"), testAccQoveryContainerRegistryDestroy("qovery_container_registry.test"))
}

func testAccContainerRegistryContractConfig(testName string, declared bool) string {
	attributes := ""
	if declared {
		attributes = `
  description     = "declared"
  config = {
    username = "declared"
  }`
	}
	return fmt.Sprintf(`
resource "qovery_container_registry" "test" {
  organization_id = "%s"
  name            = "%s"
  kind            = "DOCKER_HUB"
  url             = "https://docker.io"%s
}
`, getTestOrganizationID(), generateTestName(testName), attributes)
}

// --- helm repository -----------------------------------------------------------------------------

// TestAcc_HelmRepositoryContract covers the description of an HTTPS repository without
// credentials. The config block is covered by TestAcc_HelmRepository: q-core checks the index or
// the credentials of a helm repository on every write, so a config change needs a real account.
func TestAcc_HelmRepositoryContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_helm_repository.test"
	testName := "helm-repository-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address:  address,
		config:   func(declared bool) string { return testAccHelmRepositoryContractConfig(testName, declared) },
		declared: []testAccContractValue{{"description", "declared"}},
		defaults: []testAccContractValue{{"description", ""}},
		outOfBand: func(id string) error {
			apiPath := fmt.Sprintf("/organization/%s/helmRepository/%s", getTestOrganizationID(), id)
			return testAccEditServiceOutOfBand(apiPath, nil, func(repository map[string]any) {
				repository["description"] = "set from the Console"
			})
		},
		changed:             []testAccContractValue{{"description", "set from the Console"}},
		importStateIDPrefix: getTestOrganizationID() + ",",
		exists:              testAccQoveryHelmRepositoryExists(address),
		destroy:             testAccQoveryHelmRepositoryDestroy(address),
	})
}

func TestAcc_HelmRepositoryContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testAccServiceContractUpgradeFrom0x(t, testAccHelmRepositoryDefaultConfig("helm-repository-contract-upgrade"), testAccQoveryHelmRepositoryDestroy("qovery_helm_repository.test"))
}

func testAccHelmRepositoryContractConfig(testName string, declared bool) string {
	description := ""
	if declared {
		description = `
  description           = "declared"`
	}
	return fmt.Sprintf(`
resource "qovery_helm_repository" "test" {
  organization_id       = "%s"
  name                  = "%s"
  kind                  = "HTTPS"
  url                   = "https://gitlab.com/mulesoft-int/helm-repository/-/raw/master/"
  skip_tls_verification = false%s
}
`, getTestOrganizationID(), generateTestName(testName), description)
}
