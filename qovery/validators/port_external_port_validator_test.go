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

func makePortObject(publiclyAccessible bool, externalPort *int64) types.Object {
	attrTypes := map[string]attr.Type{
		"publicly_accessible": types.BoolType,
		"external_port":       types.Int64Type,
	}
	attrs := map[string]attr.Value{
		"publicly_accessible": types.BoolValue(publiclyAccessible),
	}
	if externalPort != nil {
		attrs["external_port"] = types.Int64Value(*externalPort)
	} else {
		attrs["external_port"] = types.Int64Null()
	}
	obj, _ := types.ObjectValue(attrTypes, attrs)
	return obj
}

func int64Ptr(i int64) *int64 { return &i }

func TestPortExternalPortValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		ConfigValue types.Object
		ExpectError bool
	}{
		{
			TestName:    "error_when_not_public_and_external_port_set",
			ConfigValue: makePortObject(false, int64Ptr(15432)),
			ExpectError: true,
		},
		{
			TestName:    "no_error_when_public_and_external_port_set",
			ConfigValue: makePortObject(true, int64Ptr(443)),
			ExpectError: false,
		},
		{
			TestName:    "no_error_when_not_public_and_external_port_not_set",
			ConfigValue: makePortObject(false, nil),
			ExpectError: false,
		},
		{
			TestName:    "no_error_when_public_and_external_port_not_set",
			ConfigValue: makePortObject(true, nil),
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := validator.ObjectRequest{
				Path:        path.Root("ports").AtListIndex(0),
				ConfigValue: tc.ConfigValue,
			}
			resp := &validator.ObjectResponse{
				Diagnostics: diag.Diagnostics{},
			}
			v := PortExternalPortValidator{}
			v.ValidateObject(context.Background(), req, resp)
			if tc.ExpectError {
				assert.True(t, resp.Diagnostics.HasError(), "expected validation error")
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "expected no validation error, got: %s", resp.Diagnostics.Errors())
			}
		})
	}
}
