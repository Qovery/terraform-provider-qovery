//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The error for a malformed import identifier must name the parts in the order ImportState parses them.
func TestCredentialsImportState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Resource resource.ResourceWithImportState
		Format   string
	}{
		{TestName: "aws", Resource: awsCredentialsResource{}, Format: "organization_id,aws_credentials_id"},
		{TestName: "eks_anywhere_vsphere", Resource: eksAnywhereVsphereCredentialsResource{}, Format: "organization_id,eks_anywhere_vsphere_credentials_id"},
		{TestName: "gcp", Resource: gcpCredentialsResource{}, Format: "organization_id,gcp_credentials_id"},
		{TestName: "scaleway", Resource: scalewayCredentialsResource{}, Format: "organization_id,scaleway_credentials_id"},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := importCredentials(t, tc.Resource, "credentials_id")
			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "format: "+tc.Format+".")

			resp = importCredentials(t, tc.Resource, "org_id,credentials_id")
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var organizationID, id types.String
			require.False(t, resp.State.GetAttribute(context.Background(), path.Root("organization_id"), &organizationID).HasError())
			require.False(t, resp.State.GetAttribute(context.Background(), path.Root("id"), &id).HasError())
			assert.Equal(t, "org_id", organizationID.ValueString())
			assert.Equal(t, "credentials_id", id.ValueString())
		})
	}
}

func importCredentials(t *testing.T, r resource.ResourceWithImportState, id string) *resource.ImportStateResponse {
	t.Helper()
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "%v", schemaResp.Diagnostics)

	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}}
	r.ImportState(ctx, resource.ImportStateRequest{ID: id}, resp)
	return resp
}
