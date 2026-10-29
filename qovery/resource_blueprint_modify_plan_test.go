//go:build unit && !integration

package qovery

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

type stubBlueprintService struct {
	blueprint.Service
	latestTag   string
	iconURI     string
	err         error
	defaults    map[string]string
	defaultsErr error
}

func (s stubBlueprintService) ResolveCatalogEntry(context.Context, string, blueprint.CatalogVersion) (blueprint.CatalogEntry, error) {
	return blueprint.CatalogEntry{LatestTag: s.latestTag, IconURI: s.iconURI}, s.err
}

func (s stubBlueprintService) GetVariableDefaults(context.Context, string, blueprint.CatalogVersion) (map[string]string, error) {
	return s.defaults, s.defaultsErr
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
	config := blueprintObjectValue(t, objectType, map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, environmentID),
		"blueprint":      tftypes.NewValue(tftypes.String, "HELM/redis/8"),
	})
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
				Config: tfsdk.Config{Schema: schema, Raw: config},
				State:  tfsdk.State{Schema: schema, Raw: state},
				Plan:   tfsdk.Plan{Schema: schema, Raw: tc.plan},
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

func TestBlueprintModifyPlanPendingRetry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaResp := &resource.SchemaResponse{}
	newBlueprintResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	schema := schemaResp.Schema
	objectType, ok := schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	values := map[string]tftypes.Value{
		"environment_id": tftypes.NewValue(tftypes.String, "84c946f2-f0af-409f-b249-849b06cdb6d5"),
		"blueprint":      tftypes.NewValue(tftypes.String, "HELM/redis/8"),
		"tag":            tftypes.NewValue(tftypes.String, "HELM/redis/8/1.0.0"),
		"catalog_url":    tftypes.NewValue(tftypes.String, "https://github.com/Qovery/service-catalog.git"),
	}
	state := blueprintObjectValue(t, objectType, values)
	config := blueprintObjectValue(t, objectType, map[string]tftypes.Value{
		"environment_id": values["environment_id"],
		"blueprint":      values["blueprint"],
	})

	t.Run("retries against the latest tag and forces a non-empty plan", func(t *testing.T) {
		t.Parallel()
		r := blueprintResource{service: stubBlueprintService{latestTag: "HELM/redis/8/1.0.2"}}
		req := resource.ModifyPlanRequest{Config: tfsdk.Config{Schema: schema, Raw: config}, State: tfsdk.State{Schema: schema, Raw: state}, Plan: tfsdk.Plan{Schema: schema, Raw: state}}
		resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schema, Raw: state}}

		r.modifyPlan(ctx, req, resp, true)

		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
		var tag string
		var catalogURL types.String
		resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root("tag"), &tag)...)
		resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root("catalog_url"), &catalogURL)...)
		assert.Equal(t, "HELM/redis/8/1.0.2", tag)
		assert.True(t, catalogURL.IsUnknown())
	})

	t.Run("retry fails rather than redeploy the old tag when the catalog cannot answer", func(t *testing.T) {
		t.Parallel()
		r := blueprintResource{service: stubBlueprintService{err: errors.New("catalog down")}}
		req := resource.ModifyPlanRequest{Config: tfsdk.Config{Schema: schema, Raw: config}, State: tfsdk.State{Schema: schema, Raw: state}, Plan: tfsdk.Plan{Schema: schema, Raw: state}}
		resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schema, Raw: state}}

		r.modifyPlan(ctx, req, resp, true)

		assert.True(t, resp.Diagnostics.HasError())
	})
}

func TestBlueprintModifyPlanIconURI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaResp := &resource.SchemaResponse{}
	newBlueprintResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	schema := schemaResp.Schema
	objectType, ok := schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	environmentID := tftypes.NewValue(tftypes.String, "84c946f2-f0af-409f-b249-849b06cdb6d5")
	catalogVersion := tftypes.NewValue(tftypes.String, "AWS/postgres/17")
	config := func(iconURI tftypes.Value) tftypes.Value {
		return blueprintObjectValue(t, objectType, map[string]tftypes.Value{
			"environment_id": environmentID,
			"blueprint":      catalogVersion,
			"icon_uri":       iconURI,
		})
	}
	plan := func(tag tftypes.Value, iconURI tftypes.Value) tftypes.Value {
		return blueprintObjectValue(t, objectType, map[string]tftypes.Value{
			"environment_id": environmentID,
			"blueprint":      catalogVersion,
			"tag":            tag,
			"icon_uri":       iconURI,
		})
	}
	nullIcon := tftypes.NewValue(tftypes.String, nil)
	unknownIcon := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	redisIcon := tftypes.NewValue(tftypes.String, "app://qovery-console/redis")
	terraformIcon := tftypes.NewValue(tftypes.String, defaultBlueprintIconURI)
	unknownTag := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	deployedTag := tftypes.NewValue(tftypes.String, "AWS/postgres/17/4.6.1")
	noState := tftypes.NewValue(objectType, nil)
	catalog := stubBlueprintService{latestTag: "AWS/postgres/17/4.6.1", iconURI: "app://qovery-console/postgresql"}

	testCases := []struct {
		name     string
		service  stubBlueprintService
		config   tftypes.Value
		state    tftypes.Value
		plan     tftypes.Value
		wantIcon types.String
	}{
		{name: "create plans the catalog icon", service: catalog, config: config(nullIcon), state: noState, plan: plan(unknownTag, unknownIcon), wantIcon: FromString("app://qovery-console/postgresql")},
		{name: "create falls back to the terraform icon when the catalog has none", service: stubBlueprintService{latestTag: "AWS/postgres/17/4.6.1"}, config: config(nullIcon), state: noState, plan: plan(unknownTag, unknownIcon), wantIcon: FromString(defaultBlueprintIconURI)},
		{name: "create keeps the configured icon", service: catalog, config: config(redisIcon), state: noState, plan: plan(unknownTag, redisIcon), wantIcon: FromString("app://qovery-console/redis")},
		{name: "create keeps a configured icon unknown until apply", service: catalog, config: config(unknownIcon), state: noState, plan: plan(unknownTag, unknownIcon), wantIcon: types.StringUnknown()},
		// UseStateForUnknown has already planned the recorded icon
		{name: "update keeps the recorded icon", service: catalog, config: config(nullIcon), state: plan(deployedTag, terraformIcon), plan: plan(deployedTag, terraformIcon), wantIcon: FromString(defaultBlueprintIconURI)},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := blueprintResource{service: tc.service}
			req := resource.ModifyPlanRequest{
				Config: tfsdk.Config{Schema: schema, Raw: tc.config},
				State:  tfsdk.State{Schema: schema, Raw: tc.state},
				Plan:   tfsdk.Plan{Schema: schema, Raw: tc.plan},
			}
			resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schema, Raw: tc.plan}}

			r.ModifyPlan(ctx, req, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var iconURI types.String
			resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root("icon_uri"), &iconURI)...)
			assert.Equal(t, tc.wantIcon, iconURI)
		})
	}
}

func TestBlueprintResolveUnknownFromCatalog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := blueprintResource{service: stubBlueprintService{latestTag: "AWS/postgres/17/4.6.1", iconURI: "app://qovery-console/postgresql"}}
	// Planned while environment_id was unknown: neither value could be resolved yet
	plan := Blueprint{
		EnvironmentID: FromString("84c946f2-f0af-409f-b249-849b06cdb6d5"),
		Blueprint:     FromString("AWS/postgres/17"),
		Tag:           types.StringUnknown(),
		IconURI:       types.StringUnknown(),
	}

	diags := r.resolveUnknownFromCatalog(ctx, &plan)

	require.False(t, diags.HasError(), diags)
	assert.Equal(t, "AWS/postgres/17/4.6.1", plan.Tag.ValueString())
	assert.Equal(t, "app://qovery-console/postgresql", plan.IconURI.ValueString())
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
