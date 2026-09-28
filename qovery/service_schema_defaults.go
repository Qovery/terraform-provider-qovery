package qovery

import "github.com/hashicorp/terraform-plugin-framework/types"

// Values q-core stores when a service request omits the attribute. The service schemas use them
// as Defaults, so removing the attribute from the configuration plans the reset, and a value
// changed outside Terraform shows up in the plan (see "HCL Config Is the Source of Truth" in
// AGENTS.md).
const (
	applicationIconURIDefault = "app://qovery-console/application"
	containerIconURIDefault   = "app://qovery-console/container"
	helmIconURIDefault        = "app://qovery-console/helm"
	databaseIconURIDefault    = "app://qovery-console/database"

	// serviceAutoDeployDefault is q-core's ServiceDomain.DEFAULT_AUTO_DEPLOY. qovery_helm uses
	// helmAutoDeployDefault instead.
	serviceAutoDeployDefault = true
	// helmAutoDeployDefault is the value every 0.x release sent for a helm service that omitted
	// auto_deploy: the helm request field is a plain boolean, so an omitted value went out as
	// false.
	helmAutoDeployDefault = false

	serviceAutoPreviewDefault = false

	// serviceEphemeralStorageDefault means no ephemeral storage: the engine only sets an
	// ephemeral storage size above 0, so the platform default applies.
	serviceEphemeralStorageDefault int64 = 0

	jobRootPathDefault = "/"
)

// ephemeralStorageFromAPI reads ephemeral_storage on application, container and job. The API
// returns null when no ephemeral storage was ever set; that reads as the 0 default, which the
// engine also treats as none.
func ephemeralStorageFromAPI(v *int32) types.Int64 {
	if v == nil {
		return types.Int64Value(serviceEphemeralStorageDefault)
	}
	return FromInt32(*v)
}

// upgradeArgumentsFrom0x turns the empty arguments list that 0.x stored for every application
// and container without arguments into null. arguments was Optional + Computed in 0.x and is
// Optional only since 1.0, so keeping [] would plan a "[] -> null" change on every service that
// omits it.
func upgradeArgumentsFrom0x(arguments types.List) types.List {
	if !arguments.IsNull() && !arguments.IsUnknown() && len(arguments.Elements()) == 0 {
		return types.ListNull(types.StringType)
	}
	return arguments
}

// Sentences appended to the description of the attributes that keep their value when removed
// from the configuration. Each one is a documented exception to the config-is-source-of-truth
// rule and is listed in the 1.0 upgrade guide.
const (
	deploymentStageIDRemovalNote = " Removing the attribute keeps the service in its current stage: the Qovery API attaches every service to a deployment stage and cannot detach it."
	gitBranchRemovalNote         = " Removing the attribute keeps the current branch: an omitted branch means the repository's default branch, which is only known once the API resolves it. Changing the repository URL while the branch is omitted resolves the new repository's default branch."
	blueprintIDRemovalNote       = " It can only be set when the service is created: the Qovery API ignores later changes, so a change is rejected at plan time, and removing the attribute keeps the recorded value."
)
