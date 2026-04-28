//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAcc_Helm_ApplyConsistencyOnSourceChange exercises the `source` arm
// of the `useStateUnlessNameChangesModifier` invalidation list. If the
// modifier reuses the cached `built_in_environment_variables` list across
// a source change, apply diverges from plan and the framework errors with
// "Provider produced inconsistent result after apply".
//
// Note: depending on the Qovery API's behaviour, changing only the chart
// version may or may not actually mutate the built-in env var list — if it
// doesn't, the test passes vacuously. The point of having it is to catch
// any future regression that breaks the modifier's invalidation logic on
// `source` changes.
//
// If the chart version below stops being available in the upstream helm
// repo, replace it with any other version returned by `helm search`.
//
// Companion: see resource_application/container/job_apply_consistency_test.go.
func TestAcc_Helm_ApplyConsistencyOnSourceChange(t *testing.T) {
	t.Parallel()
	testName := "helm-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryHelmDestroy("qovery_helm.test"),
		Steps: []resource.TestStep{
			{Config: testAccHelmApplyConsistencyConfig(testName, "1.0.0")},
			{Config: testAccHelmApplyConsistencyConfig(testName, "0.1.0")},
		},
	})
}

func testAccHelmApplyConsistencyConfig(testName, chartVersion string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_helm" "test" {
  environment_id               = qovery_environment.test.id
  name                         = "%s"
  timeout_sec                  = 600
  auto_preview                 = false
  auto_deploy                  = false
  allow_cluster_wide_resources = false

  source = {
    helm_repository = {
      helm_repository_id = qovery_helm_repository.test.id
      chart_name         = "httpbin"
      chart_version      = "%s"
    }
  }

  values_override = {}
}
`,
		testAccEnvironmentDefaultConfig(testName),
		testAccHelmRepositoryConfig(testName, "https://gitlab.com/mulesoft-int/helm-repository/-/raw/master/", "HTTPS"),
		generateTestName(testName),
		chartVersion,
	)
}
