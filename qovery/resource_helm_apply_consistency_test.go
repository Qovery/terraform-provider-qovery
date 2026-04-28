//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with `name` mutated to verify useStateUnlessNameChangesModifier
// invalidates cached built_in_environment_variables on a value-affecting
// attribute change. Source-change E2E coverage is unreliable here (the
// upstream helm repo carries a single chart version, and git_repository
// state isn't fully controlled); covered by unit tests in
// plan_modifiers_test.go instead.
func TestAcc_Helm_ApplyConsistencyOnNameChange(t *testing.T) {
	t.Parallel()
	testName := "helm-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryHelmDestroy("qovery_helm.test"),
		Steps: []resource.TestStep{
			{Config: testAccHelmApplyConsistencyConfig(testName, generateTestName(testName))},
			{Config: testAccHelmApplyConsistencyConfig(testName, generateTestName(testName)+"-renamed")},
		},
	})
}

func testAccHelmApplyConsistencyConfig(testName, name string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_helm" "test" {
  environment_id               = qovery_environment.test.id
  name                         = "%s"
  description                  = "apply consistency regression test"
  timeout_sec                  = 600
  auto_preview                 = false
  auto_deploy                  = false
  allow_cluster_wide_resources = false

  source = {
    helm_repository = {
      helm_repository_id = qovery_helm_repository.test.id
      chart_name         = "httpbin"
      chart_version      = "1.0.0"
    }
  }

  values_override = {}
}
`,
		testAccEnvironmentDefaultConfig(testName),
		testAccHelmRepositoryConfig(testName, "https://gitlab.com/mulesoft-int/helm-repository/-/raw/master/", "HTTPS"),
		name,
	)
}
