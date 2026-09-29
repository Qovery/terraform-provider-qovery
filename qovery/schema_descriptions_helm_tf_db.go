package qovery

import (
	"fmt"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

// Descriptions of qovery_helm, qovery_terraform_service and qovery_database that their data
// sources share, so the resource and the data source describe each attribute with one sentence.
// A resource appends its default, limits and write-side constraints; a data source uses the
// sentence alone.

// createdFromBlueprintIDDescription describes blueprint_id on qovery_helm and
// qovery_terraform_service. The resource appends blueprintIDRemovalNote.
func createdFromBlueprintIDDescription(kind string) string {
	return fmt.Sprintf("ID of the blueprint the %s is created from.", kind)
}

// qovery_helm.
const (
	helmTimeoutSecDescription                = "Maximum duration of the Helm operations, in seconds."
	helmAutoDeployDescription                = "Whether Qovery redeploys the Helm service on every new commit to its branch."
	helmArgumentsDescription                 = "Arguments passed to the Helm CLI."
	helmAllowClusterWideResourcesDescription = "Whether the chart can deploy resources outside the namespace of the environment, such as CRDs and cluster-scoped resources."

	helmSourceDescription               = "Source of the chart: a Helm repository (`helm_repository`) or a git repository (`git_repository`)."
	helmSourceHelmRepositoryDescription = "Chart from a Helm repository, over HTTPS or OCI."
	helmSourceHelmRepositoryIDDesc      = "ID of the `qovery_helm_repository` that hosts the chart."
	helmSourceChartNameDescription      = "Name of the chart."
	helmSourceChartVersionDescription   = "Version of the chart, for example `1.2.3`."
	helmSourceGitRepositoryDescription  = "Chart from a git repository."
	helmSourceBranchDescription         = "Branch to deploy."
	helmSourceRootPathDescription       = "Directory of the repository that holds the chart."

	helmValuesOverrideDescription      = "Values that override the defaults of the chart."
	helmValuesSetDescription           = "Values passed with `--set`, keyed by value path, for example `image.tag`."
	helmValuesSetStringDescription     = "Values passed with `--set-string`, which keeps each value a string."
	helmValuesSetJSONDescription       = "Values passed with `--set-json`, each one a JSON value."
	helmValuesFileDescription          = "Values files, given inline (`raw`) or read from a git repository (`git_repository`)."
	helmValuesRawDescription           = "Values files given inline, keyed by file name."
	helmValuesRawContentDescription    = "YAML content of the file."
	helmValuesGitRepositoryDescription = "Values files read from a git repository."
	helmValuesGitBranchDescription     = "Branch to read the files from."
	helmValuesGitPathsDescription      = "Paths of the values files in the repository."

	helmPortsDescription            = "Ports of the Helm service, keyed by port name. Each one exposes a Kubernetes service deployed by the chart."
	helmPortServiceNameDescription  = "Name of the Kubernetes service to expose."
	helmPortNamespaceDescription    = "Kubernetes namespace of the service."
	helmPortInternalPortDescription = "Port the Kubernetes service listens on."
	helmPortExternalPortDescription = "Port exposed to the internet."
)

// Defaults of qovery_helm that the schema sets with a literal. They must match the schema
// Default of arguments and of source.git_repository.root_path.
const (
	helmArgumentsDefaultNote  = "\n\t- Default: `[\"--wait\", \"--atomic\", \"--debug\"]`."
	helmSourceRootPathDefault = "/"
)

// qovery_terraform_service.
const (
	terraformServiceAutoDeployDescription = "Whether Qovery redeploys the Terraform service on every new commit to its branch."
	terraformServiceActionDescription     = "Terraform command an auto-deployment runs. `DEFAULT` follows the deployment: plan and apply on start or restart, destroy on delete, plan only on pause."

	terraformServiceGitRepositoryDescription = "Git repository that holds the Terraform code."
	terraformServiceBranchDescription        = "Branch to deploy."
	terraformServiceRootPathDescription      = "Directory of the repository that holds the Terraform code."
	terraformServiceTfvarsFilesDescription   = "Paths of the `.tfvars` files to load, for example `/environments/production/prod.tfvars`."

	terraformServiceVariablesDescription        = "Input variables of the Terraform code."
	terraformServiceVariableKeyDescription      = "Name of the variable."
	terraformServiceVariableValueDescription    = "Value of the variable."
	terraformServiceVariableIsSecretDescription = "Whether the variable is a secret."

	terraformServiceBackendDescription                = "Backend that stores the Terraform state."
	terraformServiceKubernetesBackendDescription      = "Stores the state in the Kubernetes cluster."
	terraformServiceUserProvidedBackendDescription    = "Uses the backend configured in the Terraform code."
	terraformServiceBlueprintBackendDescription       = "Qovery generates the `backend.tf` file of the service from `type` and `config`."
	terraformServiceBlueprintBackendTypeDescription   = "Type of the Terraform backend, for example `s3`, `gcs` or `azurerm`."
	terraformServiceBlueprintBackendConfigDescription = "Static settings of the backend, such as its bucket and region, without credentials."

	terraformServiceEngineDescription                 = "Engine that runs the Terraform code."
	terraformServiceEngineVersionDescription          = "Version of the engine."
	terraformServiceExplicitVersionDescription        = "Version of the engine binary, for example `1.9.0`."
	terraformServiceReadFromTerraformBlockDescription = "Whether Qovery reads the engine version from the `terraform` block of the code."

	terraformServiceJobResourcesDescription = "Resources of the Terraform job, which runs the Terraform commands."
	terraformServiceJobRAMDescription       = "Memory of the Terraform job, in MiB."
	terraformServiceJobGPUDescription       = "Number of GPUs of the Terraform job."
	terraformServiceJobStorageDescription   = "Storage of the Terraform job, in GiB."

	terraformServiceTimeoutDescription               = "Maximum duration of the Terraform operations, in seconds."
	terraformServiceUseClusterCredentialsDescription = "Whether the Terraform job authenticates to the cloud provider with the credentials of the cluster."
	terraformServiceActionExtraArgumentsDescription  = "Extra arguments of each Terraform command, keyed by command, for example `{ apply = [\"-lock=false\"] }`."
	terraformServiceCreatedAtDescription             = "Creation date of the Terraform service."
	terraformServiceUpdatedAtDescription             = "Date of the last update of the Terraform service."
)

// terraformServiceEngines lists the engine values the schema validator accepts.
var terraformServiceEngines = []string{
	string(terraformservice.EngineTerraform),
	string(terraformservice.EngineOpenTofu),
}

// qovery_database.
const (
	databaseTypeDescription          = "Engine of the database."
	databaseVersionDescription       = "Version of the engine, for example `16` for PostgreSQL."
	databaseModeDescription          = "How Qovery runs the database: `CONTAINER` as a container on the cluster, `MANAGED` as a managed service of the cloud provider, such as Amazon RDS."
	databaseAccessibilityDescription = "Network exposure of the database: `PUBLIC` makes it reachable from the internet, `PRIVATE` only from the services of its environment."
	databaseInstanceTypeDescription  = "Instance type of the database, for example `db.t3.micro` on AWS."
	databaseStorageDescription       = "Storage of the database, in GB."
	databaseExternalHostDescription  = "External host of the database, reachable from the internet when `accessibility` is `PUBLIC`."
	databasePortDescription          = "Port the database listens on."
	databaseLoginDescription         = "Username of the master user of the database, generated by Qovery."
	databasePasswordDescription      = "Password of the master user of the database, generated by Qovery."

	// databaseCannotChangeNote is appended to type and mode: the update request does not carry
	// them, so the API keeps the value of the creation.
	databaseCannotChangeNote = " It cannot change after creation."
	// databaseContainerOnlyNote is appended to cpu and memory, which the API ignores for a
	// MANAGED database.
	databaseContainerOnlyNote = " Only for a `CONTAINER` database: a `MANAGED` database uses `instance_type`."
)
