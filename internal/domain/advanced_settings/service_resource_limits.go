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
// It makes no call unless the plan sets one of the limits and a request, and writes nothing unless a
// limit needs raising and every planned limit admits its planned request.
func (c ServiceAdvancedSettingsService) RaiseResourceLimitsBeforeRequestUpdate(ctx context.Context, serviceType int, serviceID string, advancedSettingsJsonFromPlan string, plannedCPUInMilli *int32, plannedRAMInMib *int32) error {
	if advancedSettingsJsonFromPlan == "" || (plannedCPUInMilli == nil && plannedRAMInMib == nil) {
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
// planned values: those whose stored value the planned request goes above.
//
// When a planned limit does not admit its planned request, the apply fails at q-core whatever this
// does, so it returns none: the q-core error comes with nothing written ahead of it.
func limitsToRaise(current map[string]any, planned map[string]any, plannedCPUInMilli *int32, plannedRAMInMib *int32) map[string]any {
	cpuDisabled := cpuLimitDisabled
	limits := []resourceLimit{
		{key: ResourcesOverrideLimitRamInMibKey, plannedRequest: plannedRAMInMib},
		{key: ResourcesOverrideLimitCpuInMilliKey, plannedRequest: plannedCPUInMilli, disabledValue: &cpuDisabled},
	}

	raised := make(map[string]any)
	for _, limit := range limits {
		plannedLimit, inPlan := planned[limit.key]
		if !inPlan || limit.plannedRequest == nil {
			continue
		}
		request := float64(*limit.plannedRequest)
		if !limit.admits(plannedLimit, request) {
			return map[string]any{}
		}
		if !limit.admits(current[limit.key], request) {
			raised[limit.key] = plannedLimit
		}
	}
	return raised
}

// admits reports whether value, as this limit, lets the service request request. A missing or null
// limit, and a disabled one, admit any request: q-core only checks a limit that is set.
func (l resourceLimit) admits(value any, request float64) bool {
	if value == nil {
		return true
	}
	limit, isNumber := value.(float64)
	if !isNumber {
		return false
	}
	return l.isDisabled(limit) || limit >= request
}
