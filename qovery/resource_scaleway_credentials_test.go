//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

func TestAcc_ScalewayCredentials(t *testing.T) {
	t.Parallel()
	testName := "scaleway-credentials"
	var credentialsID string
	// q-core checks only the secret key of Scaleway credentials, so another project ID is accepted.
	outOfBandProjectID := uuid.NewString()
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
					getTestScalewayCredentialsOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryScalewayCredentialsExists("qovery_scaleway_credentials.test"),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_access_key", getTestScalewayCredentialsAccessKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_secret_key", getTestScalewayCredentialsSecretKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_project_id", getTestScalewayCredentialsProjectID()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_organization_id", getTestScalewayCredentialsOrganizationID()),
				),
			},
			// Update name
			{
				Config: testAccScalewayCredentialsDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					getTestScalewayCredentialsAccessKey(),
					getTestScalewayCredentialsSecretKey(),
					getTestScalewayCredentialsProjectID(),
					getTestScalewayCredentialsOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryScalewayCredentialsExists("qovery_scaleway_credentials.test"),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_access_key", getTestScalewayCredentialsAccessKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_secret_key", getTestScalewayCredentialsSecretKey()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_project_id", getTestScalewayCredentialsProjectID()),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_organization_id", getTestScalewayCredentialsOrganizationID()),
					testAccCaptureResourceID("qovery_scaleway_credentials.test", &credentialsID),
				),
			},
			// Import records the identifiers
			{
				ResourceName:            "qovery_scaleway_credentials.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"scaleway_secret_key"},
			},
			// A project changed outside Terraform shows up in the plan...
			{
				Config: testAccScalewayCredentialsDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					getTestScalewayCredentialsAccessKey(),
					getTestScalewayCredentialsSecretKey(),
					getTestScalewayCredentialsProjectID(),
					getTestScalewayCredentialsOrganizationID(),
				),
				Check: func(_ *terraform.State) error {
					apiPath := fmt.Sprintf("/organization/%s/scaleway/credentials/%s", getTestOrganizationID(), credentialsID)
					return testAccEditServiceOutOfBand(apiPath, nil, func(creds map[string]any) {
						creds["scaleway_project_id"] = outOfBandProjectID
						// The API never returns the secret key.
						creds["scaleway_secret_key"] = getTestScalewayCredentialsSecretKey()
					})
				},
				ExpectNonEmptyPlan: true,
			},
			// ... the refresh stores it, and the secret key stays in the state...
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_project_id", outOfBandProjectID),
					resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_secret_key", getTestScalewayCredentialsSecretKey()),
				),
				ExpectNonEmptyPlan: true,
			},
			// ... and the next apply reverts it.
			{
				Config: testAccScalewayCredentialsDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					getTestScalewayCredentialsAccessKey(),
					getTestScalewayCredentialsSecretKey(),
					getTestScalewayCredentialsProjectID(),
					getTestScalewayCredentialsOrganizationID(),
				),
				Check:            resource.TestCheckResourceAttr("qovery_scaleway_credentials.test", "scaleway_project_id", getTestScalewayCredentialsProjectID()),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

func TestAcc_ScalewayCredentialsUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testAccServiceContractUpgradeFrom0x(t, testAccScalewayCredentialsDefaultConfig(
		"scaleway-credentials-upgrade",
		getTestScalewayCredentialsAccessKey(),
		getTestScalewayCredentialsSecretKey(),
		getTestScalewayCredentialsProjectID(),
		getTestScalewayCredentialsOrganizationID(),
	), testAccQoveryScalewayCredentialsDestroy("qovery_scaleway_credentials.test"))
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

func testAccScalewayCredentialsDefaultConfig(testName string, accessKey string, secretKey string, projectID string, organizationID string) string {
	return fmt.Sprintf(`
resource "qovery_scaleway_credentials" "test" {
  organization_id = "%s"
  name = "%s"
  scaleway_access_key = "%s"
  scaleway_secret_key = "%s"
  scaleway_project_id = "%s"
  scaleway_organization_id = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), accessKey, secretKey, projectID, organizationID,
	)
}
