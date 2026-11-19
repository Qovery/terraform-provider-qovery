package qovery

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// autoscalingTriggerAuthAttrTypes returns the attribute types for a scaler's
// inline trigger_authentication object.
func autoscalingTriggerAuthAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"config_yaml": types.StringType,
	}
}

// autoscalingScalerAttrTypes returns the attribute types for a single scaler object.
func autoscalingScalerAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"scaler_type":            types.StringType,
		"enabled":                types.BoolType,
		"role":                   types.StringType,
		"config_json":            jsontypes.NormalizedType{},
		"config_yaml":            types.StringType,
		"trigger_authentication": types.ObjectType{AttrTypes: autoscalingTriggerAuthAttrTypes()},
	}
}

// autoscalingAttrTypes returns the attribute types for the autoscaling object.
func autoscalingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"polling_interval_seconds":        types.Int64Type,
		"cooldown_period_seconds":         types.Int64Type,
		"initial_cooldown_period_seconds": types.Int64Type,
		"idle_replica_count":              types.Int64Type,
		"advanced_yaml":                   types.StringType,
		"scalers":                         types.ListType{ElemType: types.ObjectType{AttrTypes: autoscalingScalerAttrTypes()}},
	}
}

// Backend defaults for the KEDA polling/cooldown windows. Declared as static
// schema defaults so an omitted value resolves at plan time instead of surfacing
// as a null-then-computed inconsistency after apply.
const (
	autoscalingPollingIntervalSecondsDefault = 30
	autoscalingCooldownPeriodSecondsDefault  = 300
)

// Descriptions of the autoscaling block, shared by the resource and data source schemas.
const (
	autoscalingDescription = "Event-driven autoscaling with KEDA, which requires KEDA on the cluster. " +
		"It adds to the CPU and memory autoscaling of `min_running_instances` and `max_running_instances`, and allows `min_running_instances = 0`."
	autoscalingPollingIntervalDescription  = "Seconds between two polls of the scalers."
	autoscalingCooldownPeriodDescription   = "Seconds to wait after the last trigger before scaling down."
	autoscalingScalersDescription          = "KEDA scalers that drive the autoscaling. Set at least one."
	autoscalingInitialCooldownDescription  = "Seconds to wait after the service is created before it can scale to 0 (KEDA `initialCooldownPeriod`). Unset leaves KEDA's default of 0."
	autoscalingIdleReplicaCountDescription = "Instances to keep while no scaler is active (KEDA `idleReplicaCount`). Only `0` is supported, and it must be less than `min_running_instances`: the service then scales to 0 when idle instead of to its minimum. Unset disables the idle state."
	autoscalingAdvancedYAMLDescription     = "KEDA ScaledObject `advanced` field, as YAML, e.g. `horizontalPodAutoscalerConfig.behavior` to tune how fast the service scales up and down. " +
		"Allowed top-level keys: `horizontalPodAutoscalerConfig`, `restoreToOriginalReplicaCount` and `scalingModifiers`."
	autoscalingScalerIDDescription          = "Id of the scaler."
	autoscalingScalerTypeDescription        = "Type of the KEDA scaler, for example `cpu`, `memory`, `prometheus` or `cron`."
	autoscalingScalerEnabledDescription     = "Whether the scaler is enabled."
	autoscalingScalerRoleDescription        = "Role of the scaler: `PRIMARY` or `SAFETY`."
	autoscalingScalerConfigJSONDescription  = "Scaler `metadata` only, as an object of string values, e.g. `{\"query\": \"...\", \"threshold\": \"10\"}`. Other trigger fields (`name`, `metricType`, `useCachedMetrics`) are only available through `config_yaml`. Conflicts with `config_yaml`."
	autoscalingScalerConfigYAMLDescription  = "Body of the KEDA trigger as YAML, without `type` (from `scaler_type`) and `authenticationRef` (from `trigger_authentication`). Scaler settings are under `metadata:`, next to the optional `name`, `metricType` and `useCachedMetrics`. Conflicts with `config_json`."
	autoscalingTriggerAuthDescription       = "KEDA TriggerAuthentication of the scaler."
	autoscalingTriggerAuthNameDescription   = "Name of the trigger authentication."
	autoscalingTriggerAuthConfigDescription = "Configuration of the trigger authentication, as YAML."
	// autoscalingTriggerAuthNameIgnoredNote completes the resource description:
	// the API ignores the name sent, see triggerAuthNames.
	autoscalingTriggerAuthNameIgnoredNote = " Qovery ignores this value and generates the name of the KEDA TriggerAuthentication, " +
		"so changing it has no effect on the service; the data source reports the generated name."
)

// autoscalingResourceSchema returns the resource schema attribute for the
// `autoscaling` block, shared by the application and container resources.
func autoscalingResourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: autoscalingDescription,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"polling_interval_seconds": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(autoscalingPollingIntervalDescription, autoscalingPollingIntervalSecondsDefault),
				Optional:            true,
				Computed:            true,
				// Static default matches the backend default so adding an autoscaling
				// block without this field does not yield a null plan value that the
				// API then fills in (which fails with "inconsistent result after apply").
				Default: int64default.StaticInt64(autoscalingPollingIntervalSecondsDefault),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cooldown_period_seconds": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(autoscalingCooldownPeriodDescription, autoscalingCooldownPeriodSecondsDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(autoscalingCooldownPeriodSecondsDefault),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"initial_cooldown_period_seconds": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinMaxDescription(autoscalingInitialCooldownDescription,
					autoscaling.MinInitialCooldownPeriodSeconds, autoscaling.MaxInitialCooldownPeriodSeconds, nil),
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(autoscaling.MinInitialCooldownPeriodSeconds, autoscaling.MaxInitialCooldownPeriodSeconds),
				},
			},
			"idle_replica_count": schema.Int64Attribute{
				MarkdownDescription: autoscalingIdleReplicaCountDescription,
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(autoscaling.SupportedIdleReplicaCount),
				},
			},
			"advanced_yaml": schema.StringAttribute{
				MarkdownDescription: autoscalingAdvancedYAMLDescription,
				Optional:            true,
				Validators: []validator.String{
					// The API reads a blank value as unset and returns null, which
					// would not match the configuration.
					stringvalidator.RegexMatches(regexp.MustCompile(`\S`), "must not be blank"),
				},
			},
			"scalers": schema.ListNestedAttribute{
				MarkdownDescription: autoscalingScalersDescription,
				Required:            true,
				Validators: []validator.List{
					validators.ScalerConfigExactlyOneValidator{},
				},
				PlanModifiers: []planmodifier.List{
					scalerIDsFromStateModifier{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: autoscalingScalerIDDescription,
							Computed:            true,
						},
						"scaler_type": schema.StringAttribute{
							MarkdownDescription: autoscalingScalerTypeDescription,
							Required:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: descriptions.NewBoolDefaultDescription(autoscalingScalerEnabledDescription, true),
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
						},
						"role": schema.StringAttribute{
							MarkdownDescription: autoscalingScalerRoleDescription,
							Required:            true,
							Validators: []validator.String{
								validators.NewStringEnumValidator([]string{string(autoscaling.RolePrimary), string(autoscaling.RoleSafety)}),
							},
						},
						"config_json": schema.StringAttribute{
							MarkdownDescription: autoscalingScalerConfigJSONDescription,
							Optional:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
						"config_yaml": schema.StringAttribute{
							MarkdownDescription: autoscalingScalerConfigYAMLDescription,
							Optional:            true,
						},
						"trigger_authentication": schema.SingleNestedAttribute{
							MarkdownDescription: autoscalingTriggerAuthDescription,
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: autoscalingTriggerAuthNameDescription + autoscalingTriggerAuthNameIgnoredNote,
									Required:            true,
								},
								"config_yaml": schema.StringAttribute{
									MarkdownDescription: autoscalingTriggerAuthConfigDescription,
									Optional:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// autoscalingDataSourceSchema returns the (fully computed) data source schema
// attribute for the `autoscaling` block. The data source shares the resource
// model struct, so this attribute must mirror the resource schema or reading
// state fails at runtime.
func autoscalingDataSourceSchema() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		MarkdownDescription: autoscalingDescription,
		Computed:            true,
		Attributes: map[string]dsschema.Attribute{
			"polling_interval_seconds": dsschema.Int64Attribute{
				MarkdownDescription: autoscalingPollingIntervalDescription,
				Computed:            true,
			},
			"cooldown_period_seconds": dsschema.Int64Attribute{
				MarkdownDescription: autoscalingCooldownPeriodDescription,
				Computed:            true,
			},
			"initial_cooldown_period_seconds": dsschema.Int64Attribute{
				MarkdownDescription: autoscalingInitialCooldownDescription,
				Computed:            true,
			},
			"idle_replica_count": dsschema.Int64Attribute{
				MarkdownDescription: autoscalingIdleReplicaCountDescription,
				Computed:            true,
			},
			"advanced_yaml": dsschema.StringAttribute{
				MarkdownDescription: autoscalingAdvancedYAMLDescription,
				Computed:            true,
			},
			"scalers": dsschema.ListNestedAttribute{
				MarkdownDescription: autoscalingScalersDescription,
				Computed:            true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							MarkdownDescription: autoscalingScalerIDDescription,
							Computed:            true,
						},
						"scaler_type": dsschema.StringAttribute{
							MarkdownDescription: autoscalingScalerTypeDescription,
							Computed:            true,
						},
						"enabled": dsschema.BoolAttribute{
							MarkdownDescription: autoscalingScalerEnabledDescription,
							Computed:            true,
						},
						"role": dsschema.StringAttribute{
							MarkdownDescription: autoscalingScalerRoleDescription,
							Computed:            true,
						},
						"config_json": dsschema.StringAttribute{
							MarkdownDescription: autoscalingScalerConfigJSONDescription,
							Computed:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
						"config_yaml": dsschema.StringAttribute{
							MarkdownDescription: autoscalingScalerConfigYAMLDescription,
							Computed:            true,
						},
						"trigger_authentication": dsschema.SingleNestedAttribute{
							MarkdownDescription: autoscalingTriggerAuthDescription,
							Computed:            true,
							Attributes: map[string]dsschema.Attribute{
								"name": dsschema.StringAttribute{
									MarkdownDescription: autoscalingTriggerAuthNameDescription,
									Computed:            true,
								},
								"config_yaml": dsschema.StringAttribute{
									MarkdownDescription: autoscalingTriggerAuthConfigDescription,
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// toQoveryAutoscaling converts the Terraform `autoscaling` object into the
// shared domain model. Returns nil when the block is absent (omitempty/nullable).
func toQoveryAutoscaling(o types.Object) *autoscaling.AutoscalingPolicy {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}

	attrs := o.Attributes()

	policy := &autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds:       int64ObjectAttrToInt32Pointer(attrs["polling_interval_seconds"]),
		CooldownPeriodSeconds:        int64ObjectAttrToInt32Pointer(attrs["cooldown_period_seconds"]),
		InitialCooldownPeriodSeconds: int64ObjectAttrToInt32Pointer(attrs["initial_cooldown_period_seconds"]),
		IdleReplicaCount:             int64ObjectAttrToInt32Pointer(attrs["idle_replica_count"]),
		AdvancedYAML:                 objectAttrToStringPointer(attrs["advanced_yaml"]),
	}

	if scalersList, ok := attrs["scalers"].(types.List); ok {
		policy.Scalers = scalersFromList(scalersList)
	}

	return policy
}

// scalersFromList converts the known scalers of a Terraform `scalers` list into
// the domain model. Returns nil when the list is null or unknown.
func scalersFromList(l types.List) []autoscaling.Scaler {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}

	scalers := make([]autoscaling.Scaler, 0, len(l.Elements()))
	for _, elem := range l.Elements() {
		obj, ok := elem.(types.Object)
		if !ok || obj.IsNull() || obj.IsUnknown() {
			continue
		}
		scalers = append(scalers, scalerFromObject(obj))
	}
	return scalers
}

func scalerFromObject(obj types.Object) autoscaling.Scaler {
	sAttrs := obj.Attributes()

	scaler := autoscaling.Scaler{
		ID:         objectAttrToString(sAttrs["id"]),
		ScalerType: objectAttrToString(sAttrs["scaler_type"]),
		Enabled:    objectAttrToBool(sAttrs["enabled"]),
		Role:       autoscaling.Role(objectAttrToString(sAttrs["role"])),
	}

	if cj, ok := sAttrs["config_json"].(jsontypes.Normalized); ok && !cj.IsNull() && !cj.IsUnknown() {
		scaler.Config.ConfigJSON = cj.ValueString()
	}
	scaler.Config.ConfigYAML = objectAttrToString(sAttrs["config_yaml"])

	if ta, ok := sAttrs["trigger_authentication"].(types.Object); ok && !ta.IsNull() && !ta.IsUnknown() {
		taAttrs := ta.Attributes()
		triggerAuth := &autoscaling.TriggerAuth{
			Name: objectAttrToString(taAttrs["name"]),
		}
		if cy := objectAttrToStringPointer(taAttrs["config_yaml"]); cy != nil {
			triggerAuth.ConfigYAML = cy
		}
		scaler.TriggerAuth = triggerAuth
	}

	return scaler
}

// toQoveryAutoscalingRequest converts the Terraform `autoscaling` object directly
// into the API request model. Used by the application resource, which builds the
// qovery.ApplicationRequest in the legacy client layer rather than via the DDD
// repository. Returns (nil, nil) when the block is absent.
func toQoveryAutoscalingRequest(o types.Object) (*qovery.AutoscalingPolicyRequest, error) {
	policy := toQoveryAutoscaling(o)
	if policy == nil {
		return nil, nil
	}

	req, err := autoscaling.ToQoveryRequest(*policy)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// fromAutoscaling converts the shared domain model into a Terraform object.
// Returns a null object when the policy is absent so an unset block stays unset.
// prior is the value the object replaces (the plan on Create and Update, the
// state on Read, null on import and in data sources): it supplies the order of
// the scalers, see orderScalersLikePrior, and the trigger authentication names,
// see triggerAuthNames.
func fromAutoscaling(p *autoscaling.AutoscalingPolicy, prior types.Object) types.Object {
	if p == nil {
		return types.ObjectNull(autoscalingAttrTypes())
	}

	triggerAuthAttrTypes := autoscalingTriggerAuthAttrTypes()
	scalerAttrTypes := autoscalingScalerAttrTypes()
	var priorScalers []autoscaling.Scaler
	if priorPolicy := toQoveryAutoscaling(prior); priorPolicy != nil {
		priorScalers = priorPolicy.Scalers
	}
	actual := orderScalersLikePrior(p.Scalers, priorScalers)
	names := triggerAuthNames(actual, priorScalers)

	scalerElems := make([]attr.Value, 0, len(actual))
	for i, s := range actual {
		configJSON := jsontypes.NewNormalizedNull()
		if s.Config.ConfigJSON != "" {
			configJSON = jsontypes.NewNormalizedValue(s.Config.ConfigJSON)
		}

		configYAML := types.StringNull()
		if s.Config.ConfigYAML != "" {
			configYAML = types.StringValue(s.Config.ConfigYAML)
		}

		triggerAuth := types.ObjectNull(triggerAuthAttrTypes)
		if s.TriggerAuth != nil {
			triggerAuth = types.ObjectValueMust(triggerAuthAttrTypes, map[string]attr.Value{
				"name":        types.StringValue(names[i]),
				"config_yaml": FromStringPointer(s.TriggerAuth.ConfigYAML),
			})
		}

		id := types.StringNull()
		if s.ID != "" {
			id = types.StringValue(s.ID)
		}

		scalerElems = append(scalerElems, types.ObjectValueMust(scalerAttrTypes, map[string]attr.Value{
			"id":                     id,
			"scaler_type":            types.StringValue(s.ScalerType),
			"enabled":                types.BoolValue(s.Enabled),
			"role":                   types.StringValue(string(s.Role)),
			"config_json":            configJSON,
			"config_yaml":            configYAML,
			"trigger_authentication": triggerAuth,
		}))
	}

	scalers := types.ListValueMust(types.ObjectType{AttrTypes: scalerAttrTypes}, scalerElems)

	return types.ObjectValueMust(autoscalingAttrTypes(), map[string]attr.Value{
		"polling_interval_seconds":        FromInt32Pointer(p.PollingIntervalSeconds),
		"cooldown_period_seconds":         FromInt32Pointer(p.CooldownPeriodSeconds),
		"initial_cooldown_period_seconds": FromInt32Pointer(p.InitialCooldownPeriodSeconds),
		"idle_replica_count":              FromInt32Pointer(p.IdleReplicaCount),
		"advanced_yaml":                   FromStringPointer(p.AdvancedYAML),
		"scalers":                         scalers,
	})
}

// autoscalingAttributesAddedInV2 are the autoscaling attributes that schema
// version 1 does not have: upgradeAutoscalingFromV1 sets them to null.
var autoscalingAttributesAddedInV2 = []string{"initial_cooldown_period_seconds", "idle_replica_count", "advanced_yaml"}

// withAutoscalingSchemaV1 returns s with the autoscaling attribute of schema
// version 1 of the application and container resources, whose scalers were a
// set without id. It is the prior schema of their state upgraders.
func withAutoscalingSchemaV1(s schema.Schema) schema.Schema {
	autoscalingAttr := autoscalingResourceSchema()
	scalers := autoscalingAttr.Attributes["scalers"].(schema.ListNestedAttribute)
	scalerAttrs := maps.Clone(scalers.NestedObject.Attributes)
	delete(scalerAttrs, "id")
	autoscalingAttr.Attributes = maps.Clone(autoscalingAttr.Attributes)
	for _, name := range autoscalingAttributesAddedInV2 {
		delete(autoscalingAttr.Attributes, name)
	}
	autoscalingAttr.Attributes["scalers"] = schema.SetNestedAttribute{
		Required:     true,
		NestedObject: schema.NestedAttributeObject{Attributes: scalerAttrs},
	}

	s.Attributes = maps.Clone(s.Attributes)
	s.Attributes["autoscaling"] = autoscalingAttr
	s.Version = 1
	return s
}

// upgradeAutoscalingFromV1 turns the set of scalers of a version 1 autoscaling
// object into a list of scalers with a null id. The next refresh fills the ids,
// matching the scalers by their attributes, see orderScalersLikePrior.
func upgradeAutoscalingFromV1(o types.Object) (types.Object, diag.Diagnostics) {
	if o.IsNull() || o.IsUnknown() {
		return types.ObjectNull(autoscalingAttrTypes()), nil
	}

	var diags diag.Diagnostics
	attrs := o.Attributes()
	scalerType := types.ObjectType{AttrTypes: autoscalingScalerAttrTypes()}

	scalers := types.ListNull(scalerType)
	if set, ok := attrs["scalers"].(types.Set); ok && !set.IsNull() && !set.IsUnknown() {
		elems := make([]attr.Value, 0, len(set.Elements()))
		for _, elem := range set.Elements() {
			obj, ok := elem.(types.Object)
			if !ok || obj.IsNull() || obj.IsUnknown() {
				continue
			}
			scalerAttrs := maps.Clone(obj.Attributes())
			scalerAttrs["id"] = types.StringNull()
			scaler, d := types.ObjectValue(scalerType.AttrTypes, scalerAttrs)
			diags.Append(d...)
			elems = append(elems, scaler)
		}
		if diags.HasError() {
			return types.ObjectNull(autoscalingAttrTypes()), diags
		}
		list, d := types.ListValue(scalerType, elems)
		diags.Append(d...)
		scalers = list
	}

	upgraded, d := types.ObjectValue(autoscalingAttrTypes(), map[string]attr.Value{
		"polling_interval_seconds":        attrs["polling_interval_seconds"],
		"cooldown_period_seconds":         attrs["cooldown_period_seconds"],
		"initial_cooldown_period_seconds": types.Int64Null(),
		"idle_replica_count":              types.Int64Null(),
		"advanced_yaml":                   types.StringNull(),
		"scalers":                         scalers,
	})
	diags.Append(d...)
	return upgraded, diags
}

// triggerAuthNames returns, for each scaler of actual, the name of its trigger
// authentication to store in the state ("" for a scaler without one).
//
// The API ignores the name sent for an inline trigger authentication: q-core
// generates one from the service ID and a timestamp (`<id prefix>-<timestamp>-auth`)
// and keeps it on every edit. The configured name is therefore write-only, so it
// comes from the prior scaler that matches on every other attribute. A scaler
// with no such match (after an import, in a data source, or after a change made
// outside Terraform) keeps the name the API returns.
func triggerAuthNames(actual []autoscaling.Scaler, candidates []autoscaling.Scaler) []string {
	used := make([]bool, len(candidates))

	names := make([]string, len(actual))
	for i, s := range actual {
		if s.TriggerAuth == nil {
			continue
		}
		names[i] = s.TriggerAuth.Name
		for j, c := range candidates {
			if used[j] || !sameScalerExceptTriggerAuthName(s, c) {
				continue
			}
			used[j] = true
			names[i] = c.TriggerAuth.Name
			break
		}
	}
	return names
}

// orderScalersLikePrior returns actual, the scalers the API returns in creation
// order, in the order of prior so that a refresh does not plan a reorder of the
// list. Prior scalers are matched as described in matchScalers, by id first:
// the scalers the API has not given an id to yet (Create, a scaler added on
// Update, a state upgraded from a set) are matched by their attributes. The
// scalers no prior one matches follow in API order.
func orderScalersLikePrior(actual, prior []autoscaling.Scaler) []autoscaling.Scaler {
	if len(prior) == 0 {
		return actual
	}

	used := make([]bool, len(actual))
	ordered := make([]autoscaling.Scaler, 0, len(actual))
	for _, j := range matchScalers(prior, actual, true, false) {
		if j != -1 {
			used[j] = true
			ordered = append(ordered, actual[j])
		}
	}
	for j, a := range actual {
		if !used[j] {
			ordered = append(ordered, a)
		}
	}
	return ordered
}

// matchScalers returns, for each scaler of targets, the index of the scaler of
// candidates it stands for, or -1. Each candidate is used once, in successive
// passes: same id (when byID), same attributes except id and trigger
// authentication name, same scaler_type and role, then (when byPosition) the
// remaining targets and candidates paired in order.
func matchScalers(targets, candidates []autoscaling.Scaler, byID, byPosition bool) []int {
	matched := make([]int, len(targets))
	for i := range matched {
		matched[i] = -1
	}
	used := make([]bool, len(candidates))

	pass := func(same func(t, c autoscaling.Scaler) bool) {
		for i, t := range targets {
			if matched[i] != -1 {
				continue
			}
			for j, c := range candidates {
				if !used[j] && same(t, c) {
					used[j] = true
					matched[i] = j
					break
				}
			}
		}
	}

	if byID {
		pass(func(t, c autoscaling.Scaler) bool { return t.ID != "" && t.ID == c.ID })
	}
	pass(sameScalerExceptIDAndTriggerAuthName)
	pass(func(t, c autoscaling.Scaler) bool { return t.ScalerType == c.ScalerType && t.Role == c.Role })
	if byPosition {
		pass(func(_, _ autoscaling.Scaler) bool { return true })
	}
	return matched
}

// scalerIDsFromStateModifier plans the id of each scaler from the state scaler
// it stands for (see matchScalers), not from the state scaler at the same index:
// the config order can differ from the state order (a state upgraded from a set,
// scalers reordered, added or removed in the config), and Qovery updates the
// scaler whose id is sent, editing its trigger authentication in place. A scaler
// that stands for none has an unknown id and is created.
type scalerIDsFromStateModifier struct{}

func (m scalerIDsFromStateModifier) Description(_ context.Context) string {
	return "Plans the id of each scaler from the scaler in state it stands for."
}

func (m scalerIDsFromStateModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m scalerIDsFromStateModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	state := scalersFromList(req.StateValue)

	elems := req.PlanValue.Elements()
	planned := make([]autoscaling.Scaler, len(elems))
	allKnown := true
	for i, elem := range elems {
		obj, ok := elem.(types.Object)
		if !ok || obj.IsNull() || obj.IsUnknown() || !scalerKnownExceptID(ctx, obj) {
			allKnown = false
			break
		}
		planned[i] = scalerFromObject(obj)
	}

	// With an unknown attribute, the match can change once it is known: the ids
	// stay unknown until apply, where the plan is known.
	var match []int
	if allKnown {
		match = matchScalers(planned, state, false, true)
	}

	scalerAttrTypes := autoscalingScalerAttrTypes()
	newElems := make([]attr.Value, len(elems))
	for i, elem := range elems {
		obj, ok := elem.(types.Object)
		if !ok || obj.IsNull() || obj.IsUnknown() {
			newElems[i] = elem
			continue
		}
		id := types.StringUnknown()
		if match != nil && match[i] != -1 && state[match[i]].ID != "" {
			id = types.StringValue(state[match[i]].ID)
		}
		attrs := maps.Clone(obj.Attributes())
		attrs["id"] = id
		newObj, diags := types.ObjectValue(scalerAttrTypes, attrs)
		resp.Diagnostics.Append(diags...)
		newElems[i] = newObj
	}
	if resp.Diagnostics.HasError() {
		return
	}

	planValue, diags := types.ListValue(types.ObjectType{AttrTypes: scalerAttrTypes}, newElems)
	resp.Diagnostics.Append(diags...)
	resp.PlanValue = planValue
}

// scalerKnownExceptID reports whether every attribute of a planned scaler but
// its id is known.
func scalerKnownExceptID(ctx context.Context, obj types.Object) bool {
	for name, v := range obj.Attributes() {
		if name == "id" {
			continue
		}
		tv, err := v.ToTerraformValue(ctx)
		if err != nil || !tv.IsFullyKnown() {
			return false
		}
	}
	return true
}

// sameScalerExceptIDAndTriggerAuthName reports whether a and b are equal on every
// attribute but their id and the name of their trigger authentication.
func sameScalerExceptIDAndTriggerAuthName(a, b autoscaling.Scaler) bool {
	if (a.TriggerAuth == nil) != (b.TriggerAuth == nil) {
		return false
	}
	if a.TriggerAuth != nil && ptrStringValue(a.TriggerAuth.ConfigYAML) != ptrStringValue(b.TriggerAuth.ConfigYAML) {
		return false
	}
	return a.ScalerType == b.ScalerType &&
		a.Role == b.Role &&
		a.Enabled == b.Enabled &&
		a.Config.ConfigYAML == b.Config.ConfigYAML &&
		sameJSON(a.Config.ConfigJSON, b.Config.ConfigJSON)
}

// sameScalerExceptTriggerAuthName reports whether a and b both carry a trigger
// authentication and are equal on every attribute but its name. config_json is
// compared as JSON, since the API returns it re-encoded.
func sameScalerExceptTriggerAuthName(a, b autoscaling.Scaler) bool {
	if a.TriggerAuth == nil || b.TriggerAuth == nil {
		return false
	}
	return sameScalerExceptIDAndTriggerAuthName(a, b)
}

// sameJSON reports whether a and b encode the same JSON value. Two empty strings
// are equal; an empty and a non-empty string, or an invalid document, are not.
func sameJSON(a, b string) bool {
	if a == b {
		return true
	}
	var av, bv any
	if json.Unmarshal([]byte(a), &av) != nil || json.Unmarshal([]byte(b), &bv) != nil {
		return false
	}
	return reflect.DeepEqual(av, bv)
}

// fromAutoscalingResponse converts the API response model directly into a
// Terraform object. Used by the application resource (legacy client layer).
// A serialization error (re-marshalling an already-decoded config_json map,
// which cannot realistically fail) degrades to a null object.
func fromAutoscalingResponse(res *qovery.AutoscalingPolicyResponse, prior types.Object) types.Object {
	policy, err := autoscaling.FromQoveryResponse(res)
	if err != nil {
		return types.ObjectNull(autoscalingAttrTypes())
	}
	return fromAutoscaling(policy, prior)
}

// validateAutoscalingPlan enforces, at plan time, the KEDA constraints the
// backend would otherwise only reject *after* mutating the service (leaving the
// resource partially updated / untracked). It is meant to be called from a
// resource's ModifyPlan for the application and container resources, which share
// the `min_running_instances` / `max_running_instances` / `autoscaling` schema.
//
// It enforces two rules:
//  1. When the KEDA `autoscaling` block is set, min_running_instances must be
//     strictly less than max_running_instances.
//  2. A KEDA idle_replica_count must be strictly less than min_running_instances.
//
// A direct HPA -> KEDA switch is allowed: the engine deletes the native HPA
// before it creates the ScaledObject (QOV-2561, QOV-2564).
func validateAutoscalingPlan(ctx context.Context, plan tfsdk.Plan, diags *diag.Diagnostics) {
	// No plan on destroy.
	if plan.Raw.IsNull() {
		return
	}

	var planAutoscaling types.Object
	diags.Append(plan.GetAttribute(ctx, path.Root("autoscaling"), &planAutoscaling)...)
	if diags.HasError() {
		return
	}
	hasPlanAutoscaling := !planAutoscaling.IsNull() && !planAutoscaling.IsUnknown()
	if !hasPlanAutoscaling {
		return
	}

	var planMin, planMax types.Int64
	diags.Append(plan.GetAttribute(ctx, path.Root("min_running_instances"), &planMin)...)
	diags.Append(plan.GetAttribute(ctx, path.Root("max_running_instances"), &planMax)...)
	if diags.HasError() {
		return
	}

	// Rule 1: min < max when KEDA is set.
	if !planMin.IsNull() && !planMin.IsUnknown() && !planMax.IsNull() && !planMax.IsUnknown() {
		if planMin.ValueInt64() >= planMax.ValueInt64() {
			diags.AddAttributeError(
				path.Root("min_running_instances"),
				"Invalid Running Instances Range",
				fmt.Sprintf(
					"When a KEDA `autoscaling` block is set, min_running_instances must be strictly less than max_running_instances (got min=%d max=%d). Set max_running_instances explicitly, greater than min_running_instances.",
					planMin.ValueInt64(), planMax.ValueInt64(),
				),
			)
		}
	}

	// Rule 2: an idle replica count must be less than min_running_instances.
	idle, ok := planAutoscaling.Attributes()["idle_replica_count"].(types.Int64)
	if ok && !idle.IsNull() && !idle.IsUnknown() && !planMin.IsNull() && !planMin.IsUnknown() &&
		idle.ValueInt64() >= planMin.ValueInt64() {
		diags.AddAttributeError(
			path.Root("autoscaling").AtName("idle_replica_count"),
			"Invalid Idle Replica Count",
			fmt.Sprintf(
				"autoscaling.idle_replica_count (%d) must be strictly less than min_running_instances (%d). Scaling to 0 when idle requires min_running_instances >= 1.",
				idle.ValueInt64(), planMin.ValueInt64(),
			),
		)
	}
}

func int64ObjectAttrToInt32Pointer(v attr.Value) *int32 {
	i, ok := v.(types.Int64)
	if !ok {
		return nil
	}
	return ToInt32Pointer(i)
}

func objectAttrToString(v attr.Value) string {
	s, ok := v.(types.String)
	if !ok || s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

func objectAttrToStringPointer(v attr.Value) *string {
	s, ok := v.(types.String)
	if !ok {
		return nil
	}
	return ToStringPointer(s)
}

func objectAttrToBool(v attr.Value) bool {
	b, ok := v.(types.Bool)
	if !ok || b.IsNull() || b.IsUnknown() {
		return false
	}
	return b.ValueBool()
}
