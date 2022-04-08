package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

func TestAcc_Project(t *testing.T) {
	t.Parallel()
	testAccProjectWithoutEnv(t)
	testAccProjectWithEnv(t)
}

func testAccProjectWithoutEnv(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectCongifWithoutEnv(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("qovery_project.test")),
			},
		},
	})
}

func testAccProjectWithEnv(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy("qovery_project.test_env"),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectCongifWithoutEnv(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("qovery_project.test_env")),
			},
		},
	})
}

func testAccProjectCongifWithoutEnv() string {
	return fmt.Sprintf(`
%s

resource "qovery_project" "test" {
	organization_id       = "%s"
	name                  = "project-test-without-env"
}`, testAccBaseConfig, os.Getenv("ORG_ID"))
}

func testAccProjectConfigEnv() string {
	return fmt.Sprintf(`
%s

resource "qovery_project" "test_env" {
	organization_id       = "%s"
	name                  = "project-test-with-env"
	environment_variables = []
}`, testAccBaseConfig, os.Getenv("ORG_ID"))
}

func testAccProjectExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no projectID is set")
		}

		c := testClient(os.Getenv("API_TOKEN"), "dev")
		_, _, err := c.API.ProjectsApi.ListProject(context.TODO(), rs.Primary.Attributes["organization_id"]).Execute()
		return err
	}
}

func testAccProjectDestroy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no projectID is set")
		}

		c := testClient(os.Getenv("API_TOKEN"), "dev")
		_, resp, err := c.API.ProjectsApi.ListProject(context.TODO(), rs.Primary.ID).Execute()

		if err == nil {
			return fmt.Errorf("Found project but expected it to have been deleted")
		}
		if err != nil {
			if resp.StatusCode == 404 {
				return nil
			}
			return fmt.Errorf("Unexpected error checking for deleted project: %s", resp.Status)
		}

		return err
	}
}
