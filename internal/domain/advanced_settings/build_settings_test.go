//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSettingsFromAdvancedSettings(t *testing.T) {
	t.Parallel()

	ephemeral := int32(10)

	testCases := []struct {
		TestName              string
		Settings              string
		ExpectNil             bool
		ExpectTimeoutMaxSec   int32
		ExpectCpuMaxInMilli   int32
		ExpectRamMaxInGib     int32
		ExpectEphemeral       *int32
		ExpectDisableBuildkit bool
		ExpectSkipSubmodules  bool
	}{
		{
			TestName:  "no_build_keys",
			Settings:  `{"network.ingress.proxy_body_size_mb":100}`,
			ExpectNil: true,
		},
		{
			TestName:              "all_build_keys",
			Settings:              `{"build.timeout_max_sec":3600,"build.cpu_max_in_milli":2000,"build.ram_max_in_gib":4,"build.ephemeral_storage_in_gib":10,"build.disable_buildkit_cache":true,"build.skip_git_submodules":true}`,
			ExpectTimeoutMaxSec:   3600,
			ExpectCpuMaxInMilli:   2000,
			ExpectRamMaxInGib:     4,
			ExpectEphemeral:       &ephemeral,
			ExpectDisableBuildkit: true,
			ExpectSkipSubmodules:  true,
		},
		{
			TestName:            "null_ephemeral_storage",
			Settings:            `{"build.timeout_max_sec":1800,"build.cpu_max_in_milli":4000,"build.ram_max_in_gib":8,"build.ephemeral_storage_in_gib":null,"build.disable_buildkit_cache":false,"build.skip_git_submodules":false}`,
			ExpectTimeoutMaxSec: 1800,
			ExpectCpuMaxInMilli: 4000,
			ExpectRamMaxInGib:   8,
		},
		{
			TestName:            "string_encoded_values",
			Settings:            `{"build.timeout_max_sec":"3600","build.disable_buildkit_cache":"true"}`,
			ExpectTimeoutMaxSec: 3600,
			ExpectCpuMaxInMilli: 4000,
			ExpectRamMaxInGib:   8,
			// missing keys fall back to the API defaults
			ExpectDisableBuildkit: true,
		},
		{
			TestName:            "out_of_range_value_falls_back_to_default",
			Settings:            `{"build.timeout_max_sec":3000000000}`,
			ExpectTimeoutMaxSec: 1800,
			ExpectCpuMaxInMilli: 4000,
			ExpectRamMaxInGib:   8,
		},
		{
			TestName:            "negative_value_falls_back_to_default",
			Settings:            `{"build.timeout_max_sec":-1}`,
			ExpectTimeoutMaxSec: 1800,
			ExpectCpuMaxInMilli: 4000,
			ExpectRamMaxInGib:   8,
		},
		{
			TestName:            "fractional_value_falls_back_to_default",
			Settings:            `{"build.timeout_max_sec":3600.5}`,
			ExpectTimeoutMaxSec: 1800,
			ExpectCpuMaxInMilli: 4000,
			ExpectRamMaxInGib:   8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			var settings map[string]any
			require.NoError(t, json.Unmarshal([]byte(tc.Settings), &settings))

			result := BuildSettingsFromAdvancedSettings(settings)
			if tc.ExpectNil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result)
			assert.Equal(t, tc.ExpectTimeoutMaxSec, result.GetTimeoutMaxSec())
			assert.Equal(t, tc.ExpectCpuMaxInMilli, result.GetCpuMaxInMilli())
			assert.Equal(t, tc.ExpectRamMaxInGib, result.GetRamMaxInGib())
			assert.True(t, result.EphemeralStorageInGib.IsSet())
			assert.Equal(t, tc.ExpectEphemeral, result.EphemeralStorageInGib.Get())
			assert.Equal(t, tc.ExpectDisableBuildkit, result.GetDisableBuildkitCache())
			assert.Equal(t, tc.ExpectSkipSubmodules, result.GetSkipGitSubmodules())
		})
	}
}
