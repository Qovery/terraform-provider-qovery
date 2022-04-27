package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestAcc_AWSCredentials(t *testing.T) {
	t.Parallel()
	testName := "aws-credentials"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryAWSCredentialsDestroy("qovery_aws_credentials.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAWSCredentialsDefaultConfig(
					testName,
					getTestAWSCredentialsAccessKeyID(),
					getTestAWSCredentialsSecretAccessKey(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryAWSCredentialsExists("qovery_aws_credentials.test"),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "access_key_id", getTestAWSCredentialsAccessKeyID()),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "secret_access_key", getTestAWSCredentialsSecretAccessKey()),
				),
			},
			// Update name
			{
				Config: testAccAWSCredentialsDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					getTestAWSCredentialsAccessKeyID(),
					getTestAWSCredentialsSecretAccessKey(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryAWSCredentialsExists("qovery_aws_credentials.test"),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "access_key_id", getTestAWSCredentialsAccessKeyID()),
					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "secret_access_key", getTestAWSCredentialsSecretAccessKey()),
				),
			},
		},
	})
}

// TODO: uncomment when ImportStateIdPrefix is fixed
//func TestAcc_AWSCredentialsImport(t *testing.T) {
//	t.Parallel()
//	testName := "aws-credentials-import"
//	resource.Test(t, resource.TestCase{
//		PreCheck:                 func() { testAccPreCheck(t) },
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		CheckDestroy:             testAccQoveryAWSCredentialsDestroy("qovery_aws_credentials.test"),
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: testAccAWSCredentialsDefaultConfig(
//					testName,
//					getTestAWSCredentialsAccessKeyID(),
//					getTestAWSCredentialsSecretAccessKey(),
//				),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccQoveryAWSCredentialsExists("qovery_aws_credentials.test"),
//					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "organization_id", getTestOrganizationID()),
//					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "name", generateTestName(testNamePrefix)),
//					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "access_key_id", getTestAWSCredentialsAccessKeyID()),
//					resource.TestCheckResourceAttr("qovery_aws_credentials.test", "secret_access_key", getTestAWSCredentialsSecretAccessKey()),
//				),
//			},
//			// Check Import
//			{
//				ResourceName:            "qovery_aws_credentials.test",
//				ImportState:             true,
//				ImportStateVerify:       true,
//				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
//				ImportStateVerifyIgnore: []string{"access_key_id", "secret_access_key"},
//			},
//		},
//	})
//}

func testAccQoveryAWSCredentialsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("aws_credentials not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("aws_credentials.id not found")
		}

		_, apiErr := apiClient.GetAWSCredentials(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if apiErr != nil {
			return apiErr
		}
		return nil
	}
}

func testAccQoveryAWSCredentialsDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("aws_credentials not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("aws_credentials.id not found")
		}

		_, apiErr := apiClient.GetAWSCredentials(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if apiErr == nil {
			return fmt.Errorf("found aws_credentials but expected it to be deleted")
		}
		if !apierrors.IsNotFound(apiErr) {
			return fmt.Errorf("unexpected error checking for deleted aws_credentials: %s", apiErr.Summary())
		}
		return nil
	}
}

func testAccAWSCredentialsDefaultConfig(testName string, accessKeyID string, secretAccessKey string) string {
	return fmt.Sprintf(`
resource "qovery_aws_credentials" "test" {
  organization_id = "%s"
  name = "%s"
  access_key_id = "%s"
  secret_access_key = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), accessKeyID, secretAccessKey,
	)
}
