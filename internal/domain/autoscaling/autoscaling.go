package autoscaling

import (
	"strings"

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
	// ErrInvalidInitialCooldownPeriod is returned when the initial cooldown period is out of range.
	ErrInvalidInitialCooldownPeriod = errors.New("invalid autoscaling policy: initial_cooldown_period_seconds must be between 0 and 3600")
	// ErrInvalidIdleReplicaCount is returned when the idle replica count is not the only value KEDA supports.
	ErrInvalidIdleReplicaCount = errors.New("invalid autoscaling policy: idle_replica_count must be 0, the only value supported by KEDA")
	// ErrBlankAdvancedYAML is returned when the advanced YAML is set but blank.
	ErrBlankAdvancedYAML = errors.New("invalid autoscaling policy: advanced_yaml must not be blank")
)

const (
	// MinInitialCooldownPeriodSeconds and MaxInitialCooldownPeriodSeconds bound
	// the initial cooldown period, as the API does.
	MinInitialCooldownPeriodSeconds = 0
	MaxInitialCooldownPeriodSeconds = 3600
	// SupportedIdleReplicaCount is the only idle replica count KEDA supports:
	// the HPA cannot hold a Deployment at a non-zero idle count.
	SupportedIdleReplicaCount = 0
)

// IsValid returns whether the Role is one of the allowed values.
func (r Role) IsValid() bool {
	return r == RolePrimary || r == RoleSafety
}

// AutoscalingPolicy is the domain model for a service KEDA autoscaling policy.
// It is shared between the application (legacy client layer) and container
// (full DDD) wirings so the mapping logic lives in a single place.
//
// InitialCooldownPeriodSeconds, IdleReplicaCount and AdvancedYAML map to the
// KEDA ScaledObject initialCooldownPeriod, idleReplicaCount and advanced
// fields; nil leaves them unset.
type AutoscalingPolicy struct {
	PollingIntervalSeconds       *int32
	CooldownPeriodSeconds        *int32
	InitialCooldownPeriodSeconds *int32
	IdleReplicaCount             *int32
	AdvancedYAML                 *string
	Scalers                      []Scaler
}

// Scaler is a single KEDA scaler within an AutoscalingPolicy. ID is the API id
// of an existing scaler, empty for a scaler to create.
type Scaler struct {
	ID          string
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

	if c := p.InitialCooldownPeriodSeconds; c != nil && (*c < MinInitialCooldownPeriodSeconds || *c > MaxInitialCooldownPeriodSeconds) {
		return ErrInvalidInitialCooldownPeriod
	}

	if p.IdleReplicaCount != nil && *p.IdleReplicaCount != SupportedIdleReplicaCount {
		return ErrInvalidIdleReplicaCount
	}

	if p.AdvancedYAML != nil && strings.TrimSpace(*p.AdvancedYAML) == "" {
		return ErrBlankAdvancedYAML
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
