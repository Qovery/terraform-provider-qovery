//go:build unit || !integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// --- SmartAllowApiOverride tests ---

func TestSmartAllowApiOverride_ConfigExplicitlySetFalse(t *testing.T) {
	t.Parallel()
	modifier := SmartAllowApiOverride()

	resp := &planmodifier.BoolResponse{PlanValue: types.BoolValue(false)}
	modifier.PlanModifyBool(context.Background(), planmodifier.BoolRequest{
		ConfigValue: types.BoolValue(false),
		StateValue:  types.BoolNull(),
		PlanValue:   types.BoolValue(false),
	}, resp)

	assert.Equal(t, types.BoolValue(false), resp.PlanValue,
		"should use config value when explicitly set to false")
}

func TestSmartAllowApiOverride_ConfigExplicitlySetTrue(t *testing.T) {
	t.Parallel()
	modifier := SmartAllowApiOverride()

	resp := &planmodifier.BoolResponse{PlanValue: types.BoolValue(true)}
	modifier.PlanModifyBool(context.Background(), planmodifier.BoolRequest{
		ConfigValue: types.BoolValue(true),
		StateValue:  types.BoolNull(),
		PlanValue:   types.BoolValue(true),
	}, resp)

	assert.Equal(t, types.BoolValue(true), resp.PlanValue,
		"should use config value when explicitly set to true")
}

func TestSmartAllowApiOverride_ConfigOverridesState(t *testing.T) {
	t.Parallel()
	modifier := SmartAllowApiOverride()

	resp := &planmodifier.BoolResponse{PlanValue: types.BoolValue(false)}
	modifier.PlanModifyBool(context.Background(), planmodifier.BoolRequest{
		ConfigValue: types.BoolValue(false),
		StateValue:  types.BoolValue(true),
		PlanValue:   types.BoolValue(false),
	}, resp)

	assert.Equal(t, types.BoolValue(false), resp.PlanValue,
		"config value should take precedence over state value")
}

func TestSmartAllowApiOverride_ConfigNullUsesState(t *testing.T) {
	t.Parallel()
	modifier := SmartAllowApiOverride()

	resp := &planmodifier.BoolResponse{PlanValue: types.BoolNull()}
	modifier.PlanModifyBool(context.Background(), planmodifier.BoolRequest{
		ConfigValue: types.BoolNull(),
		StateValue:  types.BoolValue(true),
		PlanValue:   types.BoolNull(),
	}, resp)

	assert.Equal(t, types.BoolValue(true), resp.PlanValue,
		"should use state value when config is null")
}

func TestSmartAllowApiOverride_ConfigNullStateValueFalse(t *testing.T) {
	t.Parallel()
	modifier := SmartAllowApiOverride()

	resp := &planmodifier.BoolResponse{PlanValue: types.BoolNull()}
	modifier.PlanModifyBool(context.Background(), planmodifier.BoolRequest{
		ConfigValue: types.BoolNull(),
		StateValue:  types.BoolValue(false),
		PlanValue:   types.BoolNull(),
	}, resp)

	assert.Equal(t, types.BoolValue(false), resp.PlanValue,
		"should use state value false when config is null")
}

func TestSmartAllowApiOverride_BothNullReturnsUnknown(t *testing.T) {
	t.Parallel()
	modifier := SmartAllowApiOverride()

	resp := &planmodifier.BoolResponse{PlanValue: types.BoolNull()}
	modifier.PlanModifyBool(context.Background(), planmodifier.BoolRequest{
		ConfigValue: types.BoolNull(),
		StateValue:  types.BoolNull(),
		PlanValue:   types.BoolNull(),
	}, resp)

	assert.True(t, resp.PlanValue.IsUnknown(),
		"should return unknown when both config and state are null (API decides)")
}

func TestSmartAllowApiOverride_Description(t *testing.T) {
	t.Parallel()
	modifier := SmartAllowApiOverride()

	assert.NotEmpty(t, modifier.Description(context.Background()))
	assert.NotEmpty(t, modifier.MarkdownDescription(context.Background()))
}

// --- UseUnknownForNullString tests ---

func TestUseUnknownForNullString_BothNull(t *testing.T) {
	t.Parallel()
	modifier := UseUnknownForNullString()

	resp := &planmodifier.StringResponse{PlanValue: types.StringNull()}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		ConfigValue: types.StringNull(),
		StateValue:  types.StringNull(),
		PlanValue:   types.StringNull(),
	}, resp)

	assert.True(t, resp.PlanValue.IsUnknown(),
		"should return unknown when both config and state are null (new list element)")
}

func TestUseUnknownForNullString_ConfigSet(t *testing.T) {
	t.Parallel()
	modifier := UseUnknownForNullString()

	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("my-name")}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		ConfigValue: types.StringValue("my-name"),
		StateValue:  types.StringNull(),
		PlanValue:   types.StringValue("my-name"),
	}, resp)

	assert.Equal(t, types.StringValue("my-name"), resp.PlanValue,
		"should not modify plan when config is set")
}

func TestUseUnknownForNullString_StateExists(t *testing.T) {
	t.Parallel()
	modifier := UseUnknownForNullString()

	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("existing-id")}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		ConfigValue: types.StringNull(),
		StateValue:  types.StringValue("existing-id"),
		PlanValue:   types.StringValue("existing-id"),
	}, resp)

	assert.Equal(t, types.StringValue("existing-id"), resp.PlanValue,
		"should not modify plan when state exists (existing element)")
}

func TestUseUnknownForNullString_Description(t *testing.T) {
	t.Parallel()
	modifier := UseUnknownForNullString()

	assert.NotEmpty(t, modifier.Description(context.Background()))
	assert.NotEmpty(t, modifier.MarkdownDescription(context.Background()))
}

// --- UseStateUnlessNameChanges tests ---

// testSchema returns a minimal schema with "name", "ports", and "built_in_environment_variables"
// attributes suitable for building tfsdk.State and tfsdk.Plan in tests.
func testSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{},
			"ports": schema.ListAttribute{
				ElementType: types.StringType,
			},
			"built_in_environment_variables": schema.ListAttribute{
				ElementType: types.StringType,
			},
		},
	}
}

var testObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"name":                          tftypes.String,
		"ports":                         tftypes.List{ElementType: tftypes.String},
		"built_in_environment_variables": tftypes.List{ElementType: tftypes.String},
	},
}

func buildTestState(name string, ports []string) tfsdk.State {
	portValues := make([]tftypes.Value, len(ports))
	for i, p := range ports {
		portValues[i] = tftypes.NewValue(tftypes.String, p)
	}
	return tfsdk.State{
		Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
			"name":                          tftypes.NewValue(tftypes.String, name),
			"ports":                         tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, portValues),
			"built_in_environment_variables": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
		}),
		Schema: testSchema(),
	}
}

func buildTestPlan(name string, ports []string) tfsdk.Plan {
	portValues := make([]tftypes.Value, len(ports))
	for i, p := range ports {
		portValues[i] = tftypes.NewValue(tftypes.String, p)
	}
	return tfsdk.Plan{
		Raw: tftypes.NewValue(testObjectType, map[string]tftypes.Value{
			"name":                          tftypes.NewValue(tftypes.String, name),
			"ports":                         tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, portValues),
			"built_in_environment_variables": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
		}),
		Schema: testSchema(),
	}
}

func TestUseStateUnlessNameChanges_NullStateReturnsUnknown(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessNameChanges()

	planList := types.ListUnknown(types.StringType)
	resp := &planmodifier.ListResponse{PlanValue: planList}
	modifier.PlanModifyList(context.Background(), planmodifier.ListRequest{
		StateValue: types.ListNull(types.StringType),
		PlanValue:  planList,
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{}, nil),
		},
	}, resp)

	assert.True(t, resp.PlanValue.IsUnknown(),
		"should leave plan as unknown when state is null (create)")
}

func TestUseStateUnlessNameChanges_NameUnchangedUsesState(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessNameChanges()

	stateList, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"QOVERY_VAR"})
	planList := types.ListUnknown(types.StringType)

	resp := &planmodifier.ListResponse{PlanValue: planList}
	modifier.PlanModifyList(context.Background(), planmodifier.ListRequest{
		StateValue: stateList,
		PlanValue:  planList,
		State:      buildTestState("my-app", []string{"port-80"}),
		Plan:       buildTestPlan("my-app", []string{"port-80"}),
		Path:       path.Root("built_in_environment_variables"),
	}, resp)

	assert.False(t, resp.PlanValue.IsUnknown(),
		"should use state value when name and ports are unchanged")
	assert.Equal(t, stateList, resp.PlanValue)
}

func TestUseStateUnlessNameChanges_NameChangedReturnsUnknown(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessNameChanges()

	stateList, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"QOVERY_VAR"})
	planList := types.ListUnknown(types.StringType)

	resp := &planmodifier.ListResponse{PlanValue: planList}
	modifier.PlanModifyList(context.Background(), planmodifier.ListRequest{
		StateValue: stateList,
		PlanValue:  planList,
		State:      buildTestState("old-name", []string{"port-80"}),
		Plan:       buildTestPlan("new-name", []string{"port-80"}),
		Path:       path.Root("built_in_environment_variables"),
	}, resp)

	assert.True(t, resp.PlanValue.IsUnknown(),
		"should leave plan as unknown when name changes (recompute)")
}

func TestUseStateUnlessNameChanges_PortsChangedReturnsUnknown(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessNameChanges()

	stateList, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"QOVERY_VAR"})
	planList := types.ListUnknown(types.StringType)

	resp := &planmodifier.ListResponse{PlanValue: planList}
	modifier.PlanModifyList(context.Background(), planmodifier.ListRequest{
		StateValue: stateList,
		PlanValue:  planList,
		State:      buildTestState("my-app", []string{"port-80"}),
		Plan:       buildTestPlan("my-app", []string{"port-80", "port-443"}),
		Path:       path.Root("built_in_environment_variables"),
	}, resp)

	assert.True(t, resp.PlanValue.IsUnknown(),
		"should leave plan as unknown when ports change (new built-in vars may appear)")
}

func TestUseStateUnlessNameChanges_Description(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessNameChanges()

	assert.NotEmpty(t, modifier.Description(context.Background()))
	assert.NotEmpty(t, modifier.MarkdownDescription(context.Background()))
}
