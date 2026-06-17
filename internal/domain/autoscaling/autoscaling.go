package autoscaling

import (
	"github.com/pkg/errors"
)

// Role represents the role a KEDA scaler plays in an autoscaling policy.
type Role string

const (
	// RolePrimary is the role of the scaler that drives the main scaling decision.
	RolePrimary Role = "PRIMARY"
	// RoleSafety is the role of a scaler acting as a safety net.
	RoleSafety Role = "SAFETY"
)

var (
	// ErrInvalidAutoscalingPolicy is returned when an AutoscalingPolicy is invalid.
	ErrInvalidAutoscalingPolicy = errors.New("invalid autoscaling policy")
	// ErrInvalidScaler is returned when a Scaler is invalid.
	ErrInvalidScaler = errors.New("invalid scaler")
	// ErrInvalidScalerType is returned when a scaler type is empty.
	ErrInvalidScalerType = errors.New("invalid scaler type: must not be empty")
	// ErrInvalidRole is returned when a scaler role is not one of the allowed values.
	ErrInvalidRole = errors.New("invalid scaler role: must be PRIMARY or SAFETY")
	// ErrInvalidConfig is returned when a scaler config does not set exactly one of config_json/config_yaml.
	ErrInvalidConfig = errors.New("invalid scaler config: exactly one of config_json or config_yaml must be set")
	// ErrInvalidTriggerAuth is returned when a trigger authentication is invalid.
	ErrInvalidTriggerAuth = errors.New("invalid trigger authentication: name must not be empty")
	// ErrNoScalers is returned when an autoscaling policy has no scalers.
	ErrNoScalers = errors.New("invalid autoscaling policy: at least one scaler is required")
)

// IsValid returns whether the Role is one of the allowed values.
func (r Role) IsValid() bool {
	return r == RolePrimary || r == RoleSafety
}

// AutoscalingPolicy is the domain model for a service KEDA autoscaling policy.
// It is shared between the application (legacy client layer) and container
// (full DDD) wirings so the mapping logic lives in a single place.
type AutoscalingPolicy struct {
	PollingIntervalSeconds *int32
	CooldownPeriodSeconds  *int32
	Scalers                []Scaler
}

// Scaler is a single KEDA scaler within an AutoscalingPolicy.
type Scaler struct {
	ScalerType  string
	Enabled     bool
	Role        Role
	Config      Config
	TriggerAuth *TriggerAuth
}

// Config holds the scaler configuration. Exactly one of ConfigJSON / ConfigYAML
// must be set (mirrors the API's config_json ⊻ config_yaml constraint).
type Config struct {
	ConfigJSON string
	ConfigYAML string
}

// TriggerAuth is an inline KEDA TriggerAuthentication attached to a scaler.
type TriggerAuth struct {
	Name       string
	ConfigYAML *string
}

// Validate returns an error to tell whether the AutoscalingPolicy is valid or not.
func (p AutoscalingPolicy) Validate() error {
	if len(p.Scalers) == 0 {
		return ErrNoScalers
	}

	for _, s := range p.Scalers {
		if err := s.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidAutoscalingPolicy.Error())
		}
	}

	return nil
}

// IsValid returns a bool to tell whether the AutoscalingPolicy is valid or not.
func (p AutoscalingPolicy) IsValid() bool {
	return p.Validate() == nil
}

// Validate returns an error to tell whether the Scaler is valid or not.
func (s Scaler) Validate() error {
	if s.ScalerType == "" {
		return ErrInvalidScalerType
	}

	if !s.Role.IsValid() {
		return ErrInvalidRole
	}

	if err := s.Config.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidScaler.Error())
	}

	if s.TriggerAuth != nil {
		if err := s.TriggerAuth.Validate(); err != nil {
			return errors.Wrap(err, ErrInvalidScaler.Error())
		}
	}

	return nil
}

// Validate returns an error if the Config does not set exactly one of ConfigJSON / ConfigYAML.
func (c Config) Validate() error {
	hasJSON := c.ConfigJSON != ""
	hasYAML := c.ConfigYAML != ""
	if hasJSON == hasYAML {
		return ErrInvalidConfig
	}

	return nil
}

// Validate returns an error to tell whether the TriggerAuth is valid or not.
func (t TriggerAuth) Validate() error {
	if t.Name == "" {
		return ErrInvalidTriggerAuth
	}

	return nil
}
