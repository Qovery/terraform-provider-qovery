package qovery

import (
	"github.com/qovery/qovery-client-go"
)

// TODO(QOV-2158): temporary bridge for the per-node-pool `spot_enabled` flag.
//
// The Qovery API accepts and returns `spot_enabled` inside the Karpenter
// stable / default / cronjob node pool overrides (QOV-2156), but the pinned
// qovery-client-go release does not model the field yet: only
// KarpenterGpuNodePoolOverride has a typed SpotEnabled. The generated structs
// do round-trip unknown JSON keys through their AdditionalProperties map
// (MarshalJSON re-emits them, UnmarshalJSON collects them), so every read and
// write of the flag goes through the helpers below and nowhere else.
//
// Once qovery-client-go is regenerated from the OpenAPI change, this file is
// the ONLY place to update: UnmarshalJSON will then put the key in the typed
// `SpotEnabled *bool` field and delete it from AdditionalProperties, so the
// getters MUST be rewritten to read the typed field (keeping the
// AdditionalProperties lookup as a fallback is not enough — after regeneration
// it is always empty) and the setters MUST assign the typed field instead of
// writing into the map.
const karpenterSpotEnabledKey = "spot_enabled"

// karpenterSpotFromAdditionalProperties reads the `spot_enabled` flag out of a
// generated model's AdditionalProperties map. It returns nil when the API did
// not send the field, which is meaningful: an absent per-pool flag means the
// pool falls back to the deprecated global `spot_enabled`.
func karpenterSpotFromAdditionalProperties(props map[string]interface{}) *bool {
	if props == nil {
		return nil
	}
	raw, ok := props[karpenterSpotEnabledKey]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case bool:
		return &v
	case *bool:
		return v
	default:
		return nil
	}
}

// karpenterSpotIntoAdditionalProperties writes the `spot_enabled` flag into a
// generated model's AdditionalProperties map, allocating it when needed.
func karpenterSpotIntoAdditionalProperties(props map[string]interface{}, enabled bool) map[string]interface{} {
	if props == nil {
		props = make(map[string]interface{}, 1)
	}
	props[karpenterSpotEnabledKey] = enabled
	return props
}

// GetStableNodePoolSpotEnabled returns the stable node pool `spot_enabled`
// flag, or nil when the pool carries no explicit value.
func GetStableNodePoolSpotEnabled(o *qovery.KarpenterStableNodePoolOverride) *bool {
	if o == nil {
		return nil
	}
	return karpenterSpotFromAdditionalProperties(o.AdditionalProperties)
}

// SetStableNodePoolSpotEnabled sets an explicit `spot_enabled` flag on the
// stable node pool override.
func SetStableNodePoolSpotEnabled(o *qovery.KarpenterStableNodePoolOverride, enabled bool) {
	if o == nil {
		return
	}
	o.AdditionalProperties = karpenterSpotIntoAdditionalProperties(o.AdditionalProperties, enabled)
}

// GetDefaultNodePoolSpotEnabled returns the default node pool `spot_enabled`
// flag, or nil when the pool carries no explicit value.
func GetDefaultNodePoolSpotEnabled(o *qovery.KarpenterDefaultNodePoolOverride) *bool {
	if o == nil {
		return nil
	}
	return karpenterSpotFromAdditionalProperties(o.AdditionalProperties)
}

// SetDefaultNodePoolSpotEnabled sets an explicit `spot_enabled` flag on the
// default node pool override.
func SetDefaultNodePoolSpotEnabled(o *qovery.KarpenterDefaultNodePoolOverride, enabled bool) {
	if o == nil {
		return
	}
	o.AdditionalProperties = karpenterSpotIntoAdditionalProperties(o.AdditionalProperties, enabled)
}

// GetCronjobNodePoolSpotEnabled returns the cronjob node pool `spot_enabled`
// flag, or nil when the pool carries no explicit value.
func GetCronjobNodePoolSpotEnabled(o *qovery.KarpenterCronjobNodePoolOverride) *bool {
	if o == nil {
		return nil
	}
	return karpenterSpotFromAdditionalProperties(o.AdditionalProperties)
}

// SetCronjobNodePoolSpotEnabled sets an explicit `spot_enabled` flag on the
// cronjob node pool override.
func SetCronjobNodePoolSpotEnabled(o *qovery.KarpenterCronjobNodePoolOverride, enabled bool) {
	if o == nil {
		return
	}
	o.AdditionalProperties = karpenterSpotIntoAdditionalProperties(o.AdditionalProperties, enabled)
}
