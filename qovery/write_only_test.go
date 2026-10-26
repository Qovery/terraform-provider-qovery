//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestValidateWriteOnly(t *testing.T) {
	t.Parallel()

	null := types.StringNull()
	set := types.StringValue("value")
	unknown := types.StringUnknown()
	testCases := []struct {
		TestName       string
		Plain          types.String
		WriteOnly      types.String
		Version        types.String
		ExpectedErrors []string
	}{
		{TestName: "success_nothing_set", Plain: null, WriteOnly: null, Version: null},
		{TestName: "success_plain_only", Plain: set, WriteOnly: null, Version: null},
		{TestName: "success_write_only_with_version", Plain: null, WriteOnly: set, Version: set},
		{
			TestName:       "error_both_forms",
			Plain:          set,
			WriteOnly:      set,
			Version:        set,
			ExpectedErrors: []string{"config.password_wo"},
		},
		{
			TestName:       "error_write_only_without_version",
			Plain:          null,
			WriteOnly:      set,
			Version:        null,
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			// An empty string is a value: the API receives it.
			TestName:       "error_empty_write_only_without_version",
			Plain:          null,
			WriteOnly:      types.StringValue(""),
			Version:        null,
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			TestName:       "error_version_without_write_only",
			Plain:          null,
			WriteOnly:      null,
			Version:        set,
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			TestName:       "error_version_with_plain",
			Plain:          set,
			WriteOnly:      null,
			Version:        set,
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			// An ephemeral value is unknown until the ephemeral resource opens.
			TestName:  "success_unknown_write_only_without_version",
			Plain:     null,
			WriteOnly: unknown,
			Version:   null,
		},
		{TestName: "success_unknown_write_only_with_plain", Plain: set, WriteOnly: unknown, Version: set},
		{TestName: "success_unknown_plain_with_write_only", Plain: unknown, WriteOnly: set, Version: set},
		{TestName: "success_unknown_version", Plain: null, WriteOnly: set, Version: unknown},
		{TestName: "success_unknown_version_without_write_only", Plain: null, WriteOnly: null, Version: unknown},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			diags := validateWriteOnly(path.Root("config"), "password", tc.Plain, tc.WriteOnly, tc.Version)
			expected := tc.ExpectedErrors
			if expected == nil {
				expected = []string{}
			}
			assert.Equal(t, expected, errorPaths(diags), "diagnostics: %v", diags)
		})
	}
}

func TestValidateWriteOnly_NamesThePaths(t *testing.T) {
	t.Parallel()

	diags := validateWriteOnly(path.Root("secrets").AtListIndex(1), "value", types.StringValue("a"), types.StringValue("b"), types.StringNull())

	assert.Equal(t, []string{"secrets[1].value_wo", "secrets[1].value_wo_version"}, errorPaths(diags))
	assert.Equal(t, "secrets[1].value_wo cannot be set with secrets[1].value: set one of them.", diags.Errors()[0].Detail())
}

func TestWriteOnlyOrPlain(t *testing.T) {
	t.Parallel()

	assert.Equal(t, new("write-only"), writeOnlyOrPlain(types.StringNull(), types.StringValue("write-only")))
	assert.Equal(t, new("plain"), writeOnlyOrPlain(types.StringValue("plain"), types.StringNull()))
	assert.Equal(t, new(""), writeOnlyOrPlain(types.StringNull(), types.StringValue("")), "an empty write-only value is sent")
	assert.Nil(t, writeOnlyOrPlain(types.StringNull(), types.StringNull()))
	assert.Nil(t, writeOnlyOrPlain(types.StringNull(), types.StringUnknown()))
}

func TestWriteOnlyFromConfig(t *testing.T) {
	t.Parallel()

	r := containerRegistryResource{}
	configBlock := func(passwordWo tftypes.Value) func(tftypes.Object) tftypes.Value {
		return func(o tftypes.Object) tftypes.Value {
			values := map[string]tftypes.Value{}
			for name, attrType := range o.AttributeTypes {
				values[name] = tftypes.NewValue(attrType, nil)
			}
			values["password_wo"] = passwordWo
			return tftypes.NewValue(o, values)
		}
	}
	kind := tftypes.NewValue(tftypes.String, "GENERIC_CR")

	testCases := []struct {
		TestName string
		Config   func(tftypes.Object) tftypes.Value
		Expected types.String
	}{
		{
			TestName: "set",
			Config:   configBlock(tftypes.NewValue(tftypes.String, "secret")),
			Expected: types.StringValue("secret"),
		},
		{
			TestName: "null",
			Config:   configBlock(tftypes.NewValue(tftypes.String, nil)),
			Expected: types.StringNull(),
		},
		{
			TestName: "null_parent",
			Config:   func(o tftypes.Object) tftypes.Value { return tftypes.NewValue(o, nil) },
			Expected: types.StringNull(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			got := writeOnlyFromConfig(context.Background(), credentialsConfig(t, r, kind, tc.Config), registryPasswordWoPath, &diags)
			assert.False(t, diags.HasError(), "diagnostics: %v", diags)
			assert.Equal(t, tc.Expected, got)
		})
	}
}
