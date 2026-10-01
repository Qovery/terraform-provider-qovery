package qovery

import (
	"strconv"
	"strings"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
)

// Descriptions shared by the organization-level resources, the project and environment resources
// and their data sources: a resource and its data source read the same base sentence from here.

// storedDescriptionDescription describes the description attribute whose schema Default is
// storedDescriptionDefault, an empty string.
func storedDescriptionDescription(kind string) string {
	return descriptions.NewStringDefaultDescription(descriptionDescription(kind), strconv.Quote(storedDescriptionDefault))
}

// usedByKinds appends the kinds of container registry or helm repository that use a config key.
func usedByKinds[K ~string](description string, kinds ...K) string {
	names := make([]string, len(kinds))
	for i, kind := range kinds {
		names[i] = "`" + string(kind) + "`"
	}
	list := names[len(names)-1]
	if len(names) > 1 {
		list = strings.Join(names[:len(names)-1], ", ") + " and " + list
	}
	return description + " Used by " + list + "."
}

// Environment.
const (
	environmentProjectIDDescription = "ID of the project."
	environmentClusterIDDescription = "ID of the cluster the environment deploys to."
	environmentModeDescription      = "Mode of the environment."
)

// Annotations and labels groups.
const (
	annotationsGroupAnnotationsDescription = "Kubernetes annotations of the group, as a map of key to value."
	annotationsGroupScopesDescription      = "Kubernetes resources the annotations apply to."
	labelsGroupLabelsDescription           = "Kubernetes labels of the group."
	labelsGroupLabelKeyDescription         = "Key of the label."
	labelsGroupLabelValueDescription       = "Value of the label."
	labelsGroupLabelPropagateDescription   = "Whether Qovery also applies the label to the cloud provider resources, for example as AWS tags or GCP labels."
)

// annotationsGroupScopes lists the scopes the Qovery API accepts.
var annotationsGroupScopes = clientEnumToStringArray(qovery.AllowedOrganizationAnnotationsGroupScopeEnumEnumValues)

// API token.
const (
	apiTokenRoleIDDescription = "ID of the role that sets the permissions of the token, built-in or custom."
	apiTokenTokenDescription  = "Value of the API token."
)

// Cloud provider credentials.
const (
	credentialsAWSAccessKeyIDDescription              = "AWS access key ID."
	credentialsAWSSecretAccessKeyDescription          = "AWS secret access key."
	credentialsAWSRoleARNDescription                  = "ARN of the AWS IAM role Qovery assumes, for example `arn:aws:iam::123456789012:role/QoveryRole`."
	credentialsGCPJSONKeyDescription                  = "JSON key of the GCP service account."
	credentialsGCPServiceAccountEmailDescription      = "Email of the GCP service account Qovery impersonates through Workload Identity Federation, for example `qovery@my-project.iam.gserviceaccount.com`."
	credentialsGCPWorkloadIdentityProviderDescription = "Full resource name of the Workload Identity Federation provider, for example `projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider`."
	credentialsScalewayAccessKeyDescription           = "Scaleway API access key."
	credentialsScalewaySecretKeyDescription           = "Scaleway API secret key."
	credentialsScalewayProjectIDDescription           = "ID of the Scaleway project."
)

// Container registry and helm repository.
const (
	registryURLDescription                       = "URL of the container registry, for example `https://docker.io` for Docker Hub."
	registryKindDescription                      = "Kind of the container registry."
	registryRegionDescription                    = "Region of the registry, for example `us-east-1` or `fr-par`."
	registryUsernameDescription                  = "Username Qovery authenticates with."
	registryPasswordDescription                  = "Password or access token of `username`."
	helmRepositoryURLDescription                 = "URL of the helm repository, for example `https://charts.example.com`."
	helmRepositoryKindDescription                = "Kind of the helm repository."
	helmRepositorySkipTLSVerificationDescription = "Whether Qovery skips the verification of the TLS certificate of the repository."

	// registryWorkloadIdentityNote marks the config keys of a GCP registry that authenticates
	// through Workload Identity Federation.
	registryWorkloadIdentityNote = " Used by `" + string(registry.KindGcpArtifactRegistry) + "` with `gcp_credentials_type`."
)

// Blueprint.
const (
	blueprintCatalogEntryDescription = "Catalog entry of the blueprint, as `<provider>/<service_family>/<service_version>`, for example `AWS/postgres/17`."
	blueprintTagDescription          = "Catalog release of the blueprint, for example `AWS/postgres/17/4.1.0`."
	blueprintVariablesDescription    = "Variables of the blueprint, as a map of name to value."
	blueprintServiceIDDescription    = "ID of the terraform or helm service that runs the blueprint."
	blueprintServiceTypeDescription  = "Type of the service that runs the blueprint: `" + string(blueprint.ServiceTypeTerraform) + "` or `" + string(blueprint.ServiceTypeHelm) + "`."
	blueprintCatalogURLDescription   = "URL of the blueprint catalog entry."
)

// Custom role.
const (
	customRoleClusterPermissionsDescription     = "Permissions of the role on clusters."
	customRoleClusterIDDescription              = "ID of the cluster."
	customRoleClusterPermissionDescription      = "Permission of the role on the cluster."
	customRoleProjectPermissionsDescription     = "Permissions of the role on projects."
	customRoleProjectIDDescription              = "ID of the project."
	customRoleIsAdminDescription                = "Whether the role has admin rights on the whole project."
	customRoleEnvironmentPermissionsDescription = "Permissions of the role on each environment type of the project."
	customRoleEnvironmentTypeDescription        = "Environment type."
	customRoleEnvironmentPermissionDescription  = "Permission of the role on the environments of this type."
)

// Deployment and deployment stage.
const (
	deploymentIDDescription            = "ID of the deployment, as a UUID."
	deploymentEnvironmentIDDescription = "ID of the environment to deploy."
	deploymentVersionDescription       = "Version of the deployment, as a UUID."
	deploymentDesiredStateDescription  = "Desired state of the environment."
	deploymentReadsNothingNote         = " Always `null`: this data source reads nothing from Qovery."
	deploymentStageIsAfterDescription  = "ID of the deployment stage this stage moves right after."
	deploymentStageIsBeforeDescription = "ID of the deployment stage this stage moves right before."

	// deploymentStageOrderNotReturnedNote documents is_after and is_before, which the resource
	// keeps from the state.
	deploymentStageOrderNotReturnedNote = " The API does not return it, so a reorder made outside Terraform does not show up in the plan."
)

// Git token.
const (
	gitTokenTypeDescription               = "Git provider of the token."
	gitTokenBitbucketWorkspaceDescription = "Bitbucket workspace the token has access to."
	gitTokenTokenDescription              = "Value of the token: a personal access token or an app token of the git provider."
)

// Organization and organization member.
const (
	organizationPlanDescription         = "Subscription plan of the organization."
	organizationMemberIDDescription     = "ID of the member: the invitation ID while the invitation is pending, then the user ID."
	organizationMemberEmailDescription  = "Email of the member."
	organizationMemberRoleIDDescription = "ID of the role of the member, built-in or custom."
	organizationMemberUserIDDescription = "User ID of the member, `null` until the invitation is accepted."
	organizationMemberStatusDescription = "Status of the invitation: `" + member.StatusPending + "`, `" + member.StatusExpired + "` or `" + member.StatusAccepted + "`."
)
