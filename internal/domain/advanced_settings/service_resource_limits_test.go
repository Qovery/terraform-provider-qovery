//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
)

func int32Ptr(v int32) *int32 {
	return &v
}

func decodeSettings(t *testing.T, raw string) map[string]any {
	t.Helper()
	settings := make(map[string]any)
	require.NoError(t, json.Unmarshal([]byte(raw), &settings))
	return settings
}

func TestLimitsToRaise(t *testing.T) {
	t.Parallel()

	const (
		ram = ResourcesOverrideLimitRamInMibKey
		cpu = ResourcesOverrideLimitCpuInMilliKey
	)

	testCases := []struct {
		TestName   string
		Current    string
		Planned    string
		PlannedCPU *int32
		PlannedRAM *int32
		Expected   map[string]any
	}{
		{
			TestName:   "raise_ram_above_stored_limit",
			Current:    `{"` + ram + `": 1024, "` + cpu + `": null}`,
			Planned:    `{"` + ram + `": 2048}`,
			PlannedCPU: int32Ptr(500),
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{ram: float64(2048)},
		},
		{
			TestName:   "raise_cpu_above_stored_limit",
			Current:    `{"` + ram + `": null, "` + cpu + `": 500}`,
			Planned:    `{"` + cpu + `": 1000}`,
			PlannedCPU: int32Ptr(1000),
			PlannedRAM: int32Ptr(512),
			Expected:   map[string]any{cpu: float64(1000)},
		},
		{
			TestName:   "raise_ram_and_lower_cpu",
			Current:    `{"` + ram + `": 1024, "` + cpu + `": 2000}`,
			Planned:    `{"` + ram + `": 2048, "` + cpu + `": 500}`,
			PlannedCPU: int32Ptr(250),
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{ram: float64(2048)},
		},
		{
			TestName:   "lower_request_and_limit",
			Current:    `{"` + ram + `": 2048}`,
			Planned:    `{"` + ram + `": 1024}`,
			PlannedRAM: int32Ptr(1024),
			Expected:   map[string]any{},
		},
		{
			TestName:   "raise_request_within_stored_limit",
			Current:    `{"` + ram + `": 2048}`,
			Planned:    `{"` + ram + `": 2048}`,
			PlannedRAM: int32Ptr(1536),
			Expected:   map[string]any{},
		},
		{
			TestName:   "request_equal_to_stored_limit",
			Current:    `{"` + ram + `": 1024}`,
			Planned:    `{"` + ram + `": 2048}`,
			PlannedRAM: int32Ptr(1024),
			Expected:   map[string]any{},
		},
		{
			TestName:   "limit_not_in_plan",
			Current:    `{"` + ram + `": 1024}`,
			Planned:    `{}`,
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{},
		},
		{
			TestName:   "no_stored_limit",
			Current:    `{"` + ram + `": null}`,
			Planned:    `{"` + ram + `": 2048}`,
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{},
		},
		{
			TestName:   "stored_limit_missing_from_api",
			Current:    `{}`,
			Planned:    `{"` + ram + `": 2048}`,
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{},
		},
		{
			TestName:   "stored_cpu_limit_disabled",
			Current:    `{"` + cpu + `": -1}`,
			Planned:    `{"` + cpu + `": 1000}`,
			PlannedCPU: int32Ptr(1000),
			Expected:   map[string]any{},
		},
		{
			TestName:   "planned_limit_null",
			Current:    `{"` + ram + `": 1024}`,
			Planned:    `{"` + ram + `": null}`,
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{ram: nil},
		},
		{
			TestName:   "planned_cpu_limit_disabled",
			Current:    `{"` + cpu + `": 500}`,
			Planned:    `{"` + cpu + `": -1}`,
			PlannedCPU: int32Ptr(1000),
			Expected:   map[string]any{cpu: float64(-1)},
		},
		{
			TestName:   "planned_ram_limit_minus_one_is_not_disabled",
			Current:    `{"` + ram + `": 1024}`,
			Planned:    `{"` + ram + `": -1}`,
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{},
		},
		{
			TestName:   "planned_limit_below_planned_request",
			Current:    `{"` + ram + `": 1024}`,
			Planned:    `{"` + ram + `": 1536}`,
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{},
		},
		{
			TestName:   "invalid_cpu_pair_blocks_ram_raise",
			Current:    `{"` + ram + `": 1024, "` + cpu + `": 500}`,
			Planned:    `{"` + ram + `": 2048, "` + cpu + `": 800}`,
			PlannedCPU: int32Ptr(1000),
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{},
		},
		{
			TestName:   "planned_limit_not_a_number",
			Current:    `{"` + ram + `": 1024}`,
			Planned:    `{"` + ram + `": "2048"}`,
			PlannedRAM: int32Ptr(2048),
			Expected:   map[string]any{},
		},
		{
			TestName: "planned_request_unknown",
			Current:  `{"` + ram + `": 1024}`,
			Planned:  `{"` + ram + `": 2048}`,
			Expected: map[string]any{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			raised := limitsToRaise(decodeSettings(t, tc.Current), decodeSettings(t, tc.Planned), tc.PlannedCPU, tc.PlannedRAM)
			assert.Equal(t, tc.Expected, raised)
		})
	}
}

func TestRaiseResourceLimitsBeforeRequestUpdate_PutBody(t *testing.T) {
	t.Parallel()

	const currentSettings = `{
		"build.timeout_max_sec": 1800,
		"build.skip_git_submodules": false,
		"network.ingress.proxy_body_size_mb": 100,
		"resources.override.limit.ram_in_mib": 1024,
		"resources.override.limit.cpu_in_milli": 2000
	}`

	testCases := []struct {
		TestName        string
		PlanJson        string
		PlannedCPU      *int32
		PlannedRAM      *int32
		ExpectedPutBody map[string]any
	}{
		{
			TestName:   "only_the_raised_limit_changes",
			PlanJson:   `{"resources.override.limit.ram_in_mib": 2048, "resources.override.limit.cpu_in_milli": 500, "network.ingress.proxy_body_size_mb": 200}`,
			PlannedCPU: int32Ptr(250),
			PlannedRAM: int32Ptr(2048),
			ExpectedPutBody: map[string]any{
				"network.ingress.proxy_body_size_mb":    float64(100),
				"resources.override.limit.ram_in_mib":   float64(2048),
				"resources.override.limit.cpu_in_milli": float64(2000),
			},
		},
		{
			TestName:        "nothing_to_raise",
			PlanJson:        `{"resources.override.limit.ram_in_mib": 512}`,
			PlannedCPU:      int32Ptr(500),
			PlannedRAM:      int32Ptr(512),
			ExpectedPutBody: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			transport := &putCapturingTransport{getBody: currentSettings}
			cfg := qovery.NewConfiguration()
			cfg.Servers = qovery.ServerConfigurations{{URL: "https://api.qovery.test"}}
			cfg.HTTPClient = &http.Client{Transport: transport}

			err := NewServiceAdvancedSettingsService(cfg).RaiseResourceLimitsBeforeRequestUpdate(context.Background(), domain.APPLICATION, "app-1", tc.PlanJson, tc.PlannedCPU, tc.PlannedRAM)
			require.NoError(t, err)

			if tc.ExpectedPutBody == nil {
				assert.Nil(t, transport.putBody, "no PUT is sent when no limit needs raising")
				return
			}
			var putBody map[string]any
			require.NoError(t, json.Unmarshal(transport.putBody, &putBody))
			assert.Equal(t, tc.ExpectedPutBody, putBody)
		})
	}
}

func TestRaiseResourceLimitsBeforeRequestUpdate_NoCallWithoutPlannedLimitOrRequest(t *testing.T) {
	t.Parallel()

	const plannedLimits = `{"resources.override.limit.ram_in_mib": 8192, "resources.override.limit.cpu_in_milli": 4000}`

	testCases := []struct {
		TestName   string
		PlanJson   string
		PlannedCPU *int32
		PlannedRAM *int32
	}{
		{TestName: "empty_plan", PlanJson: "", PlannedCPU: int32Ptr(4000), PlannedRAM: int32Ptr(8192)},
		{TestName: "plan_without_limits", PlanJson: `{"network.ingress.proxy_body_size_mb": 200}`, PlannedCPU: int32Ptr(4000), PlannedRAM: int32Ptr(8192)},
		{TestName: "no_planned_request", PlanJson: plannedLimits},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			cfg, transport := newRecordingConfig()
			err := NewServiceAdvancedSettingsService(cfg).RaiseResourceLimitsBeforeRequestUpdate(context.Background(), domain.JOB, "job-1", tc.PlanJson, tc.PlannedCPU, tc.PlannedRAM)
			require.NoError(t, err)
			assert.Empty(t, transport.requests)
		})
	}
}
