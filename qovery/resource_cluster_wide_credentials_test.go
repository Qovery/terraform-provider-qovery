//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"
)

const testAccSlackWebhookURL = "https://hooks.slack.com/services/T00000000/B00000000/terraform-provider-test"

// TestAcc_ClusterWideCredentials covers using_wide_credentials on an AWS EKS cluster in READY state
// (no cloud infra is provisioned). q-core only allows restricted permissions for an allowlisted
// organization, and the wide-scope account must be AWS credentials with an assume role (STS): the
// test needs TEST_AWS_ROLE_WIDE_CREDENTIALS_ID.
func TestAcc_ClusterWideCredentials(t *testing.T) {
	if getTestAWSWideCredentialsID() == "" {
		t.Skip("TEST_AWS_ROLE_WIDE_CREDENTIALS_ID is not set")
	}
	t.Parallel()
	testName := "cluster-wide-credentials"
	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create with restricted permissions and a slack webhook
			{
				Config:           testAccClusterConfigWithWideCredentials(testName, true, testAccSlackWebhookURL),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					testAccCaptureResourceID("qovery_cluster.test", &clusterID),
					resource.TestCheckResourceAttr("qovery_cluster.test", "using_wide_credentials.wide_scope_cloud_account_id", getTestAWSWideCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "using_wide_credentials.slack_webhook_url", testAccSlackWebhookURL),
					resource.TestCheckResourceAttr("qovery_cluster.test", "using_wide_credentials.require_wide_access_to_deploy", "true"),
				),
			},
			// The webhook removed outside Terraform shows on refresh, and the apply sets it back
			{
				Config: testAccClusterConfigWithWideCredentials(testName, true, testAccSlackWebhookURL),
				Check: resource.ComposeTestCheckFunc(
					testAccSetClusterSlackWebhookOutOfBand(&clusterID, nil),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:           testAccClusterConfigWithWideCredentials(testName, true, testAccSlackWebhookURL),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_cluster.test", "using_wide_credentials.slack_webhook_url", testAccSlackWebhookURL),
					testAccCheckClusterWidePermissionsInAPI(&clusterID, true, true),
				),
			},
			// Removing slack_webhook_url from the config removes the webhook
			{
				Config: testAccClusterConfigWithWideCredentials(testName, true, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("qovery_cluster.test", plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "using_wide_credentials.slack_webhook_url"),
					testAccCheckClusterWidePermissionsInAPI(&clusterID, true, false),
				),
			},
			// Import
			{
				ResourceName:            "qovery_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"advanced_settings_json", "min_running_nodes", "max_running_nodes"},
			},
			// Removing the block turns restricted permissions off
			{
				Config: testAccClusterConfigWithWideCredentials(testName, false, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("qovery_cluster.test", plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "using_wide_credentials"),
					testAccCheckClusterWidePermissionsInAPI(&clusterID, false, false),
				),
			},
			// Adding it back to an existing cluster sets the wide-scope account through the cloud
			// provider info. require_wide_access_to_deploy has a null prior value then: it must be
			// planned unknown, as the edit sets it to false.
			{
				Config: testAccClusterConfigWithWideCredentials(testName, true, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("qovery_cluster.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue("qovery_cluster.test", tfjsonpath.New("using_wide_credentials").AtMapKey("require_wide_access_to_deploy")),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_cluster.test", "using_wide_credentials.wide_scope_cloud_account_id", getTestAWSWideCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "using_wide_credentials.require_wide_access_to_deploy", "false"),
					testAccCheckClusterWidePermissionsInAPI(&clusterID, true, false),
				),
			},
		},
	})
}

// testAccSetClusterSlackWebhookOutOfBand sets the slack webhook of a cluster with restricted
// permissions through the API, nil to remove it.
func testAccSetClusterSlackWebhookOutOfBand(clusterID *string, slackWebhookURL *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		return testAccEditClusterOutOfBand(*clusterID, nil, func(request *qovery.ClusterRequest) {
			widePermissions := qovery.WidePermissionsRequestDto{UsingRestrictivePermissions: qovery.PtrBool(true)}
			if slackWebhookURL != nil {
				widePermissions.SetSlackWebhookUrl(*slackWebhookURL)
			}
			request.WidePermissions = &widePermissions
		})
	}
}

// testAccCheckClusterWidePermissionsInAPI checks the restricted permissions the API holds for the
// cluster, and whether it has a slack webhook.
func testAccCheckClusterWidePermissionsInAPI(clusterID *string, restricted bool, hasSlackWebhook bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		cluster, err := testAccReadClusterFromAPI(*clusterID)
		if err != nil {
			return err
		}
		widePermissions := cluster.GetWidePermissions()
		if got := widePermissions.GetUsingRestrictivePermissions(); got != restricted {
			return fmt.Errorf("expected using_restrictive_permissions %t, got %t", restricted, got)
		}
		wideScopeCloudAccountID := widePermissions.GetWideScopeCloudAccountId()
		if restricted && wideScopeCloudAccountID != getTestAWSWideCredentialsID() {
			return fmt.Errorf("expected wide_scope_cloud_account_id %q, got %q", getTestAWSWideCredentialsID(), wideScopeCloudAccountID)
		}
		if !restricted && wideScopeCloudAccountID != "" {
			return fmt.Errorf("expected no wide_scope_cloud_account_id, got %q", wideScopeCloudAccountID)
		}
		if got := widePermissions.GetSlackWebhookUrl() != ""; got != hasSlackWebhook {
			return fmt.Errorf("expected a slack webhook: %t, got: %t", hasSlackWebhook, got)
		}
		return nil
	}
}

func testAccClusterConfigWithWideCredentials(testName string, withWideCredentials bool, slackWebhookURL string) string {
	wideCredentials := ""
	if withWideCredentials {
		slackWebhookLine := ""
		if slackWebhookURL != "" {
			slackWebhookLine = fmt.Sprintf("\n    slack_webhook_url           = %q", slackWebhookURL)
		}
		wideCredentials = fmt.Sprintf(`
  using_wide_credentials = {
    wide_scope_cloud_account_id = %q%s
  }
`, getTestAWSWideCredentialsID(), slackWebhookLine)
	}
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"
%s
  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large"] },
          { key = "InstanceFamily", operator = "In", values = ["t3a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), wideCredentials)
}
