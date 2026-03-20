//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNormalizeJSONValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    any
		expected any
	}{
		{
			testName: "string_true_to_bool",
			input:    "true",
			expected: true,
		},
		{
			testName: "string_false_to_bool",
			input:    "false",
			expected: false,
		},
		{
			testName: "bool_true_unchanged",
			input:    true,
			expected: true,
		},
		{
			testName: "bool_false_unchanged",
			input:    false,
			expected: false,
		},
		{
			testName: "numeric_string_to_float",
			input:    "42",
			expected: float64(42),
		},
		{
			testName: "float_string_to_float",
			input:    "3.14",
			expected: float64(3.14),
		},
		{
			testName: "regular_string_unchanged",
			input:    "hello",
			expected: "hello",
		},
		{
			testName: "empty_string_unchanged",
			input:    "",
			expected: "",
		},
		{
			testName: "float64_unchanged",
			input:    float64(42),
			expected: float64(42),
		},
		{
			testName: "nil_unchanged",
			input:    nil,
			expected: nil,
		},
		{
			testName: "slice_unchanged",
			input:    []any{"a", "b"},
			expected: []any{"a", "b"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()
			result := normalizeJSONValue(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("normalizeJSONValue(%v) = %v (%T), want %v (%T)",
					tc.input, result, result, tc.expected, tc.expected)
			}
		})
	}
}

func TestComputeOverriddenSettings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName              string
		current               map[string]any
		defaults              map[string]any
		state                 map[string]any
		isTriggeredFromImport bool
		expected              map[string]any
	}{
		{
			testName: "string_bool_from_api_matches_bool_default",
			current:  map[string]any{"static_ip": "true"},
			defaults: map[string]any{"static_ip": true},
			state:    map[string]any{"static_ip": true},
			expected: map[string]any{"static_ip": true},
		},
		{
			testName: "string_bool_from_api_differs_from_default",
			current:  map[string]any{"static_ip": "true"},
			defaults: map[string]any{"static_ip": false},
			state:    map[string]any{"static_ip": true},
			expected: map[string]any{"static_ip": true},
		},
		{
			testName: "bool_override_detected",
			current:  map[string]any{"static_ip": true},
			defaults: map[string]any{"static_ip": false},
			state:    map[string]any{"static_ip": true},
			expected: map[string]any{"static_ip": true},
		},
		{
			testName: "setting_not_in_state_skipped",
			current:  map[string]any{"static_ip": true, "other": "value"},
			defaults: map[string]any{"static_ip": false, "other": "default"},
			state:    map[string]any{"static_ip": true},
			expected: map[string]any{"static_ip": true},
		},
		{
			testName:              "import_includes_all_non_default",
			current:               map[string]any{"static_ip": true, "other": "value"},
			defaults:              map[string]any{"static_ip": false, "other": "default"},
			state:                 map[string]any{},
			isTriggeredFromImport: true,
			expected:              map[string]any{"static_ip": true, "other": "value"},
		},
		{
			testName:              "import_skips_default_values",
			current:               map[string]any{"static_ip": false, "name": "hello"},
			defaults:              map[string]any{"static_ip": false, "name": "hello"},
			state:                 map[string]any{},
			isTriggeredFromImport: true,
			expected:              map[string]any{},
		},
		{
			testName: "state_value_normalized_when_matching_default",
			current:  map[string]any{"static_ip": false},
			defaults: map[string]any{"static_ip": false},
			state:    map[string]any{"static_ip": "false"},
			expected: map[string]any{"static_ip": false},
		},
		{
			testName: "numeric_string_normalized",
			current:  map[string]any{"retention": "365"},
			defaults: map[string]any{"retention": float64(365)},
			state:    map[string]any{"retention": float64(365)},
			expected: map[string]any{"retention": float64(365)},
		},
		{
			testName: "multiple_overrides",
			current:  map[string]any{"a": true, "b": "custom", "c": float64(10)},
			defaults: map[string]any{"a": false, "b": "default", "c": float64(5)},
			state:    map[string]any{"a": true, "b": "custom", "c": float64(10)},
			expected: map[string]any{"a": true, "b": "custom", "c": float64(10)},
		},
		{
			testName: "empty_state_no_overrides",
			current:  map[string]any{"a": true},
			defaults: map[string]any{"a": false},
			state:    map[string]any{},
			expected: map[string]any{},
		},
		{
			testName: "key_in_current_not_in_defaults_treated_as_override",
			current:  map[string]any{"new_setting": "value"},
			defaults: map[string]any{},
			state:    map[string]any{"new_setting": "value"},
			expected: map[string]any{"new_setting": "value"},
		},
		{
			testName:              "import_key_in_current_not_in_defaults",
			current:               map[string]any{"new_setting": "value"},
			defaults:              map[string]any{},
			state:                 map[string]any{},
			isTriggeredFromImport: true,
			expected:              map[string]any{"new_setting": "value"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()
			result := computeOverriddenSettings(tc.current, tc.defaults, tc.state, tc.isTriggeredFromImport)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("computeOverriddenSettings() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestComputeOverriddenSettingsJSONRoundTrip verifies that the output of
// computeOverriddenSettings produces correct JSON when marshaled — the same
// path used by ReadClusterAdvancedSettings and ReadServiceAdvancedSettings.
func TestComputeOverriddenSettingsJSONRoundTrip(t *testing.T) {
	t.Parallel()

	// Simulate the customer's scenario: API returns bool true, default is bool false,
	// state has bool true. The output JSON should have boolean true, not string "true".
	current := map[string]any{"qovery.static_ip_mode": true, "other.setting": "value"}
	defaults := map[string]any{"qovery.static_ip_mode": false, "other.setting": "default"}
	state := map[string]any{"qovery.static_ip_mode": true, "other.setting": "value"}

	result := computeOverriddenSettings(current, defaults, state, false)

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back and verify types are preserved
	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	staticIP, ok := parsed["qovery.static_ip_mode"]
	if !ok {
		t.Fatal("qovery.static_ip_mode missing from result")
	}
	if _, isBool := staticIP.(bool); !isBool {
		t.Errorf("qovery.static_ip_mode should be bool, got %T (%v)", staticIP, staticIP)
	}
	if staticIP != true {
		t.Errorf("qovery.static_ip_mode = %v, want true", staticIP)
	}
}

// TestComputeOverriddenSettingsStringBoolRoundTrip verifies the fix for the
// customer issue: API returns "true" (string) instead of true (bool).
func TestComputeOverriddenSettingsStringBoolRoundTrip(t *testing.T) {
	t.Parallel()

	// API returns string "true", default is bool false, state has bool true
	current := map[string]any{"qovery.static_ip_mode": "true"}
	defaults := map[string]any{"qovery.static_ip_mode": false}
	state := map[string]any{"qovery.static_ip_mode": true}

	result := computeOverriddenSettings(current, defaults, state, false)

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// The JSON should contain true (boolean), not "true" (string)
	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	staticIP := parsed["qovery.static_ip_mode"]
	if _, isBool := staticIP.(bool); !isBool {
		t.Errorf("qovery.static_ip_mode should be bool after normalization, got %T (%v)", staticIP, staticIP)
	}
}
