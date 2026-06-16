package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
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
		"polling_interval_seconds": types.Int64Type,
		"cooldown_period_seconds":  types.Int64Type,
		"scalers":                  types.SetType{ElemType: types.ObjectType{AttrTypes: autoscalingScalerAttrTypes()}},
	}
}

const autoscalingDescription = "Event-driven autoscaling (KEDA) configuration. " +
	"KEDA is additive to the CPU/memory HPA (min/max_running_instances) and unlocks " +
	"scale-to-zero (min_running_instances = 0). Requires KEDA to be enabled on the cluster."

// autoscalingResourceSchema returns the resource schema attribute for the
// `autoscaling` block, shared by the application and container resources.
func autoscalingResourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:         autoscalingDescription,
		MarkdownDescription: autoscalingDescription,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"polling_interval_seconds": schema.Int64Attribute{
				Description: "Interval in seconds between each KEDA polling of the scalers.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cooldown_period_seconds": schema.Int64Attribute{
				Description: "Period in seconds to wait after the last trigger before scaling back down.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"scalers": schema.SetNestedAttribute{
				Description: "List of KEDA scalers driving the autoscaling. At least one scaler is required.",
				Required:    true,
				Validators: []validator.Set{
					validators.ScalerConfigExactlyOneValidator{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"scaler_type": schema.StringAttribute{
							Description: "Type of the KEDA scaler (e.g. cpu, memory, prometheus, cron).",
							Required:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the scaler is enabled. Defaults to true.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
						"role": schema.StringAttribute{
							Description: "Role of the scaler: PRIMARY or SAFETY.",
							Required:    true,
							Validators: []validator.String{
								validators.NewStringEnumValidator([]string{string(autoscaling.RolePrimary), string(autoscaling.RoleSafety)}),
							},
						},
						"config_json": schema.StringAttribute{
							Description: "Scaler configuration as JSON. Mutually exclusive with config_yaml.",
							Optional:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
						"config_yaml": schema.StringAttribute{
							Description: "Scaler configuration as raw YAML. Mutually exclusive with config_json.",
							Optional:    true,
						},
						"trigger_authentication": schema.SingleNestedAttribute{
							Description: "Inline KEDA TriggerAuthentication for this scaler.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Name of the trigger authentication.",
									Required:    true,
								},
								"config_yaml": schema.StringAttribute{
									Description: "Raw KEDA TriggerAuthentication YAML configuration.",
									Optional:    true,
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
		Description:         autoscalingDescription,
		MarkdownDescription: autoscalingDescription,
		Computed:            true,
		Attributes: map[string]dsschema.Attribute{
			"polling_interval_seconds": dsschema.Int64Attribute{
				Description: "Interval in seconds between each KEDA polling of the scalers.",
				Computed:    true,
			},
			"cooldown_period_seconds": dsschema.Int64Attribute{
				Description: "Period in seconds to wait after the last trigger before scaling back down.",
				Computed:    true,
			},
			"scalers": dsschema.SetNestedAttribute{
				Description: "List of KEDA scalers driving the autoscaling.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"scaler_type": dsschema.StringAttribute{
							Description: "Type of the KEDA scaler.",
							Computed:    true,
						},
						"enabled": dsschema.BoolAttribute{
							Description: "Whether the scaler is enabled.",
							Computed:    true,
						},
						"role": dsschema.StringAttribute{
							Description: "Role of the scaler: PRIMARY or SAFETY.",
							Computed:    true,
						},
						"config_json": dsschema.StringAttribute{
							Description: "Scaler configuration as JSON.",
							Computed:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
						"config_yaml": dsschema.StringAttribute{
							Description: "Scaler configuration as raw YAML.",
							Computed:    true,
						},
						"trigger_authentication": dsschema.SingleNestedAttribute{
							Description: "Inline KEDA TriggerAuthentication for this scaler.",
							Computed:    true,
							Attributes: map[string]dsschema.Attribute{
								"name": dsschema.StringAttribute{
									Description: "Name of the trigger authentication.",
									Computed:    true,
								},
								"config_yaml": dsschema.StringAttribute{
									Description: "Raw KEDA TriggerAuthentication YAML configuration.",
									Computed:    true,
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
		PollingIntervalSeconds: int64ObjectAttrToInt32Pointer(attrs["polling_interval_seconds"]),
		CooldownPeriodSeconds:  int64ObjectAttrToInt32Pointer(attrs["cooldown_period_seconds"]),
	}

	scalersSet, ok := attrs["scalers"].(types.Set)
	if !ok || scalersSet.IsNull() || scalersSet.IsUnknown() {
		return policy
	}

	scalers := make([]autoscaling.Scaler, 0, len(scalersSet.Elements()))
	for _, elem := range scalersSet.Elements() {
		obj, ok := elem.(types.Object)
		if !ok || obj.IsNull() || obj.IsUnknown() {
			continue
		}
		sAttrs := obj.Attributes()

		scaler := autoscaling.Scaler{
			ScalerType: objectAttrToString(sAttrs["scaler_type"]),
			Enabled:    objectAttrToBool(sAttrs["enabled"]),
			Role:       autoscaling.Role(objectAttrToString(sAttrs["role"])),
		}

		if cj, ok := sAttrs["config_json"].(jsontypes.Normalized); ok && !cj.IsNull() && !cj.IsUnknown() {
			scaler.Config.ConfigJSON = cj.ValueString()
		}
		if cy := objectAttrToString(sAttrs["config_yaml"]); cy != "" {
			scaler.Config.ConfigYAML = cy
		}

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

		scalers = append(scalers, scaler)
	}
	policy.Scalers = scalers

	return policy
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
func fromAutoscaling(p *autoscaling.AutoscalingPolicy) types.Object {
	if p == nil {
		return types.ObjectNull(autoscalingAttrTypes())
	}

	scalerElems := make([]attr.Value, 0, len(p.Scalers))
	for _, s := range p.Scalers {
		configJSON := jsontypes.NewNormalizedNull()
		if s.Config.ConfigJSON != "" {
			configJSON = jsontypes.NewNormalizedValue(s.Config.ConfigJSON)
		}

		configYAML := types.StringNull()
		if s.Config.ConfigYAML != "" {
			configYAML = types.StringValue(s.Config.ConfigYAML)
		}

		triggerAuth := types.ObjectNull(autoscalingTriggerAuthAttrTypes())
		if s.TriggerAuth != nil {
			triggerAuth = types.ObjectValueMust(autoscalingTriggerAuthAttrTypes(), map[string]attr.Value{
				"name":        types.StringValue(s.TriggerAuth.Name),
				"config_yaml": fromStringPointerNull(s.TriggerAuth.ConfigYAML),
			})
		}

		scalerElems = append(scalerElems, types.ObjectValueMust(autoscalingScalerAttrTypes(), map[string]attr.Value{
			"scaler_type":            types.StringValue(s.ScalerType),
			"enabled":                types.BoolValue(s.Enabled),
			"role":                   types.StringValue(string(s.Role)),
			"config_json":            configJSON,
			"config_yaml":            configYAML,
			"trigger_authentication": triggerAuth,
		}))
	}

	scalers := types.SetValueMust(types.ObjectType{AttrTypes: autoscalingScalerAttrTypes()}, scalerElems)

	return types.ObjectValueMust(autoscalingAttrTypes(), map[string]attr.Value{
		"polling_interval_seconds": int32PointerToInt64Value(p.PollingIntervalSeconds),
		"cooldown_period_seconds":  int32PointerToInt64Value(p.CooldownPeriodSeconds),
		"scalers":                  scalers,
	})
}

// fromAutoscalingResponse converts the API response model directly into a
// Terraform object. Used by the application resource (legacy client layer).
// A serialization error (re-marshalling an already-decoded config_json map,
// which cannot realistically fail) degrades to a null object.
func fromAutoscalingResponse(res *qovery.AutoscalingPolicyResponse) types.Object {
	policy, err := autoscaling.FromQoveryResponse(res)
	if err != nil {
		return types.ObjectNull(autoscalingAttrTypes())
	}
	return fromAutoscaling(policy)
}

func int64ObjectAttrToInt32Pointer(v attr.Value) *int32 {
	i, ok := v.(types.Int64)
	if !ok || i.IsNull() || i.IsUnknown() {
		return nil
	}
	val := int32(i.ValueInt64())
	return &val
}

func int32PointerToInt64Value(v *int32) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
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
	if !ok || s.IsNull() || s.IsUnknown() {
		return nil
	}
	val := s.ValueString()
	return &val
}

func objectAttrToBool(v attr.Value) bool {
	b, ok := v.(types.Bool)
	if !ok || b.IsNull() || b.IsUnknown() {
		return false
	}
	return b.ValueBool()
}

func fromStringPointerNull(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}
