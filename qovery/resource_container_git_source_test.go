//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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

func testAccContainerGitSourceConfig(testName string, buildSettings string) string {
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
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(), buildSettings,
	)
}
