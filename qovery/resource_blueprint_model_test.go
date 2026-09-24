//go:build unit && !integration

package qovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

func blueprintModelStringMap(t *testing.T, values map[string]string) types.Map {
	t.Helper()
	m, diags := types.MapValueFrom(context.Background(), types.StringType, values)
	require.False(t, diags.HasError())
	return m
}

func TestConvertDomainBlueprintToBlueprint(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	serviceID := uuid.NewString()
	value := func(s string) *string { return &s }
	bp := &blueprint.Blueprint{
		ID:            uuid.New(),
		EnvironmentID: uuid.New(),
		Name:          "my-db",
		Tag:           "aws/postgres/17/1.0.0",
		CatalogURL:    "https://catalog/aws/postgres",
		ServiceType:   blueprint.ServiceTypeTerraform,
		ServiceID:     &serviceID,
		Variables: []blueprint.Variable{
			{Name: "instance_type", Value: value("db.t3.small")},
			{Name: "storage_gb", Value: value("20")},
			{Name: blueprintImportIdentifierVariable, Value: value("my-db-id")},
			{Name: "api_key", IsSecret: true},
			{Name: "db_token", IsSecret: true},
		},
	}
	overrides := &BlueprintSpecOverrides{CPU: FromString("500m")}

	t.Run("tracks declared variables only and keeps unreadable values from prior", func(t *testing.T) {
		prior := Blueprint{
			IconURI:         FromString("https://cdn/icon.svg"),
			Variables:       blueprintModelStringMap(t, map[string]string{"instance_type": "db.t3.micro", "removed": "x"}),
			SecretVariables: blueprintModelStringMap(t, map[string]string{"api_key": "test-value-1", "gone": "y"}),
			SpecOverrides:   overrides,
			Deploy:          types.BoolValue(false),
		}

		state, diags := convertDomainBlueprintToBlueprint(ctx, bp, prior, false)
		require.False(t, diags.HasError())
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"instance_type": "db.t3.small"}), state.Variables)
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"api_key": "test-value-1"}), state.SecretVariables)
		assert.Equal(t, "https://cdn/icon.svg", state.IconURI.ValueString())
		assert.Equal(t, overrides, state.SpecOverrides)
		assert.False(t, state.Deploy.ValueBool())
		assert.Equal(t, serviceID, state.ServiceID.ValueString())
		assert.Equal(t, "TERRAFORM", state.ServiceType.ValueString())
		assert.Equal(t, bp.EnvironmentID.String(), state.EnvironmentID.ValueString())
	})

	t.Run("omitted maps stay null", func(t *testing.T) {
		prior := Blueprint{Variables: types.MapNull(types.StringType), SecretVariables: types.MapNull(types.StringType)}
		state, diags := convertDomainBlueprintToBlueprint(ctx, bp, prior, false)
		require.False(t, diags.HasError())
		assert.True(t, state.Variables.IsNull())
		assert.True(t, state.SecretVariables.IsNull())
		assert.Equal(t, defaultBlueprintIconURI, state.IconURI.ValueString())
		assert.True(t, state.Deploy.ValueBool())
	})
}

func TestBlueprintSchemasAreValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	resourceResp := &resource.SchemaResponse{}
	newBlueprintResource().Schema(ctx, resource.SchemaRequest{}, resourceResp)
	require.False(t, resourceResp.Diagnostics.HasError(), resourceResp.Diagnostics)
	require.False(t, resourceResp.Schema.ValidateImplementation(ctx).HasError())

	dataSourceResp := &datasource.SchemaResponse{}
	newBlueprintDataSource().Schema(ctx, datasource.SchemaRequest{}, dataSourceResp)
	require.False(t, dataSourceResp.Diagnostics.HasError(), dataSourceResp.Diagnostics)
	require.False(t, dataSourceResp.Schema.ValidateImplementation(ctx).HasError())
}

func TestBlueprintConfigEqualIgnoringDeploy(t *testing.T) {
	t.Parallel()
	state := Blueprint{
		EnvironmentID:   FromString(uuid.NewString()),
		Name:            FromString("my-db"),
		Tag:             FromString("HELM/redis/8/1.0.2"),
		IconURI:         FromString(defaultBlueprintIconURI),
		Variables:       blueprintModelStringMap(t, map[string]string{"memory_limit": "256Mi"}),
		SecretVariables: types.MapNull(types.StringType),
		SpecOverrides:   &BlueprintSpecOverrides{CPU: FromString("500m")},
		Deploy:          types.BoolValue(true),
		CatalogURL:      FromString("https://catalog/helm/redis"),
	}

	onlyDeploy := state
	onlyDeploy.Deploy = types.BoolValue(false)
	onlyDeploy.CatalogURL = types.StringUnknown()
	assert.True(t, onlyDeploy.configEqualIgnoringDeploy(state))

	otherVariable := onlyDeploy
	otherVariable.Variables = blueprintModelStringMap(t, map[string]string{"memory_limit": "512Mi"})
	assert.False(t, otherVariable.configEqualIgnoringDeploy(state))

	otherCase := onlyDeploy
	otherCase.Blueprint = FromString("helm/redis/8")
	state.Blueprint = FromString("HELM/redis/8")
	assert.True(t, otherCase.configEqualIgnoringDeploy(state))

	otherOverride := onlyDeploy
	otherOverride.SpecOverrides = &BlueprintSpecOverrides{CPU: FromString("1")}
	assert.False(t, otherOverride.configEqualIgnoringDeploy(state))
}

func TestBlueprintToUpdateRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := Blueprint{
		Name:            FromString("my-db"),
		Tag:             FromString("aws/postgres/17/1.0.1"),
		IconURI:         FromString(defaultBlueprintIconURI),
		Variables:       blueprintModelStringMap(t, map[string]string{"instance_type": "db.t3.small"}),
		SecretVariables: types.MapNull(types.StringType),
	}
	state := Blueprint{
		Variables:       blueprintModelStringMap(t, map[string]string{"instance_type": "db.t3.micro", "storage_gb": "20"}),
		SecretVariables: blueprintModelStringMap(t, map[string]string{"api_key": "test-value-1"}),
		SpecOverrides:   &BlueprintSpecOverrides{Timeout: types.Int64Value(600)},
	}

	request, diags := plan.toUpdateRequest(ctx, state)
	require.False(t, diags.HasError())
	assert.Equal(t, map[string]string{"instance_type": "db.t3.small"}, request.Variables)
	assert.Empty(t, request.SecretVariables)
	assert.Nil(t, request.SpecOverrides)
	assert.Equal(t, []string{"instance_type", "storage_gb", "api_key"}, request.PreviousVariableNames)
	require.NotNil(t, request.PreviousSpecOverrides)
	assert.Equal(t, int32(600), *request.PreviousSpecOverrides.Timeout)
}

func TestConvertDomainBlueprintToBlueprintVersionAndFailures(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value := func(s string) *string { return &s }
	newBlueprint := func(dispatch blueprint.DispatchStatus) *blueprint.Blueprint {
		return &blueprint.Blueprint{
			ID:               uuid.New(),
			EnvironmentID:    uuid.New(),
			Name:             "renamed",
			Tag:              "HELM/redis/8/1.0.3",
			LatestDeployment: &blueprint.Dispatch{ID: "dispatch-2", Status: dispatch},
			Variables: []blueprint.Variable{
				{Name: "memory_limit", Value: value("1Gi")},
				{Name: "region", Value: value("eu-west-3")},
				{Name: "replicas", Value: value("3")},
			},
		}
	}
	prior := Blueprint{
		Blueprint:       FromString("HELM/redis/8"),
		Name:            FromString("my-redis"),
		Tag:             FromString("HELM/redis/8/1.0.2"),
		Variables:       blueprintModelStringMap(t, map[string]string{"memory_limit": "256Mi"}),
		SecretVariables: blueprintModelStringMap(t, map[string]string{"api_key": "test-value-1"}),
	}

	t.Run("successful apply refreshes from the API", func(t *testing.T) {
		state, diags := convertDomainBlueprintToBlueprint(ctx, newBlueprint(blueprint.DispatchStatusRunning), prior, false)
		require.False(t, diags.HasError())
		assert.Equal(t, "renamed", state.Name.ValueString())
		assert.Equal(t, "HELM/redis/8/1.0.3", state.Tag.ValueString())
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"memory_limit": "1Gi"}), state.Variables)
		assert.Equal(t, "HELM/redis/8", state.Blueprint.ValueString())
	})

	t.Run("failed apply keeps the last applied values", func(t *testing.T) {
		state, diags := convertDomainBlueprintToBlueprint(ctx, newBlueprint(blueprint.DispatchStatusFailed), prior, false)
		require.False(t, diags.HasError())
		assert.Equal(t, prior.SecretVariables, state.SecretVariables)
		assert.Equal(t, "my-redis", state.Name.ValueString())
		assert.Equal(t, "HELM/redis/8/1.0.2", state.Tag.ValueString())
		assert.Equal(t, prior.Variables, state.Variables)
	})

	t.Run("pending apply keeps the last applied values even if the dispatch succeeded", func(t *testing.T) {
		state, diags := convertDomainBlueprintToBlueprint(ctx, newBlueprint(blueprint.DispatchStatusRunning), prior, true)
		require.False(t, diags.HasError())
		assert.Equal(t, "my-redis", state.Name.ValueString())
		assert.Equal(t, "HELM/redis/8/1.0.2", state.Tag.ValueString())
		assert.Equal(t, prior.Variables, state.Variables)
	})
}

func TestRequiresReplaceIfOtherBlueprintService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCases := []struct {
		name  string
		state types.String
		plan  types.String
		want  bool
	}{
		{name: "major upgrade stays in place", state: FromString("AWS/postgres/16"), plan: FromString("AWS/postgres/17"), want: false},
		{name: "other family replaces", state: FromString("AWS/postgres/17"), plan: FromString("AWS/mysql/8"), want: true},
		{name: "other provider replaces", state: FromString("AWS/redis/7"), plan: FromString("GCP/redis/7"), want: true},
		{name: "unknown plan does not replace", state: FromString("AWS/postgres/17"), plan: types.StringUnknown(), want: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.StringRequest{
				StateValue:  tc.state,
				PlanValue:   tc.plan,
				ConfigValue: tc.plan,
				State:       tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})},
				Plan:        tfsdk.Plan{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})},
			}
			resp := &planmodifier.StringResponse{PlanValue: tc.plan}
			requiresReplaceIfOtherBlueprintService().PlanModifyString(ctx, req, resp)
			assert.Equal(t, tc.want, resp.RequiresReplace)
		})
	}
}

func TestBlueprintVersionAttributesSameVersion(t *testing.T) {
	t.Parallel()
	environmentID := uuid.NewString()
	current := blueprintVersionAttributes{
		EnvironmentID: FromString(environmentID),
		Blueprint:     FromString("AWS/postgres/17"),
		Tag:           FromString("AWS/postgres/17/4.1.0"),
	}

	assert.True(t, current.sameVersion(blueprintVersionAttributes{EnvironmentID: FromString(environmentID), Blueprint: FromString("aws/postgres/17")}))
	assert.False(t, current.sameVersion(blueprintVersionAttributes{EnvironmentID: FromString(uuid.NewString()), Blueprint: FromString("AWS/postgres/17")}))
	assert.False(t, current.sameVersion(blueprintVersionAttributes{EnvironmentID: FromString(environmentID), Blueprint: FromString("AWS/postgres/18")}))

	noTag := current
	noTag.Tag = types.StringNull()
	assert.False(t, noTag.sameVersion(blueprintVersionAttributes{EnvironmentID: FromString(environmentID), Blueprint: FromString("AWS/postgres/17")}))
}
