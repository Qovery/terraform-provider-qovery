//go:build unit && !integration

package qovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
)

func i32(v int32) *int32 { return &v }
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

	obj := fromAutoscaling(original)
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

	obj := fromAutoscaling(nil)
	assert.True(t, obj.IsNull())
	assert.Nil(t, toQoveryAutoscaling(obj))
}
