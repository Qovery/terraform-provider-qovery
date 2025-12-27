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

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

func TestAcc_GitToken(t *testing.T) {
	t.Parallel()
	testName := "git-token"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryGitTokenDestroy("qovery_git_token.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGitTokenDefaultConfig(testName, "GITHUB"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGitTokenExists("qovery_git_token.test"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_git_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_git_token.test", "type", "GITHUB"),
					resource.TestCheckResourceAttrSet("qovery_git_token.test", "token"),
				),
			},
			// Update name
			{
				Config: testAccGitTokenDefaultConfig(testName+"-updated", "GITHUB"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGitTokenExists("qovery_git_token.test"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_git_token.test", "name", generateTestName(testName+"-updated")),
					resource.TestCheckResourceAttr("qovery_git_token.test", "type", "GITHUB"),
				),
			},
			// Check Import (token field is sensitive, so use ImportStateVerifyIgnore)
			{
				ResourceName:            "qovery_git_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAcc_GitTokenGitHub(t *testing.T) {
	t.Parallel()
	testName := "git-token-github"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryGitTokenDestroy("qovery_git_token.test"),
		Steps: []resource.TestStep{
			// Create GitHub token
			{
				Config: testAccGitTokenDefaultConfig(testName, "GITHUB"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGitTokenExists("qovery_git_token.test"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_git_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_git_token.test", "type", "GITHUB"),
				),
			},
		},
	})
}

func TestAcc_GitTokenGitLab(t *testing.T) {
	t.Parallel()
	testName := "git-token-gitlab"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryGitTokenDestroy("qovery_git_token.test"),
		Steps: []resource.TestStep{
			// Create GitLab token
			{
				Config: testAccGitTokenDefaultConfig(testName, "GITLAB"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGitTokenExists("qovery_git_token.test"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_git_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_git_token.test", "type", "GITLAB"),
				),
			},
		},
	})
}

func TestAcc_GitTokenWithDescription(t *testing.T) {
	t.Parallel()
	testName := "git-token-desc"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryGitTokenDestroy("qovery_git_token.test"),
		Steps: []resource.TestStep{
			// Create with description
			{
				Config: testAccGitTokenConfigWithDescription(testName, "GITHUB", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGitTokenExists("qovery_git_token.test"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_git_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_git_token.test", "type", "GITHUB"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "description", "Initial description"),
				),
			},
			// Update description
			{
				Config: testAccGitTokenConfigWithDescription(testName, "GITHUB", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGitTokenExists("qovery_git_token.test"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_git_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_git_token.test", "type", "GITHUB"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAcc_GitToken_Import(t *testing.T) {
	t.Parallel()
	testName := "git-token-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryGitTokenDestroy("qovery_git_token.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGitTokenDefaultConfig(testName, "GITHUB"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGitTokenExists("qovery_git_token.test"),
					resource.TestCheckResourceAttr("qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_git_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_git_token.test", "type", "GITHUB"),
				),
			},
			// Check Import (token field is sensitive, so use ImportStateVerifyIgnore)
			{
				ResourceName:            "qovery_git_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccQoveryGitTokenExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("git token not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("git_token.id not found")
		}

		_, err := qoveryServices.GitToken.Get(context.TODO(), rs.Primary.Attributes["organization_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryGitTokenDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("git token not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("git_token.id not found")
		}

		_, err := qoveryServices.GitToken.Get(context.TODO(), rs.Primary.Attributes["organization_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found git token but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted git token: %s", err.Error())
		}
		return nil
	}
}

func testAccGitTokenDefaultConfig(testName string, tokenType string) string {
	return fmt.Sprintf(`
resource "qovery_git_token" "test" {
  organization_id = "%s"
  name            = "%s"
  type            = "%s"
  token           = "ghp_test_token_value_for_testing_purposes_only"
}
`, getTestOrganizationID(), generateTestName(testName), tokenType)
}

func testAccGitTokenConfigWithDescription(testName string, tokenType string, description string) string {
	return fmt.Sprintf(`
resource "qovery_git_token" "test" {
  organization_id = "%s"
  name            = "%s"
  type            = "%s"
  token           = "ghp_test_token_value_for_testing_purposes_only"
  description     = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), tokenType, description)
}
