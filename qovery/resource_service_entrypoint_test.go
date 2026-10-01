//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// A Console save of the service settings writes entrypoint = "" where the API held none. "" means
// the image's entrypoint, like null, so a configuration that omits entrypoint must plan no change.
// 1.0.0-rc.1 planned entrypoint = "" -> null after every such save.

// testAccEmptyEntrypointTarget is the service an empty entrypoint test runs against.
type testAccEmptyEntrypointTarget struct {
	address string
	// config renders the service without entrypoint.
	config  string
	apiPath func(serviceID string) string
	// toRequest turns the GET response of the service into its edit request, nil when both share
	// their shape.
	toRequest func(service map[string]any)
	// entrypointHolder returns the object of the service JSON that holds the entrypoint.
	entrypointHolder func(service map[string]any) map[string]any
	// stateKey is the flatmap key of the entrypoint in state.
	stateKey string
	// dataSource reads the service through its data source, as data.<type>.test.
	dataSource string
	exists     resource.TestCheckFunc
	destroy    resource.TestCheckFunc
}

func testAccEmptyEntrypointFromConsole(t *testing.T, target testAccEmptyEntrypointTarget) {
	var serviceID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             target.destroy,
		Steps: []resource.TestStep{
			// 1. The service without entrypoint.
			{
				Config: target.config,
				Check: resource.ComposeAggregateTestCheckFunc(
					target.exists,
					testAccCaptureResourceID(target.address, &serviceID),
					resource.TestCheckNoResourceAttr(target.address, target.stateKey),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Save the service from the Console, which writes entrypoint = "": nothing to plan.
			{
				Config: target.config,
				Check: func(_ *terraform.State) error {
					return testAccSetEmptyEntrypointOutOfBand(target, serviceID)
				},
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 3. The refresh stores no entrypoint.
			{
				RefreshState: true,
				Check:        resource.TestCheckNoResourceAttr(target.address, target.stateKey),
			},
			// 4. The data source reports the API value as-is.
			{
				Config: target.config + target.dataSource,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data."+target.address, testAccContractJSONPath(target.stateKey), knownvalue.StringExact("")),
				},
			},
			// 5. Import records no entrypoint either.
			{
				ResourceName:      target.address,
				ImportState:       true,
				ImportStateVerify: true,
				// An unset advanced_settings_json is stored as "" after apply and read as "{}" on
				// import, under its own contract (QOV-2028).
				ImportStateVerifyIgnore: []string{"advanced_settings_json"},
			},
		},
	})
}

// testAccSetEmptyEntrypointOutOfBand writes entrypoint = "" through the API, the way a Console
// save of the service settings does, and checks the API holds "" afterwards: if it normalized ""
// to null, the test would prove nothing.
func testAccSetEmptyEntrypointOutOfBand(target testAccEmptyEntrypointTarget, serviceID string) error {
	apiPath := target.apiPath(serviceID)
	if err := testAccEditServiceOutOfBand(apiPath, target.toRequest, func(service map[string]any) {
		target.entrypointHolder(service)["entrypoint"] = ""
	}); err != nil {
		return err
	}

	var service map[string]any
	if err := testAccQoveryAPIJSON(http.MethodGet, apiPath, nil, &service); err != nil {
		return err
	}
	entrypoint, ok := target.entrypointHolder(service)["entrypoint"]
	if !ok || entrypoint != "" {
		return fmt.Errorf("%s: the API holds entrypoint %v after the out-of-band edit, want \"\"", apiPath, entrypoint)
	}
	return nil
}

// testAccServiceDataSource reads <resourceType>.test through its data source.
func testAccServiceDataSource(resourceType string) string {
	return fmt.Sprintf(`
data "%[1]s" "test" {
  id = %[1]s.test.id
}
`, resourceType)
}

func TestAcc_ContainerEmptyEntrypointFromConsole(t *testing.T) {
	t.Parallel()
	const address = "qovery_container.test"
	testAccEmptyEntrypointFromConsole(t, testAccEmptyEntrypointTarget{
		address: address,
		config:  testAccContainerContractConfig("container-empty-entrypoint", false),
		apiPath: func(id string) string { return "/container/" + id },
		toRequest: func(cont map[string]any) {
			if registry, ok := cont["registry"].(map[string]any); ok {
				cont["registry_id"] = registry["id"]
			}
		},
		entrypointHolder: func(cont map[string]any) map[string]any { return cont },
		stateKey:         "entrypoint",
		dataSource:       testAccServiceDataSource("qovery_container"),
		exists:           testAccQoveryContainerExists(address),
		destroy:          testAccQoveryContainerDestroy(address),
	})
}

func TestAcc_ApplicationEmptyEntrypointFromConsole(t *testing.T) {
	t.Parallel()
	const address = "qovery_application.test"
	testAccEmptyEntrypointFromConsole(t, testAccEmptyEntrypointTarget{
		address:          address,
		config:           testAccApplicationContractConfig("application-empty-entrypoint", false),
		apiPath:          func(id string) string { return "/application/" + id },
		entrypointHolder: func(app map[string]any) map[string]any { return app },
		stateKey:         "entrypoint",
		dataSource:       testAccServiceDataSource("qovery_application"),
		exists:           testAccQoveryApplicationExists(address),
		destroy:          testAccQoveryApplicationDestroy(address),
	})
}

func TestAcc_CronJobEmptyEntrypointFromConsole(t *testing.T) {
	t.Parallel()
	const address = "qovery_job.test"
	testAccEmptyEntrypointFromConsole(t, testAccEmptyEntrypointTarget{
		address: address,
		config:  testAccCronJobContractConfig("cron-job-empty-entrypoint", false),
		apiPath: func(id string) string { return "/job/" + id },
		entrypointHolder: func(job map[string]any) map[string]any {
			return jsonObject(jsonObject(job, "schedule"), "cronjob")
		},
		stateKey:   "schedule.cronjob.command.entrypoint",
		dataSource: testAccServiceDataSource("qovery_job"),
		exists:     testAccQoveryJobExists(address),
		destroy:    testAccQoveryJobDestroy(address),
	})
}

func TestAcc_LifecycleJobEmptyEntrypointFromConsole(t *testing.T) {
	t.Parallel()
	const address = "qovery_job.test"
	testAccEmptyEntrypointFromConsole(t, testAccEmptyEntrypointTarget{
		address: address,
		config:  testAccLifecycleJobContractConfig("lifecycle-job-empty-entrypoint", false, false, ""),
		apiPath: func(id string) string { return "/job/" + id },
		entrypointHolder: func(job map[string]any) map[string]any {
			return jsonObject(jsonObject(job, "schedule"), "on_start")
		},
		stateKey:   "schedule.on_start.entrypoint",
		dataSource: testAccServiceDataSource("qovery_job"),
		exists:     testAccQoveryJobExists(address),
		destroy:    testAccQoveryJobDestroy(address),
	})
}
