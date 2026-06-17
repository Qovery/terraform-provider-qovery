//go:build unit || !integration

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

var scalerObjectType = map[string]attr.Type{
	"config_json": types.StringType,
	"config_yaml": types.StringType,
	"enabled":     types.BoolType,
}

// makeScalerObject builds a scaler object carrying only the attributes the
// validator inspects. Pass nil to leave an attribute null (a null `enabled`
// means the scaler is enabled, mirroring the schema default of true).
func makeScalerObject(configJSON, configYAML *string) types.Object {
	return makeScalerObjectEnabled(configJSON, configYAML, nil)
}

func makeScalerObjectEnabled(configJSON, configYAML *string, enabled *bool) types.Object {
	attrs := map[string]attr.Value{
		"config_json": types.StringNull(),
		"config_yaml": types.StringNull(),
		"enabled":     types.BoolNull(),
	}
	if configJSON != nil {
		attrs["config_json"] = types.StringValue(*configJSON)
	}
	if configYAML != nil {
		attrs["config_yaml"] = types.StringValue(*configYAML)
	}
	if enabled != nil {
		attrs["enabled"] = types.BoolValue(*enabled)
	}
	obj, _ := types.ObjectValue(scalerObjectType, attrs)
	return obj
}

func TestScalerConfigExactlyOneValidator(t *testing.T) {
	t.Parallel()

	elemType := types.ObjectType{AttrTypes: scalerObjectType}

	testCases := []struct {
		TestName    string
		ConfigValue types.Set
		ExpectError bool
	}{
		{
			TestName:    "null_set_skips",
			ConfigValue: types.SetNull(elemType),
			ExpectError: false,
		},
		{
			TestName:    "empty_set_errors",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{}),
			ExpectError: true,
		},
		{
			TestName: "json_only_ok",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObject(new(`{"a":1}`), nil),
			}),
			ExpectError: false,
		},
		{
			TestName: "yaml_only_ok",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObject(nil, new("a: 1")),
			}),
			ExpectError: false,
		},
		{
			TestName: "both_set_errors",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObject(new(`{"a":1}`), new("a: 1")),
			}),
			ExpectError: true,
		},
		{
			TestName: "neither_set_errors",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObject(nil, nil),
			}),
			ExpectError: true,
		},
		{
			TestName: "one_valid_one_invalid_errors",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObject(new(`{"a":1}`), nil),
				makeScalerObject(nil, nil),
			}),
			ExpectError: true,
		},
		{
			TestName: "single_scaler_disabled_errors",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObjectEnabled(new(`{"a":1}`), nil, new(false)),
			}),
			ExpectError: true,
		},
		{
			TestName: "all_scalers_disabled_errors",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObjectEnabled(new(`{"a":1}`), nil, new(false)),
				makeScalerObjectEnabled(nil, new("a: 1"), new(false)),
			}),
			ExpectError: true,
		},
		{
			TestName: "one_enabled_among_disabled_ok",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObjectEnabled(new(`{"a":1}`), nil, new(false)),
				makeScalerObjectEnabled(nil, new("a: 1"), new(true)),
			}),
			ExpectError: false,
		},
		{
			TestName: "single_scaler_enabled_ok",
			ConfigValue: types.SetValueMust(elemType, []attr.Value{
				makeScalerObjectEnabled(new(`{"a":1}`), nil, new(true)),
			}),
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := validator.SetRequest{
				Path:        path.Root("autoscaling").AtName("scalers"),
				ConfigValue: tc.ConfigValue,
			}
			resp := &validator.SetResponse{Diagnostics: diag.Diagnostics{}}
			ScalerConfigExactlyOneValidator{}.ValidateSet(context.Background(), req, resp)
			if tc.ExpectError {
				assert.True(t, resp.Diagnostics.HasError(), "expected validation error")
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "expected no error, got: %s", resp.Diagnostics.Errors())
			}
		})
	}
}
