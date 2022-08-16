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

func TestAcc_ScalewayCredentials(t *testing.T) {
	t.Parallel()
	testName := "scaleway-credentials"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryScalewayCredentialsDestroy("qovery_scaleway_credentials.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccScalewayCredentialsDefaultConfig(
					testName,
					getTestScalewayCredentialsAccessKey(),
					getTestScalewayCredentialsSecretKey(),
					getTestScalewayCredentialsProjectID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryScalewayCredentialsExists("qovery_scaleway_credentials.test"),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_access_key", getTestScalewayCredentialsAccessKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_secret_key", getTestScalewayCredentialsSecretKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_project_id", getTestScalewayCredentialsProjectID()),
				),
			},
			// Update name
			{
				Config: testAccScalewayCredentialsDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					getTestScalewayCredentialsAccessKey(),
					getTestScalewayCredentialsSecretKey(),
					getTestScalewayCredentialsProjectID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryScalewayCredentialsExists("qovery_scaleway_credentials.test"),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_access_key", getTestScalewayCredentialsAccessKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_secret_key", getTestScalewayCredentialsSecretKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_project_id", getTestScalewayCredentialsProjectID()),
				),
			},
			// Check Import
			{
				ResourceName:            "qovery_scaleway_credentials.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"scaleway_access_key", "scaleway_secret_key", "scaleway_project_id"},
			},
		},
	})
}

func testAccQoveryScalewayCredentialsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("scaleway_credentials not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("scaleway_credentials.id not found")
		}

		_, err := qoveryServices.CredentialsScaleway.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryScalewayCredentialsDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("scaleway_credentials not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("scaleway_credentials.id not found")
		}

		_, err := qoveryServices.CredentialsScaleway.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found scaleway_credentials but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted scaleway_credentials: %s", err.Error())
		}
		return nil
	}
}

func testAccScalewayCredentialsDefaultConfig(testName string, accessKey string, secretKey string, projectID string) string {
	return fmt.Sprintf(`
resource "qovery_scaleway_credentials" "test" {
  organization_id = "%s"
  name = "%s"
  scaleway_access_key = "%s"
  scaleway_secret_key = "%s"
  scaleway_project_id = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), accessKey, secretKey, projectID,
	)
}
