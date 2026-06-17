package autoscaling

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
)

// ToQoveryRequest converts the domain AutoscalingPolicy into the API request
// model. The API exposes autoscaling as a oneOf wrapper discriminated by
// `mode`; KEDA is currently the only variant, so the wrapper is always built
// from a KedaAutoscalingRequest with Mode=KEDA.
func ToQoveryRequest(p AutoscalingPolicy) (qovery.AutoscalingPolicyRequest, error) {
	scalers := make([]qovery.KedaScalerRequest, 0, len(p.Scalers))
	for _, s := range p.Scalers {
		scaler := qovery.NewKedaScalerRequest(s.ScalerType, qovery.KedaScalerRole(s.Role))
		scaler.Enabled = &s.Enabled

		switch {
		case s.Config.ConfigJSON != "":
			var configJSON map[string]any
			if err := json.Unmarshal([]byte(s.Config.ConfigJSON), &configJSON); err != nil {
				return qovery.AutoscalingPolicyRequest{}, errors.Wrap(err, "failed to parse scaler config_json")
			}
			scaler.ConfigJson = configJSON
		case s.Config.ConfigYAML != "":
			scaler.ConfigYaml = &s.Config.ConfigYAML
		}

		if s.TriggerAuth != nil {
			triggerAuth := qovery.NewKedaTriggerAuthenticationRequest(s.TriggerAuth.Name)
			triggerAuth.ConfigYaml = s.TriggerAuth.ConfigYAML
			scaler.TriggerAuthentication = triggerAuth
		}

		scalers = append(scalers, *scaler)
	}

	keda := qovery.NewKedaAutoscalingRequest(qovery.AUTOSCALINGMODE_KEDA, scalers)
	keda.PollingIntervalSeconds = p.PollingIntervalSeconds
	keda.CooldownPeriodSeconds = p.CooldownPeriodSeconds

	return qovery.KedaAutoscalingRequestAsAutoscalingPolicyRequest(keda), nil
}

// FromQoveryResponse converts the API response model into the domain
// AutoscalingPolicy. Response-only fields (id, created_at, updated_at,
// service_id) are dropped. Returns (nil, nil) when autoscaling is absent.
func FromQoveryResponse(res *qovery.AutoscalingPolicyResponse) (*AutoscalingPolicy, error) {
	if res == nil || res.KedaAutoscalingResponse == nil {
		return nil, nil
	}

	keda := res.KedaAutoscalingResponse

	scalers := make([]Scaler, 0, len(keda.Scalers))
	for _, s := range keda.Scalers {
		config := Config{}
		if len(s.ConfigJson) > 0 {
			configJSON, err := json.Marshal(s.ConfigJson)
			if err != nil {
				return nil, errors.Wrap(err, "failed to serialize scaler config_json")
			}
			config.ConfigJSON = string(configJSON)
		} else if yaml := s.ConfigYaml.Get(); yaml != nil {
			config.ConfigYAML = *yaml
		}

		var triggerAuth *TriggerAuth
		if s.TriggerAuthentication != nil {
			triggerAuth = &TriggerAuth{
				Name:       s.TriggerAuthentication.Name,
				ConfigYAML: s.TriggerAuthentication.ConfigYaml,
			}
		}

		scalers = append(scalers, Scaler{
			ScalerType:  s.ScalerType,
			Enabled:     s.Enabled,
			Role:        Role(s.Role),
			Config:      config,
			TriggerAuth: triggerAuth,
		})
	}

	pollingInterval := keda.PollingIntervalSeconds
	cooldownPeriod := keda.CooldownPeriodSeconds

	return &AutoscalingPolicy{
		PollingIntervalSeconds: &pollingInterval,
		CooldownPeriodSeconds:  &cooldownPeriod,
		Scalers:                scalers,
	}, nil
}
