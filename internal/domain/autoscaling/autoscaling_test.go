//go:build unit && !integration

package autoscaling_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/qovery/qovery-client-go"
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

func TestToQoveryRequest_ScalerID(t *testing.T) {
	policy := autoscaling.AutoscalingPolicy{
		Scalers: []autoscaling.Scaler{
			{ID: "scaler-1", ScalerType: "cron", Role: autoscaling.RolePrimary, Config: autoscaling.Config{ConfigYAML: "a: 1"}},
			{ScalerType: "cpu", Role: autoscaling.RoleSafety, Config: autoscaling.Config{ConfigYAML: "b: 2"}},
		},
	}

	req, err := autoscaling.ToQoveryRequest(policy)
	require.NoError(t, err)
	scalers := req.KedaAutoscalingRequest.Scalers
	require.Len(t, scalers, 2)
	require.NotNil(t, scalers[0].Id, "an existing scaler is sent with its id")
	assert.Equal(t, "scaler-1", *scalers[0].Id)
	assert.Nil(t, scalers[1].Id, "a new scaler is sent without id")
}

func TestFromQoveryResponse_ScalerID(t *testing.T) {
	yaml := "a: 1"
	scaler := qovery.NewKedaScalerResponse("scaler-1", time.Now(), "cron", true, qovery.KEDASCALERROLE_PRIMARY)
	scaler.ConfigYaml = *qovery.NewNullableString(&yaml)
	keda := qovery.NewKedaAutoscalingResponse("policy-1", time.Now(), "service-1", qovery.AUTOSCALINGMODE_KEDA, 30, 300, []qovery.KedaScalerResponse{*scaler})
	res := qovery.KedaAutoscalingResponseAsAutoscalingPolicyResponse(keda)

	got, err := autoscaling.FromQoveryResponse(&res)
	require.NoError(t, err)
	require.Len(t, got.Scalers, 1)
	assert.Equal(t, "scaler-1", got.Scalers[0].ID)
	assert.Equal(t, "a: 1", got.Scalers[0].Config.ConfigYAML)
}

func TestAutoscalingPolicy_Validate_ScaledObjectSettings(t *testing.T) {
	t.Parallel()

	scalers := []autoscaling.Scaler{{
		ScalerType: "cpu",
		Enabled:    true,
		Role:       autoscaling.RolePrimary,
		Config:     autoscaling.Config{ConfigYAML: "value: \"80\""},
	}}

	testCases := []struct {
		TestName    string
		Policy      autoscaling.AutoscalingPolicy
		ExpectedErr error
	}{
		{TestName: "success_unset", Policy: autoscaling.AutoscalingPolicy{}},
		{TestName: "success_initial_cooldown_min", Policy: autoscaling.AutoscalingPolicy{InitialCooldownPeriodSeconds: ptrInt32(0)}},
		{TestName: "success_initial_cooldown_max", Policy: autoscaling.AutoscalingPolicy{InitialCooldownPeriodSeconds: ptrInt32(3600)}},
		{TestName: "success_idle_replica_count", Policy: autoscaling.AutoscalingPolicy{IdleReplicaCount: ptrInt32(0)}},
		{TestName: "success_advanced_yaml", Policy: autoscaling.AutoscalingPolicy{AdvancedYAML: ptrStr("restoreToOriginalReplicaCount: true")}},
		{TestName: "error_negative_initial_cooldown", Policy: autoscaling.AutoscalingPolicy{InitialCooldownPeriodSeconds: ptrInt32(-1)}, ExpectedErr: autoscaling.ErrInvalidInitialCooldownPeriod},
		{TestName: "error_initial_cooldown_too_long", Policy: autoscaling.AutoscalingPolicy{InitialCooldownPeriodSeconds: ptrInt32(3601)}, ExpectedErr: autoscaling.ErrInvalidInitialCooldownPeriod},
		{TestName: "error_idle_replica_count_not_zero", Policy: autoscaling.AutoscalingPolicy{IdleReplicaCount: ptrInt32(1)}, ExpectedErr: autoscaling.ErrInvalidIdleReplicaCount},
		{TestName: "error_blank_advanced_yaml", Policy: autoscaling.AutoscalingPolicy{AdvancedYAML: ptrStr(" \n")}, ExpectedErr: autoscaling.ErrBlankAdvancedYAML},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			policy := tc.Policy
			policy.Scalers = scalers
			err := policy.Validate()
			if tc.ExpectedErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tc.ExpectedErr)
		})
	}
}

func TestToQoveryRequest_ScaledObjectSettings(t *testing.T) {
	t.Parallel()

	scalers := []autoscaling.Scaler{{ScalerType: "cpu", Role: autoscaling.RolePrimary, Config: autoscaling.Config{ConfigYAML: "a: 1"}}}

	t.Run("set", func(t *testing.T) {
		t.Parallel()

		req, err := autoscaling.ToQoveryRequest(autoscaling.AutoscalingPolicy{
			InitialCooldownPeriodSeconds: ptrInt32(120),
			IdleReplicaCount:             ptrInt32(0),
			AdvancedYAML:                 ptrStr("restoreToOriginalReplicaCount: true"),
			Scalers:                      scalers,
		})
		require.NoError(t, err)

		body, err := json.Marshal(req.KedaAutoscalingRequest)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"initial_cooldown_period_seconds":120`)
		assert.Contains(t, string(body), `"idle_replica_count":0`)
		assert.Contains(t, string(body), `"advanced_yaml":"restoreToOriginalReplicaCount: true"`)
	})

	// Unset settings are sent as an explicit null so that removing them from the
	// configuration clears them.
	t.Run("unset", func(t *testing.T) {
		t.Parallel()

		req, err := autoscaling.ToQoveryRequest(autoscaling.AutoscalingPolicy{Scalers: scalers})
		require.NoError(t, err)

		body, err := json.Marshal(req.KedaAutoscalingRequest)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"initial_cooldown_period_seconds":null`)
		assert.Contains(t, string(body), `"idle_replica_count":null`)
		assert.Contains(t, string(body), `"advanced_yaml":null`)
	})
}

func TestFromQoveryResponse_ScaledObjectSettings(t *testing.T) {
	t.Parallel()

	keda := qovery.NewKedaAutoscalingResponse("policy-1", time.Now(), "service-1", qovery.AUTOSCALINGMODE_KEDA, 30, 300, []qovery.KedaScalerResponse{})
	keda.InitialCooldownPeriodSeconds = *qovery.NewNullableInt32(ptrInt32(120))
	keda.IdleReplicaCount = *qovery.NewNullableInt32(ptrInt32(0))
	keda.AdvancedYaml = *qovery.NewNullableString(ptrStr("restoreToOriginalReplicaCount: true"))
	res := qovery.KedaAutoscalingResponseAsAutoscalingPolicyResponse(keda)

	got, err := autoscaling.FromQoveryResponse(&res)
	require.NoError(t, err)
	assert.Equal(t, ptrInt32(120), got.InitialCooldownPeriodSeconds)
	assert.Equal(t, ptrInt32(0), got.IdleReplicaCount)
	assert.Equal(t, ptrStr("restoreToOriginalReplicaCount: true"), got.AdvancedYAML)

	unset := qovery.KedaAutoscalingResponseAsAutoscalingPolicyResponse(
		qovery.NewKedaAutoscalingResponse("policy-1", time.Now(), "service-1", qovery.AUTOSCALINGMODE_KEDA, 30, 300, []qovery.KedaScalerResponse{}),
	)
	got, err = autoscaling.FromQoveryResponse(&unset)
	require.NoError(t, err)
	assert.Nil(t, got.InitialCooldownPeriodSeconds)
	assert.Nil(t, got.IdleReplicaCount)
	assert.Nil(t, got.AdvancedYAML)
}
