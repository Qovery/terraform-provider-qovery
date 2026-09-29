package advanced_settings

import (
	"github.com/qovery/qovery-client-go"
)

const (
	BuildTimeoutMaxSecKey         = "build.timeout_max_sec"
	BuildCpuMaxInMilliKey         = "build.cpu_max_in_milli"
	BuildRamMaxInGibKey           = "build.ram_max_in_gib"
	BuildEphemeralStorageInGibKey = "build.ephemeral_storage_in_gib"
	BuildDisableBuildkitCacheKey  = "build.disable_buildkit_cache"
	BuildSkipGitSubmodulesKey     = "build.skip_git_submodules"
)

// BuildSettingsKeys lists the advanced settings keys that mirror the typed build settings.
// The API keeps them in sync with the build settings of the service in both directions.
var BuildSettingsKeys = []string{
	BuildTimeoutMaxSecKey,
	BuildCpuMaxInMilliKey,
	BuildRamMaxInGibKey,
	BuildEphemeralStorageInGibKey,
	BuildDisableBuildkitCacheKey,
	BuildSkipGitSubmodulesKey,
}

// BuildSettingsFromAdvancedSettings extracts the build settings from the build.* keys of a service's
// advanced settings. It returns nil when none of those keys is present. A missing or malformed key
// falls back to the API default.
func BuildSettingsFromAdvancedSettings(settings map[string]any) *qovery.BuildSettings {
	found := false
	for _, key := range BuildSettingsKeys {
		if _, ok := settings[key]; ok {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	result := qovery.NewBuildSettings()
	if v, ok := int32Setting(settings, BuildTimeoutMaxSecKey); ok {
		result.SetTimeoutMaxSec(v)
	}
	if v, ok := int32Setting(settings, BuildCpuMaxInMilliKey); ok {
		result.SetCpuMaxInMilli(v)
	}
	if v, ok := int32Setting(settings, BuildRamMaxInGibKey); ok {
		result.SetRamMaxInGib(v)
	}
	if v, ok := int32Setting(settings, BuildEphemeralStorageInGibKey); ok {
		result.SetEphemeralStorageInGib(v)
	} else {
		result.SetEphemeralStorageInGibNil()
	}
	if v, ok := boolSetting(settings, BuildDisableBuildkitCacheKey); ok {
		result.SetDisableBuildkitCache(v)
	}
	if v, ok := boolSetting(settings, BuildSkipGitSubmodulesKey); ok {
		result.SetSkipGitSubmodules(v)
	}
	return result
}

func int32Setting(settings map[string]any, key string) (int32, bool) {
	v, ok := normalizeJSONValue(settings[key]).(float64)
	if !ok {
		return 0, false
	}
	return int32(v), true
}

func boolSetting(settings map[string]any, key string) (bool, bool) {
	v, ok := normalizeJSONValue(settings[key]).(bool)
	return v, ok
}
