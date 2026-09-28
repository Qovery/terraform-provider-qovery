//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestPlan returns a plan of sch holding the given root attributes, the others null.
func newTestPlan(t *testing.T, sch schema.Schema, attributes map[string]attr.Value) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	plan := tfsdk.Plan{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
	for name, value := range attributes {
		require.False(t, plan.SetAttribute(ctx, path.Root(name), value).HasError())
	}
	return plan
}

// --- PortNameDefault ---

var testPortAttrTypes = map[string]attr.Type{
	"name":          types.StringType,
	"internal_port": types.Int64Type,
}

func testPortsSchema() schema.Schema {
	return schema.Schema{Attributes: map[string]schema.Attribute{
		"ports": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"name":          schema.StringAttribute{Optional: true, Computed: true},
				"internal_port": schema.Int64Attribute{Required: true},
			}},
		},
	}}
}

func TestPortNameDefault(t *testing.T) {
	t.Parallel()

	portsPlan := func(internalPort types.Int64) tfsdk.Plan {
		port := types.ObjectValueMust(testPortAttrTypes, map[string]attr.Value{
			"name":          types.StringUnknown(),
			"internal_port": internalPort,
		})
		return newTestPlan(t, testPortsSchema(), map[string]attr.Value{
			"ports": types.ListValueMust(types.ObjectType{AttrTypes: testPortAttrTypes}, []attr.Value{port}),
		})
	}
	destroyPlan := tfsdk.Plan{Schema: testPortsSchema(), Raw: tftypes.NewValue(testPortsSchema().Type().TerraformType(context.Background()), nil)}

	testCases := []struct {
		TestName    string
		Config      types.String
		Plan        tfsdk.Plan
		PlanValue   types.String
		ExpectValue types.String
	}{
		{TestName: "configured_name_is_kept", Config: types.StringValue("web"), Plan: portsPlan(types.Int64Value(8080)), PlanValue: types.StringValue("web"), ExpectValue: types.StringValue("web")},
		{TestName: "omitted_name_plans_p_internal_port", Config: types.StringNull(), Plan: portsPlan(types.Int64Value(8080)), PlanValue: types.StringUnknown(), ExpectValue: types.StringValue("p8080")},
		{TestName: "omitted_name_replaces_the_state_value", Config: types.StringNull(), Plan: portsPlan(types.Int64Value(3000)), PlanValue: types.StringValue("renamed-in-console"), ExpectValue: types.StringValue("p3000")},
		{TestName: "unknown_internal_port_plans_unknown", Config: types.StringNull(), Plan: portsPlan(types.Int64Unknown()), PlanValue: types.StringNull(), ExpectValue: types.StringUnknown()},
		{TestName: "destroy_is_left_alone", Config: types.StringNull(), Plan: destroyPlan, PlanValue: types.StringNull(), ExpectValue: types.StringNull()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.StringResponse{PlanValue: tc.PlanValue}
			PortNameDefault().PlanModifyString(context.Background(), planmodifier.StringRequest{
				Path:        path.Root("ports").AtListIndex(0).AtName("name"),
				ConfigValue: tc.Config,
				Plan:        tc.Plan,
				PlanValue:   tc.PlanValue,
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.ExpectValue, resp.PlanValue)
		})
	}
}

// --- JobIconUriDefault and JobLifecycleTypeDefault ---

var (
	testCronjobAttrTypes  = map[string]attr.Type{"schedule": types.StringType}
	testOnStartAttrTypes  = map[string]attr.Type{"entrypoint": types.StringType}
	testScheduleAttrTypes = map[string]attr.Type{
		"cronjob":        types.ObjectType{AttrTypes: testCronjobAttrTypes},
		"on_start":       types.ObjectType{AttrTypes: testOnStartAttrTypes},
		"lifecycle_type": types.StringType,
	}
)

func testJobScheduleSchema() schema.Schema {
	return schema.Schema{Attributes: map[string]schema.Attribute{
		"icon_uri": schema.StringAttribute{Optional: true, Computed: true},
		"schedule": schema.SingleNestedAttribute{
			Required: true,
			Attributes: map[string]schema.Attribute{
				"cronjob": schema.SingleNestedAttribute{Optional: true, Attributes: map[string]schema.Attribute{
					"schedule": schema.StringAttribute{Required: true},
				}},
				"on_start": schema.SingleNestedAttribute{Optional: true, Attributes: map[string]schema.Attribute{
					"entrypoint": schema.StringAttribute{Optional: true},
				}},
				"lifecycle_type": schema.StringAttribute{Optional: true, Computed: true},
			},
		},
	}}
}

// testJobSchedulePlan returns a job plan whose schedule is a cron job (cron), a lifecycle job
// (lifecycle) or unknown (neither).
func testJobSchedulePlan(t *testing.T, kind string) tfsdk.Plan {
	cronjob := types.ObjectNull(testCronjobAttrTypes)
	onStart := types.ObjectNull(testOnStartAttrTypes)
	switch kind {
	case "cron":
		cronjob = types.ObjectValueMust(testCronjobAttrTypes, map[string]attr.Value{"schedule": types.StringValue("*/5 * * * *")})
	case "lifecycle":
		onStart = types.ObjectValueMust(testOnStartAttrTypes, map[string]attr.Value{"entrypoint": types.StringNull()})
	case "unknown_cronjob":
		cronjob = types.ObjectUnknown(testCronjobAttrTypes)
	default:
		return newTestPlan(t, testJobScheduleSchema(), map[string]attr.Value{
			"schedule": types.ObjectUnknown(testScheduleAttrTypes),
		})
	}
	return newTestPlan(t, testJobScheduleSchema(), map[string]attr.Value{
		"schedule": types.ObjectValueMust(testScheduleAttrTypes, map[string]attr.Value{
			"cronjob":        cronjob,
			"on_start":       onStart,
			"lifecycle_type": types.StringUnknown(),
		}),
	})
}

func TestJobIconUriDefault(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Config      types.String
		Schedule    string
		PlanValue   types.String
		ExpectValue types.String
	}{
		{TestName: "configured_icon_is_kept", Config: types.StringValue("app://qovery-console/custom"), Schedule: "cron", PlanValue: types.StringValue("app://qovery-console/custom"), ExpectValue: types.StringValue("app://qovery-console/custom")},
		{TestName: "cron_job_plans_the_cron_icon", Config: types.StringNull(), Schedule: "cron", PlanValue: types.StringUnknown(), ExpectValue: types.StringValue(jobCronIconURIDefault)},
		{TestName: "lifecycle_job_plans_the_lifecycle_icon", Config: types.StringNull(), Schedule: "lifecycle", PlanValue: types.StringValue("app://qovery-console/set-in-console"), ExpectValue: types.StringValue(jobLifecycleIconURIDefault)},
		{TestName: "unknown_schedule_plans_unknown", Config: types.StringNull(), Schedule: "unknown", PlanValue: types.StringNull(), ExpectValue: types.StringUnknown()},
		{TestName: "unknown_cronjob_plans_unknown", Config: types.StringNull(), Schedule: "unknown_cronjob", PlanValue: types.StringNull(), ExpectValue: types.StringUnknown()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.StringResponse{PlanValue: tc.PlanValue}
			JobIconUriDefault().PlanModifyString(context.Background(), planmodifier.StringRequest{
				Path:        path.Root("icon_uri"),
				ConfigValue: tc.Config,
				Plan:        testJobSchedulePlan(t, tc.Schedule),
				PlanValue:   tc.PlanValue,
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.ExpectValue, resp.PlanValue)
		})
	}
}

func TestJobLifecycleTypeDefault(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Config      types.String
		Schedule    string
		PlanValue   types.String
		ExpectValue types.String
	}{
		{TestName: "configured_type_is_kept", Config: types.StringValue("TERRAFORM"), Schedule: "lifecycle", PlanValue: types.StringValue("TERRAFORM"), ExpectValue: types.StringValue("TERRAFORM")},
		{TestName: "lifecycle_job_plans_generic", Config: types.StringNull(), Schedule: "lifecycle", PlanValue: types.StringUnknown(), ExpectValue: types.StringValue(jobLifecycleTypeDefault)},
		{TestName: "lifecycle_job_resets_a_console_type", Config: types.StringNull(), Schedule: "lifecycle", PlanValue: types.StringValue("CLOUDFORMATION"), ExpectValue: types.StringValue(jobLifecycleTypeDefault)},
		{TestName: "cron_job_plans_null", Config: types.StringNull(), Schedule: "cron", PlanValue: types.StringUnknown(), ExpectValue: types.StringNull()},
		{TestName: "unknown_cronjob_plans_unknown", Config: types.StringNull(), Schedule: "unknown_cronjob", PlanValue: types.StringNull(), ExpectValue: types.StringUnknown()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.StringResponse{PlanValue: tc.PlanValue}
			JobLifecycleTypeDefault().PlanModifyString(context.Background(), planmodifier.StringRequest{
				Path:        path.Root("schedule").AtName("lifecycle_type"),
				ConfigValue: tc.Config,
				Plan:        testJobSchedulePlan(t, tc.Schedule),
				PlanValue:   tc.PlanValue,
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.ExpectValue, resp.PlanValue)
		})
	}
}

// --- CustomDomainsBoolDefaults ---

var testCustomDomainType = types.ObjectType{AttrTypes: customDomainAttrTypes}

func testCustomDomain(id, domain types.String, generateCertificate, useCdn types.Bool) attr.Value {
	return types.ObjectValueMust(customDomainAttrTypes, map[string]attr.Value{
		"id":                   id,
		"domain":               domain,
		"validation_domain":    types.StringNull(),
		"status":               types.StringNull(),
		"generate_certificate": generateCertificate,
		"use_cdn":              useCdn,
	})
}

func testCustomDomains(domains ...attr.Value) types.Set {
	return types.SetValueMust(testCustomDomainType, domains)
}

func TestCustomDomainsBoolDefaults(t *testing.T) {
	t.Parallel()

	known := types.StringValue("domain-id")
	app := types.StringValue("app.example.com")
	api := types.StringValue("api.example.com")

	testCases := []struct {
		TestName string
		Config   types.Set
		Plan     types.Set
		Expect   types.Set
	}{
		{
			TestName: "omitted_flags_plan_false",
			Config:   testCustomDomains(testCustomDomain(types.StringNull(), app, types.BoolNull(), types.BoolNull())),
			Plan:     testCustomDomains(testCustomDomain(types.StringUnknown(), app, types.BoolUnknown(), types.BoolUnknown())),
			Expect:   testCustomDomains(testCustomDomain(types.StringUnknown(), app, types.BoolValue(false), types.BoolValue(false))),
		},
		{
			TestName: "configured_flags_are_kept",
			Config:   testCustomDomains(testCustomDomain(types.StringNull(), app, types.BoolValue(true), types.BoolValue(true))),
			Plan:     testCustomDomains(testCustomDomain(known, app, types.BoolValue(true), types.BoolValue(true))),
			Expect:   testCustomDomains(testCustomDomain(known, app, types.BoolValue(true), types.BoolValue(true))),
		},
		{
			TestName: "console_set_flag_is_reset_when_omitted",
			Config:   testCustomDomains(testCustomDomain(types.StringNull(), app, types.BoolValue(true), types.BoolNull())),
			Plan:     testCustomDomains(testCustomDomain(known, app, types.BoolValue(true), types.BoolValue(true))),
			Expect:   testCustomDomains(testCustomDomain(known, app, types.BoolValue(true), types.BoolValue(false))),
		},
		{
			TestName: "elements_are_matched_by_domain",
			Config: testCustomDomains(
				testCustomDomain(types.StringNull(), app, types.BoolValue(true), types.BoolNull()),
				testCustomDomain(types.StringNull(), api, types.BoolNull(), types.BoolValue(true)),
			),
			Plan: testCustomDomains(
				testCustomDomain(types.StringUnknown(), api, types.BoolUnknown(), types.BoolValue(true)),
				testCustomDomain(types.StringUnknown(), app, types.BoolValue(true), types.BoolUnknown()),
			),
			Expect: testCustomDomains(
				testCustomDomain(types.StringUnknown(), app, types.BoolValue(true), types.BoolValue(false)),
				testCustomDomain(types.StringUnknown(), api, types.BoolValue(false), types.BoolValue(true)),
			),
		},
		{
			TestName: "unknown_domain_is_left_alone",
			Config:   testCustomDomains(testCustomDomain(types.StringNull(), types.StringUnknown(), types.BoolNull(), types.BoolNull())),
			Plan:     testCustomDomains(testCustomDomain(types.StringUnknown(), types.StringUnknown(), types.BoolUnknown(), types.BoolUnknown())),
			Expect:   testCustomDomains(testCustomDomain(types.StringUnknown(), types.StringUnknown(), types.BoolUnknown(), types.BoolUnknown())),
		},
		{
			TestName: "null_set_is_left_alone",
			Config:   types.SetNull(testCustomDomainType),
			Plan:     types.SetNull(testCustomDomainType),
			Expect:   types.SetNull(testCustomDomainType),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.SetResponse{PlanValue: tc.Plan}
			CustomDomainsBoolDefaults("generate_certificate", "use_cdn").PlanModifySet(context.Background(), planmodifier.SetRequest{
				Path:        path.Root("custom_domains"),
				ConfigValue: tc.Config,
				PlanValue:   tc.Plan,
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.True(t, tc.Expect.Equal(resp.PlanValue), "got %s, want %s", resp.PlanValue, tc.Expect)
		})
	}
}

// TestCustomDomainsBoolDefaults_OnlyListedAttributes covers helm, whose generate_certificate is
// Required: only use_cdn gets a default.
func TestCustomDomainsBoolDefaults_OnlyListedAttributes(t *testing.T) {
	t.Parallel()

	app := types.StringValue("app.example.com")
	plan := testCustomDomains(testCustomDomain(types.StringUnknown(), app, types.BoolUnknown(), types.BoolUnknown()))

	resp := &planmodifier.SetResponse{PlanValue: plan}
	CustomDomainsBoolDefaults("use_cdn").PlanModifySet(context.Background(), planmodifier.SetRequest{
		Path:        path.Root("custom_domains"),
		ConfigValue: testCustomDomains(testCustomDomain(types.StringNull(), app, types.BoolNull(), types.BoolNull())),
		PlanValue:   plan,
	}, resp)

	expect := testCustomDomains(testCustomDomain(types.StringUnknown(), app, types.BoolUnknown(), types.BoolValue(false)))
	assert.True(t, expect.Equal(resp.PlanValue), "got %s, want %s", resp.PlanValue, expect)
}

// --- RejectChangeAfterCreate ---

func TestRejectChangeAfterCreate(t *testing.T) {
	t.Parallel()

	sch := schema.Schema{Attributes: map[string]schema.Attribute{
		"blueprint_id": schema.StringAttribute{Optional: true, Computed: true},
	}}
	existing := tfsdk.State{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(context.Background()), map[string]tftypes.Value{
		"blueprint_id": tftypes.NewValue(tftypes.String, "blueprint-a"),
	})}
	creating := tfsdk.State{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(context.Background()), nil)}

	testCases := []struct {
		TestName    string
		State       tfsdk.State
		StateValue  types.String
		PlanValue   types.String
		ExpectError bool
	}{
		{TestName: "create_accepts_any_value", State: creating, StateValue: types.StringNull(), PlanValue: types.StringValue("blueprint-a")},
		{TestName: "unchanged_value_is_accepted", State: existing, StateValue: types.StringValue("blueprint-a"), PlanValue: types.StringValue("blueprint-a")},
		{TestName: "unknown_value_is_accepted", State: existing, StateValue: types.StringValue("blueprint-a"), PlanValue: types.StringUnknown()},
		{TestName: "different_value_is_rejected", State: existing, StateValue: types.StringValue("blueprint-a"), PlanValue: types.StringValue("blueprint-b"), ExpectError: true},
		{TestName: "setting_it_after_create_is_rejected", State: existing, StateValue: types.StringNull(), PlanValue: types.StringValue("blueprint-b"), ExpectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.StringResponse{PlanValue: tc.PlanValue}
			RejectChangeAfterCreate(blueprintIDChangeReason).PlanModifyString(context.Background(), planmodifier.StringRequest{
				Path:       path.Root("blueprint_id"),
				State:      tc.State,
				StateValue: tc.StateValue,
				Plan:       newTestPlan(t, sch, map[string]attr.Value{"blueprint_id": tc.PlanValue}),
				PlanValue:  tc.PlanValue,
			}, resp)

			assert.Equal(t, tc.ExpectError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.PlanValue, resp.PlanValue, "the modifier must never rewrite the plan")
		})
	}
}

func TestRejectChangeAfterCreate_Message(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sch := schema.Schema{Attributes: map[string]schema.Attribute{
		"blueprint_id": schema.StringAttribute{Optional: true, Computed: true},
	}}
	state := tfsdk.State{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), map[string]tftypes.Value{
		"blueprint_id": tftypes.NewValue(tftypes.String, "blueprint-a"),
	})}

	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("blueprint-b")}
	RejectChangeAfterCreate(blueprintIDChangeReason).PlanModifyString(ctx, planmodifier.StringRequest{
		Path:       path.Root("blueprint_id"),
		State:      state,
		StateValue: types.StringValue("blueprint-a"),
		Plan:       newTestPlan(t, sch, map[string]attr.Value{"blueprint_id": types.StringValue("blueprint-b")}),
		PlanValue:  types.StringValue("blueprint-b"),
	}, resp)

	require.True(t, resp.Diagnostics.HasError())
	detail := resp.Diagnostics.Errors()[0].Detail()
	assert.Contains(t, detail, blueprintIDChangeReason)
	assert.Contains(t, detail, `"blueprint-a"`, "the error names the value to declare")
}

// TestJobLifecycleType_ChangeRejected runs the lifecycle_type modifiers in schema order: the
// default planned for an omitted type is rejected when the job was created with another type,
// because q-core cannot change it.
func TestJobLifecycleType_ChangeRejected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	sch := testJobScheduleSchema()
	lifecyclePlan := testJobSchedulePlan(t, "lifecycle")
	existing := tfsdk.State{Schema: sch, Raw: lifecyclePlan.Raw}

	testCases := []struct {
		TestName    string
		Config      types.String
		StateValue  types.String
		ExpectError bool
	}{
		{TestName: "omitted_type_of_a_generic_job_plans_nothing", Config: types.StringNull(), StateValue: types.StringValue("GENERIC")},
		{TestName: "omitted_type_of_a_terraform_job_is_rejected", Config: types.StringNull(), StateValue: types.StringValue("TERRAFORM"), ExpectError: true},
		{TestName: "declared_type_of_a_terraform_job_is_kept", Config: types.StringValue("TERRAFORM"), StateValue: types.StringValue("TERRAFORM")},
		{TestName: "changed_type_is_rejected", Config: types.StringValue("CLOUDFORMATION"), StateValue: types.StringValue("TERRAFORM"), ExpectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.StringRequest{
				Path:        path.Root("schedule").AtName("lifecycle_type"),
				ConfigValue: tc.Config,
				State:       existing,
				StateValue:  tc.StateValue,
				Plan:        lifecyclePlan,
				PlanValue:   tc.Config,
			}
			if tc.Config.IsNull() {
				req.PlanValue = types.StringUnknown()
			}
			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
			for _, m := range []planmodifier.String{JobLifecycleTypeDefault(), RejectChangeAfterCreate(lifecycleTypeChangeReason)} {
				req.PlanValue = resp.PlanValue
				m.PlanModifyString(ctx, req, resp)
			}

			assert.Equal(t, tc.ExpectError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}

// --- UseStateUnlessRepositoryChanges ---

var testGitRepositoryAttrTypes = map[string]attr.Type{
	"url":    types.StringType,
	"branch": types.StringType,
}

func testGitRepositorySchema() schema.Schema {
	return schema.Schema{Attributes: map[string]schema.Attribute{
		"git_repository": schema.SingleNestedAttribute{
			Required: true,
			Attributes: map[string]schema.Attribute{
				"url":    schema.StringAttribute{Required: true},
				"branch": schema.StringAttribute{Optional: true, Computed: true},
			},
		},
	}}
}

func testGitRepositoryData(t *testing.T, url, branch types.String) (tfsdk.Plan, tfsdk.State) {
	t.Helper()
	plan := newTestPlan(t, testGitRepositorySchema(), map[string]attr.Value{
		"git_repository": types.ObjectValueMust(testGitRepositoryAttrTypes, map[string]attr.Value{"url": url, "branch": branch}),
	})
	return plan, tfsdk.State{Schema: plan.Schema, Raw: plan.Raw}
}

func TestUseStateUnlessRepositoryChanges(t *testing.T) {
	t.Parallel()

	const (
		repositoryA = "https://github.com/Qovery/test_http_server.git"
		repositoryB = "https://github.com/Qovery/helm_chart_engine_testing.git"
	)

	testCases := []struct {
		TestName    string
		Config      types.String
		StateURL    types.String
		StateValue  types.String
		PlannedURL  types.String
		ExpectValue types.String
	}{
		{TestName: "omitted_branch_keeps_the_resolved_one", Config: types.StringNull(), StateURL: types.StringValue(repositoryA), StateValue: types.StringValue("master"), PlannedURL: types.StringValue(repositoryA), ExpectValue: types.StringValue("master")},
		{TestName: "new_repository_resolves_its_default_branch", Config: types.StringNull(), StateURL: types.StringValue(repositoryA), StateValue: types.StringValue("master"), PlannedURL: types.StringValue(repositoryB), ExpectValue: types.StringUnknown()},
		{TestName: "unknown_repository_leaves_the_branch_unknown", Config: types.StringNull(), StateURL: types.StringValue(repositoryA), StateValue: types.StringValue("master"), PlannedURL: types.StringUnknown(), ExpectValue: types.StringUnknown()},
		{TestName: "declared_branch_is_left_alone", Config: types.StringValue("main"), StateURL: types.StringValue(repositoryA), StateValue: types.StringValue("master"), PlannedURL: types.StringValue(repositoryA), ExpectValue: types.StringValue("main")},
		{TestName: "no_recorded_branch_is_left_alone", Config: types.StringNull(), StateURL: types.StringValue(repositoryA), StateValue: types.StringNull(), PlannedURL: types.StringValue(repositoryA), ExpectValue: types.StringUnknown()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			planValue := tc.Config
			if tc.Config.IsNull() {
				planValue = types.StringUnknown()
			}
			plan, _ := testGitRepositoryData(t, tc.PlannedURL, planValue)
			_, state := testGitRepositoryData(t, tc.StateURL, tc.StateValue)

			resp := &planmodifier.StringResponse{PlanValue: planValue}
			UseStateUnlessRepositoryChanges().PlanModifyString(context.Background(), planmodifier.StringRequest{
				Path:        path.Root("git_repository").AtName("branch"),
				ConfigValue: tc.Config,
				Plan:        plan,
				PlanValue:   planValue,
				State:       state,
				StateValue:  tc.StateValue,
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.ExpectValue, resp.PlanValue)
		})
	}
}
