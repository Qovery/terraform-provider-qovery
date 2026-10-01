package advanced_settings

import (
	"reflect"
	"strconv"
)

// normalizeJSONValue converts string-encoded booleans and numbers to their
// native Go types, matching what json.Unmarshal produces for JSON primitives.
// This ensures type-consistent comparisons between API responses and state
// values, preventing spurious diffs when the API returns "true" (string)
// instead of true (boolean).
//
// Note: Numeric strings like "42" are converted to float64 to match
// json.Unmarshal behavior. This is safe because the advanced settings API
// schema defines values as typed (boolean, number, string) — a value
// that is semantically a string will not look like a bare number.
func normalizeJSONValue(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}

	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	return v
}

// computeOverriddenSettings compares current API settings against defaults and
// state and returns the settings Terraform tracks:
//
//   - a key already present in state and reported by the API is kept, with the
//     remote value: a change made outside Terraform to a tracked key — including
//     a reset back to the API default — is reflected on refresh and re-planned
//     (QOV-2028);
//   - a key absent from state is skipped, so settings overridden only outside
//     Terraform are not pulled into state (advanced_settings_json is a partial
//     override map and the API has no ownership flag to tell user-set keys from
//     Qovery-provisioned ones);
//   - when isTriggeredFromImport is true, every non-default key is included
//     regardless of state.
//
// Values are normalized before comparison to handle type mismatches (e.g., API
// returns "true" as string instead of boolean true).
func computeOverriddenSettings(
	current map[string]any,
	defaults map[string]any,
	state map[string]any,
	isTriggeredFromImport bool,
) map[string]any {
	overridden := make(map[string]any)
	for name, value := range current {
		defaultValue := defaults[name]
		stateValue, inState := state[name]
		if !isTriggeredFromImport && !inState {
			continue
		}
		normalizedCurrent := normalizeJSONValue(value)
		normalizedDefault := normalizeJSONValue(defaultValue)
		normalizedState := normalizeJSONValue(stateValue)
		switch {
		case inState && reflect.DeepEqual(normalizedState, normalizedCurrent):
			// State already reflects the remote value (after normalization): emit
			// the state's *raw* scalar form rather than the normalized one. The
			// config drives advanced_settings_json as a JSON string, so a customer
			// who wrote "true"/"60" (string) in jsonencode keeps those values in
			// state; coercing them to true/60 (native) on refresh would produce a
			// permanent diff (QOV-2027). Normalization is still used for the
			// comparison so real drift is detected — it just no longer rewrites
			// the stored encoding.
			overridden[name] = stateValue
		case inState || !reflect.DeepEqual(normalizedDefault, normalizedCurrent):
			// Real drift on a tracked key, or a non-default key on import: surface
			// the remote value. A tracked key whose remote value went back to the
			// default is kept with that default rather than dropped, so it stays
			// tracked and later out-of-band changes remain visible (QOV-2028).
			overridden[name] = normalizedCurrent
		}
	}

	// Preserve "unknown" overrides: keys present in state but absent from both the API
	// response (current) and the defaults. The API does not recognize these keys for this
	// service type, so it never returns them. Dropping them here would cause a perpetual
	// diff: state loses the key on refresh, then the next plan re-adds it from config.
	// Carrying the state value forward keeps the refreshed value equal to the configured
	// value. On import state is empty, so this loop is a no-op and import behavior is
	// unchanged.
	for name, stateValue := range state {
		if _, inCurrent := current[name]; inCurrent {
			continue
		}
		if _, inDefaults := defaults[name]; inDefaults {
			continue
		}
		// Preserve the state's raw scalar form for the same reason as above
		// (QOV-2027): the API never returns these keys, so the state value is the
		// only source of truth and must round-trip unchanged against the config.
		overridden[name] = stateValue
	}

	return overridden
}
