//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"
)

const containerGitSourceAddress = "qovery_container.test"

func TestAcc_ContainerSwitchToGitRepositoryAndBack(t *testing.T) {
	t.Parallel()
	testName := "container-git-source"
	var containerID string
	sameID := statecheck.CompareValue(compare.ValuesSame())
	// CompareValue checks one recorded value per AddStateValue call, so each step needs its own call.
	keepsID := func() []statecheck.StateCheck {
		return []statecheck.StateCheck{sameID.AddStateValue(containerGitSourceAddress, tfjsonpath.New("id"))}
	}
	inPlaceUpdate := resource.ConfigPlanChecks{
		PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(containerGitSourceAddress, plancheck.ResourceActionUpdate)},
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy(containerGitSourceAddress),
		Steps: []resource.TestStep{
			{
				Config:            testAccContainerDefaultConfig(testName),
				ConfigPlanChecks:  testAccEmptyPlanAfterApply,
				ConfigStateChecks: keepsID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(containerGitSourceAddress),
					testAccCaptureResourceID(containerGitSourceAddress, &containerID),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "image_name", containerImageName),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "true"),
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "source.git_repository.url"),
				),
			},
			{
				Config:            testAccContainerGitSourceConfig(testName, containerGitSourceBuildSettings),
				ConfigPlanChecks:  inPlaceUpdate,
				ConfigStateChecks: keepsID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(containerGitSourceAddress),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.branch", "master"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.root_path", "/"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.git_token_id", getTestQoverySandboxGitTokenID()),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "source.git_repository.provider"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "build_settings.timeout_max_sec", "900"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_preview", "false"),
					resource.TestCheckResourceAttrSet(containerGitSourceAddress, "registry_id"),
					resource.TestCheckResourceAttrSet(containerGitSourceAddress, "image_name"),
					resource.TestCheckResourceAttrSet(containerGitSourceAddress, "tag"),
				),
			},
			// A source changed outside Terraform shows as a difference on refresh, and an apply restores it.
			{
				PreConfig: func() {
					err := testAccEditServiceOutOfBand("/container/"+containerID, withoutContainerImageFields, func(container map[string]any) {
						git := container["source"].(map[string]any)["git_repository"].(map[string]any)
						container["source"] = map[string]any{"git_repository": map[string]any{
							"url":             git["url"],
							"branch":          "test",
							"root_path":       git["root_path"],
							"git_token_id":    git["git_token_id"],
							"provider":        git["provider"],
							"dockerfile_path": git["dockerfile_path"],
						}}
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config:             testAccContainerGitSourceConfig(testName, containerGitSourceBuildSettings),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:            testAccContainerGitSourceConfig(testName, containerGitSourceBuildSettings),
				ConfigPlanChecks:  inPlaceUpdate,
				ConfigStateChecks: keepsID(),
				Check:             resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.branch", "master"),
			},
			// Build settings changed outside Terraform show as a difference on refresh.
			{
				PreConfig: func() {
					err := testAccEditServiceOutOfBand("/container/"+containerID, func(container map[string]any) {
						withoutContainerImageFields(container)
						delete(container, "source")
					}, func(container map[string]any) {
						container["build_settings"].(map[string]any)["timeout_max_sec"] = 600
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config:             testAccContainerGitSourceConfig(testName, containerGitSourceBuildSettings),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Removing build_settings from the config resets the build settings.
			{
				Config:            testAccContainerGitSourceConfig(testName, ""),
				ConfigPlanChecks:  inPlaceUpdate,
				ConfigStateChecks: keepsID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.url", applicationRepositoryURL),
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "build_settings.timeout_max_sec"),
				),
			},
			{
				ResourceName:      containerGitSourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:            testAccContainerDefaultConfig(testName),
				ConfigPlanChecks:  inPlaceUpdate,
				ConfigStateChecks: keepsID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(containerGitSourceAddress),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "image_name", containerImageName),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "tag", containerTag),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "true"),
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "source.git_repository.url"),
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "build_settings.timeout_max_sec"),
				),
			},
			{
				Config:   testAccContainerDefaultConfig(testName),
				PlanOnly: true,
			},
		},
	})
}

func TestAcc_ContainerCreateWithGitRepositorySource(t *testing.T) {
	t.Parallel()
	testName := "container-git-source-create"
	sameID := statecheck.CompareValue(compare.ValuesSame())
	// CompareValue checks one recorded value per AddStateValue call, so each step needs its own call.
	keepsID := func() []statecheck.StateCheck {
		return []statecheck.StateCheck{sameID.AddStateValue(containerGitSourceAddress, tfjsonpath.New("id"))}
	}
	inPlaceUpdate := resource.ConfigPlanChecks{
		PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(containerGitSourceAddress, plancheck.ResourceActionUpdate)},
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy(containerGitSourceAddress),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerGitSourceConfig(testName, containerGitSourceBuildSettings),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(containerGitSourceAddress, plancheck.ResourceActionCreate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: keepsID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(containerGitSourceAddress),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.branch", "master"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.root_path", "/"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "build_settings.timeout_max_sec", "900"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_preview", "false"),
					resource.TestCheckResourceAttrSet(containerGitSourceAddress, "registry_id"),
					resource.TestCheckResourceAttrSet(containerGitSourceAddress, "image_name"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "tag", "pending-first-build"),
				),
			},
			{
				Config:   testAccContainerGitSourceConfig(testName, containerGitSourceBuildSettings),
				PlanOnly: true,
			},
			{
				ResourceName:      containerGitSourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
				// An import leaves build_settings null: the block is tracked only once the config sets it.
				ImportStateVerifyIgnore: []string{"build_settings"},
			},
			{
				Config:            testAccContainerDefaultConfig(testName),
				ConfigPlanChecks:  inPlaceUpdate,
				ConfigStateChecks: keepsID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(containerGitSourceAddress, "image_name", containerImageName),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "tag", containerTag),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "true"),
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "source.git_repository.url"),
				),
			},
			{
				Config:            testAccContainerGitSourceConfig(testName, containerGitSourceBuildSettings),
				ConfigPlanChecks:  inPlaceUpdate,
				ConfigStateChecks: keepsID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "build_settings.timeout_max_sec", "900"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "false"),
				),
			},
		},
	})
}

func TestAcc_ContainerGitRepositorySourceAutoFlags(t *testing.T) {
	t.Parallel()
	testName := "container-git-source-auto-flags"
	var containerID string
	autoFlags := func(value bool) []plancheck.PlanCheck {
		return []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(containerGitSourceAddress, plancheck.ResourceActionUpdate),
			plancheck.ExpectKnownValue(containerGitSourceAddress, tfjsonpath.New("auto_deploy"), knownvalue.Bool(value)),
			plancheck.ExpectKnownValue(containerGitSourceAddress, tfjsonpath.New("auto_preview"), knownvalue.Bool(value)),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy(containerGitSourceAddress),
		Steps: []resource.TestStep{
			{
				Config:           testAccContainerGitSourceConfig(testName, containerGitSourceAutoFlags),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(containerGitSourceAddress),
					testAccCaptureResourceID(containerGitSourceAddress, &containerID),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "source.git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "true"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_preview", "true"),
				),
			},
			// Flags changed outside Terraform show as a difference on refresh, and an apply restores them.
			{
				PreConfig: func() {
					err := testAccEditServiceOutOfBand("/container/"+containerID, func(container map[string]any) {
						withoutContainerImageFields(container)
						delete(container, "source")
					}, func(container map[string]any) {
						container["auto_deploy"] = false
						container["auto_preview"] = false
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config:             testAccContainerGitSourceConfig(testName, containerGitSourceAutoFlags),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccContainerGitSourceConfig(testName, containerGitSourceAutoFlags),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             autoFlags(true),
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "true"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_preview", "true"),
				),
			},
			{
				ResourceName:      containerGitSourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Removing the flags from the config plans their default, false, and the apply sends false
			// explicitly: an update of a Git container without them would keep the stored true.
			{
				Config: testAccContainerGitSourceConfig(testName, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             autoFlags(false),
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "auto_preview", "false"),
				),
			},
		},
	})
}

func TestAcc_ContainerGitRepositorySourceDeploymentRestrictions(t *testing.T) {
	t.Parallel()
	testName := "container-git-source-restrictions"
	var containerID string
	inPlaceUpdate := resource.ConfigPlanChecks{
		PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(containerGitSourceAddress, plancheck.ResourceActionUpdate)},
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
	excludeDocs := resource.TestCheckTypeSetElemNestedAttrs(containerGitSourceAddress, "deployment_restrictions.*",
		map[string]string{"mode": "EXCLUDE", "type": "PATH", "value": "docs/"})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy(containerGitSourceAddress),
		Steps: []resource.TestStep{
			{
				Config:           testAccContainerGitSourceConfig(testName, containerGitSourceExcludeDocs),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(containerGitSourceAddress),
					testAccCaptureResourceID(containerGitSourceAddress, &containerID),
					resource.TestCheckResourceAttr(containerGitSourceAddress, "deployment_restrictions.#", "1"),
					excludeDocs,
					testAccContainerDeploymentRestrictionCount(1),
				),
			},
			// Changing a rule replaces it, and adding one creates it.
			{
				Config:           testAccContainerGitSourceConfig(testName, containerGitSourceTwoRules),
				ConfigPlanChecks: inPlaceUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(containerGitSourceAddress, "deployment_restrictions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(containerGitSourceAddress, "deployment_restrictions.*",
						map[string]string{"mode": "EXCLUDE", "type": "PATH", "value": "documentation/"}),
					resource.TestCheckTypeSetElemNestedAttrs(containerGitSourceAddress, "deployment_restrictions.*",
						map[string]string{"mode": "MATCH", "type": "PATH", "value": "src/"}),
					testAccContainerDeploymentRestrictionCount(2),
				),
			},
			// A rule added outside Terraform shows as a difference on refresh, and an apply deletes it.
			{
				PreConfig: func() {
					_, res, err := qoveryAPIClient.ServiceMainCallsAPI.CreateServiceDeploymentRestriction(context.TODO(), containerID).
						ServiceDeploymentRestrictionRequest(qovery.ServiceDeploymentRestrictionRequest{
							Mode:  qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH,
							Type:  qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH,
							Value: "api/",
						}).Execute()
					if err != nil || res.StatusCode >= 400 {
						t.Fatalf("cannot add a deployment restriction to container %s: %v", containerID, err)
					}
				},
				Config:             testAccContainerGitSourceConfig(testName, containerGitSourceTwoRules),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:           testAccContainerGitSourceConfig(testName, containerGitSourceTwoRules),
				ConfigPlanChecks: inPlaceUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(containerGitSourceAddress, "deployment_restrictions.#", "2"),
					testAccContainerDeploymentRestrictionCount(2),
				),
			},
			{
				ResourceName:      containerGitSourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Removing the block deletes the rules.
			{
				Config:           testAccContainerGitSourceConfig(testName, ""),
				ConfigPlanChecks: inPlaceUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "deployment_restrictions.#"),
					testAccContainerDeploymentRestrictionCount(0),
				),
			},
			{
				Config:           testAccContainerGitSourceConfig(testName, containerGitSourceExcludeDocs),
				ConfigPlanChecks: inPlaceUpdate,
				Check:            testAccContainerDeploymentRestrictionCount(1),
			},
			// A registry image takes no rules: the switch keeps them only until the block is removed.
			{
				Config:      testAccContainerImageConfigWithAttributes(testName, containerGitSourceExcludeDocs),
				ExpectError: regexp.MustCompile("deployment_restrictions needs source.git_repository"),
			},
			{
				Config:           testAccContainerDefaultConfig(testName),
				ConfigPlanChecks: inPlaceUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "source.git_repository.url"),
					resource.TestCheckNoResourceAttr(containerGitSourceAddress, "deployment_restrictions.#"),
					testAccContainerDeploymentRestrictionCount(0),
				),
			},
		},
	})
}

// testAccContainerDeploymentRestrictionCount checks how many deployment restrictions Qovery holds for the test container.
func testAccContainerDeploymentRestrictionCount(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[containerGitSourceAddress]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("%s: id not found in state", containerGitSourceAddress)
		}
		rules, res, err := qoveryAPIClient.ServiceMainCallsAPI.ListServiceDeploymentRestrictions(context.TODO(), rs.Primary.ID).Execute()
		if err != nil || res.StatusCode >= 400 {
			return fmt.Errorf("cannot list the deployment restrictions of container %s: %v", rs.Primary.ID, err)
		}
		if len(rules.Results) != expected {
			return fmt.Errorf("container %s: expected %d deployment restrictions in Qovery, got %d", rs.Primary.ID, expected, len(rules.Results))
		}
		return nil
	}
}

// withoutContainerImageFields drops the image of a container read from the API, which a request on a
// container with a Git repository source must not repeat next to a source.
func withoutContainerImageFields(container map[string]any) {
	for _, key := range []string{"registry", "image_name", "tag"} {
		delete(container, key)
	}
}

const containerGitSourceBuildSettings = `
  build_settings = {
    timeout_max_sec = 900
  }`

const containerGitSourceAutoFlags = `
  auto_deploy = true
  auto_preview = true`

const containerGitSourceExcludeDocs = `
  deployment_restrictions = [
    { mode = "EXCLUDE", type = "PATH", value = "docs/" },
  ]`

const containerGitSourceTwoRules = `
  deployment_restrictions = [
    { mode = "EXCLUDE", type = "PATH", value = "documentation/" },
    { mode = "MATCH", type = "PATH", value = "src/" },
  ]`

func testAccContainerImageConfigWithAttributes(testName string, attributes string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"%s
  healthchecks = {}
  advanced_settings_json = jsonencode({"network.ingress.proxy_body_size_mb": 101})
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, attributes,
	)
}

func testAccContainerGitSourceConfig(testName string, attributes string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  # The registry outlives the container: it runs the image again once switched back.
  depends_on = [qovery_container_registry.test]
  source = {
    git_repository = {
      url = "%s"
      branch = "master"
      git_token_id = "%s"
      dockerfile_path = "Dockerfile"
    }
  }%s
  healthchecks = {}
  advanced_settings_json = jsonencode({"network.ingress.proxy_body_size_mb": 101})
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(), attributes,
	)
}
