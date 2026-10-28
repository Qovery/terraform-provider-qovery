package advanced_settings

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
