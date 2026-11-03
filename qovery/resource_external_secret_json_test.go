//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/qovery/qovery-client-go"
)

// TEST_SECRET_MANAGER_ACCESS_ID must name an access configured on the test cluster.
// Resolution happens at deployment, so these lifecycle tests do not read a real secret value.
func TestAcc_JSONExternalSecret(t *testing.T) {
	t.Parallel()
	accessID := os.Getenv("TEST_SECRET_MANAGER_ACCESS_ID")
	if accessID == "" {
		t.Skip("TEST_SECRET_MANAGER_ACCESS_ID is required for external-secret acceptance tests")
	}
	targets := []struct {
		kind    string
		scope   qovery.APIVariableScopeEnum
		config  func(string) string
		destroy resource.TestCheckFunc
	}{
		{"application", qovery.APIVARIABLESCOPEENUM_APPLICATION, testAccApplicationDefaultConfig, testAccQoveryApplicationDestroy("qovery_application.test")},
		{"container", qovery.APIVARIABLESCOPEENUM_CONTAINER, testAccContainerDefaultConfig, testAccQoveryContainerDestroy("qovery_container.test")},
	}
	for _, target := range targets {
		t.Run(target.kind, func(t *testing.T) {
			t.Parallel()
			address := "qovery_" + target.kind + ".test"
			dataAddress := "data.qovery_" + target.kind + ".json"
			var serviceID string
			config := func(selector string) string {
				selection := ""
				if selector != "" {
					selection = fmt.Sprintf("json_target_key = %q", selector)
				}
				block := fmt.Sprintf("\n external_secrets = [{key = \"JSON_PASSWORD\", reference = \"/terraform-provider/tests/json\", secret_manager_access_id = %q, %s}]\n", accessID, selection)
				base := target.config("json-external-secret-" + target.kind)
				marker := fmt.Sprintf("resource \"qovery_%s\" \"test\" {", target.kind)
				return strings.Replace(base, marker, marker+block, 1) + fmt.Sprintf("\ndata \"qovery_%s\" \"json\" { id = %s.id }\n", target.kind, address)
			}
			check := func(selector string) resource.TestCheckFunc {
				if selector == "" {
					return resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckNoResourceAttr(address, "external_secrets.0.json_target_key"),
						resource.TestCheckNoResourceAttr(dataAddress, "external_secrets.0.json_target_key"),
					)
				}
				return resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "external_secrets.0.json_target_key", selector),
					resource.TestCheckResourceAttr(dataAddress, "external_secrets.0.json_target_key", selector),
				)
			}
			const first = "credentials.password"
			const second = "accounts.0.password"
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             target.destroy,
				Steps: []resource.TestStep{
					{
						Config: config(first), ConfigPlanChecks: testAccEmptyPlanAfterApply,
						Check: resource.ComposeAggregateTestCheckFunc(check(first), testAccCaptureResourceID(address, &serviceID)),
					},
					{Config: config(second), ConfigPlanChecks: testAccEmptyPlanAfterApply, Check: check(second)},
					{ResourceName: address, ImportState: true, ImportStateVerify: true},
					{Config: config(second), PlanOnly: true, ExpectNonEmptyPlan: true, PreConfig: func() {
						testAccChangeJSONSelector(t, target.scope, serviceID, first)
					}},
					{Config: config(second), ConfigPlanChecks: testAccEmptyPlanAfterApply, Check: check(second)},
					{Config: config(""), ConfigPlanChecks: testAccEmptyPlanAfterApply, Check: check("")},
					{ResourceName: address, ImportState: true, ImportStateVerify: true},
				},
			})
		})
	}
}

func testAccChangeJSONSelector(t *testing.T, scope qovery.APIVariableScopeEnum, serviceID, selector string) {
	t.Helper()
	vars, _, err := qoveryAPIClient.VariableMainCallsAPI.ListVariables(context.Background()).ParentId(serviceID).Scope(scope).Execute()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vars.GetResults() {
		if v.Key != "JSON_PASSWORD" {
			continue
		}
		_, _, err = qoveryAPIClient.VariableMainCallsAPI.EditVariable(context.Background(), v.Id).VariableEditRequest(qovery.VariableEditRequest{
			Key: v.Key, Value: v.Value, SecretManagerAccessId: v.SecretManagerAccessId,
			Description: *qovery.NewNullableString(v.Description), JsonTargetKey: *qovery.NewNullableString(&selector),
		}).Execute()
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	t.Fatal("JSON_PASSWORD external secret was not returned by the API")
}
