//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"reflect"
	"testing"
)

func TestComputeUnknownKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName         string
		validKeys        map[string]struct{}
		advancedSettings map[string]any
		expected         []string
	}{
		{
			testName:         "all_keys_valid_returns_empty",
			validKeys:        map[string]struct{}{"network.ingress.cors_enabled": {}},
			advancedSettings: map[string]any{"network.ingress.cors_enabled": true},
			expected:         []string{},
		},
		{
			testName:         "unknown_keys_returned_sorted",
			validKeys:        map[string]struct{}{"network.ingress.cors_enabled": {}},
			advancedSettings: map[string]any{"security.service_account_name": "x", "network.dns.ndots": float64(1)},
			expected:         []string{"network.dns.ndots", "security.service_account_name"},
		},
		{
			testName:         "mixed_valid_and_unknown",
			validKeys:        map[string]struct{}{"network.ingress.cors_enabled": {}},
			advancedSettings: map[string]any{"network.ingress.cors_enabled": true, "network.dns.ndots": float64(1)},
			expected:         []string{"network.dns.ndots"},
		},
		{
			testName:         "empty_settings_returns_empty",
			validKeys:        map[string]struct{}{"a": {}},
			advancedSettings: map[string]any{},
			expected:         []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()
			result := computeUnknownKeys(tc.validKeys, tc.advancedSettings)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("computeUnknownKeys() = %v, want %v", result, tc.expected)
			}
		})
	}
}
