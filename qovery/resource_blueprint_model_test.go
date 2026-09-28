//go:build unit && !integration

package qovery

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	t.Run("without defaults, tracks declared variables only and keeps unreadable values from prior", func(t *testing.T) {
		prior := Blueprint{
			IconURI:         FromString("app://qovery-console/postgresql"),
			Variables:       blueprintModelStringMap(t, map[string]string{"instance_type": "db.t3.micro", "removed": "x"}),
			SecretVariables: blueprintModelStringMap(t, map[string]string{"api_key": "test-value-1", "gone": "y"}),
			SpecOverrides:   overrides,
			Deploy:          types.BoolValue(false),
		}

		state, diags := convertDomainBlueprintToBlueprint(ctx, bp, prior, false, nil)
		require.False(t, diags.HasError())
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"instance_type": "db.t3.small"}), state.Variables)
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"api_key": "test-value-1"}), state.SecretVariables)
		assert.Equal(t, "app://qovery-console/postgresql", state.IconURI.ValueString(), "no icon from the API: prior stands")
		assert.Equal(t, overrides, state.SpecOverrides)
		assert.False(t, state.Deploy.ValueBool())
		assert.Equal(t, serviceID, state.ServiceID.ValueString())
		assert.Equal(t, "TERRAFORM", state.ServiceType.ValueString())
		assert.Equal(t, bp.EnvironmentID.String(), state.EnvironmentID.ValueString())
	})

	t.Run("omitted maps stay null", func(t *testing.T) {
		prior := Blueprint{Variables: types.MapNull(types.StringType), SecretVariables: types.MapNull(types.StringType)}
		state, diags := convertDomainBlueprintToBlueprint(ctx, bp, prior, false, nil)
		require.False(t, diags.HasError())
		assert.True(t, state.Variables.IsNull())
		assert.True(t, state.SecretVariables.IsNull())
		assert.Equal(t, defaultBlueprintIconURI, state.IconURI.ValueString())
		assert.True(t, state.Deploy.ValueBool())
	})

	defaults := map[string]string{"instance_type": "db.t3.micro", "storage_gb": "20"}

	t.Run("with defaults, also tracks undeclared variables that differ from their default", func(t *testing.T) {
		withConsoleValue := *bp
		withConsoleValue.Variables = append([]blueprint.Variable{{Name: "backup_window", Value: value("03:00")}}, bp.Variables...)
		prior := Blueprint{
			Variables:       blueprintModelStringMap(t, map[string]string{"storage_gb": "20"}),
			SecretVariables: types.MapNull(types.StringType),
		}

		state, diags := convertDomainBlueprintToBlueprint(ctx, &withConsoleValue, prior, false, defaults)
		require.False(t, diags.HasError())
		// storage_gb is declared though equal to its default; instance_type differs from its default;
		// backup_window has no default; import_identifier is platform-set
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"storage_gb": "20", "instance_type": "db.t3.small", "backup_window": "03:00"}), state.Variables)
	})

	t.Run("with defaults, variables all at their default stay null when omitted", func(t *testing.T) {
		atDefaults := *bp
		atDefaults.Variables = []blueprint.Variable{{Name: "instance_type", Value: value("db.t3.micro")}, {Name: "storage_gb", Value: value("20")}}
		prior := Blueprint{Variables: types.MapNull(types.StringType), SecretVariables: types.MapNull(types.StringType)}

		state, diags := convertDomainBlueprintToBlueprint(ctx, &atDefaults, prior, false, defaults)
		require.False(t, diags.HasError())
		assert.True(t, state.Variables.IsNull())
	})

	t.Run("the icon of the service wins over prior", func(t *testing.T) {
		withIcon := *bp
		withIcon.IconURI = value("app://qovery-console/redis")
		prior := Blueprint{IconURI: FromString(defaultBlueprintIconURI)}

		state, diags := convertDomainBlueprintToBlueprint(ctx, &withIcon, prior, false, nil)
		require.False(t, diags.HasError())
		assert.Equal(t, "app://qovery-console/redis", state.IconURI.ValueString())
	})

	t.Run("import records what the API returns", func(t *testing.T) {
		imported := *bp
		imported.IconURI = value("app://qovery-console/postgresql")
		prior := Blueprint{
			ID:              FromString(bp.ID.String()),
			Blueprint:       types.StringNull(),
			Tag:             types.StringNull(),
			IconURI:         types.StringNull(),
			Variables:       types.MapNull(types.StringType),
			SecretVariables: types.MapNull(types.StringType),
			Deploy:          types.BoolNull(),
		}

		state, diags := convertDomainBlueprintToBlueprint(ctx, &imported, prior, false, defaults)
		require.False(t, diags.HasError())
		assert.Equal(t, "aws/postgres/17", state.Blueprint.ValueString())
		assert.Equal(t, "app://qovery-console/postgresql", state.IconURI.ValueString())
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"instance_type": "db.t3.small"}), state.Variables)
		assert.True(t, state.SecretVariables.IsNull(), "secret values cannot be read back")
		assert.Nil(t, state.SpecOverrides)
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

func TestBlueprintIconURIRejectsChangeAfterCreate(t *testing.T) {
	t.Parallel()
	resp := &resource.SchemaResponse{}
	newBlueprintResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	iconURI, ok := resp.Schema.Attributes["icon_uri"].(schema.StringAttribute)
	require.True(t, ok)
	assert.Contains(t, iconURI.PlanModifiers, RejectChangeAfterCreate(blueprintIconChangeReason))
}

func TestBlueprintVariableDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	bp := &blueprint.Blueprint{EnvironmentID: uuid.New(), Tag: "HELM/redis/8/1.0.2"}

	t.Run("defaults of the deployed catalog version", func(t *testing.T) {
		var diags diag.Diagnostics
		r := blueprintResource{service: stubBlueprintService{defaults: map[string]string{"memory_limit": "512Mi"}}}
		assert.Equal(t, map[string]string{"memory_limit": "512Mi"}, r.variableDefaults(ctx, bp, &diags))
		assert.Empty(t, diags)
	})

	t.Run("a catalog error only warns and tracks declared variables", func(t *testing.T) {
		var diags diag.Diagnostics
		r := blueprintResource{service: stubBlueprintService{defaultsErr: errors.New("catalog unavailable")}}
		assert.Nil(t, r.variableDefaults(ctx, bp, &diags))
		require.Len(t, diags, 1)
		assert.Equal(t, diag.SeverityWarning, diags[0].Severity())
		assert.Contains(t, diags[0].Detail(), "catalog unavailable")
	})

	t.Run("an unparseable tag only warns", func(t *testing.T) {
		var diags diag.Diagnostics
		r := blueprintResource{service: stubBlueprintService{}}
		assert.Nil(t, r.variableDefaults(ctx, &blueprint.Blueprint{Tag: "redis"}, &diags))
		require.Len(t, diags, 1)
		assert.Equal(t, diag.SeverityWarning, diags[0].Severity())
	})
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
				{Name: "api_key", IsSecret: true},
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
		state, diags := convertDomainBlueprintToBlueprint(ctx, newBlueprint(blueprint.DispatchStatusRunning), prior, false, nil)
		require.False(t, diags.HasError())
		assert.Equal(t, "renamed", state.Name.ValueString())
		assert.Equal(t, "HELM/redis/8/1.0.3", state.Tag.ValueString())
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"memory_limit": "1Gi"}), state.Variables)
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"api_key": "test-value-1"}), state.SecretVariables)
		assert.Equal(t, "HELM/redis/8", state.Blueprint.ValueString())
	})

	t.Run("a deploy that failed outside Terraform still reports the saved settings", func(t *testing.T) {
		state, diags := convertDomainBlueprintToBlueprint(ctx, newBlueprint(blueprint.DispatchStatusFailed), prior, false, nil)
		require.False(t, diags.HasError())
		assert.Equal(t, "renamed", state.Name.ValueString())
		assert.Equal(t, "HELM/redis/8/1.0.3", state.Tag.ValueString())
		assert.Equal(t, blueprintModelStringMap(t, map[string]string{"memory_limit": "1Gi"}), state.Variables)
	})

	t.Run("pending apply keeps the last applied values even if the dispatch succeeded", func(t *testing.T) {
		upgraded := newBlueprint(blueprint.DispatchStatusRunning)
		upgraded.Tag = "HELM/redis/9/1.0.0"
		state, diags := convertDomainBlueprintToBlueprint(ctx, upgraded, prior, true, map[string]string{})
		require.False(t, diags.HasError())
		assert.Equal(t, "HELM/redis/8", state.Blueprint.ValueString())
		assert.Equal(t, "my-redis", state.Name.ValueString())
		assert.Equal(t, "HELM/redis/8/1.0.2", state.Tag.ValueString())
		assert.Equal(t, prior.Variables, state.Variables)
		assert.Equal(t, prior.SecretVariables, state.SecretVariables)
	})
}

func TestBlueprintVersionValue(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name  string
		prior types.String
		tag   string
		want  types.String
	}{
		{name: "derived from the tag without prior", prior: types.StringNull(), tag: "HELM/redis/8/1.0.2", want: FromString("HELM/redis/8")},
		{name: "prior spelling kept for the same version", prior: FromString("helm/Redis/8"), tag: "HELM/redis/8/1.0.2", want: FromString("helm/Redis/8")},
		{name: "a version changed outside Terraform wins over prior", prior: FromString("HELM/redis/8"), tag: "HELM/redis/9/1.0.0", want: FromString("HELM/redis/9")},
		{name: "unparseable prior is replaced", prior: FromString("redis-8"), tag: "HELM/redis/8/1.0.2", want: FromString("HELM/redis/8")},
		{name: "unparseable tag keeps prior", prior: FromString("HELM/redis/8"), tag: "redis", want: FromString("HELM/redis/8")},
		{name: "unparseable tag without prior", prior: types.StringUnknown(), tag: "redis", want: types.StringNull()},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, blueprintVersionValue(tc.prior, tc.tag))
		})
	}
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
		{name: "no prior state does not replace", state: types.StringNull(), plan: FromString("AWS/postgres/17"), want: false},
		{name: "unparseable value replaces", state: FromString("AWS/postgres/17"), plan: FromString("postgres-17"), want: true},
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
