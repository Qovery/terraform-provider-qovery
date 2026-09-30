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

func TestInt64MinValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		ConfigValue   types.Int64
		ErrorContains string
	}{
		{TestName: "null", ConfigValue: types.Int64Null()},
		{TestName: "unknown", ConfigValue: types.Int64Unknown()},
		{TestName: "minimum", ConfigValue: types.Int64Value(60)},
		{TestName: "above_minimum", ConfigValue: types.Int64Value(61)},
		{TestName: "error_below_minimum", ConfigValue: types.Int64Value(59), ErrorContains: "Number value must be at least 60, got: 59."},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			req := validator.Int64Request{
				Path:        path.Root("timeout_seconds"),
				ConfigValue: tc.ConfigValue,
			}
			resp := &validator.Int64Response{}
			Int64MinValidator{Min: 60}.ValidateInt64(context.Background(), req, resp)

			if tc.ErrorContains == "" {
				assert.False(t, resp.Diagnostics.HasError(), "unexpected error: %v", resp.Diagnostics)
				return
			}
			assert.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.ErrorContains)
		})
	}
}
