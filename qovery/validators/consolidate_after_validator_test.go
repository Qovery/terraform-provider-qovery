//go:build unit && !integration
// +build unit,!integration

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestConsolidateAfterValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		ConfigValue   types.String
		ErrorContains string
	}{
		{TestName: "null", ConfigValue: types.StringNull()},
		{TestName: "unknown", ConfigValue: types.StringUnknown()},
		{TestName: "seconds", ConfigValue: types.StringValue("30s")},
		{TestName: "minutes", ConfigValue: types.StringValue("10m")},
		{TestName: "minutes_that_are_not_whole_hours", ConfigValue: types.StringValue("90m")},
		{TestName: "hours", ConfigValue: types.StringValue("1h")},
		{TestName: "maximum", ConfigValue: types.StringValue("24h")},
		// q-core returns zero seconds as "0h".
		{TestName: "zero", ConfigValue: types.StringValue("0h")},
		{TestName: "error_empty", ConfigValue: types.StringValue(""), ErrorContains: "must be <number><unit>"},
		{TestName: "error_no_unit", ConfigValue: types.StringValue("30"), ErrorContains: "must be <number><unit>"},
		{TestName: "error_unsupported_unit", ConfigValue: types.StringValue("1d"), ErrorContains: "must be <number><unit>"},
		{TestName: "error_go_duration", ConfigValue: types.StringValue("1h30m"), ErrorContains: "must be <number><unit>"},
		{TestName: "error_negative", ConfigValue: types.StringValue("-1m"), ErrorContains: "must be <number><unit>"},
		{TestName: "error_above_24h", ConfigValue: types.StringValue("25h"), ErrorContains: "must not exceed 24h"},
		{TestName: "error_above_24h_in_seconds", ConfigValue: types.StringValue("86401s"), ErrorContains: "must not exceed 24h"},
		{TestName: "error_number_overflows", ConfigValue: types.StringValue("99999999999999999999s"), ErrorContains: "must not exceed 24h"},
		{TestName: "error_minutes_that_are_whole_hours", ConfigValue: types.StringValue("60m"), ErrorContains: `Write "1h" instead`},
		{TestName: "error_seconds_that_are_whole_minutes", ConfigValue: types.StringValue("120s"), ErrorContains: `Write "2m" instead`},
		{TestName: "error_seconds_that_are_whole_hours", ConfigValue: types.StringValue("86400s"), ErrorContains: `Write "24h" instead`},
		{TestName: "error_leading_zero", ConfigValue: types.StringValue("05m"), ErrorContains: `Write "5m" instead`},
		{TestName: "error_zero_seconds", ConfigValue: types.StringValue("0s"), ErrorContains: `Write "0h" instead`},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:        path.Root("consolidate_after"),
				ConfigValue: tc.ConfigValue,
			}
			resp := &validator.StringResponse{}
			NewConsolidateAfterValidator().ValidateString(context.Background(), req, resp)

			if tc.ErrorContains == "" {
				assert.False(t, resp.Diagnostics.HasError(), "unexpected error: %v", resp.Diagnostics)
				return
			}
			assert.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.ErrorContains)
		})
	}
}
