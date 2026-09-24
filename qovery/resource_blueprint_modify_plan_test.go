//go:build unit && !integration

package qovery

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

type stubBlueprintService struct {
	blueprint.Service
	latestTag string
	err       error
}

func (s stubBlueprintService) ResolveLatestTag(context.Context, string, blueprint.CatalogVersion) (string, error) {
	return s.latestTag, s.err
}

func blueprintObjectValue(t *testing.T, objectType tftypes.Object, values map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}
	for name, value := range values {
		attributes[name] = value
	}
	return tftypes.NewValue(objectType, attributes)
}

func TestBlueprintModifyPlanTag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaResp := &resource.SchemaResponse{}
	newBlueprintResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	schema := schemaResp.Schema
	objectType, ok := schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	environmentID := "84c946f2-f0af-409f-b249-849b06cdb6d5"
	state := blueprintObjectValue(t, objectType, map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, environmentID),
		"blueprint":      tftypes.NewValue(tftypes.String, "HELM/redis/8"),
		"tag":            tftypes.NewValue(tftypes.String, "HELM/redis/8/1.0.0"),
	})
	plan := func(tag tftypes.Value, extra map[string]tftypes.Value) tftypes.Value {
		values := map[string]tftypes.Value{
			"environment_id": tftypes.NewValue(tftypes.String, environmentID),
			"blueprint":      tftypes.NewValue(tftypes.String, "HELM/redis/8"),
			"tag":            tag,
		}
		for name, value := range extra {
			values[name] = value
		}
		return blueprintObjectValue(t, objectType, values)
	}
	unknownTag := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	noOpPlan := plan(tftypes.NewValue(tftypes.String, "HELM/redis/8/1.0.0"), nil)
	deployOnlyPlan := plan(unknownTag, map[string]tftypes.Value{"deploy": tftypes.NewValue(tftypes.Bool, false)})
	renamePlan := plan(unknownTag, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "renamed")})

	testCases := []struct {
		name      string
		service   stubBlueprintService
		plan      tftypes.Value
		wantTag   string
		wantError bool
	}{
		{name: "plans the latest tag", service: stubBlueprintService{latestTag: "HELM/redis/8/1.0.2"}, plan: noOpPlan, wantTag: "HELM/redis/8/1.0.2"},
		{name: "no-op plan keeps the deployed tag when the catalog cannot answer", service: stubBlueprintService{err: errors.New("catalog down")}, plan: noOpPlan, wantTag: "HELM/redis/8/1.0.0"},
		{name: "deploy-only change keeps the deployed tag when the catalog cannot answer", service: stubBlueprintService{err: errors.New("catalog down")}, plan: deployOnlyPlan, wantTag: "HELM/redis/8/1.0.0"},
		{name: "update fails rather than deploy a tag that may not be the latest", service: stubBlueprintService{err: errors.New("catalog down")}, plan: renamePlan, wantError: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := blueprintResource{service: tc.service}
			req := resource.ModifyPlanRequest{
				State: tfsdk.State{Schema: schema, Raw: state},
				Plan:  tfsdk.Plan{Schema: schema, Raw: tc.plan},
			}
			resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schema, Raw: tc.plan}}

			r.ModifyPlan(ctx, req, resp)

			if tc.wantError {
				assert.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var tag string
			resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root("tag"), &tag)...)
			assert.Equal(t, tc.wantTag, tag)
		})
	}
}

func TestCanKeepDeployedTag(t *testing.T) {
	t.Parallel()
	current := &blueprintVersionAttributes{
		EnvironmentID: FromString("84c946f2-f0af-409f-b249-849b06cdb6d5"),
		Blueprint:     FromString("HELM/redis/8"),
		Tag:           FromString("HELM/redis/8/1.0.0"),
	}
	planned := blueprintVersionAttributes{EnvironmentID: current.EnvironmentID, Blueprint: current.Blueprint}

	assert.True(t, canKeepDeployedTag(current, planned, false))
	assert.False(t, canKeepDeployedTag(current, planned, true), "a pending retry deploys, so it needs the latest tag")
	assert.False(t, canKeepDeployedTag(nil, planned, false), "a create has no deployed tag")
}
