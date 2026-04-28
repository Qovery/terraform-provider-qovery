//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with the helm `source` mutated. A regression in
// useStateUnlessNameChangesModifier surfaces here as "Provider produced
// inconsistent result after apply". If the chart version below stops being
// available upstream, swap it for any other version from `helm search`.
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
