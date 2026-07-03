//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Coverage for the source/git_repository/image_name/tag/registry_id branches
// of UseStateUnlessNameChanges. Acceptance tests only exercise `name` and
// `git_repository.root_path` end-to-end; single-tag ECR fixtures and a
// single upstream helm version block the rest, so they're covered here.

var extendedTestObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String,
		"source": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"tag": tftypes.String,
			},
		},
		"git_repository": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"branch": tftypes.String,
			},
		},
		"values_override": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"commit": tftypes.String,
			},
		},
		"schedule": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"lifecycle_type": tftypes.String,
			},
		},
		"image_name":                     tftypes.String,
		"tag":                            tftypes.String,
		"registry_id":                    tftypes.String,
		"mode":                           tftypes.String,
		"built_in_environment_variables": tftypes.List{ElementType: tftypes.String},
	},
}

func extendedTestSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{},
			"source": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"tag": schema.StringAttribute{},
				},
			},
			"git_repository": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"branch": schema.StringAttribute{},
				},
			},
			"values_override": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"commit": schema.StringAttribute{},
				},
			},
			"schedule": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"lifecycle_type": schema.StringAttribute{},
				},
			},
			"image_name":  schema.StringAttribute{},
			"tag":         schema.StringAttribute{},
			"registry_id": schema.StringAttribute{},
			"mode":        schema.StringAttribute{},
			"built_in_environment_variables": schema.ListAttribute{
				ElementType: types.StringType,
			},
		},
	}
}

type extendedTestArgs struct {
	name         string
	sourceTag    string
	gitBranch    string
	valuesCommit string
	scheduleType string
	imageName    string
	tag          string
	registryID   string
	mode         string
}

func makeExtendedRaw(a extendedTestArgs) tftypes.Value {
	return tftypes.NewValue(extendedTestObjectType, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, a.name),
		"source": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{"tag": tftypes.String},
		}, map[string]tftypes.Value{
			"tag": tftypes.NewValue(tftypes.String, a.sourceTag),
		}),
		"git_repository": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{"branch": tftypes.String},
		}, map[string]tftypes.Value{
			"branch": tftypes.NewValue(tftypes.String, a.gitBranch),
		}),
		"values_override": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{"commit": tftypes.String},
		}, map[string]tftypes.Value{
			"commit": tftypes.NewValue(tftypes.String, a.valuesCommit),
		}),
		"schedule": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{"lifecycle_type": tftypes.String},
		}, map[string]tftypes.Value{
			"lifecycle_type": tftypes.NewValue(tftypes.String, a.scheduleType),
		}),
		"image_name":                     tftypes.NewValue(tftypes.String, a.imageName),
		"tag":                            tftypes.NewValue(tftypes.String, a.tag),
		"registry_id":                    tftypes.NewValue(tftypes.String, a.registryID),
		"mode":                           tftypes.NewValue(tftypes.String, a.mode),
		"built_in_environment_variables": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
	})
}

func TestUseStateUnlessNameChanges_ValueAffectingAttrChanges(t *testing.T) {
	t.Parallel()

	base := extendedTestArgs{
		name:         "my-app",
		sourceTag:    "1.0.0",
		gitBranch:    "main",
		valuesCommit: "abc123",
		scheduleType: "GENERIC",
		imageName:    "img",
		tag:          "1.0.0",
		registryID:   "reg-1",
		mode:         "PRODUCTION",
	}

	cases := []struct {
		name        string
		mutate      func(*extendedTestArgs)
		wantUnknown bool
	}{
		{"source_tag", func(a *extendedTestArgs) { a.sourceTag = "2.0.0" }, true},
		{"git_branch", func(a *extendedTestArgs) { a.gitBranch = "develop" }, true},
		{"values_override_commit", func(a *extendedTestArgs) { a.valuesCommit = "def456" }, true},
		{"schedule_type", func(a *extendedTestArgs) { a.scheduleType = "TERRAFORM" }, true},
		{"image_name", func(a *extendedTestArgs) { a.imageName = "other-img" }, true},
		{"tag", func(a *extendedTestArgs) { a.tag = "2.0.0" }, true},
		{"registry_id", func(a *extendedTestArgs) { a.registryID = "reg-2" }, true},
		{"mode", func(a *extendedTestArgs) { a.mode = "STAGING" }, true},
		{"unchanged", func(a *extendedTestArgs) {}, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			planArgs := base
			tc.mutate(&planArgs)

			stateList, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"QOVERY_VAR"})
			planList := types.ListUnknown(types.StringType)
			sch := extendedTestSchema()

			resp := &planmodifier.ListResponse{PlanValue: planList}
			UseStateUnlessNameChanges().PlanModifyList(context.Background(), planmodifier.ListRequest{
				StateValue: stateList,
				PlanValue:  planList,
				State:      tfsdk.State{Raw: makeExtendedRaw(base), Schema: sch},
				Plan:       tfsdk.Plan{Raw: makeExtendedRaw(planArgs), Schema: sch},
				Path:       path.Root("built_in_environment_variables"),
			}, resp)

			assert.Equal(t, tc.wantUnknown, resp.PlanValue.IsUnknown())
		})
	}
}

// --- UseStateUnlessPortsChange tests ---

func TestUseStateUnlessPortsChange_NullStateReturnsUnknown(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessPortsChange()

	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		StateValue: types.StringNull(),
		PlanValue:  types.StringUnknown(),
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{}, nil),
		},
	}, resp)

	assert.True(t, resp.PlanValue.IsUnknown(),
		"should leave plan as unknown when state is null (create)")
}

func TestUseStateUnlessPortsChange_PortsUnchangedPreservesNull(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessPortsChange()

	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		StateValue: types.StringNull(),
		PlanValue:  types.StringUnknown(),
		State:      buildTestState("my-app", []string{"port-80"}),
		Plan:       buildTestPlan("my-app", []string{"port-80"}),
	}, resp)

	assert.True(t, resp.PlanValue.IsNull(),
		"should preserve null state when ports unchanged (no external_host)")
}

func TestUseStateUnlessPortsChange_PortsUnchangedPreservesHost(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessPortsChange()

	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		StateValue: types.StringValue("my-host.example.com"),
		PlanValue:  types.StringUnknown(),
		State:      buildTestState("my-app", []string{"port-80"}),
		Plan:       buildTestPlan("my-app", []string{"port-80"}),
	}, resp)

	assert.Equal(t, types.StringValue("my-host.example.com"), resp.PlanValue,
		"should preserve existing host when ports unchanged")
}

func TestUseStateUnlessPortsChange_PortsChangedReturnsUnknown(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessPortsChange()

	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		StateValue: types.StringNull(),
		PlanValue:  types.StringUnknown(),
		State:      buildTestState("my-app", []string{"port-80"}),
		Plan:       buildTestPlan("my-app", []string{"port-80", "port-443"}),
	}, resp)

	assert.True(t, resp.PlanValue.IsUnknown(),
		"should recompute when ports change (external_host may appear)")
}

func TestUseStateUnlessPortsChange_Description(t *testing.T) {
	t.Parallel()
	modifier := UseStateUnlessPortsChange()

	assert.NotEmpty(t, modifier.Description(context.Background()))
	assert.NotEmpty(t, modifier.MarkdownDescription(context.Background()))
}

// --- RejectExistingVpcChange tests ---

// existingVpcTestAttrTypes mixes the three child types found in the real
// existing-VPC blocks so one object exercises every comparison branch.
var existingVpcTestAttrTypes = map[string]attr.Type{
	"vpc_id":  types.StringType,
	"subnets": types.ListType{ElemType: types.StringType},
	"private": types.BoolType,
}

func existingVpcObject(vpcID types.String, subnets types.List, private types.Bool) types.Object {
	return types.ObjectValueMust(existingVpcTestAttrTypes, map[string]attr.Value{
		"vpc_id":  vpcID,
		"subnets": subnets,
		"private": private,
	})
}

// TestRejectExistingVpcChange_Behavior exercises the block-level modifier
// semantics: presence changes and known child value changes after creation are
// rejected without forcing replacement, while creation, no-ops, per-type null
// equivalences (null≡empty list/string, null≡false), and unknown (deferred)
// values are allowed.
func TestRejectExistingVpcChange_Behavior(t *testing.T) {
	t.Parallel()

	subnetA := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-a")})
	subnetAB := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-a"), types.StringValue("subnet-b")})
	subnetBA := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-b"), types.StringValue("subnet-a")})
	subnetAA := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-a"), types.StringValue("subnet-a")})
	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	nullList := types.ListNull(types.StringType)
	withUnknownElement := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-a"), types.StringUnknown()})

	vpcA := types.StringValue("vpc-a")
	base := existingVpcObject(vpcA, subnetA, types.BoolValue(true))
	nullObject := types.ObjectNull(existingVpcTestAttrTypes)

	raw := types.StringValue("non-empty-resource")
	existingState := buildState(&raw)
	updatePlan := buildPlan(&raw)

	testCases := []struct {
		TestName    string
		State       tfsdk.State
		Plan        tfsdk.Plan
		StateValue  types.Object
		PlanValue   types.Object
		ExpectError bool
	}{
		{"create_allows_initial_block", buildState(nil), updatePlan, nullObject, base, false},
		{"destroy_plan_skipped", existingState, buildPlan(nil), base, nullObject, false},
		{"unchanged_allowed", existingState, updatePlan, base, base, false},
		{"both_null_allowed", existingState, updatePlan, nullObject, nullObject, false},
		{"unknown_plan_skipped", existingState, updatePlan, base, types.ObjectUnknown(existingVpcTestAttrTypes), false},
		{"removal_rejected", existingState, updatePlan, base, nullObject, true},
		{"addition_after_creation_rejected", existingState, updatePlan, nullObject, base, true},
		{"unknown_child_skipped", existingState, updatePlan, base,
			existingVpcObject(types.StringUnknown(), subnetA, types.BoolValue(true)), false},
		{"string_change_rejected", existingState, updatePlan, base,
			existingVpcObject(types.StringValue("vpc-b"), subnetA, types.BoolValue(true)), true},
		{"string_cleared_after_creation_rejected", existingState, updatePlan, base,
			existingVpcObject(types.StringNull(), subnetA, types.BoolValue(true)), true},
		{"string_null_and_empty_equivalent", existingState, updatePlan,
			existingVpcObject(types.StringNull(), subnetA, types.BoolValue(true)),
			existingVpcObject(types.StringValue(""), subnetA, types.BoolValue(true)), false},
		{"bool_change_rejected", existingState, updatePlan, base,
			existingVpcObject(vpcA, subnetA, types.BoolValue(false)), true},
		{"bool_set_true_after_creation_rejected", existingState, updatePlan,
			existingVpcObject(vpcA, subnetA, types.BoolNull()),
			existingVpcObject(vpcA, subnetA, types.BoolValue(true)), true},
		{"bool_null_and_false_equivalent", existingState, updatePlan,
			existingVpcObject(vpcA, subnetA, types.BoolNull()),
			existingVpcObject(vpcA, subnetA, types.BoolValue(false)), false},
		{"list_element_added_rejected", existingState, updatePlan, base,
			existingVpcObject(vpcA, subnetAB, types.BoolValue(true)), true},
		{"list_multiplicity_change_rejected", existingState, updatePlan,
			existingVpcObject(vpcA, subnetAB, types.BoolValue(true)),
			existingVpcObject(vpcA, subnetAA, types.BoolValue(true)), true},
		{"list_removed_after_creation_rejected", existingState, updatePlan, base,
			existingVpcObject(vpcA, nullList, types.BoolValue(true)), true},
		{"list_reorder_rejected", existingState, updatePlan,
			existingVpcObject(vpcA, subnetAB, types.BoolValue(true)),
			existingVpcObject(vpcA, subnetBA, types.BoolValue(true)), true},
		{"list_null_and_empty_equivalent", existingState, updatePlan,
			existingVpcObject(vpcA, nullList, types.BoolValue(true)),
			existingVpcObject(vpcA, emptyList, types.BoolValue(true)), false},
		{"list_unknown_element_skipped", existingState, updatePlan, base,
			existingVpcObject(vpcA, withUnknownElement, types.BoolValue(true)), false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ObjectResponse{PlanValue: tc.PlanValue}
			RejectExistingVpcChange().PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
				Path:       path.Root("features").AtName("existing_vpc"),
				State:      tc.State,
				StateValue: tc.StateValue,
				Plan:       tc.Plan,
				PlanValue:  tc.PlanValue,
			}, resp)

			assert.False(t, resp.RequiresReplace, "modifier must never force replacement")
			assert.Equal(t, tc.ExpectError, resp.Diagnostics.HasError(),
				"unexpected diagnostics outcome: %v", resp.Diagnostics)
		})
	}
}

// TestRejectExistingVpcChange_DiagnosticDetails asserts that a child change is
// reported on the changed child's path, and that an order-only list difference
// gets the reorder-specific diagnostic (which tells the user to match the
// state order) rather than the generic immutability error.
func TestRejectExistingVpcChange_DiagnosticDetails(t *testing.T) {
	t.Parallel()

	subnetAB := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-a"), types.StringValue("subnet-b")})
	subnetBA := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-b"), types.StringValue("subnet-a")})
	vpcA := types.StringValue("vpc-a")
	raw := types.StringValue("non-empty-resource")
	blockPath := path.Root("features").AtName("existing_vpc")

	stateValue := existingVpcObject(vpcA, subnetAB, types.BoolValue(true))
	planValue := existingVpcObject(vpcA, subnetBA, types.BoolValue(true))

	resp := &planmodifier.ObjectResponse{PlanValue: planValue}
	RejectExistingVpcChange().PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
		Path:       blockPath,
		State:      buildState(&raw),
		StateValue: stateValue,
		Plan:       buildPlan(&raw),
		PlanValue:  planValue,
	}, resp)

	require.Len(t, resp.Diagnostics.Errors(), 1, "exactly the changed child must be reported")
	diagErr := resp.Diagnostics.Errors()[0]
	assert.Equal(t, "Existing VPC list order changed", diagErr.Summary(),
		"reorder must get the order-specific diagnostic, not the generic immutability error")

	withPath, ok := diagErr.(diag.DiagnosticWithPath)
	require.True(t, ok, "diagnostic must carry an attribute path")
	assert.True(t, blockPath.AtName("subnets").Equal(withPath.Path()),
		"diagnostic must point at the changed child attribute, got %s", withPath.Path())
}
