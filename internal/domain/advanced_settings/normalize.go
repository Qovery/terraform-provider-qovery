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
// state, returning only settings that differ from defaults or are present in
// state. Values are normalized before comparison to handle type mismatches
// (e.g., API returns "true" as string instead of boolean true).
//
// When isTriggeredFromImport is true, all non-default settings are included
// regardless of whether they exist in state.
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
		if !reflect.DeepEqual(normalizedDefault, normalizedCurrent) {
			overridden[name] = normalizedCurrent
		} else if inState {
			overridden[name] = normalizeJSONValue(stateValue)
		}
	}
	return overridden
}
