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

func TestAcc_GcpCredentials(t *testing.T) {
	t.Parallel()
	testName := "gcp-credentials"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryGCPCredentialsDestroy("qovery_gcp_credentials.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGCPCredentialsDefaultConfig(
					testName,
					getTestGCPCredentials(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGCPCredentialsExists("qovery_gcp_credentials.test"),
					resource.TestCheckResourceAttr("qovery_gcp_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_gcp_credentials.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_gcp_credentials.test", "gcp_credentials", getTestGCPCredentials()),
				),
			},
			// Update name
			{
				Config: testAccGCPCredentialsDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					getTestGCPCredentials(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryGCPCredentialsExists("qovery_gcp_credentials.test"),
					resource.TestCheckResourceAttr("qovery_gcp_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_gcp_credentials.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttr("qovery_gcp_credentials.test", "gcp_credentials", getTestGCPCredentials()),
				),
			},
			// Check Import
			{
				ResourceName:            "qovery_gcp_credentials.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"gcp_credentials"},
			},
		},
	})
}

func testAccQoveryGCPCredentialsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("gcp_credentials not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("gcp_credentials.id not found")
		}

		_, err := qoveryServices.CredentialsGcp.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryGCPCredentialsDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("gcp_credentials not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("gcp_credentials.id not found")
		}

		_, err := qoveryServices.CredentialsGcp.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found gcp_credentials but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted gcp_credentials: %s", err.Error())
		}
		return nil
	}
}

func testAccGCPCredentialsDefaultConfig(testName string, gcpCredentials string) string {
	return fmt.Sprintf(`
resource "qovery_gcp_credentials" "test" {
  organization_id = "%s"
  name = "%s"
  gcp_credentials = chomp(<<-EOF
%s
EOF
  )
}
`, getTestOrganizationID(), generateTestName(testName), gcpCredentials,
	)
}
