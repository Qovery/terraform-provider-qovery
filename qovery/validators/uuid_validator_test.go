//go:build unit && !integration
// +build unit,!integration

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestUUIDValidator_ValidateString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		ConfigValue types.String
		ExpectError bool
	}{
		{TestName: "success_uuid", ConfigValue: types.StringValue("6e2c3b8f-3d1a-4f0b-9a7e-2c5d8e1f4a6b")},
		// Built-in organization roles have placeholder ids that are not RFC 4122 UUIDs.
		{TestName: "success_builtin_role_placeholder_id", ConfigValue: types.StringValue("11111111-1111-1111-1111-111111111111")},
		{TestName: "success_null", ConfigValue: types.StringNull()},
		{TestName: "success_unknown", ConfigValue: types.StringUnknown()},
		{TestName: "error_name_instead_of_id", ConfigValue: types.StringValue("my-cluster"), ExpectError: true},
		{TestName: "error_empty", ConfigValue: types.StringValue(""), ExpectError: true},
		{TestName: "error_truncated_uuid", ConfigValue: types.StringValue("6e2c3b8f-3d1a-4f0b-9a7e-2c5d8e1f4a6"), ExpectError: true},
		{TestName: "error_surrounding_space", ConfigValue: types.StringValue(" 6e2c3b8f-3d1a-4f0b-9a7e-2c5d8e1f4a6b"), ExpectError: true},
		// uuid.Parse accepts these spellings, but the API returns the lowercase form with dashes,
		// so they fail at apply.
		{TestName: "error_uppercase_uuid", ConfigValue: types.StringValue("6E2C3B8F-3D1A-4F0B-9A7E-2C5D8E1F4A6B"), ExpectError: true},
		{TestName: "error_braced_uuid", ConfigValue: types.StringValue("{6e2c3b8f-3d1a-4f0b-9a7e-2c5d8e1f4a6b}"), ExpectError: true},
		{TestName: "error_urn_uuid", ConfigValue: types.StringValue("urn:uuid:6e2c3b8f-3d1a-4f0b-9a7e-2c5d8e1f4a6b"), ExpectError: true},
		{TestName: "error_uuid_without_dashes", ConfigValue: types.StringValue("6e2c3b8f3d1a4f0b9a7e2c5d8e1f4a6b"), ExpectError: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{Path: path.Root("cluster_id"), ConfigValue: tc.ConfigValue}
			resp := &validator.StringResponse{}
			NewUUIDValidator().ValidateString(context.Background(), req, resp)
			assert.Equal(t, tc.ExpectError, resp.Diagnostics.HasError(), "diagnostics: %v", resp.Diagnostics)
		})
	}
}

func TestUUIDValidator_ValidateSet(t *testing.T) {
	t.Parallel()

	const validUUID = "6e2c3b8f-3d1a-4f0b-9a7e-2c5d8e1f4a6b"
	testCases := []struct {
		TestName    string
		ConfigValue types.Set
		ErrorCount  int
	}{
		{TestName: "success_uuids", ConfigValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue(validUUID)})},
		{TestName: "success_empty", ConfigValue: types.SetValueMust(types.StringType, []attr.Value{})},
		{TestName: "success_null", ConfigValue: types.SetNull(types.StringType)},
		{TestName: "success_unknown", ConfigValue: types.SetUnknown(types.StringType)},
		{TestName: "success_unknown_element", ConfigValue: types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()})},
		{
			TestName: "error_each_invalid_element",
			ConfigValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(validUUID),
				types.StringValue("first"),
				types.StringValue("second"),
			}),
			ErrorCount: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := validator.SetRequest{Path: path.Root("labels_group_ids"), ConfigValue: tc.ConfigValue}
			resp := &validator.SetResponse{}
			NewUUIDValidator().ValidateSet(context.Background(), req, resp)
			assert.Equal(t, tc.ErrorCount, resp.Diagnostics.ErrorsCount(), "diagnostics: %v", resp.Diagnostics)
		})
	}
}
