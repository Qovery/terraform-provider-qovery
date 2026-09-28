//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// The service contract tests cover the attributes QOV-2327 moved to the config-is-source-of-truth
// rule on each service resource: removing them from the configuration plans the q-core default,
// a change made outside Terraform shows on refresh and the next apply reverts it, import records
// the remote value, and a state written by the last 0.x release upgrades to an empty plan.

// testAccContractValue is the value of one attribute of a service, by flatmap key (for example
// ports.0.name, or arguments.# for a list). A nil value means null.
type testAccContractValue struct {
	key   string
	value any
}

// testAccServiceContractTarget is the service a contract test runs against.
type testAccServiceContractTarget struct {
	address string
	// config renders the service with every contract attribute declared with a non-default value
	// (declared) or omitted.
	config func(declared bool) string
	// declared are the values the declared configuration sets; defaults are the values planned
	// when the attributes are omitted.
	declared []testAccContractValue
	defaults []testAccContractValue
	// outOfBandOnDeclared makes the out-of-band steps run on the declared configuration instead of
	// the one that omits the attributes.
	outOfBandOnDeclared bool
	// outOfBand changes the service the way the Qovery Console does, and changed are the values
	// the refresh must then report.
	outOfBand func(serviceID string) error
	changed   []testAccContractValue
	// importStateVerifyIgnore lists the attributes import cannot record.
	importStateVerifyIgnore []string
	exists                  resource.TestCheckFunc
	destroy                 resource.TestCheckFunc
}

// testAccServiceContract runs the contract steps. Every step that applies is followed by an empty
// refreshed plan, which proves the API holds what the state holds.
func testAccServiceContract(t *testing.T, target testAccServiceContractTarget) {
	var serviceID string
	// An unset advanced_settings_json is stored as "" after apply and read as "{}" on import,
	// under its own contract (QOV-2028).
	target.importStateVerifyIgnore = append(target.importStateVerifyIgnore, "advanced_settings_json")

	outOfBandConfig, outOfBandValues := target.config(false), target.defaults
	if target.outOfBandOnDeclared {
		outOfBandConfig, outOfBandValues = target.config(true), target.declared
	}
	outOfBandSteps := []resource.TestStep{
		// A change made outside Terraform shows in the plan, which reverts it...
		{
			Config: outOfBandConfig,
			Check: func(_ *terraform.State) error {
				return target.outOfBand(serviceID)
			},
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PostApplyPostRefresh: append(
					[]plancheck.PlanCheck{plancheck.ExpectResourceAction(target.address, plancheck.ResourceActionUpdate)},
					testAccContractPlanChecks(target.address, outOfBandValues)...,
				),
			},
			ExpectNonEmptyPlan: true,
		},
		// ... and the refresh stores the remote value.
		{
			RefreshState:       true,
			Check:              testAccContractStateChecks(target.address, target.changed),
			ExpectNonEmptyPlan: true,
		},
		// The corrective apply reverts it.
		{
			Config:           outOfBandConfig,
			Check:            testAccContractStateChecks(target.address, outOfBandValues),
			ConfigPlanChecks: testAccEmptyPlanAfterApply,
		},
	}

	steps := []resource.TestStep{
		// 1. Every contract attribute declared with a non-default value.
		{
			Config: target.config(true),
			Check: resource.ComposeAggregateTestCheckFunc(
				target.exists,
				testAccCaptureResourceID(target.address, &serviceID),
				testAccContractStateChecks(target.address, target.declared),
			),
			ConfigPlanChecks: testAccEmptyPlanAfterApply,
		},
		// 2. Import records the declared remote values.
		{
			ResourceName:            target.address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: target.importStateVerifyIgnore,
		},
	}
	if target.outOfBandOnDeclared {
		steps = append(steps, outOfBandSteps...)
	}
	steps = append(steps,
		// Removing the attributes plans the q-core defaults, and the apply writes them.
		resource.TestStep{
			Config: target.config(false),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: append(
					[]plancheck.PlanCheck{plancheck.ExpectResourceAction(target.address, plancheck.ResourceActionUpdate)},
					testAccContractPlanChecks(target.address, target.defaults)...,
				),
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
			Check: testAccContractStateChecks(target.address, target.defaults),
		},
	)
	if !target.outOfBandOnDeclared {
		steps = append(steps, outOfBandSteps...)
	}
	steps = append(steps,
		// Import records the remote defaults.
		resource.TestStep{
			ResourceName:            target.address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: target.importStateVerifyIgnore,
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             target.destroy,
		Steps:                    steps,
	})
}

// testAccServiceContractUpgradeFrom0x applies config with the last 0.x release, then with this
// provider: the plan, which refreshes the 0.x state first, must be empty, as terraform plan shows
// it after the upgrade. A plan without refresh would compare the configuration with values 0.x
// stored and 1.0 reads differently from the same API value, such as a null ephemeral_storage.
func testAccServiceContractUpgradeFrom0x(t *testing.T, config string, destroy resource.TestCheckFunc) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: destroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: testAccLastProvider0xFromRegistry,
				Config:            config,
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// --- values --------------------------------------------------------------------------------------

func testAccContractStateChecks(address string, values []testAccContractValue) resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 0, len(values))
	for _, v := range values {
		if v.value == nil {
			checks = append(checks, resource.TestCheckNoResourceAttr(address, v.key))
			continue
		}
		checks = append(checks, resource.TestCheckResourceAttr(address, v.key, fmt.Sprint(v.value)))
	}
	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func testAccContractPlanChecks(address string, values []testAccContractValue) []plancheck.PlanCheck {
	checks := make([]plancheck.PlanCheck, 0, len(values))
	for _, v := range values {
		checks = append(checks, plancheck.ExpectKnownValue(address, testAccContractJSONPath(v.key), testAccContractKnownValue(v.value)))
	}
	return checks
}

// testAccContractJSONPath turns a flatmap key into the path of the attribute in the plan.
func testAccContractJSONPath(key string) tfjsonpath.Path {
	parts := strings.Split(strings.TrimSuffix(key, ".#"), ".")
	p := tfjsonpath.New(parts[0])
	for _, part := range parts[1:] {
		if index, err := strconv.Atoi(part); err == nil {
			p = p.AtSliceIndex(index)
			continue
		}
		p = p.AtMapKey(part)
	}
	return p
}

func testAccContractKnownValue(value any) knownvalue.Check {
	switch v := value.(type) {
	case nil:
		return knownvalue.Null()
	case bool:
		return knownvalue.Bool(v)
	case int:
		return knownvalue.Int64Exact(int64(v))
	case string:
		return knownvalue.StringExact(v)
	}
	panic(fmt.Sprintf("unsupported contract value %T", value))
}

// --- API -----------------------------------------------------------------------------------------

// testAccQoveryAPIJSON sends a request to the Qovery API with the test client's host and token.
// The out-of-band edits send JSON bodies rather than client-go models: the GET response of a
// service is reused as its edit request, and q-core ignores the fields only the response has.
func testAccQoveryAPIJSON(method, apiPath string, body any, out any) error {
	cfg := qoveryAPIClient.GetConfig()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(context.TODO(), method, strings.TrimSuffix(cfg.Servers[0].URL, "/")+apiPath, reader)
	if err != nil {
		return err
	}
	for key, value := range cfg.DefaultHeader {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, apiPath, err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("%s %s: HTTP %d: %s", method, apiPath, res.StatusCode, data)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

// testAccEditServiceOutOfBand edits a service through the API, bypassing Terraform: it reads the
// service, turns the response into an edit request with toRequest (nil when both share their
// shape), applies change and sends it back.
func testAccEditServiceOutOfBand(apiPath string, toRequest func(map[string]any), change func(map[string]any)) error {
	var service map[string]any
	if err := testAccQoveryAPIJSON(http.MethodGet, apiPath, nil, &service); err != nil {
		return err
	}
	if toRequest != nil {
		toRequest(service)
	}
	change(service)
	return testAccQoveryAPIJSON(http.MethodPut, apiPath, service, nil)
}

// testAccEditCustomDomainsOutOfBand sets generate_certificate and use_cdn on every custom domain
// of the service at apiPath.
func testAccEditCustomDomainsOutOfBand(apiPath string, generateCertificate, useCdn bool) error {
	var list struct {
		Results []struct {
			ID     string `json:"id"`
			Domain string `json:"domain"`
		} `json:"results"`
	}
	if err := testAccQoveryAPIJSON(http.MethodGet, apiPath+"/customDomain", nil, &list); err != nil {
		return err
	}
	if len(list.Results) == 0 {
		return fmt.Errorf("%s: no custom domain", apiPath)
	}
	for _, d := range list.Results {
		body := map[string]any{"domain": d.Domain, "generate_certificate": generateCertificate, "use_cdn": useCdn}
		if err := testAccQoveryAPIJSON(http.MethodPut, apiPath+"/customDomain/"+d.ID, body, nil); err != nil {
			return err
		}
	}
	return nil
}

// jsonObject returns the object at key of m, creating it when missing.
func jsonObject(m map[string]any, key string) map[string]any {
	if o, ok := m[key].(map[string]any); ok {
		return o
	}
	o := map[string]any{}
	m[key] = o
	return o
}

// jsonFirst returns the first object of the list at key of m.
func jsonFirst(m map[string]any, key string) (map[string]any, error) {
	list, ok := m[key].([]any)
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("no %s in the API response", key)
	}
	o, ok := list[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected %s in the API response", key)
	}
	return o, nil
}

// testAccContractDomain is a custom domain unique to the test run.
func testAccContractDomain(testName string) string {
	return fmt.Sprintf("%s-%s.example.com", testName, testNameSuffix[:8])
}

// --- application ---------------------------------------------------------------------------------

// TestAcc_ApplicationContract covers icon_uri, auto_deploy, ephemeral_storage, arguments,
// ports.name and the custom domain flags on the application.
func TestAcc_ApplicationContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_application.test"
	testName := "application-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address: address,
		config:  func(declared bool) string { return testAccApplicationContractConfig(testName, declared) },
		declared: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-declared"},
			{"auto_deploy", false},
			{"ephemeral_storage", 4},
			{"arguments.0", "--declared"},
			{"ports.0.name", "declared"},
			{"custom_domains.0.generate_certificate", true},
			{"custom_domains.0.use_cdn", true},
		},
		defaults: []testAccContractValue{
			{"icon_uri", "app://qovery-console/application"},
			{"auto_deploy", true},
			{"ephemeral_storage", 0},
			{"arguments.#", nil},
			{"ports.0.name", "p8080"},
			{"custom_domains.0.generate_certificate", false},
			{"custom_domains.0.use_cdn", false},
		},
		outOfBand: func(id string) error {
			apiPath := "/application/" + id
			err := testAccEditServiceOutOfBand(apiPath, nil, func(app map[string]any) {
				app["icon_uri"] = "app://qovery-console/contract-console"
				app["auto_deploy"] = false
				app["ephemeral_storage_in_gib"] = 2
				app["arguments"] = []string{"--console"}
				if port, err := jsonFirst(app, "ports"); err == nil {
					port["name"] = "console"
				}
			})
			if err != nil {
				return err
			}
			return testAccEditCustomDomainsOutOfBand(apiPath, true, true)
		},
		changed: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-console"},
			{"auto_deploy", false},
			{"ephemeral_storage", 2},
			{"arguments.0", "--console"},
			{"ports.0.name", "console"},
			{"custom_domains.0.generate_certificate", true},
			{"custom_domains.0.use_cdn", true},
		},
		exists:  testAccQoveryApplicationExists(address),
		destroy: testAccQoveryApplicationDestroy(address),
	})
}

func TestAcc_ApplicationContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testName := "application-contract-upgrade"
	testAccServiceContractUpgradeFrom0x(t, testAccApplicationContractConfig(testName, false), testAccQoveryApplicationDestroy("qovery_application.test"))
}

func testAccApplicationContractConfig(testName string, declared bool) string {
	attributes := `
  ports          = [{ internal_port = 8080, external_port = 443, publicly_accessible = true }]
  custom_domains = [{ domain = "` + testAccContractDomain(testName) + `" }]`
	if declared {
		attributes = `
  icon_uri          = "app://qovery-console/contract-declared"
  auto_deploy       = false
  ephemeral_storage = 4
  arguments         = ["--declared"]
  ports             = [{ name = "declared", internal_port = 8080, external_port = 443, publicly_accessible = true }]
  custom_domains    = [{ domain = "` + testAccContractDomain(testName) + `", generate_certificate = true, use_cdn = true }]`
	}
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id  = qovery_environment.test.id
  name            = "%s"
  build_mode      = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url          = "%s"
    git_token_id = "%s"
  }
  healthchecks = {}%s
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(), attributes)
}

// --- container -----------------------------------------------------------------------------------

// TestAcc_ContainerContract covers icon_uri, auto_deploy, auto_preview, ephemeral_storage,
// arguments, ports.name, ports.protocol and the custom domain flags on the container.
func TestAcc_ContainerContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_container.test"
	testName := "container-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address: address,
		config:  func(declared bool) string { return testAccContainerContractConfig(testName, declared) },
		declared: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-declared"},
			{"auto_deploy", false},
			{"auto_preview", true},
			{"ephemeral_storage", 4},
			{"arguments.0", "--declared"},
			{"ports.0.name", "declared"},
			{"ports.0.protocol", "GRPC"},
			{"custom_domains.0.generate_certificate", true},
			{"custom_domains.0.use_cdn", true},
		},
		defaults: []testAccContractValue{
			{"icon_uri", "app://qovery-console/container"},
			{"auto_deploy", true},
			{"auto_preview", false},
			{"ephemeral_storage", 0},
			{"arguments.#", nil},
			{"ports.0.name", "p8080"},
			{"ports.0.protocol", "HTTP"},
			{"custom_domains.0.generate_certificate", false},
			{"custom_domains.0.use_cdn", false},
		},
		outOfBand: func(id string) error {
			apiPath := "/container/" + id
			err := testAccEditServiceOutOfBand(apiPath, nil, func(cont map[string]any) {
				cont["icon_uri"] = "app://qovery-console/contract-console"
				cont["auto_deploy"] = false
				cont["auto_preview"] = true
				cont["ephemeral_storage_in_gib"] = 2
				cont["arguments"] = []string{"--console"}
				if registry, ok := cont["registry"].(map[string]any); ok {
					cont["registry_id"] = registry["id"]
				}
				if port, err := jsonFirst(cont, "ports"); err == nil {
					port["name"] = "console"
					port["protocol"] = "GRPC"
				}
			})
			if err != nil {
				return err
			}
			return testAccEditCustomDomainsOutOfBand(apiPath, true, true)
		},
		changed: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-console"},
			{"auto_deploy", false},
			{"auto_preview", true},
			{"ephemeral_storage", 2},
			{"arguments.0", "--console"},
			{"ports.0.name", "console"},
			{"ports.0.protocol", "GRPC"},
			{"custom_domains.0.generate_certificate", true},
			{"custom_domains.0.use_cdn", true},
		},
		exists:  testAccQoveryContainerExists(address),
		destroy: testAccQoveryContainerDestroy(address),
	})
}

func TestAcc_ContainerContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testName := "container-contract-upgrade"
	testAccServiceContractUpgradeFrom0x(t, testAccContainerContractConfig(testName, false), testAccQoveryContainerDestroy("qovery_container.test"))
}

func testAccContainerContractConfig(testName string, declared bool) string {
	attributes := `
  ports          = [{ internal_port = 8080, external_port = 443, publicly_accessible = true }]
  custom_domains = [{ domain = "` + testAccContractDomain(testName) + `" }]`
	if declared {
		attributes = `
  icon_uri          = "app://qovery-console/contract-declared"
  auto_deploy       = false
  auto_preview      = true
  ephemeral_storage = 4
  arguments         = ["--declared"]
  ports             = [{ name = "declared", internal_port = 8080, external_port = 443, publicly_accessible = true, protocol = "GRPC" }]
  custom_domains    = [{ domain = "` + testAccContractDomain(testName) + `", generate_certificate = true, use_cdn = true }]`
	}
	return fmt.Sprintf(`
%s
%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id    = qovery_container_registry.test.id
  name           = "%s"
  image_name     = "%s"
  tag            = "%s"
  healthchecks   = {}%s
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, attributes)
}

// --- job -----------------------------------------------------------------------------------------

// TestAcc_CronJobContract covers icon_uri, auto_deploy, auto_preview, ephemeral_storage and the
// cron job entrypoint on a job built from an image.
func TestAcc_CronJobContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_job.test"
	testName := "cron-job-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address: address,
		config:  func(declared bool) string { return testAccCronJobContractConfig(testName, declared) },
		declared: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-declared"},
			{"auto_deploy", false},
			{"auto_preview", true},
			{"ephemeral_storage", 4},
			{"schedule.cronjob.command.entrypoint", "/bin/declared"},
			{"schedule.lifecycle_type", nil},
		},
		defaults: []testAccContractValue{
			{"icon_uri", "app://qovery-console/cron-job"},
			{"auto_deploy", true},
			{"auto_preview", false},
			{"ephemeral_storage", 0},
			{"schedule.cronjob.command.entrypoint", nil},
			{"schedule.lifecycle_type", nil},
		},
		outOfBand: func(id string) error {
			return testAccEditServiceOutOfBand("/job/"+id, nil, func(job map[string]any) {
				job["icon_uri"] = "app://qovery-console/contract-console"
				job["auto_deploy"] = false
				job["auto_preview"] = true
				job["ephemeral_storage_in_gib"] = 2
				jsonObject(jsonObject(job, "schedule"), "cronjob")["entrypoint"] = "/bin/console"
			})
		},
		changed: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-console"},
			{"auto_deploy", false},
			{"auto_preview", true},
			{"ephemeral_storage", 2},
			{"schedule.cronjob.command.entrypoint", "/bin/console"},
		},
		exists:  testAccQoveryJobExists(address),
		destroy: testAccQoveryJobDestroy(address),
	})
}

func TestAcc_CronJobContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testName := "cron-job-contract-upgrade"
	testAccServiceContractUpgradeFrom0x(t, testAccCronJobContractConfig(testName, false), testAccQoveryJobDestroy("qovery_job.test"))
}

func testAccCronJobContractConfig(testName string, declared bool) string {
	attributes, entrypoint := "", ""
	if declared {
		attributes = `
  icon_uri          = "app://qovery-console/contract-declared"
  auto_deploy       = false
  auto_preview      = true
  ephemeral_storage = 4`
		entrypoint = `
        entrypoint = "/bin/declared"`
	}
	return fmt.Sprintf(`
%s
%s

resource "qovery_job" "test" {
  environment_id       = qovery_environment.test.id
  name                 = "%s"
  max_duration_seconds = 300
  max_nb_restart       = 0
  healthchecks         = {}%s
  source = {
    image = {
      registry_id = qovery_container_registry.test.id
      name        = "%s"
      tag         = "%s"
    }
  }
  schedule = {
    cronjob = {
      schedule = "%s"
      command = {%s
        arguments = ["arg1"]
      }
    }
  }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), attributes, jobImageName, jobImageTag, jobScheduleCronString, entrypoint)
}

// TestAcc_LifecycleJobContract covers icon_uri, lifecycle_type, the on_start entrypoint and
// root_path on a lifecycle job built from a Dockerfile. q-core cannot change the lifecycle type of
// an existing job, so the declared configuration sets the default GENERIC and the change is
// covered by TestAcc_LifecycleJobTypeChangeRejected.
func TestAcc_LifecycleJobContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_job.test"
	testName := "lifecycle-job-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address: address,
		config: func(declared bool) string {
			lifecycleType := ""
			if declared {
				lifecycleType = "GENERIC"
			}
			return testAccLifecycleJobContractConfig(testName, declared, false, lifecycleType)
		},
		declared: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-declared"},
			{"schedule.lifecycle_type", "GENERIC"},
			{"schedule.on_start.entrypoint", "/bin/declared"},
			{"source.docker.git_repository.root_path", "/simple_app"},
		},
		defaults: []testAccContractValue{
			{"icon_uri", "app://qovery-console/lifecycle-job"},
			{"schedule.lifecycle_type", "GENERIC"},
			{"schedule.on_start.entrypoint", nil},
			{"source.docker.git_repository.root_path", "/"},
		},
		outOfBand: func(id string) error {
			return testAccEditServiceOutOfBand("/job/"+id, nil, func(job map[string]any) {
				job["icon_uri"] = "app://qovery-console/contract-console"
				jsonObject(jsonObject(job, "schedule"), "on_start")["entrypoint"] = "/bin/console"
				jsonObject(jsonObject(jsonObject(job, "source"), "docker"), "git_repository")["root_path"] = "/several_services"
			})
		},
		changed: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-console"},
			{"schedule.lifecycle_type", "GENERIC"},
			{"schedule.on_start.entrypoint", "/bin/console"},
			{"source.docker.git_repository.root_path", "/several_services"},
		},
		exists:  testAccQoveryJobExists(address),
		destroy: testAccQoveryJobDestroy(address),
	})
}

// TestAcc_LifecycleJobContractUpgradeFrom0x declares root_path: 0.x sent "" when it was omitted,
// which q-core now rejects on create. The read of a stored "" is covered by the unit tests.
func TestAcc_LifecycleJobContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testName := "lifecycle-job-contract-upgrade"
	testAccServiceContractUpgradeFrom0x(t, testAccLifecycleJobContractConfig(testName, false, true, ""), testAccQoveryJobDestroy("qovery_job.test"))
}

// TestAcc_LifecycleJobTypeChangeRejected checks that a change of the lifecycle type of an existing
// job fails at plan time, whether the type is changed or omitted, since q-core rejects it at apply.
func TestAcc_LifecycleJobTypeChangeRejected(t *testing.T) {
	t.Parallel()
	const address = "qovery_job.test"
	testName := "lifecycle-job-type-change"
	rejected := regexp.MustCompile(`Cannot change schedule.lifecycle_type after creation`)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryJobDestroy(address),
		Steps: []resource.TestStep{
			{
				Config:           testAccLifecycleJobContractConfig(testName, false, false, "TERRAFORM"),
				Check:            resource.TestCheckResourceAttr(address, "schedule.lifecycle_type", "TERRAFORM"),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// Omitting the type plans the GENERIC default, which q-core would reject.
			{
				Config:      testAccLifecycleJobContractConfig(testName, false, false, ""),
				ExpectError: rejected,
			},
			{
				Config:      testAccLifecycleJobContractConfig(testName, false, false, "CLOUDFORMATION"),
				ExpectError: rejected,
			},
			// The recorded type is still accepted.
			{
				Config:   testAccLifecycleJobContractConfig(testName, false, false, "TERRAFORM"),
				PlanOnly: true,
			},
		},
	})
}

// testAccLifecycleJobContractConfig builds the job from an inline Dockerfile in the Helm chart
// repository: q-core checks that root_path is a folder of the repository, and that a
// dockerfile_path exists under it. lifecycleType is the declared lifecycle_type, "" to omit it.
func testAccLifecycleJobContractConfig(testName string, declared bool, for0x bool, lifecycleType string) string {
	icon, entrypoint, rootPath := "", "", ""
	if for0x {
		rootPath = `
        root_path    = "/"`
	}
	if lifecycleType != "" {
		lifecycleType = fmt.Sprintf(`
    lifecycle_type = "%s"`, lifecycleType)
	}
	if declared {
		icon = `
  icon_uri = "app://qovery-console/contract-declared"`
		entrypoint = `
      entrypoint = "/bin/declared"`
		rootPath = `
        root_path    = "/simple_app"`
	}
	return fmt.Sprintf(`
%s

resource "qovery_job" "test" {
  environment_id       = qovery_environment.test.id
  name                 = "%s"
  max_duration_seconds = 300
  max_nb_restart       = 0
  healthchecks         = {}%s
  source = {
    docker = {
      dockerfile_raw = "FROM busybox"
      git_repository = {
        url          = "%s"
        branch       = "%s"
        git_token_id = "%s"%s
      }
    }
  }
  schedule = {%s
    on_start = {%s
      arguments = ["arg1"]
    }
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), icon, helmGitRepositoryURL, "main", getTestQoverySandboxGitTokenID(), rootPath, lifecycleType, entrypoint)
}

// --- helm ----------------------------------------------------------------------------------------

// helmGitRepositoryURL is the public repository holding the test Helm charts.
const helmGitRepositoryURL = "https://github.com/Qovery/helm_chart_engine_testing.git"

// TestAcc_HelmContract covers icon_uri, auto_deploy, auto_preview, ports.protocol, the custom
// domain use_cdn and both git token ids on the helm service. The configuration that omits them
// also covers two create and apply errors 0.x hit: a port without protocol, and set = {} with
// set_string unset.
func TestAcc_HelmContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_helm.test"
	testName := "helm-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address: address,
		config:  func(declared bool) string { return testAccHelmContractConfig(testName, declared, false) },
		declared: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-declared"},
			{"auto_deploy", true},
			{"auto_preview", true},
			{"ports.web.protocol", "GRPC"},
			{"custom_domains.0.use_cdn", true},
			{"source.git_repository.git_token_id", getTestQoverySandboxGitTokenID()},
			{"values_override.file.git_repository.git_token_id", getTestQoverySandboxGitTokenID()},
		},
		defaults: []testAccContractValue{
			{"icon_uri", "app://qovery-console/helm"},
			{"auto_deploy", false},
			{"auto_preview", false},
			{"ports.web.protocol", "HTTP"},
			{"custom_domains.0.use_cdn", false},
			{"source.git_repository.git_token_id", nil},
			{"values_override.file.git_repository.git_token_id", nil},
		},
		outOfBand: func(id string) error {
			apiPath := "/helm/" + id
			err := testAccEditServiceOutOfBand(apiPath, testAccHelmRequestFromResponse, func(helm map[string]any) {
				helm["icon_uri"] = "app://qovery-console/contract-console"
				helm["auto_deploy"] = true
				helm["auto_preview"] = true
				if port, err := jsonFirst(helm, "ports"); err == nil {
					port["protocol"] = "GRPC"
				}
				jsonObject(jsonObject(helm, "source"), "git_repository")["git_token_id"] = getTestQoverySandboxGitTokenID()
				file := jsonObject(jsonObject(helm, "values_override"), "file")
				jsonObject(jsonObject(file, "git"), "git_repository")["git_token_id"] = getTestQoverySandboxGitTokenID()
			})
			if err != nil {
				return err
			}
			return testAccEditCustomDomainsOutOfBand(apiPath, true, true)
		},
		changed: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-console"},
			{"auto_deploy", true},
			{"auto_preview", true},
			{"ports.web.protocol", "GRPC"},
			{"custom_domains.0.use_cdn", true},
			{"source.git_repository.git_token_id", getTestQoverySandboxGitTokenID()},
			{"values_override.file.git_repository.git_token_id", getTestQoverySandboxGitTokenID()},
		},
		exists:  testAccQoveryHelmExists(address),
		destroy: testAccQoveryHelmDestroy(address),
	})
}

// TestAcc_HelmContractUpgradeFrom0x declares the port protocol, which 0.x required, and leaves
// set unset, whose empty value 0.x read back with the wrong shape.
func TestAcc_HelmContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testName := "helm-contract-upgrade"
	testAccServiceContractUpgradeFrom0x(t, testAccHelmContractConfig(testName, false, true), testAccQoveryHelmDestroy("qovery_helm.test"))
}

func testAccHelmContractConfig(testName string, declared bool, for0x bool) string {
	attributes, protocol, useCdn, gitToken, set := "", "", "", "", `
    set = {}`
	if declared {
		attributes = `
  icon_uri     = "app://qovery-console/contract-declared"
  auto_deploy  = true
  auto_preview = true`
		protocol = `, protocol = "GRPC"`
		useCdn = `, use_cdn = true`
		gitToken = fmt.Sprintf(`
      git_token_id = "%s"`, getTestQoverySandboxGitTokenID())
	}
	if for0x {
		protocol = `, protocol = "HTTP"`
		set = ""
	}
	return fmt.Sprintf(`
%s

resource "qovery_helm" "test" {
  environment_id               = qovery_environment.test.id
  name                         = "%s"
  description                  = "helm contract"
  allow_cluster_wide_resources = false%s
  source = {
    git_repository = {
      url       = "%s"
      branch    = "main"
      root_path = "/simple_app"%s
    }
  }
  values_override = {%s
    file = {
      git_repository = {
        url    = "%s"
        branch = "main"
        paths  = ["/simple_app/values.yaml"]%s
      }
    }
  }
  ports = {
    web = { service_name = "simple-app", internal_port = 80, external_port = 443%s }
  }
  custom_domains = [{ domain = "%s", generate_certificate = true%s }]
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), attributes, helmGitRepositoryURL, gitToken, set, helmGitRepositoryURL, gitToken, protocol, testAccContractDomain(testName), useCdn)
}

// testAccHelmRequestFromResponse turns the helm source of a GET response into the shape of the
// edit request.
func testAccHelmRequestFromResponse(helm map[string]any) {
	source, _ := helm["source"].(map[string]any)
	switch {
	case source["git"] != nil:
		repository := jsonObject(jsonObject(source, "git"), "git_repository")
		helm["source"] = map[string]any{"git_repository": map[string]any{
			"url":          repository["url"],
			"branch":       repository["branch"],
			"root_path":    repository["root_path"],
			"git_token_id": repository["git_token_id"],
		}}
	case source["repository"] != nil:
		repository := jsonObject(source, "repository")
		helm["source"] = map[string]any{"helm_repository": map[string]any{
			"repository":    jsonObject(repository, "repository")["id"],
			"chart_name":    repository["chart_name"],
			"chart_version": repository["chart_version"],
		}}
	}
}

// --- terraform_service ---------------------------------------------------------------------------

// TestAcc_TerraformServiceContract covers the non-secret variables: a value changed outside
// Terraform shows on refresh, where 0.x kept the state value.
func TestAcc_TerraformServiceContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_terraform_service.test"
	testName := "terraform-service-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address:             address,
		config:              func(declared bool) string { return testAccTerraformServiceContractConfig(testName, declared) },
		outOfBandOnDeclared: true,
		declared: []testAccContractValue{
			{"variables.0.key", "REGION"},
			{"variables.0.value", "eu-west-3"},
		},
		defaults: []testAccContractValue{
			{"variables.#", nil},
		},
		outOfBand: func(id string) error {
			return testAccEditServiceOutOfBand("/terraform/"+id, testAccTerraformRequestFromResponse, func(tf map[string]any) {
				jsonObject(tf, "terraform_variables_source")["tf_vars"] = []map[string]any{
					{"key": "REGION", "value": "us-east-1", "secret": false},
				}
			})
		},
		changed: []testAccContractValue{
			{"variables.0.key", "REGION"},
			{"variables.0.value", "us-east-1"},
		},
		importStateVerifyIgnore: []string{"updated_at"},
		exists:                  testAccQoveryTerraformServiceExists(address),
		destroy:                 testAccQoveryTerraformServiceDestroy(address),
	})
}

func TestAcc_TerraformServiceContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testName := "terraform-service-contract-upgrade"
	testAccServiceContractUpgradeFrom0x(t, testAccTerraformServiceContractConfig(testName, true), testAccQoveryTerraformServiceDestroy("qovery_terraform_service.test"))
}

func testAccTerraformServiceContractConfig(testName string, declared bool) string {
	variables := ""
	if declared {
		variables = `
  variables = [{ key = "REGION", value = "eu-west-3" }]`
	}
	return fmt.Sprintf(`
%s

resource "qovery_terraform_service" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  description    = "terraform service contract"
  auto_deploy    = true

  git_repository = {
    url          = "https://github.com/Qovery/terraform-examples.git"
    branch       = "main"
    root_path    = "/"
    git_token_id = "%s"
  }

  tfvars_files = []%s

  backend = {
    kubernetes = {}
  }

  engine = "TERRAFORM"

  engine_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 20
  }

  timeout_seconds         = 1800
  use_cluster_credentials = false
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), getTestQoverySandboxGitTokenID(), variables)
}

// testAccTerraformRequestFromResponse turns the files source of a GET response into the shape of
// the edit request.
func testAccTerraformRequestFromResponse(tf map[string]any) {
	source, _ := tf["terraform_files_source"].(map[string]any)
	if git, ok := source["git"].(map[string]any); ok {
		repository := jsonObject(git, "git_repository")
		tf["terraform_files_source"] = map[string]any{"git_repository": map[string]any{
			"url":          repository["url"],
			"branch":       repository["branch"],
			"root_path":    repository["root_path"],
			"git_token_id": repository["git_token_id"],
		}}
	}
	if _, ok := tf["auto_deploy_config"]; !ok {
		tf["auto_deploy_config"] = map[string]any{"auto_deploy": tf["auto_deploy"], "terraform_action": "DEFAULT"}
	}
}

// --- database ------------------------------------------------------------------------------------

// TestAcc_DatabaseContract covers icon_uri on a CONTAINER database. Its first step checks that a
// MANAGED database without instance_type fails at plan time.
func TestAcc_DatabaseContract(t *testing.T) {
	t.Parallel()
	const address = "qovery_database.test"
	testName := "database-contract"
	testAccServiceContract(t, testAccServiceContractTarget{
		address: address,
		config:  func(declared bool) string { return testAccDatabaseContractConfig(testName, declared) },
		declared: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-declared"},
		},
		defaults: []testAccContractValue{
			{"icon_uri", "app://qovery-console/database"},
		},
		outOfBand: func(id string) error {
			return testAccEditServiceOutOfBand("/database/"+id, nil, func(db map[string]any) {
				db["icon_uri"] = "app://qovery-console/contract-console"
			})
		},
		changed: []testAccContractValue{
			{"icon_uri", "app://qovery-console/contract-console"},
		},
		exists:  testAccQoveryDatabaseExists(address),
		destroy: testAccQoveryDatabaseDestroy(address),
	})
}

func TestAcc_DatabaseContractUpgradeFrom0x(t *testing.T) {
	t.Parallel()
	testName := "database-contract-upgrade"
	testAccServiceContractUpgradeFrom0x(t, testAccDatabaseContractConfig(testName, false), testAccQoveryDatabaseDestroy("qovery_database.test"))
}

func testAccDatabaseContractConfig(testName string, declared bool) string {
	icon := ""
	if declared {
		icon = `
  icon_uri = "app://qovery-console/contract-declared"`
	}
	return fmt.Sprintf(`
%s

resource "qovery_database" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  type           = "REDIS"
  version        = "6.2"
  mode           = "CONTAINER"%s
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), icon)
}

// TestAcc_DatabaseManagedRequiresInstanceType checks at plan time that a MANAGED database declares
// instance_type, which the Qovery API requires.
func TestAcc_DatabaseManagedRequiresInstanceType(t *testing.T) {
	t.Parallel()
	testName := "database-managed-instance-type"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "qovery_database" "test" {
  environment_id = "00000000-0000-0000-0000-000000000000"
  name           = "%s"
  type           = "POSTGRESQL"
  version        = "16"
  mode           = "MANAGED"
}
`, generateTestName(testName)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Missing instance_type`),
			},
		},
	})
}
