package qovery

import "fmt"

// Descriptions shared by the resources and data sources that expose the same attribute, so its
// wording lives in one place. Each helper returns the sentence both describe; a resource appends
// its default and limits with the descriptions package helpers, and a data source uses the
// sentence alone.

// managingChangesGuideURL is the Registry guide that explains how the provider plans changes and
// lists the attributes that keep their value when removed.
const managingChangesGuideURL = "https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/managing-changes"

// Sentences appended to the description of the attributes that keep their value when removed
// from the configuration. Each one is a documented exception to the config-is-source-of-truth
// rule, listed in the managing-changes guide.
const (
	deploymentStageIDRemovalNote = " Removing it keeps the service in its current stage, because Qovery cannot detach a service from its stage."
	gitBranchRemovalNote         = " When omitted, Qovery uses the repository's default branch, and removing it keeps the current branch."
	blueprintIDRemovalNote       = " It can only be set at creation: changing it fails at plan time, and removing it keeps the recorded value."
)

// recreatesOnChange is appended to the description of the attributes whose change replaces the
// resource.
func recreatesOnChange(kind string) string {
	return " Changing it recreates the " + kind + "."
}

func idDescription(kind string) string {
	return fmt.Sprintf("ID of the %s.", kind)
}

func nameDescription(kind string) string {
	return fmt.Sprintf("Name of the %s.", kind)
}

func descriptionDescription(kind string) string {
	return fmt.Sprintf("Description of the %s.", kind)
}

const (
	environmentIDDescription     = "ID of the environment."
	organizationIDDescription    = "ID of the organization."
	deploymentStageIDDescription = "ID of the deployment stage of the service. Stages set the order in which the services of an environment deploy."
	isSkippedDescription         = "Whether environment-wide deployments skip the service. It stays in its deployment stage."
)

// advancedSettingsJSONDescription describes advanced_settings_json on a resource. The attribute
// is a documented exception to the config-is-source-of-truth rule: the QOV-2028 contract
// implemented by computeOverriddenSettings, explained in the managing-changes guide.
func advancedSettingsJSONDescription(defaultSettingsOperation string) string {
	return "Advanced settings to override, as a JSON string built with `jsonencode()`. " +
		"The [Qovery API documentation](https://api-doc.qovery.com/#tag/" + defaultSettingsOperation + ") lists them with their defaults. " +
		"Terraform manages only the keys you set, and removing a key keeps its current value: see [Advanced settings](" + managingChangesGuideURL + "#advanced-settings)."
}

func dataSourceAdvancedSettingsJSONDescription(kind string) string {
	return fmt.Sprintf("Advanced settings of the %s, as a JSON string.", kind)
}

func iconURIDescription(kind string) string {
	return fmt.Sprintf("Icon of the %s in the Qovery Console.", kind)
}

func autoPreviewDescription(kind string) string {
	return fmt.Sprintf("Whether Qovery creates a preview environment with the %s for each pull request.", kind)
}

func cpuDescription(kind string) string {
	return fmt.Sprintf("CPU of the %s, in millicores (1000 = 1 vCPU).", kind)
}

func memoryDescription(kind string) string {
	return fmt.Sprintf("Memory of the %s, in MB.", kind)
}

func ephemeralStorageDescription(kind string) string {
	return fmt.Sprintf("Ephemeral storage of the %s, in GiB. `0` sets none, so the platform default applies.", kind)
}

func minRunningInstancesDescription(kind string) string {
	return fmt.Sprintf("Minimum number of running instances of the %s.", kind)
}

func maxRunningInstancesDescription(kind string) string {
	return fmt.Sprintf("Maximum number of running instances of the %s.", kind)
}

const (
	entrypointDescription = "Command that replaces the `ENTRYPOINT` of the image."
	argumentsDescription  = "Arguments that replace the `CMD` of the image."
	// omittedSetsNoneNote is appended on resources to the lists whose omission means none.
	omittedSetsNoneNote = " Omitting it sets none."
)

const applicationAutoDeployDescription = "Whether Qovery redeploys the application on every new commit to its branch."

// Git repository and Docker build attributes of application, job, helm and terraform_service.
const (
	gitRepositoryURLDescription       = "URL of the git repository, for example `https://github.com/my-org/my-app.git`."
	gitRepositoryRootPathDescription  = "Directory of the repository to build from, for monorepos."
	gitRepositoryTokenIDDescription   = "ID of the `qovery_git_token` used to access a private repository."
	dockerfilePathDescription         = "Path of the Dockerfile, relative to `root_path`, for example `Dockerfile`."
	dockerTargetBuildStageDescription = "Stage of a multi-stage Dockerfile to build."
)

func externalHostDescription(kind string) string {
	return fmt.Sprintf("Public host of the %s. Set only when the %s has a publicly accessible port.", kind, kind)
}

func internalHostDescription(kind string) string {
	return fmt.Sprintf("Internal host of the %s, reachable from the other services of the environment.", kind)
}

// groupIDsDescription describes labels_group_ids and annotations_group_ids on a resource: the
// configuration manages the whole set. groupKind is "labels" or "annotations"; target names what
// the group applies to, for example "the application's pods".
func groupIDsDescription(groupKind, target string) string {
	return fmt.Sprintf("IDs of the %s groups applied to %s. Omitting it detaches every %s group.", groupKind, target, groupKind)
}

func dataSourceGroupIDsDescription(groupKind, kind string) string {
	return fmt.Sprintf("IDs of the %s groups attached to the %s.", groupKind, kind)
}

// variableListDescription holds the descriptions of a variable list attribute and of its nested
// attributes. Every service, environment and project, and their data sources, read them from
// variableListDescriptions.
type variableListDescription struct {
	List                  string
	ID                    string
	Key                   string
	Value                 string
	Description           string
	MountPath             string
	Reference             string
	SecretManagerAccessID string
}

// variableListDescriptions returns the descriptions of the variable list attribute named
// attribute on the resource kind owner. It panics on an unknown attribute, which Schema calls in
// every test and at provider start.
func variableListDescriptions(attribute, owner string) variableListDescription {
	variable := func(list, item string) variableListDescription {
		return variableListDescription{
			List:        fmt.Sprintf(list, owner),
			ID:          "ID of the " + item + ".",
			Key:         "Name of the " + item + ".",
			Value:       "Value of the " + item + ".",
			Description: "Description of the " + item + ".",
		}
	}
	alias := func(list, aliased string) variableListDescription {
		return variableListDescription{
			List:        fmt.Sprintf(list, owner),
			ID:          "ID of the alias.",
			Key:         "Name of the alias.",
			Value:       "Name of the " + aliased + " to alias.",
			Description: "Description of the alias.",
		}
	}
	override := func(list, overridden string) variableListDescription {
		return variableListDescription{
			List:        fmt.Sprintf(list, owner),
			ID:          "ID of the override.",
			Key:         "Name of the " + overridden + " to override.",
			Value:       "Value that replaces the inherited one.",
			Description: "Description of the override.",
		}
	}
	file := func(list, item string) variableListDescription {
		d := variable(list, item)
		d.Value = "Content of the file."
		d.MountPath = "Path where the file is mounted."
		return d
	}
	external := func(list, item string) variableListDescription {
		return variableListDescription{
			List:                  fmt.Sprintf(list, owner),
			ID:                    "ID of the " + item + ".",
			Key:                   "Name of the " + item + ".",
			Description:           "Description of the " + item + ".",
			MountPath:             "Absolute path where the file is mounted.",
			Reference:             "Reference of the secret in the secret manager, such as its name or ARN.",
			SecretManagerAccessID: "ID of the cluster's secret manager access that reads the secret.",
		}
	}

	switch attribute {
	case "environment_variables":
		return variable("Environment variables of the %s.", "environment variable")
	case "built_in_environment_variables":
		return variable("Environment variables Qovery defines for the %s.", "environment variable")
	case "environment_variable_aliases":
		return alias("Environment variable aliases of the %s. An alias gives an existing variable another name.", "variable")
	case "environment_variable_overrides":
		return override("Environment variable overrides of the %s. An override replaces the value of a variable inherited from a broader scope.", "variable")
	case "secrets":
		return variable("Secrets of the %s.", "secret")
	case "secret_aliases":
		return alias("Secret aliases of the %s. An alias gives an existing secret another name.", "secret")
	case "secret_overrides":
		return override("Secret overrides of the %s. An override replaces the value of a secret inherited from a broader scope.", "secret")
	case "environment_variable_files":
		return file("Environment variable files of the %s, each mounted as a file.", "variable")
	case "secret_files":
		return file("Secret files of the %s, each mounted as a file.", "secret")
	case "external_secrets":
		return external("External secrets of the %s, read from an external secret manager such as AWS Secrets Manager.", "external secret")
	case "external_secret_files":
		return external("External secret files of the %s, read from an external secret manager and mounted as files.", "external secret file")
	}
	panic("variableListDescriptions: unknown attribute " + attribute)
}

// portDescription holds the descriptions of the ports attribute of application, container and
// helm, and of its nested attributes.
type portDescription struct {
	List               string
	ID                 string
	Name               string
	InternalPort       string
	ExternalPort       string
	PubliclyAccessible string
	Protocol           string
	IsDefault          string
}

func portDescriptions(kind string) portDescription {
	return portDescription{
		List:               fmt.Sprintf("Ports of the %s. A publicly accessible port needs an `external_port`.", kind),
		ID:                 "ID of the port.",
		Name:               "Name of the port.",
		InternalPort:       fmt.Sprintf("Port the %s listens on.", kind),
		ExternalPort:       "Port exposed to the internet. Required when `publicly_accessible` is `true`.",
		PubliclyAccessible: "Whether the port is exposed to the internet.",
		Protocol:           "Protocol of the port.",
		// Documented exception to the config-is-source-of-truth rule.
		IsDefault: fmt.Sprintf("Whether the root domain of the %s routes to this port. Qovery marks one port as the default, so an omitted value keeps the one Qovery chose.", kind),
	}
}

// dataSourcePortIsDefaultDescription describes ports.is_default on a data source.
func dataSourcePortIsDefaultDescription(kind string) string {
	return fmt.Sprintf("Whether the root domain of the %s routes to this port.", kind)
}

// portNameDefaultNote documents the PortNameDefault plan modifier.
const portNameDefaultNote = "\n\t- Default: `p<internal_port>`, for example `p8080`."

// storageDescription holds the descriptions of the storage attribute of application and
// container, and of its nested attributes.
type storageDescription struct {
	List       string
	ID         string
	Type       string
	Size       string
	MountPoint string
}

func storageDescriptions(kind string) storageDescription {
	return storageDescription{
		List:       fmt.Sprintf("Persistent volumes of the %s. Their data survives restarts.", kind),
		ID:         "ID of the storage.",
		Type:       "Type of the storage.",
		Size:       "Size of the storage, in GB.",
		MountPoint: "Path where the storage is mounted.",
	}
}

// customDomainDescription holds the descriptions of the custom_domains attribute of application,
// container and helm, and of its nested attributes.
type customDomainDescription struct {
	List                string
	ID                  string
	Domain              string
	GenerateCertificate string
	UseCDN              string
	ValidationDomain    string
	Status              string
}

func customDomainDescriptions(kind string) customDomainDescription {
	return customDomainDescription{
		List:                fmt.Sprintf("Custom domains of the %s. Each one needs a CNAME record that points to its `validation_domain`.", kind),
		ID:                  "ID of the custom domain.",
		Domain:              "Custom domain, for example `app.example.com`.",
		GenerateCertificate: "Whether Qovery issues and renews a Let's Encrypt TLS certificate for the domain.",
		UseCDN:              "Whether the domain is behind a CDN such as Cloudflare. Qovery then only checks that the domain resolves to an IP, not to the service's load balancer.",
		ValidationDomain:    "Domain the CNAME record must point to.",
		Status:              "Status of the custom domain.",
	}
}

// deploymentRestrictionDescription holds the descriptions of the deployment_restrictions
// attribute of application, helm and job, and of its nested attributes.
type deploymentRestrictionDescription struct {
	List  string
	ID    string
	Mode  string
	Type  string
	Value string
}

func deploymentRestrictionDescriptions(kind string) deploymentRestrictionDescription {
	return deploymentRestrictionDescription{
		List:  fmt.Sprintf("Deployment restrictions of the %s: a new commit deploys it only if the files it changes pass them.", kind),
		ID:    "ID of the deployment restriction.",
		Mode:  "`MATCH` deploys only when a changed file matches `value`; `EXCLUDE` ignores the changed files that match it.",
		Type:  "Type of the restriction. Only `PATH` is supported.",
		Value: "Path the changed files are compared with, for example `src/`.",
	}
}
