//go:build unit && !integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
)

func i32(v int32) *int32  { return &v }
func sp(v string) *string { return &v }

func TestAutoscaling_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers: []autoscaling.Scaler{
			{
				ScalerType: "prometheus",
				Enabled:    true,
				Role:       autoscaling.RolePrimary,
				Config:     autoscaling.Config{ConfigJSON: `{"query":"up","threshold":"1"}`},
				TriggerAuth: &autoscaling.TriggerAuth{
					Name:       "auth",
					ConfigYAML: sp("foo: bar"),
				},
			},
			{
				ScalerType: "cron",
				Enabled:    false,
				Role:       autoscaling.RoleSafety,
				Config:     autoscaling.Config{ConfigYAML: "start: 0 0 * * *"},
			},
		},
	}

	obj := fromAutoscaling(original, types.ObjectNull(autoscalingAttrTypes()))
	require.False(t, obj.IsNull())

	got := toQoveryAutoscaling(obj)
	require.NotNil(t, got)

	require.NotNil(t, got.PollingIntervalSeconds)
	assert.Equal(t, int32(30), *got.PollingIntervalSeconds)
	require.NotNil(t, got.CooldownPeriodSeconds)
	assert.Equal(t, int32(300), *got.CooldownPeriodSeconds)
	require.Len(t, got.Scalers, 2)

	byType := map[string]autoscaling.Scaler{}
	for _, s := range got.Scalers {
		byType[s.ScalerType] = s
	}

	prom := byType["prometheus"]
	assert.True(t, prom.Enabled)
	assert.Equal(t, autoscaling.RolePrimary, prom.Role)
	assert.JSONEq(t, `{"query":"up","threshold":"1"}`, prom.Config.ConfigJSON)
	assert.Empty(t, prom.Config.ConfigYAML)
	require.NotNil(t, prom.TriggerAuth)
	assert.Equal(t, "auth", prom.TriggerAuth.Name)
	require.NotNil(t, prom.TriggerAuth.ConfigYAML)
	assert.Equal(t, "foo: bar", *prom.TriggerAuth.ConfigYAML)

	cron := byType["cron"]
	assert.False(t, cron.Enabled)
	assert.Equal(t, autoscaling.RoleSafety, cron.Role)
	assert.Equal(t, "start: 0 0 * * *", cron.Config.ConfigYAML)
	assert.Empty(t, cron.Config.ConfigJSON)
	assert.Nil(t, cron.TriggerAuth)
}

func TestAutoscaling_NilRoundTrip(t *testing.T) {
	t.Parallel()

	obj := fromAutoscaling(nil, types.ObjectNull(autoscalingAttrTypes()))
	assert.True(t, obj.IsNull())
	assert.Nil(t, toQoveryAutoscaling(obj))
}

// promScaler returns a prometheus PRIMARY scaler on query, with a trigger
// authentication named name (none when name is "").
func promScaler(query, name string) autoscaling.Scaler {
	s := autoscaling.Scaler{
		ScalerType: "prometheus",
		Enabled:    true,
		Role:       autoscaling.RolePrimary,
		Config:     autoscaling.Config{ConfigJSON: `{"query":"` + query + `"}`},
	}
	if name != "" {
		s.TriggerAuth = &autoscaling.TriggerAuth{Name: name, ConfigYAML: sp("secretTargetRef: []")}
	}
	return s
}

func TestTriggerAuthNames(t *testing.T) {
	t.Parallel()

	reformatted := promScaler("up", "configured")
	reformatted.Config.ConfigJSON = "{ \"query\" : \"up\" }"

	otherAuthConfig := promScaler("up", "configured")
	otherAuthConfig.TriggerAuth.ConfigYAML = sp("secretTargetRef: [{parameter: apiKey}]")

	testCases := []struct {
		TestName string
		Actual   []autoscaling.Scaler
		Prior    *autoscaling.AutoscalingPolicy
		Expected []string
	}{
		{
			TestName: "no_prior_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    nil,
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "matching_scaler_keeps_configured_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{promScaler("up", "configured")}},
			Expected: []string{"configured"},
		},
		{
			TestName: "config_json_compared_as_json",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{reformatted}},
			Expected: []string{"configured"},
		},
		{
			TestName: "each_scaler_keeps_its_own_name",
			Actual: []autoscaling.Scaler{
				promScaler("up", "a676e575-1-auth"),
				promScaler("down", "a676e575-2-auth"),
			},
			Prior: &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{
				promScaler("down", "configured-down"),
				promScaler("up", "configured-up"),
			}},
			Expected: []string{"configured-up", "configured-down"},
		},
		{
			TestName: "scalers_equal_but_name_each_take_one_name",
			Actual: []autoscaling.Scaler{
				promScaler("up", "a676e575-1-auth"),
				promScaler("up", "a676e575-2-auth"),
			},
			Prior: &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{
				promScaler("up", "first"),
				promScaler("up", "second"),
			}},
			Expected: []string{"first", "second"},
		},
		{
			TestName: "scaler_changed_outside_terraform_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("down", "a676e575-1-auth")},
			Prior:    &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{promScaler("up", "configured")}},
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "trigger_auth_config_changed_outside_terraform_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{otherAuthConfig}},
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "trigger_auth_added_outside_terraform_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{promScaler("up", "")}},
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "scaler_without_trigger_auth_has_no_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "")},
			Prior:    &autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{promScaler("up", "configured")}},
			Expected: []string{""},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expected, triggerAuthNames(tc.Actual, tc.Prior))
		})
	}
}

// The state keeps the configured name of an inline trigger authentication,
// which the API replaces with a generated one.
func TestFromAutoscaling_KeepsConfiguredTriggerAuthName(t *testing.T) {
	t.Parallel()

	prior := fromAutoscaling(&autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers:                []autoscaling.Scaler{promScaler("up", "vocca-livekit-eu-staging-datadog-calls")},
	}, types.ObjectNull(autoscalingAttrTypes()))

	actual := &autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers:                []autoscaling.Scaler{promScaler("up", "a676e575-1793957741362-auth")},
	}

	assert.True(t, fromAutoscaling(actual, prior).Equal(prior))
	assert.Equal(t, "a676e575-1793957741362-auth", actual.Scalers[0].TriggerAuth.Name, "the API policy must not be modified")
}
