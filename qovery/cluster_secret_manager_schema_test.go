//go:build unit && !integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func secretManagerAuthSchemaAttribute(t *testing.T, name string) schema.StringAttribute {
	t.Helper()

	resp := &resource.SchemaResponse{}
	clusterResource{}.Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())

	accesses, ok := resp.Schema.Attributes["secret_manager_accesses"].(schema.SetNestedAttribute)
	require.True(t, ok)
	auth, ok := accesses.NestedObject.Attributes["authentication"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	attribute, ok := auth.Attributes[name].(schema.StringAttribute)
	require.True(t, ok)
	return attribute
}

func validateStringAttribute(attribute schema.StringAttribute, value types.String) bool {
	resp := &validator.StringResponse{}
	for _, v := range attribute.Validators {
		v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("attribute"), ConfigValue: value}, resp)
	}
	return !resp.Diagnostics.HasError()
}

// An empty assume role value would be planned as "" but read back as null once the API ignores it, which Terraform
// reports as an inconsistent result: the schema has to reject it at plan time.
func TestClusterSchema_SecretManagerAssumeRoleRejectsEmptyStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		attribute string
		valid     string
	}{
		{attribute: "assume_role_arn", valid: "arn:aws:iam::222222222222:role/reader"},
		{attribute: "assume_role_external_id", valid: "prod-eu-1"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.attribute, func(t *testing.T) {
			t.Parallel()
			attribute := secretManagerAuthSchemaAttribute(t, tc.attribute)

			assert.False(t, validateStringAttribute(attribute, types.StringValue("")), "an empty string must be rejected")
			assert.True(t, validateStringAttribute(attribute, types.StringValue(tc.valid)))
			assert.True(t, validateStringAttribute(attribute, types.StringNull()), "unset stays allowed")
		})
	}
}
