//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

const (
	ecrRegion = "eu-west-3"
)

func TestAcc_HelmRepository(t *testing.T) {
	t.Parallel()
	testName := "helm-repository"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryHelmRepositoryDestroy("qovery_helm_repository.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccHelmRepositoryDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryHelmRepositoryExists("qovery_helm_repository.test"),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "kind", "OCI_ECR"),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "url", strings.Replace(getTestAwsEcrURL(), "https", "oci", 1)),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "config.region", ecrRegion),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "config.access_key_id", getTestAWSCredentialsAccessKeyID()),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "config.secret_access_key", getTestAWSCredentialsSecretAccessKey()),
				),
			},
			// Add description
			{
				Config: testAccHelmRepositoryDefaultConfigWithDescription(
					testName,
					"this is a description",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryHelmRepositoryExists("qovery_helm_repository.test"),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "kind", "OCI_ECR"),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "url", strings.Replace(getTestAwsEcrURL(), "https", "oci", 1)),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "description", "this is a description"),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "config.region", ecrRegion),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "config.access_key_id", getTestAWSCredentialsAccessKeyID()),
					resource.TestCheckResourceAttr("qovery_helm_repository.test", "config.secret_access_key", getTestAWSCredentialsSecretAccessKey()),
				),
			},
			// Check Import
			{
				ResourceName:            "qovery_helm_repository.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"config"},
			},
		},
	})
}

func testAccQoveryHelmRepositoryExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("helm repository not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("helm_repository.id not found")
		}

		_, err := qoveryServices.HelmRepository.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryHelmRepositoryDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("helm repository not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("helm_repository.id not found")
		}

		_, err := qoveryServices.HelmRepository.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found helm repository but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted helm repository: %s", err.Error())
		}
		return nil
	}
}

func testAccHelmRepositoryDefaultConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_helm_repository" "test" {
  organization_id = "%s"
  name = "%s"
  kind = "OCI_ECR"
  url = "%s"
  config = {
    region = "%s"
    access_key_id = "%s"
    secret_access_key = "%s"
  }
  skip_tls_verification = false
}
`, getTestOrganizationID(), generateTestName(testName), strings.Replace(getTestAwsEcrURL(), "https://", "oci://", 1), awsECRRegion, getTestAWSCredentialsAccessKeyID(), getTestAWSCredentialsSecretAccessKey(),
	)
}

func testAccHelmRepositoryConfig(testName string, url string, kind string) string {
	return fmt.Sprintf(`
resource "qovery_helm_repository" "test" {
  organization_id = "%s"
  name = "%s"
  url = "%s"
  kind = "%s"
  skip_tls_verification = false
}
`, getTestOrganizationID(), generateTestName(testName), url, kind,
	)
}

func testAccHelmRepositoryDefaultConfigWithDescription(testName string, description string) string {
	return fmt.Sprintf(`
resource "qovery_helm_repository" "test" {
  organization_id = "%s"
  name = "%s"
  kind = "OCI_ECR"
  url = "%s"
  config = {
    region = "%s"
    access_key_id = "%s"
    secret_access_key = "%s"
  }
  description = "%s"
  skip_tls_verification = false
}
`, getTestOrganizationID(), generateTestName(testName), strings.Replace(getTestAwsEcrURL(), "https://", "oci://", 1), ecrRegion, getTestAWSCredentialsAccessKeyID(), getTestAWSCredentialsSecretAccessKey(), description,
	)
}
