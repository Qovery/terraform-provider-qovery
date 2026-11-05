package advanced_settings

import (
	"context"
	"encoding/json"
)

const (
	ResourcesOverrideLimitRamInMibKey   = "resources.override.limit.ram_in_mib"
	ResourcesOverrideLimitCpuInMilliKey = "resources.override.limit.cpu_in_milli"

	// cpuLimitDisabled is the CPU limit that turns the limit off (q-core ServiceDomain.CPU_LIMIT_DISABLED).
	cpuLimitDisabled = float64(-1)
)

// RaiseResourceLimitsBeforeRequestUpdate stores the planned RAM and CPU limits of a service ahead of
// the service edit, for each limit the planned request goes above.
//
// q-core checks each endpoint against the value the other one stored: the service edit checks the
// new request against the stored limit, and the advanced settings update checks the new limit against
// the stored request. So raising a request and its limit together fails when the service edit comes
// first, and when RAM and CPU move in opposite directions no order of the two calls passes. Storing
// the raised limits first, then editing the service, then writing all the advanced settings passes
// every check. Remove this once the q-core service edit accepts the limits.
//
// It makes no call unless the plan sets one of the limits, and writes nothing unless a limit needs
// raising.
func (c ServiceAdvancedSettingsService) RaiseResourceLimitsBeforeRequestUpdate(ctx context.Context, serviceType int, serviceID string, advancedSettingsJsonFromPlan string, plannedCPUInMilli *int32, plannedRAMInMib *int32) error {
	if advancedSettingsJsonFromPlan == "" {
		return nil
	}

	planned := make(map[string]any)
	if err := json.Unmarshal([]byte(advancedSettingsJsonFromPlan), &planned); err != nil {
		return err
	}
	_, plansRAMLimit := planned[ResourcesOverrideLimitRamInMibKey]
	_, plansCPULimit := planned[ResourcesOverrideLimitCpuInMilliKey]
	if !plansRAMLimit && !plansCPULimit {
		return nil
	}

	urlAdvancedSettings, err := c.computeServiceAdvancedSettingsUrl(serviceType, serviceID)
	if err != nil {
		return err
	}

	current, err := c.getServiceAdvancedSettings(ctx, *urlAdvancedSettings)
	if err != nil {
		return err
	}

	raised := limitsToRaise(current, planned, plannedCPUInMilli, plannedRAMInMib)
	if len(raised) == 0 {
		return nil
	}

	// Only the raised limits change: the other planned settings are written after the service edit.
	settings := withoutBuildSettingsKeys(current)
	for k, v := range raised {
		settings[k] = v
	}
	return c.putServiceAdvancedSettings(ctx, *urlAdvancedSettings, settings)
}

// resourceLimit is one advanced setting that caps a resource request of the service.
type resourceLimit struct {
	key            string
	plannedRequest *int32
	// disabledValue is the limit value that turns the limit off, for the resources that have one.
	disabledValue *float64
}

func (l resourceLimit) isDisabled(value float64) bool {
	return l.disabledValue != nil && value == *l.disabledValue
}

// limitsToRaise returns the limits to store before the service edit, keyed by setting, with their
// planned values.
func limitsToRaise(current map[string]any, planned map[string]any, plannedCPUInMilli *int32, plannedRAMInMib *int32) map[string]any {
	cpuDisabled := cpuLimitDisabled
	limits := []resourceLimit{
		{key: ResourcesOverrideLimitRamInMibKey, plannedRequest: plannedRAMInMib},
		{key: ResourcesOverrideLimitCpuInMilliKey, plannedRequest: plannedCPUInMilli, disabledValue: &cpuDisabled},
	}

	raised := make(map[string]any)
	for _, limit := range limits {
		if value, ok := limit.valueToRaise(current, planned); ok {
			raised[limit.key] = value
		}
	}
	return raised
}

// valueToRaise returns the planned limit when the service edit would reject the planned request
// against the stored limit and the planned limit admits the planned request. Every other case is left
// to the usual calls, so an invalid plan gets the q-core error and nothing is written ahead of it.
func (l resourceLimit) valueToRaise(current map[string]any, planned map[string]any) (any, bool) {
	if l.plannedRequest == nil {
		return nil, false
	}
	plannedLimit, inPlan := planned[l.key]
	if !inPlan {
		return nil, false
	}

	// Without a stored limit the service edit does not check the request.
	storedLimit, isSet := current[l.key].(float64)
	if !isSet || l.isDisabled(storedLimit) {
		return nil, false
	}

	request := float64(*l.plannedRequest)
	if request <= storedLimit {
		return nil, false
	}

	// A null limit removes it.
	if plannedLimit == nil {
		return nil, true
	}
	value, isNumber := plannedLimit.(float64)
	if !isNumber || (value < request && !l.isDisabled(value)) {
		return nil, false
	}
	return value, true
}
