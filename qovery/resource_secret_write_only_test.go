//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// The secrets of a project, the domain write path, keep their write-only values out of the state.
// The steps cover a mixed list, the switch from value, an unchanged version, a version bump, a
// description change, a reorder, import, and the adoption of value_wo after import. The ephemeral
// password changes on every plan, so each empty plan after apply shows that a new value alone
// plans nothing. The ids show that no step deletes and recreates a secret.
func TestAcc_ProjectSecretWriteOnly(t *testing.T) {
	t.Parallel()
	const addr = "qovery_project.test"
	testName := "project-secret-wo"
	plainValue := "plain-" + uuid.NewString()
	var projectID, idA, idB string
	config := func(secrets ...string) string {
		return testAccEphemeralRandomPassword + fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  secrets = [%s]
}
`, getTestOrganizationID(), generateTestName(testName), strings.Join(secrets, ",\n    "))
	}
	reordered := config(testAccWriteOnlySecret("B", "2", ""), testAccWriteOnlySecret("A", "1", "rotated with the version"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        testAccRandomProvider,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		CheckDestroy:             testAccQoveryProjectDestroy(addr),
		Steps: []resource.TestStep{
			// A plain entry stores its value, a write-only entry only its version.
			{
				Config: config(testAccPlainSecret("A", plainValue), testAccWriteOnlySecret("B", "1", "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists(addr),
					resource.TestCheckResourceAttr(addr, "secrets.#", "2"),
					resource.TestCheckResourceAttr(addr, "secrets.0.key", "A"),
					resource.TestCheckResourceAttr(addr, "secrets.0.value", plainValue),
					resource.TestCheckResourceAttr(addr, "secrets.1.key", "B"),
					resource.TestCheckNoResourceAttr(addr, "secrets.1.value"),
					resource.TestCheckNoResourceAttr(addr, "secrets.1.value_wo"),
					resource.TestCheckResourceAttr(addr, "secrets.1.value_wo_version", "1"),
					testAccCaptureResourceID(addr, &projectID),
					testAccCaptureAttribute(addr, "secrets.0.id", &idA),
					testAccCaptureAttribute(addr, "secrets.1.id", &idB),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// Switching A to value_wo updates it in place and removes its value from the state.
			{
				Config:           config(testAccWriteOnlySecret("A", "1", ""), testAccWriteOnlySecret("B", "1", "")),
				ConfigPlanChecks: testAccUpdateThenEmptyPlan(addr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "secrets.0.id", &idA),
					resource.TestCheckNoResourceAttr(addr, "secrets.0.value"),
					resource.TestCheckResourceAttr(addr, "secrets.0.value_wo_version", "1"),
					testAccCheckStateHasNoValue(plainValue),
				),
			},
			// New values alone plan nothing.
			{
				Config:   config(testAccWriteOnlySecret("A", "1", ""), testAccWriteOnlySecret("B", "1", "")),
				PlanOnly: true,
			},
			// Bumping the version of B sends its current value.
			{
				Config:           config(testAccWriteOnlySecret("A", "1", ""), testAccWriteOnlySecret("B", "2", "")),
				ConfigPlanChecks: testAccUpdateThenEmptyPlan(addr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "secrets.1.id", &idB),
					resource.TestCheckResourceAttr(addr, "secrets.1.value_wo_version", "2"),
				),
			},
			// Changing the description of A sends its current value too.
			{
				Config:           config(testAccWriteOnlySecret("A", "1", "rotated with the version"), testAccWriteOnlySecret("B", "2", "")),
				ConfigPlanChecks: testAccUpdateThenEmptyPlan(addr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "secrets.0.id", &idA),
					resource.TestCheckResourceAttr(addr, "secrets.0.description", "rotated with the version"),
				),
			},
			// A reorder only moves the entries in the state.
			{
				Config:           reordered,
				ConfigPlanChecks: testAccUpdateThenEmptyPlan(addr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "secrets.0.key", "B"),
					resource.TestCheckResourceAttrPtr(addr, "secrets.0.id", &idB),
					resource.TestCheckResourceAttr(addr, "secrets.1.key", "A"),
					resource.TestCheckResourceAttrPtr(addr, "secrets.1.id", &idA),
				),
			},
			// Import records the keys, neither the values nor the versions...
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secrets"},
				ImportStateCheck:        testAccImportedSecretsWithoutVersion(2),
			},
			// ... so a project imported into a fresh state...
			{
				Config: testAccEphemeralRandomPassword + fmt.Sprintf(`
removed {
  from = %s

  lifecycle {
    destroy = false
  }
}
`, addr),
				Check: testAccCheckAbsentFromState(addr),
			},
			{
				Config:             reordered,
				ResourceName:       addr,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc:  func(*terraform.State) (string, error) { return projectID, nil },
				ImportStateCheck:   testAccImportedSecretsWithoutVersion(2),
			},
			// ... adopts value_wo in one update that keeps the secrets.
			{
				Config:           reordered,
				ConfigPlanChecks: testAccUpdateThenEmptyPlan(addr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "id", &projectID),
					resource.TestCheckResourceAttrPtr(addr, "secrets.0.id", &idB),
					resource.TestCheckResourceAttrPtr(addr, "secrets.1.id", &idA),
					resource.TestCheckResourceAttr(addr, "secrets.0.value_wo_version", "2"),
					resource.TestCheckResourceAttr(addr, "secrets.1.value_wo_version", "1"),
				),
			},
		},
	})
}

// ignore_changes on the key of a secrets entry keeps the old key in the plan while the configuration
// sets value_wo under the new one. The apply fails instead of sending an empty value, and the
// secret keeps its value.
func TestAcc_ProjectSecretWriteOnlyIgnoredKey(t *testing.T) {
	t.Parallel()
	const addr = "qovery_project.test"
	testName := "project-secret-wo-ignored-key"
	var idToken string
	config := func(key, version, lifecycle string) string {
		return testAccEphemeralRandomPassword + fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  secrets = [%s]
  %s
}
`, getTestOrganizationID(), generateTestName(testName), testAccWriteOnlySecret(key, version, ""), lifecycle)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        testAccRandomProvider,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		CheckDestroy:             testAccQoveryProjectDestroy(addr),
		Steps: []resource.TestStep{
			{
				Config: config("TOKEN", "1", ""),
				Check:  testAccCaptureAttribute(addr, "secrets.0.id", &idToken),
			},
			{
				Config:      config("NEW_TOKEN", "2", "lifecycle {\n    ignore_changes = [secrets[0].key]\n  }"),
				ExpectError: regexp.MustCompile(`The\s+plan\s+creates\s+or\s+updates\s+the\s+secret\s+"TOKEN"`),
			},
			// The failed apply sent nothing: the state still holds TOKEN at version 1, and the
			// plan still holds the version bump.
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "secrets.0.key", "TOKEN"),
					resource.TestCheckResourceAttrPtr(addr, "secrets.0.id", &idToken),
					resource.TestCheckResourceAttr(addr, "secrets.0.value_wo_version", "1"),
				),
			},
		},
	})
}

// The secrets of an application, the client write path, keep their write-only values out of the
// state.
func TestAcc_ApplicationSecretWriteOnly(t *testing.T) {
	t.Parallel()
	const addr = "qovery_application.test"
	testName := "application-secret-wo"
	plainValue := "plain-" + uuid.NewString()
	var idA, idB string
	config := func(secrets ...string) string {
		return testAccEphemeralRandomPassword + fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
    git_token_id = "%s"
  }
  healthchecks = {}
  secrets = [%s]
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(), strings.Join(secrets, ",\n    "))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        testAccRandomProvider,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		CheckDestroy:             testAccQoveryApplicationDestroy(addr),
		Steps: []resource.TestStep{
			{
				Config: config(testAccPlainSecret("A", plainValue), testAccWriteOnlySecret("B", "1", "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists(addr),
					resource.TestCheckResourceAttr(addr, "secrets.0.value", plainValue),
					resource.TestCheckNoResourceAttr(addr, "secrets.1.value"),
					resource.TestCheckResourceAttr(addr, "secrets.1.value_wo_version", "1"),
					testAccCaptureAttribute(addr, "secrets.0.id", &idA),
					testAccCaptureAttribute(addr, "secrets.1.id", &idB),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// Switching A to value_wo and bumping B update both in place.
			{
				Config:           config(testAccWriteOnlySecret("A", "1", ""), testAccWriteOnlySecret("B", "2", "")),
				ConfigPlanChecks: testAccUpdateThenEmptyPlan(addr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "secrets.0.id", &idA),
					resource.TestCheckResourceAttrPtr(addr, "secrets.1.id", &idB),
					resource.TestCheckNoResourceAttr(addr, "secrets.0.value"),
					resource.TestCheckResourceAttr(addr, "secrets.1.value_wo_version", "2"),
					testAccCheckStateHasNoValue(plainValue),
				),
			},
			// A reorder only moves the entries in the state.
			{
				Config:           config(testAccWriteOnlySecret("B", "2", ""), testAccWriteOnlySecret("A", "1", "")),
				ConfigPlanChecks: testAccUpdateThenEmptyPlan(addr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(addr, "secrets.0.id", &idB),
					resource.TestCheckResourceAttrPtr(addr, "secrets.1.id", &idA),
				),
			},
		},
	})
}

// A project whose secrets 1.2.2 stored as a set reads them as a list after the upgrade. A config
// in the order of the state plans nothing, any other order plans a reorder and nothing else. The
// state order of a set is not known in advance, so both orders run.
func TestAcc_ProjectSecretsUpgradeFrom1_2(t *testing.T) {
	t.Parallel()
	for _, order := range [][]string{{"A", "B"}, {"B", "A"}} {
		order := order
		t.Run(strings.Join(order, ""), func(t *testing.T) {
			t.Parallel()
			const addr = "qovery_project.test"
			testName := "project-secrets-upgrade-" + strings.ToLower(strings.Join(order, ""))
			entries := make([]string, 0, len(order))
			for _, key := range order {
				entries = append(entries, testAccPlainSecret(key, "value-"+key))
			}
			config := fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  secrets = [%s]
}
`, getTestOrganizationID(), generateTestName(testName), strings.Join(entries, ",\n    "))

			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				CheckDestroy: testAccQoveryProjectDestroy(addr),
				Steps: []resource.TestStep{
					{
						ExternalProviders: testAccProvider1_2FromRegistry,
						Config:            config,
						Check:             testAccQoveryProjectExists(addr),
					},
					{
						ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
						Config:                   config,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply:             []plancheck.PlanCheck{expectAtMostSecretsReordered{addr: addr}},
							PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
						},
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(addr, "secrets.0.key", order[0]),
							resource.TestCheckResourceAttr(addr, "secrets.1.key", order[1]),
						),
					},
				},
			})
		})
	}
}

// expectAtMostSecretsReordered checks the plan of addr: no change when the configuration lists the
// secrets in the order of the state, else an update that only reorders them. The ids of the
// secrets and the computed attributes the update plans unknown are not compared.
type expectAtMostSecretsReordered struct {
	addr string
}

func (e expectAtMostSecretsReordered) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, rc := range req.Plan.ResourceChanges {
		if rc.Address != e.addr {
			continue
		}
		before, _ := rc.Change.Before.(map[string]any)
		after, _ := rc.Change.After.(map[string]any)
		afterUnknown, _ := rc.Change.AfterUnknown.(map[string]any)

		if reflect.DeepEqual(secretKeysOf(before["secrets"]), secretKeysOf(after["secrets"])) {
			if !rc.Change.Actions.NoOp() {
				resp.Error = fmt.Errorf("%s: the configuration lists the secrets in the state order, expected no change, got %v on %v, sensitivity changed: %t", e.addr, rc.Change.Actions, changedAttributes(before, after, afterUnknown), !reflect.DeepEqual(rc.Change.BeforeSensitive, rc.Change.AfterSensitive))
			}
			return
		}
		if !rc.Change.Actions.Update() {
			resp.Error = fmt.Errorf("%s: expected an update that reorders the secrets, got %v", e.addr, rc.Change.Actions)
			return
		}
		if !reflect.DeepEqual(secretsWithoutIDs(before["secrets"]), secretsWithoutIDs(after["secrets"])) {
			resp.Error = fmt.Errorf("%s: the secrets changed beyond their order: %v -> %v", e.addr, before["secrets"], after["secrets"])
			return
		}
		for name, value := range before {
			if unknown, _ := afterUnknown[name].(bool); name == "secrets" || unknown {
				continue
			}
			if !reflect.DeepEqual(value, after[name]) {
				resp.Error = fmt.Errorf("%s: %s changed: %v -> %v", e.addr, name, value, after[name])
				return
			}
		}
		return
	}
	resp.Error = fmt.Errorf("%s: not in the plan", e.addr)
}

// changedAttributes lists the attributes of a plan whose value changes, with their values.
func changedAttributes(before, after, afterUnknown map[string]any) []string {
	changed := []string{}
	for name, value := range before {
		if unknown, _ := afterUnknown[name].(bool); unknown {
			changed = append(changed, name+" (unknown)")
			continue
		}
		if !reflect.DeepEqual(value, after[name]) {
			changed = append(changed, fmt.Sprintf("%s: %v -> %v", name, value, after[name]))
		}
	}
	sort.Strings(changed)
	return changed
}

// secretKeysOf returns the keys of a secrets value of a plan, in order.
func secretKeysOf(secrets any) []string {
	entries, _ := secrets.([]any)
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		fields, _ := entry.(map[string]any)
		key, _ := fields["key"].(string)
		keys = append(keys, key)
	}
	return keys
}

// secretsWithoutIDs returns the entries of a secrets value of a plan without their id, sorted by
// key.
func secretsWithoutIDs(secrets any) []map[string]any {
	entries, _ := secrets.([]any)
	result := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		fields, _ := entry.(map[string]any)
		withoutID := map[string]any{}
		for name, value := range fields {
			if name != "id" {
				withoutID[name] = value
			}
		}
		result = append(result, withoutID)
	}
	sort.Slice(result, func(i, j int) bool {
		first, _ := result[i]["key"].(string)
		second, _ := result[j]["key"].(string)
		return first < second
	})
	return result
}

// testAccPlainSecret declares a secrets entry with a plain value.
func testAccPlainSecret(key, value string) string {
	return fmt.Sprintf(`{ key = %q, value = %q }`, key, value)
}

// testAccWriteOnlySecret declares a secrets entry whose value_wo is the ephemeral random password.
func testAccWriteOnlySecret(key, version, description string) string {
	entry := fmt.Sprintf(`{ key = %q, value_wo = ephemeral.random_password.secret.result, value_wo_version = %q`, key, version)
	if description != "" {
		entry += fmt.Sprintf(`, description = %q`, description)
	}
	return entry + " }"
}

// testAccUpdateThenEmptyPlan expects addr to be updated in place, then the configuration to plan
// nothing.
func testAccUpdateThenEmptyPlan(addr string) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
}

// testAccImportedSecretsWithoutVersion checks that the import recorded count secrets, without
// value or version.
func testAccImportedSecretsWithoutVersion(count int) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		attributes := states[0].Attributes
		if got := attributes["secrets.#"]; got != fmt.Sprint(count) {
			return fmt.Errorf("import recorded %s secrets, expected %d", got, count)
		}
		for name, value := range attributes {
			if strings.HasPrefix(name, "secrets.") && (strings.HasSuffix(name, ".value") || strings.HasSuffix(name, ".value_wo_version")) {
				return fmt.Errorf("import recorded %s = %q, expected none", name, value)
			}
		}
		return nil
	}
}

// testAccCaptureAttribute stores the value of an attribute of resourceName in into.
func testAccCaptureAttribute(resourceName, attribute string, into *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s: not in the state", resourceName)
		}
		value, ok := rs.Primary.Attributes[attribute]
		if !ok || value == "" {
			return fmt.Errorf("%s: %s is not set", resourceName, attribute)
		}
		*into = value
		return nil
	}
}
