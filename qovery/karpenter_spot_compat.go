package qovery

import (
	"github.com/qovery/qovery-client-go"
)

// Accessors for the per-node-pool `spot_enabled` flag of the Karpenter stable, default and
// cronjob node pool overrides.
//
// The generated field is a `NullableBool`, which has three states where the provider only cares
// about two: a concrete value, versus none. In a response, absent and explicitly-null both mean
// the API omitted a value equal to the global flag it derived, so both collapse to a nil *bool
// here and the read path resolves nil to that global. In a request, an unset field is left out of
// the body — the generated ToMap only emits the key when the field is set — and the API would
// then hand the pool its global flag. The conversion layer therefore always calls a setter.
//
// The three override types are distinct generated structs with identical accessors, so these
// wrappers exist to keep the conversion layer uniform across them.

// karpenterSpotValue collapses a generated GetSpotEnabledOk() result onto a plain *bool. The
// value is copied so callers cannot write through the pointer into the API model.
func karpenterSpotValue(value *bool, isSet bool) *bool {
	if !isSet || value == nil {
		return nil
	}
	enabled := *value
	return &enabled
}

// GetStableNodePoolSpotEnabled returns the stable node pool `spot_enabled` flag, or nil when the
// pool carries no explicit value.
func GetStableNodePoolSpotEnabled(o *qovery.KarpenterStableNodePoolOverride) *bool {
	return karpenterSpotValue(o.GetSpotEnabledOk())
}

// SetStableNodePoolSpotEnabled sets an explicit `spot_enabled` flag on the stable node pool
// override.
func SetStableNodePoolSpotEnabled(o *qovery.KarpenterStableNodePoolOverride, enabled bool) {
	if o == nil {
		return
	}
	o.SetSpotEnabled(enabled)
}

// GetDefaultNodePoolSpotEnabled returns the default node pool `spot_enabled` flag, or nil when
// the pool carries no explicit value.
func GetDefaultNodePoolSpotEnabled(o *qovery.KarpenterDefaultNodePoolOverride) *bool {
	return karpenterSpotValue(o.GetSpotEnabledOk())
}

// SetDefaultNodePoolSpotEnabled sets an explicit `spot_enabled` flag on the default node pool
// override.
func SetDefaultNodePoolSpotEnabled(o *qovery.KarpenterDefaultNodePoolOverride, enabled bool) {
	if o == nil {
		return
	}
	o.SetSpotEnabled(enabled)
}

// GetCronjobNodePoolSpotEnabled returns the cronjob node pool `spot_enabled` flag, or nil when
// the pool carries no explicit value.
func GetCronjobNodePoolSpotEnabled(o *qovery.KarpenterCronjobNodePoolOverride) *bool {
	return karpenterSpotValue(o.GetSpotEnabledOk())
}

// SetCronjobNodePoolSpotEnabled sets an explicit `spot_enabled` flag on the cronjob node pool
// override.
func SetCronjobNodePoolSpotEnabled(o *qovery.KarpenterCronjobNodePoolOverride, enabled bool) {
	if o == nil {
		return
	}
	o.SetSpotEnabled(enabled)
}
