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

func TestJSONTargetKeyValidator(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		value   types.String
		invalid bool
	}{
		{"null", types.StringNull(), false},
		{"unknown", types.StringUnknown(), false},
		{"nested", types.StringValue("accounts.0.password"), false},
		{"empty", types.StringValue(""), true},
		{"invalid_member", types.StringValue("a..b"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var resp validator.StringResponse
			JSONTargetKeyValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path: path.Root("json_target_key"), ConfigValue: tc.value,
			}, &resp)
			assert.Equal(t, tc.invalid, resp.Diagnostics.HasError())
		})
	}
}
