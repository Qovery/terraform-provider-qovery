//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"encoding/json"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The pinned qovery-client-go does not model the per node pool spot_enabled flag yet, so the
// provider round-trips it through the generated AdditionalProperties maps. These tests pin the
// wire format down: they must keep passing once the client is regenerated and the compat
// helpers switch to the typed field.

func TestKarpenterSpotCompat_MarshalsSpotEnabledInsideEachOverride(t *testing.T) {
	t.Parallel()

	stable := qovery.KarpenterStableNodePoolOverride{}
	SetStableNodePoolSpotEnabled(&stable, false)

	defaultOverride := qovery.KarpenterDefaultNodePoolOverride{}
	SetDefaultNodePoolSpotEnabled(&defaultOverride, true)

	cronjob := qovery.KarpenterCronjobNodePoolOverride{}
	SetCronjobNodePoolSpotEnabled(&cronjob, true)

	nodePool := qovery.KarpenterNodePool{
		Requirements: []qovery.KarpenterNodePoolRequirement{{
			Key:      qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH,
			Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN,
			Values:   []string{"AMD64"},
		}},
		StableOverride:  &stable,
		DefaultOverride: &defaultOverride,
		CronjobOverride: &cronjob,
	}

	raw, err := json.Marshal(nodePool)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	assert.Equal(t, map[string]any{"spot_enabled": false}, decoded["stable_override"])
	assert.Equal(t, map[string]any{"spot_enabled": true}, decoded["default_override"])
	assert.Equal(t, map[string]any{"spot_enabled": true}, decoded["cronjob_override"])
}

func TestKarpenterSpotCompat_ReadsSpotEnabledFromApiResponse(t *testing.T) {
	t.Parallel()

	// Shape of the response the API returns once per node pool spot is deployed: every
	// override carries spot_enabled, next to the fields the client already models.
	const payload = `{
		"spot_enabled": true,
		"disk_size_in_gib": 50,
		"default_service_architecture": "AMD64",
		"qovery_node_pools": {
			"requirements": [{"key": "Arch", "operator": "In", "values": ["AMD64"]}],
			"stable_override": {
				"spot_enabled": false,
				"limits": {"enabled": true, "max_cpu_in_vcpu": 10, "max_memory_in_gibibytes": 20, "max_gpu": 0}
			},
			"default_override": {"spot_enabled": true},
			"cronjob_override": {"spot_enabled": true}
		}
	}`

	var parameters qovery.ClusterFeatureKarpenterParameters
	require.NoError(t, json.Unmarshal([]byte(payload), &parameters))

	nodePools := parameters.QoveryNodePools
	require.NotNil(t, nodePools.StableOverride)
	require.NotNil(t, nodePools.DefaultOverride)
	require.NotNil(t, nodePools.CronjobOverride)

	assert.Equal(t, false, *GetStableNodePoolSpotEnabled(nodePools.StableOverride))
	assert.Equal(t, true, *GetDefaultNodePoolSpotEnabled(nodePools.DefaultOverride))
	assert.Equal(t, true, *GetCronjobNodePoolSpotEnabled(nodePools.CronjobOverride))

	// The fields the client already models keep being parsed into their typed fields.
	require.NotNil(t, nodePools.StableOverride.Limits)
	assert.Equal(t, int32(10), nodePools.StableOverride.Limits.MaxCpuInVcpu)
}

func TestKarpenterSpotCompat_AbsentSpotEnabledReadsAsNil(t *testing.T) {
	t.Parallel()

	// An override without spot_enabled is how every cluster looked before the migration, and
	// the absence is meaningful: that node pool falls back to the global flag.
	const payload = `{
		"requirements": [{"key": "Arch", "operator": "In", "values": ["AMD64"]}],
		"stable_override": {"limits": {"enabled": true, "max_cpu_in_vcpu": 10, "max_memory_in_gibibytes": 20, "max_gpu": 0}},
		"default_override": {},
		"cronjob_override": {}
	}`

	var nodePools qovery.KarpenterNodePool
	require.NoError(t, json.Unmarshal([]byte(payload), &nodePools))

	assert.Nil(t, GetStableNodePoolSpotEnabled(nodePools.StableOverride))
	assert.Nil(t, GetDefaultNodePoolSpotEnabled(nodePools.DefaultOverride))
	assert.Nil(t, GetCronjobNodePoolSpotEnabled(nodePools.CronjobOverride))
	assert.Nil(t, GetStableNodePoolSpotEnabled(nil))
	assert.Nil(t, GetDefaultNodePoolSpotEnabled(nil))
	assert.Nil(t, GetCronjobNodePoolSpotEnabled(nil))
}

func TestKarpenterSpotCompat_SettersDoNotDropExistingProperties(t *testing.T) {
	t.Parallel()

	// Anything else the API sent back on the override must survive the round-trip.
	stable := qovery.KarpenterStableNodePoolOverride{
		AdditionalProperties: map[string]interface{}{"some_future_field": "kept"},
	}
	SetStableNodePoolSpotEnabled(&stable, true)

	raw, err := json.Marshal(stable)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, map[string]any{"spot_enabled": true, "some_future_field": "kept"}, decoded)
}
