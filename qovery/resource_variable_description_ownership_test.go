//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/qovery/qovery-client-go"
)

// testAccDescribedVariables maps each variable set attribute to the key of the one variable the
// description ownership tests declare in it.
var testAccDescribedVariables = map[string]string{
	"environment_variables":      "DESC_VAR",
	"secrets":                    "DESC_SECRET",
	"environment_variable_files": "DESC_FILE",
	"secret_files":               "DESC_SECRET_FILE",
}

// The API never returns secret values and does not always return the mount path of a secret
// file, so import cannot record them.
var testAccDescribedVariablesImportIgnore = []string{"secrets.0.value", "secret_files.0.value", "secret_files.0.mount_path"}

// testAccVariableDescriptionsTarget is the service a variable description ownership test runs
// against.
type testAccVariableDescriptionsTarget struct {
	address           string
	dataSourceAddress string
	scope             qovery.APIVariableScopeEnum
	// config renders the service with the variables of testAccDescribedVariables, all with the
	// given description; "" leaves the description out.
	config  func(testName, description string, withDataSource bool) string
	exists  resource.TestCheckFunc
	destroy resource.TestCheckFunc
}

// TestAcc_ApplicationVariableDescriptionOwnership covers variable, secret and file descriptions
// on the application read path.
func TestAcc_ApplicationVariableDescriptionOwnership(t *testing.T) {
	t.Parallel()
	const address = "qovery_application.test"
	testAccVariableDescriptionOwnership(t, "application-variable-description-ownership", testAccVariableDescriptionsTarget{
		address:           address,
		dataSourceAddress: "data.qovery_application.test",
		scope:             qovery.APIVARIABLESCOPEENUM_APPLICATION,
		config: func(testName, description string, withDataSource bool) string {
			return testAccVariableDescriptionsConfig(testName, description, withDataSource, "qovery_application", fmt.Sprintf(`
  environment_id  = qovery_environment.test.id
  name            = "%s"
  build_mode      = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url          = "%s"
    git_token_id = "%s"
  }
  healthchecks = {}`, generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID()))
		},
		exists:  testAccQoveryApplicationExists(address),
		destroy: testAccQoveryApplicationDestroy(address),
	})
}

// TestAcc_ContainerVariableDescriptionOwnership covers variable, secret and file descriptions on
// the domain read path shared by container, job, helm, environment and project.
func TestAcc_ContainerVariableDescriptionOwnership(t *testing.T) {
	t.Parallel()
	const address = "qovery_container.test"
	testAccVariableDescriptionOwnership(t, "container-variable-description-ownership", testAccVariableDescriptionsTarget{
		address:           address,
		dataSourceAddress: "data.qovery_container.test",
		scope:             qovery.APIVARIABLESCOPEENUM_CONTAINER,
		config: func(testName, description string, withDataSource bool) string {
			return testAccContainerRegistryDefaultConfig(testName) + testAccVariableDescriptionsConfig(testName, description, withDataSource, "qovery_container", fmt.Sprintf(`
  environment_id = qovery_environment.test.id
  registry_id    = qovery_container_registry.test.id
  name           = "%s"
  image_name     = "%s"
  tag            = "%s"
  healthchecks   = {}`, generateTestName(testName), containerImageName, containerTag))
		},
		exists:  testAccQoveryContainerExists(address),
		destroy: testAccQoveryContainerDestroy(address),
	})
}

// TestAcc_ApplicationVariableDescriptionUpgradeFrom0x upgrades from the last 0.x release an
// application whose variables declare no description.
func TestAcc_ApplicationVariableDescriptionUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	const address = "qovery_application.test"
	testAccVariableDescriptionUpgradeFrom0x(t, "application-variable-description-upgrade", testAccVariableDescriptionsTarget{
		address: address,
		config: func(testName, description string, withDataSource bool) string {
			return testAccVariableDescriptionsConfig(testName, description, withDataSource, "qovery_application", fmt.Sprintf(`
  environment_id  = qovery_environment.test.id
  name            = "%s"
  build_mode      = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url          = "%s"
    git_token_id = "%s"
  }
  healthchecks = {}`, generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID()))
		},
		destroy: testAccQoveryApplicationDestroy(address),
	})
}

// TestAcc_ContainerVariableDescriptionUpgradeFrom0x upgrades from the last 0.x release a container
// whose variables declare no description.
func TestAcc_ContainerVariableDescriptionUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	const address = "qovery_container.test"
	testAccVariableDescriptionUpgradeFrom0x(t, "container-variable-description-upgrade", testAccVariableDescriptionsTarget{
		address: address,
		config: func(testName, description string, withDataSource bool) string {
			return testAccContainerRegistryDefaultConfig(testName) + testAccVariableDescriptionsConfig(testName, description, withDataSource, "qovery_container", fmt.Sprintf(`
  environment_id = qovery_environment.test.id
  registry_id    = qovery_container_registry.test.id
  name           = "%s"
  image_name     = "%s"
  tag            = "%s"
  healthchecks   = {}`, generateTestName(testName), containerImageName, containerTag))
		},
		destroy: testAccQoveryContainerDestroy(address),
	})
}

// testAccVariableDescriptionUpgradeFrom0x applies with the last 0.x release a service whose
// variables declare no description and whose group ids are omitted. 0.x stored those
// descriptions and group ids as null; the API holds "" and no group, which 1.0 also reads as
// null, so the first 1.0 plan is empty.
func testAccVariableDescriptionUpgradeFrom0x(t *testing.T, testName string, target testAccVariableDescriptionsTarget) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: target.destroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: testAccLastProvider0xFromRegistry,
				Config:            target.config(testName, "", false),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   target.config(testName, "", false),
				PlanOnly:                 true,
			},
		},
	})
}

// testAccVariableDescriptionOwnership runs the ownership steps: a description set outside Terraform
// shows in the plan and the next apply clears it, removing a description from the configuration
// plans its removal, import records the remote descriptions and the data source reports them.
// Every step checks the descriptions the API holds, not only the state.
func testAccVariableDescriptionOwnership(t *testing.T, testName string, target testAccVariableDescriptionsTarget) {
	const (
		declared  = "declared"
		inConsole = "set in the Console"
	)
	config := func(description string, withDataSource bool) string {
		return target.config(testName, description, withDataSource)
	}
	var serviceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             target.destroy,
		Steps: []resource.TestStep{
			// 1. Descriptions omitted: stored as null, none in the API.
			{
				Config: config("", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					target.exists,
					testAccCaptureResourceID(target.address, &serviceID),
					testAccCheckVariableDescriptionsInState(target.address, nil),
					testAccCheckVariableDescriptionsInAPI(target.scope, &serviceID, ""),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Descriptions set outside Terraform show in the plan while the config omits them.
			{
				Config: config("", false),
				Check:  testAccSetVariableDescriptionsOutOfBand(target.scope, &serviceID, inConsole),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(target.address, plancheck.ResourceActionUpdate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 3. The corrective apply clears them.
			{
				Config: config("", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableDescriptionsInState(target.address, nil),
					testAccCheckVariableDescriptionsInAPI(target.scope, &serviceID, ""),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 4. Declare the descriptions.
			{
				Config: config(declared, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableDescriptionsInState(target.address, qovery.PtrString(declared)),
					testAccCheckVariableDescriptionsInAPI(target.scope, &serviceID, declared),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 5. Import records the remote descriptions.
			{
				ResourceName:            target.address,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: testAccDescribedVariablesImportIgnore,
			},
			// 6. The data source reports them.
			{
				Config:           config(declared, true),
				Check:            testAccCheckVariableDescriptionsInState(target.dataSourceAddress, qovery.PtrString(declared)),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 7. Removing the descriptions from the config plans their removal and the apply
			// clears them.
			{
				Config: config("", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(target.address, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableDescriptionsInState(target.address, nil),
					testAccCheckVariableDescriptionsInAPI(target.scope, &serviceID, ""),
				),
			},
		},
	})
}

// --- configuration -------------------------------------------------------------------------------

// testAccVariableDescriptionsConfig renders a service of resourceType with body and the variables
// of testAccDescribedVariables, next to the environment it lives in.
func testAccVariableDescriptionsConfig(testName, description string, withDataSource bool, resourceType, body string) string {
	describe := ""
	if description != "" {
		describe = fmt.Sprintf(", description = %q", description)
	}

	dataSource := ""
	if withDataSource {
		dataSource = fmt.Sprintf(`
data "%s" "test" {
  id = %s.test.id
}
`, resourceType, resourceType)
	}

	return fmt.Sprintf(`
%s

resource "%s" "test" {%s
  environment_variables      = [{ key = "DESC_VAR", value = "value"%s }]
  secrets                    = [{ key = "DESC_SECRET", value = "secret"%s }]
  environment_variable_files = [{ key = "DESC_FILE", value = "file", mount_path = "/etc/desc/file"%s }]
  secret_files               = [{ key = "DESC_SECRET_FILE", value = "secret-file", mount_path = "/etc/desc/secret-file"%s }]
}
%s`, testAccEnvironmentDefaultConfig(testName), resourceType, body, describe, describe, describe, describe, dataSource)
}

// --- state checks --------------------------------------------------------------------------------

// testAccCheckVariableDescriptionsInState checks the description of every variable of
// testAccDescribedVariables in the state of address: absent (null) when want is nil.
func testAccCheckVariableDescriptionsInState(address string, want *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[address]
		if !ok {
			return fmt.Errorf("%s: not found in state", address)
		}
		for attribute, key := range testAccDescribedVariables {
			prefix, err := testAccSetElementPrefix(rs.Primary.Attributes, attribute, key)
			if err != nil {
				return fmt.Errorf("%s: %w", address, err)
			}
			got, isSet := rs.Primary.Attributes[prefix+"description"]
			switch {
			case want == nil && isSet:
				return fmt.Errorf("%s: %s %s description = %q, want null", address, attribute, key, got)
			case want != nil && (!isSet || got != *want):
				return fmt.Errorf("%s: %s %s description = %q (set: %t), want %q", address, attribute, key, got, isSet, *want)
			}
		}
		return nil
	}
}

// testAccSetElementPrefix returns the flatmap prefix ("<attribute>.<index>.") of the element of a
// set attribute whose key is key.
func testAccSetElementPrefix(attributes map[string]string, attribute, key string) (string, error) {
	for name, value := range attributes {
		if value != key || !strings.HasPrefix(name, attribute+".") || !strings.HasSuffix(name, ".key") {
			continue
		}
		prefix := strings.TrimSuffix(name, "key")
		if strings.Count(prefix, ".") == 2 {
			return prefix, nil
		}
	}
	return "", fmt.Errorf("%s: no element with key %s", attribute, key)
}

// --- API checks and out-of-band changes ----------------------------------------------------------

// testAccListServiceVariables returns the variables of testAccDescribedVariables declared on the
// service, by key.
func testAccListServiceVariables(scope qovery.APIVariableScopeEnum, serviceID string) (map[string]qovery.VariableResponse, error) {
	list, res, err := qoveryAPIClient.VariableMainCallsAPI.ListVariables(context.TODO()).
		ParentId(serviceID).
		Scope(scope).
		Execute()
	if err != nil || (res != nil && res.StatusCode >= 400) {
		return nil, fmt.Errorf("failed to list variables of service %s: %v", serviceID, err)
	}
	byKey := make(map[string]qovery.VariableResponse, len(testAccDescribedVariables))
	for _, v := range list.GetResults() {
		if v.Scope == scope {
			byKey[v.Key] = v
		}
	}
	for _, key := range testAccDescribedVariables {
		if _, ok := byKey[key]; !ok {
			return nil, fmt.Errorf("service %s: variable %s not found in API", serviceID, key)
		}
	}
	return byKey, nil
}

// testAccCheckVariableDescriptionsInAPI checks that every variable of testAccDescribedVariables
// holds want as its description in the API ("" for none).
func testAccCheckVariableDescriptionsInAPI(scope qovery.APIVariableScopeEnum, serviceID *string, want string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		byKey, err := testAccListServiceVariables(scope, *serviceID)
		if err != nil {
			return err
		}
		for _, key := range testAccDescribedVariables {
			v := byKey[key]
			if got := v.GetDescription(); got != want {
				return fmt.Errorf("service %s: variable %s description in API = %q, want %q", *serviceID, key, got, want)
			}
		}
		return nil
	}
}

// testAccSetVariableDescriptionsOutOfBand sets the description of every variable of
// testAccDescribedVariables through the API, bypassing Terraform, the way the Qovery Console
// does: the edit omits the value, which the API keeps.
func testAccSetVariableDescriptionsOutOfBand(scope qovery.APIVariableScopeEnum, serviceID *string, description string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		byKey, err := testAccListServiceVariables(scope, *serviceID)
		if err != nil {
			return err
		}
		for _, key := range testAccDescribedVariables {
			v := byKey[key]
			_, res, err := qoveryAPIClient.VariableMainCallsAPI.EditVariable(context.TODO(), v.Id).
				VariableEditRequest(qovery.VariableEditRequest{
					Key:         v.Key,
					Description: *qovery.NewNullableString(&description),
				}).
				Execute()
			if err != nil || (res != nil && res.StatusCode >= 400) {
				return fmt.Errorf("failed to edit variable %s of service %s out of band: %v", key, *serviceID, err)
			}
		}
		return nil
	}
}
