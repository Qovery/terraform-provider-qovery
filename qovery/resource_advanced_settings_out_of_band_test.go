//go:build integration && !unit

package qovery_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
)

// TestAcc_ApplicationAdvancedSettingsResetOutOfBand reproduces QOV-2028: an
// advanced setting tracked in state that is reset to its API default outside
// Terraform (e.g. from the Qovery Console) must be reflected on refresh, and the
// key must stay tracked so that later out-of-band changes remain visible.
//
// The application resource is used because it shares computeOverriddenSettings
// with the cluster resource and is far cheaper to provision. build.timeout_max_sec
// is the setting already declared by testAccApplicationDefaultConfig (1700).
//
// Before the fix the refresh kept the stale state value when the remote value
// went back to the default, so the plan after the out-of-band reset was empty and
// the drift was invisible: step 2 fails on that code.
func TestAcc_ApplicationAdvancedSettingsResetOutOfBand(t *testing.T) {
	t.Parallel()
	testName := "application-adv-settings-reset-oob"
	const (
		resourceName = "qovery_application.test"
		settingKey   = "build.timeout_max_sec"
		configured   = int32(1700)
		changed      = int32(1600)
	)

	var applicationID string
	var defaultValue int32

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			defaultValue = testAccApplicationDefaultBuildTimeoutMaxSec(t)
			if defaultValue == configured || defaultValue == changed {
				t.Fatalf("API default for %s is %d; the test needs a default distinct from %d and %d", settingKey, defaultValue, configured, changed)
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy(resourceName),
		Steps: []resource.TestStep{
			// 1. Apply: the key becomes tracked in state.
			{
				Config: testAccApplicationDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists(resourceName),
					testAccCheckAdvancedSettingsJSON(resourceName, settingKey, configured),
					testAccCaptureResourceID(resourceName, &applicationID),
				),
			},
			// 2. Reset the key to its default out of band. The harness plans (with refresh)
			// after this step: the plan must be non-empty (default -> 1700).
			{
				Config: testAccApplicationDefaultConfig(testName),
				Check: func(_ *terraform.State) error {
					return testAccSetApplicationAdvancedSettingOutOfBand(applicationID, settingKey, defaultValue)
				},
				ExpectNonEmptyPlan: true,
			},
			// 3. Refresh only: state now holds the default, not the stale 1700.
			{
				RefreshState: true,
				Check: func(s *terraform.State) error {
					return testAccCheckAdvancedSettingsJSON(resourceName, settingKey, defaultValue)(s)
				},
				ExpectNonEmptyPlan: true,
			},
			// 4. Another out-of-band change, then refresh only: the key is still
			// tracked after the reset, so the new value is picked up.
			{
				PreConfig: func() {
					if err := testAccSetApplicationAdvancedSettingOutOfBand(applicationID, settingKey, changed); err != nil {
						t.Fatal(err)
					}
				},
				RefreshState:       true,
				Check:              testAccCheckAdvancedSettingsJSON(resourceName, settingKey, changed),
				ExpectNonEmptyPlan: true,
			},
			// 5. Corrective apply converges back to the configured value.
			{
				Config: testAccApplicationDefaultConfig(testName),
				Check:  testAccCheckAdvancedSettingsJSON(resourceName, settingKey, configured),
			},
			// 6. Stable afterwards.
			{
				Config:             testAccApplicationDefaultConfig(testName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccApplicationDefaultBuildTimeoutMaxSec returns the API default of
// build.timeout_max_sec for applications.
func testAccApplicationDefaultBuildTimeoutMaxSec(t *testing.T) int32 {
	t.Helper()
	defaults, res, err := qoveryAPIClient.ApplicationsAPI.GetDefaultApplicationAdvancedSettings(context.TODO()).Execute()
	if err != nil {
		t.Fatalf("failed to fetch default application advanced settings: %s", err)
	}
	if res != nil && res.StatusCode >= 400 {
		t.Fatalf("failed to fetch default application advanced settings: HTTP %d", res.StatusCode)
	}
	if defaults == nil || defaults.BuildTimeoutMaxSec == nil {
		t.Fatal("default application advanced settings do not contain build.timeout_max_sec")
	}
	return *defaults.BuildTimeoutMaxSec
}

// testAccSetApplicationAdvancedSettingOutOfBand writes one advanced setting
// directly through the API, bypassing Terraform, the way the Qovery Console does.
// The update merges the other keys from the current remote values.
func testAccSetApplicationAdvancedSettingOutOfBand(applicationID string, key string, value int32) error {
	if applicationID == "" {
		return fmt.Errorf("application id was not captured from state")
	}
	body, err := json.Marshal(map[string]any{key: value})
	if err != nil {
		return err
	}
	svc := advanced_settings.NewServiceAdvancedSettingsService(qoveryAPIClient.GetConfig())
	if err := svc.UpdateServiceAdvancedSettings(domain.APPLICATION, applicationID, string(body)); err != nil {
		return fmt.Errorf("failed to set %s=%d out of band: %w", key, value, err)
	}
	return nil
}

// testAccCheckAdvancedSettingsJSON asserts that advanced_settings_json in state
// contains key with the given numeric value.
func testAccCheckAdvancedSettingsJSON(resourceName string, key string, want int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s: not found in state", resourceName)
		}
		raw := rs.Primary.Attributes["advanced_settings_json"]
		var settings map[string]any
		if err := json.Unmarshal([]byte(raw), &settings); err != nil {
			return fmt.Errorf("%s: advanced_settings_json is not valid JSON (%q): %w", resourceName, raw, err)
		}
		got, ok := settings[key]
		if !ok {
			return fmt.Errorf("%s: %s missing from advanced_settings_json %s", resourceName, key, raw)
		}
		num, ok := got.(float64)
		if !ok || num != float64(want) {
			return fmt.Errorf("%s: %s = %v, want %d (advanced_settings_json = %s)", resourceName, key, got, want, raw)
		}
		return nil
	}
}

// testAccCaptureResourceID stores the resource's ID for use in later steps that
// have no access to state (PreConfig).
func testAccCaptureResourceID(resourceName string, into *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("%s: id not found in state", resourceName)
		}
		*into = rs.Primary.ID
		return nil
	}
}
