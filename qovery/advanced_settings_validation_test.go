//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/pkg/errors"
)

// testAdvancedSettingsConfig builds a tfsdk.Config holding a single advanced_settings_json
// string attribute with the given raw value.
func testAdvancedSettingsConfig(value tftypes.Value) tfsdk.Config {
	return tfsdk.Config{
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{advancedSettingsJSONAttr: tftypes.String},
			},
			map[string]tftypes.Value{advancedSettingsJSONAttr: value},
		),
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				advancedSettingsJSONAttr: schema.StringAttribute{Optional: true},
			},
		},
	}
}

func TestWarnUnknownAdvancedSettingsKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName         string
		value            tftypes.Value
		lookup           func(advancedSettingsJson string) ([]string, error)
		expectedWarnings []string
	}{
		{
			testName: "unknown_keys_produce_one_warning_each",
			value:    tftypes.NewValue(tftypes.String, `{"bogus.key": 1, "other.typo": true}`),
			lookup: func(string) ([]string, error) {
				return []string{"bogus.key", "other.typo"}, nil
			},
			expectedWarnings: []string{"bogus.key", "other.typo"},
		},
		{
			testName: "no_unknown_keys_no_warning",
			value:    tftypes.NewValue(tftypes.String, `{"known.key": 1}`),
			lookup: func(string) ([]string, error) {
				return nil, nil
			},
			expectedWarnings: nil,
		},
		{
			testName: "null_attribute_no_warning_lookup_not_called",
			value:    tftypes.NewValue(tftypes.String, nil),
			lookup: func(string) ([]string, error) {
				panic("lookup must not be called for a null attribute")
			},
			expectedWarnings: nil,
		},
		{
			testName: "unknown_value_no_warning_lookup_not_called",
			value:    tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			lookup: func(string) ([]string, error) {
				panic("lookup must not be called for an unknown attribute")
			},
			expectedWarnings: nil,
		},
		{
			testName: "lookup_error_degrades_to_no_warning",
			value:    tftypes.NewValue(tftypes.String, `{"bogus.key": 1}`),
			lookup: func(string) ([]string, error) {
				return nil, errors.New("api unreachable")
			},
			expectedWarnings: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			warnUnknownAdvancedSettingsKeys(context.Background(), tc.lookup, testAdvancedSettingsConfig(tc.value), &diags)

			if diags.HasError() {
				t.Fatalf("expected no errors, got %v", diags.Errors())
			}

			warnings := diags.Warnings()
			if len(warnings) != len(tc.expectedWarnings) {
				t.Fatalf("got %d warnings, want %d: %v", len(warnings), len(tc.expectedWarnings), warnings)
			}
			for i, key := range tc.expectedWarnings {
				if !strings.Contains(warnings[i].Summary(), key) {
					t.Errorf("warning %d summary %q does not mention key %q", i, warnings[i].Summary(), key)
				}
			}
		})
	}
}

func TestWarnUnknownClusterAdvancedSettings_NilServiceIsNoop(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	cfg := testAdvancedSettingsConfig(tftypes.NewValue(tftypes.String, `{"bogus.key": 1}`))

	warnUnknownClusterAdvancedSettings(context.Background(), nil, cfg, &diags)

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics with nil service, got %v", diags)
	}
}
