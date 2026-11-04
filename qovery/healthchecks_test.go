//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthChecks_StartupProbe(t *testing.T) {
	t.Parallel()

	startupProbe := &Probe{
		InitialDelaySeconds: types.Int64Value(10),
		PeriodSeconds:       types.Int64Value(5),
		TimeoutSeconds:      types.Int64Value(2),
		SuccessThreshold:    types.Int64Value(1),
		FailureThreshold:    types.Int64Value(60),
		Type: ProbeType{
			Http: &ProbeHttp{
				Port:   types.Int64Value(8080),
				Path:   types.StringValue("/started"),
				Scheme: types.StringValue("HTTP"),
			},
		},
	}

	testCases := []struct {
		TestName             string
		HealthChecks         HealthChecks
		ExpectedStartupProbe string
	}{
		{
			TestName:     "configured_startup_probe_is_sent",
			HealthChecks: HealthChecks{StartupProbe: startupProbe},
			ExpectedStartupProbe: `{
				"type": {"http": {"port": 8080, "path": "/started", "scheme": "HTTP"}, "tcp": null},
				"initial_delay_seconds": 10,
				"period_seconds": 5,
				"timeout_seconds": 2,
				"success_threshold": 1,
				"failure_threshold": 60
			}`,
		},
		{
			// An edit that leaves the startup probe out removes it, so the provider sends null to clear it.
			TestName:             "missing_startup_probe_is_sent_as_null",
			HealthChecks:         HealthChecks{},
			ExpectedStartupProbe: `null`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tc.HealthChecks.toHealthchecksRequest())
			require.NoError(t, err)

			var request map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(body, &request))
			require.Contains(t, request, "startup_probe")
			assert.JSONEq(t, tc.ExpectedStartupProbe, string(request["startup_probe"]))
		})
	}
}

func TestConvertHealthchecksResponseToDomain_StartupProbe(t *testing.T) {
	t.Parallel()

	t.Run("startup_probe_is_read_back", func(t *testing.T) {
		t.Parallel()

		var response qovery.Healthcheck
		require.NoError(t, json.Unmarshal([]byte(`{
			"startup_probe": {
				"type": {"exec": {"command": ["cat", "/tmp/started"]}},
				"initial_delay_seconds": 10,
				"period_seconds": 5,
				"timeout_seconds": 2,
				"success_threshold": 1,
				"failure_threshold": 60
			}
		}`), &response))

		healthchecks := convertHealthchecksResponseToDomain(response)

		assert.Nil(t, healthchecks.ReadinessProbe)
		assert.Nil(t, healthchecks.LivenessProbe)
		assert.Equal(t, &Probe{
			InitialDelaySeconds: types.Int64Value(10),
			PeriodSeconds:       types.Int64Value(5),
			TimeoutSeconds:      types.Int64Value(2),
			SuccessThreshold:    types.Int64Value(1),
			FailureThreshold:    types.Int64Value(60),
			Type: ProbeType{
				Exec: &ProbeExec{
					Command: FromStringArray([]string{"cat", "/tmp/started"}),
				},
			},
		}, healthchecks.StartupProbe)
	})

	t.Run("no_startup_probe_is_null", func(t *testing.T) {
		t.Parallel()

		for _, body := range []string{`{}`, `{"startup_probe": null}`} {
			var response qovery.Healthcheck
			require.NoError(t, json.Unmarshal([]byte(body), &response))

			assert.Nil(t, convertHealthchecksResponseToDomain(response).StartupProbe, body)
		}
	})
}
