//go:build unit && !integration

package autoscaling_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
)

func ptrInt32(v int32) *int32 { return &v }
func ptrStr(v string) *string { return &v }

func TestRole_IsValid(t *testing.T) {
	assert.True(t, autoscaling.RolePrimary.IsValid())
	assert.True(t, autoscaling.RoleSafety.IsValid())
	assert.False(t, autoscaling.Role("OTHER").IsValid())
}

func TestConfig_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		config    autoscaling.Config
		expectErr bool
	}{
		{"json only", autoscaling.Config{ConfigJSON: `{"a":1}`}, false},
		{"yaml only", autoscaling.Config{ConfigYAML: "a: 1"}, false},
		{"both set", autoscaling.Config{ConfigJSON: `{"a":1}`, ConfigYAML: "a: 1"}, true},
		{"neither set", autoscaling.Config{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAutoscalingPolicy_Validate(t *testing.T) {
	validScaler := autoscaling.Scaler{
		ScalerType: "cpu",
		Enabled:    true,
		Role:       autoscaling.RolePrimary,
		Config:     autoscaling.Config{ConfigJSON: `{"value":"80"}`},
	}

	t.Run("valid", func(t *testing.T) {
		p := autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{validScaler}}
		assert.NoError(t, p.Validate())
	})

	t.Run("no scalers", func(t *testing.T) {
		p := autoscaling.AutoscalingPolicy{}
		assert.ErrorIs(t, p.Validate(), autoscaling.ErrNoScalers)
	})

	t.Run("invalid role", func(t *testing.T) {
		s := validScaler
		s.Role = autoscaling.Role("BAD")
		p := autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{s}}
		assert.Error(t, p.Validate())
	})

	t.Run("empty scaler type", func(t *testing.T) {
		s := validScaler
		s.ScalerType = ""
		p := autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{s}}
		assert.Error(t, p.Validate())
	})

	t.Run("trigger auth without name", func(t *testing.T) {
		s := validScaler
		s.TriggerAuth = &autoscaling.TriggerAuth{}
		p := autoscaling.AutoscalingPolicy{Scalers: []autoscaling.Scaler{s}}
		assert.Error(t, p.Validate())
	})
}

func TestRoundTrip_ConfigJSON(t *testing.T) {
	policy := autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: ptrInt32(30),
		CooldownPeriodSeconds:  ptrInt32(300),
		Scalers: []autoscaling.Scaler{
			{
				ScalerType: "prometheus",
				Enabled:    true,
				Role:       autoscaling.RolePrimary,
				Config:     autoscaling.Config{ConfigJSON: `{"query":"up","threshold":"1"}`},
				TriggerAuth: &autoscaling.TriggerAuth{
					Name:       "auth",
					ConfigYAML: ptrStr("foo: bar"),
				},
			},
		},
	}

	req, err := autoscaling.ToQoveryRequest(policy)
	require.NoError(t, err)
	require.NotNil(t, req.KedaAutoscalingRequest)
	assert.Equal(t, "KEDA", string(req.KedaAutoscalingRequest.Mode))
	require.Len(t, req.KedaAutoscalingRequest.Scalers, 1)

	scaler := req.KedaAutoscalingRequest.Scalers[0]
	assert.Equal(t, "prometheus", scaler.ScalerType)
	assert.NotNil(t, scaler.ConfigJson)
	assert.Equal(t, "up", scaler.ConfigJson["query"])
	require.NotNil(t, scaler.TriggerAuthentication)
	assert.Equal(t, "auth", scaler.TriggerAuthentication.Name)
}

func TestToQoveryRequest_ConfigYAML(t *testing.T) {
	policy := autoscaling.AutoscalingPolicy{
		Scalers: []autoscaling.Scaler{
			{
				ScalerType: "cron",
				Enabled:    false,
				Role:       autoscaling.RoleSafety,
				Config:     autoscaling.Config{ConfigYAML: "start: 0 0 * * *"},
			},
		},
	}

	req, err := autoscaling.ToQoveryRequest(policy)
	require.NoError(t, err)
	scaler := req.KedaAutoscalingRequest.Scalers[0]
	require.NotNil(t, scaler.ConfigYaml)
	assert.Equal(t, "start: 0 0 * * *", *scaler.ConfigYaml)
	assert.Nil(t, scaler.ConfigJson)
	require.NotNil(t, scaler.Enabled)
	assert.False(t, *scaler.Enabled)
}

func TestToQoveryRequest_InvalidJSON(t *testing.T) {
	policy := autoscaling.AutoscalingPolicy{
		Scalers: []autoscaling.Scaler{
			{ScalerType: "x", Role: autoscaling.RolePrimary, Config: autoscaling.Config{ConfigJSON: "{not json"}},
		},
	}

	_, err := autoscaling.ToQoveryRequest(policy)
	assert.Error(t, err)
}

func TestFromQoveryResponse_Nil(t *testing.T) {
	got, err := autoscaling.FromQoveryResponse(nil)
	require.NoError(t, err)
	assert.Nil(t, got)
}
