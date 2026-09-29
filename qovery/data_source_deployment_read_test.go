//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploymentDataSourceReadEchoesTheConfiguration(t *testing.T) {
	t.Parallel()

	const deploymentID = "5f0d6a8e-1f7b-4c5e-9a39-3f4c7f1b2d10"
	const version = "0b1e8f5a-7c1d-4a0e-8f2b-6d9c3e4a5b61"

	testCases := []struct {
		TestName string
		Version  *string
	}{
		{TestName: "id_only"},
		{TestName: "id_and_version", Version: new(version)},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := deploymentDataSource{}

			schemaResp := &datasource.SchemaResponse{}
			ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
			require.False(t, schemaResp.Diagnostics.HasError())
			objType := schemaResp.Schema.Type().TerraformType(ctx)

			var versionValue tftypes.Value
			if tc.Version == nil {
				versionValue = tftypes.NewValue(tftypes.String, nil)
			} else {
				versionValue = tftypes.NewValue(tftypes.String, *tc.Version)
			}

			req := datasource.ReadRequest{Config: tfsdk.Config{
				Schema: schemaResp.Schema,
				Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
					"id":             tftypes.NewValue(tftypes.String, deploymentID),
					"environment_id": tftypes.NewValue(tftypes.String, nil),
					"version":        versionValue,
					"desired_state":  tftypes.NewValue(tftypes.String, nil),
				}),
			}}
			resp := &datasource.ReadResponse{State: tfsdk.State{
				Schema: schemaResp.Schema,
				Raw:    tftypes.NewValue(objType, nil),
			}}

			ds.Read(ctx, req, resp)

			require.False(t, resp.Diagnostics.HasError(), "the read must not fail: %v", resp.Diagnostics)
			var state NewDeploymentTerraform
			require.False(t, resp.State.Get(ctx, &state).HasError())
			assert.Equal(t, types.StringValue(deploymentID), state.Id)
			assert.Equal(t, types.StringPointerValue(tc.Version), state.Version)
			assert.True(t, state.EnvironmentId.IsNull())
			assert.True(t, state.DesiredState.IsNull())
		})
	}
}
