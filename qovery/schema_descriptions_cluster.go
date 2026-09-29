package qovery

import (
	"fmt"
	"strings"
)

// Descriptions of the qovery_cluster resource and data source attributes. Each const or helper
// holds the sentence both describe; the resource appends its defaults and write-side
// constraints, and the data source uses the sentence alone.

const (
	clusterCredentialsIDDescription  = "ID of the cloud provider credentials of the cluster, such as a `qovery_aws_credentials` resource."
	clusterCloudProviderDescription  = "Cloud provider of the cluster."
	clusterRegionDescription         = "Region of the cluster, for example `us-east-2` on AWS, or `on-premise` for a `PARTIALLY_MANAGED` cluster."
	clusterKubernetesModeDescription = "How Qovery manages the Kubernetes cluster: `MANAGED` creates and operates it, `SELF_MANAGED` deploys to a cluster operated outside Qovery, and `PARTIALLY_MANAGED` deploys to an EKS Anywhere cluster."
	clusterProductionDescription     = "Whether the cluster is a production cluster."
	clusterStateDescription          = "State of the cluster."
	// clusterStateValuesNote explains the values of clusterStates, as updateClusterStatus applies
	// them.
	clusterStateValuesNote = " `DEPLOYED` deploys it, `STOPPED` stops it, and `READY` does not deploy it."
)

// Node sizing: instance_type, disk_size, min_running_nodes and max_running_nodes. Their defaults
// come from the clusterNodeSizingDefaultModifier plan modifiers.
const (
	clusterInstanceTypeDescription    = "Instance type of the cluster nodes, for example `t3a.xlarge` on AWS, `DEV1-L` on Scaleway or `Standard_B2s_v2` on Azure."
	clusterDiskSizeDescription        = "Disk size of the cluster nodes, in GB."
	clusterMinRunningNodesDescription = "Minimum number of nodes of the cluster."
	clusterMaxRunningNodesDescription = "Maximum number of nodes of the cluster."

	// clusterNodeSizingNote is the constraint of the four sizing attributes on the resource: the
	// clusters whose nodes Qovery does not size, where clusterNodeSizingDefaultModifier warns.
	clusterNodeSizingNote = " Karpenter, GCP, `SELF_MANAGED` and `PARTIALLY_MANAGED` clusters ignore it, and the plan warns when it is set."

	// The data source reports the API values, which are placeholders where Qovery ignores them.
	dataSourceClusterInstanceTypeNote = " Karpenter clusters report `KARPENTER`, and GCP clusters `AUTO_PILOT`."
	dataSourceClusterDiskSizeNote     = " GCP clusters report `0`."
	dataSourceClusterNodeCountNote    = " Karpenter, GCP and self-managed clusters, which do not use it, report a placeholder."
)

// clusterInstanceTypeDefaultNote lists the instance type ClusterInstanceTypeDefault plans on each
// cloud provider.
var clusterInstanceTypeDefaultNote = fmt.Sprintf("\n\t- Default: `%s` on AWS, `%s` on Scaleway, `%s` on Azure.",
	clusterInstanceTypeDefaults["AWS"], clusterInstanceTypeDefaults["SCW"], clusterInstanceTypeDefaults["AZURE"])

// Features.
const (
	clusterFeaturesDescription  = "Features of the cluster: its VPC, static IPs, Karpenter and GKE KMS key."
	clusterVpcSubnetDescription = "CIDR block of the VPC of an AWS `MANAGED` cluster."
	clusterStaticIPDescription  = "Whether the cluster nodes or NAT gateways use static IP addresses."
	// clusterStaticIPNote is the Qovery API rule that rejectStaticIPDisable anticipates.
	clusterStaticIPNote = " It cannot change once an AWS, GCP or Azure cluster has been deployed."

	clusterNatGatewaysDescription                 = "Static egress IPs of the NAT gateways, on a GCP cluster."
	clusterNatGatewaysStaticIPsEnabledDescription = "Whether the NAT gateways use reserved static egress IPs."
	clusterNatGatewaysStaticIPsEnabledNote        = " Requires `static_ip = true`."
	clusterNatGatewaysStaticIPsCountDescription   = "Number of static egress IPs of the NAT gateways."

	// existingVpcImmutableNote documents RejectExistingVpcChange.
	existingVpcImmutableNote = " Adding, changing or removing it after creation fails at plan time."

	clusterExistingVpcDescription             = "Existing AWS VPC to deploy the cluster into, instead of a VPC Qovery creates."
	clusterExistingVpcIDDescription           = "ID of the existing VPC, for example `vpc-0123456789abcdef0`."
	clusterExistingVpcEksSubnetsNote          = " They must auto-assign public IPv4 addresses (`map_public_ip_on_launch = true`)."
	clusterExistingVpcFargateSubnetsNote      = " They must reach the internet through a NAT gateway."
	clusterExistingVpcPrivateNodesDescription = "Whether the EKS nodes run in private subnets, which reach the internet through a NAT gateway."

	clusterGcpExistingVpcDescription                = "Existing GCP VPC network to deploy the cluster into, instead of a network Qovery creates."
	clusterGcpExistingVpcNameDescription            = "Name of the existing VPC network, for example `my-existing-vpc`."
	clusterGcpExistingVpcProjectIDDescription       = "ID of the GCP project that owns the VPC network, when it is not the project of the credentials (Shared VPC)."
	clusterGcpExistingVpcSubnetworkDescription      = "Name of the subnetwork of the VPC network for the GKE nodes."
	clusterGcpExistingVpcServicesRangeDescription   = "Name of the secondary IP range of the subnetwork for the GKE services."
	clusterGcpExistingVpcPodsRangeDescription       = "Name of the secondary IP range of the subnetwork for the GKE pods."
	clusterGcpExistingVpcExtraPodsRangesDescription = "Names of additional secondary IP ranges for the GKE pods."
	clusterGcpExistingVpcPrivateNodesDescription    = "Whether the GKE nodes are private, without public IP addresses."

	clusterGkeKmsKeyDescription = "Resource name of the Cloud KMS key that encrypts the boot disks, etcd, storage buckets and volumes of a GCP cluster."
	// clusterGkeKmsKeyNote documents rejectGkeKmsKeyChange.
	clusterGkeKmsKeyNote = " Setting, changing or removing it after creation fails at plan time."
)

// existingVpcSubnetsDescription describes a subnet list of features.existing_vpc: zone is the
// availability zone letter, target what the subnets are for.
func existingVpcSubnetsDescription(zone, target string) string {
	return fmt.Sprintf("IDs of the subnets of availability zone %s for %s.", strings.ToUpper(zone), target)
}

// existingVpcFargateSubnetsDescription describes a Fargate subnet list of features.existing_vpc.
func existingVpcFargateSubnetsDescription(zone string) string {
	return fmt.Sprintf("IDs of the private subnets of availability zone %s for EKS Fargate, which Karpenter requires.", strings.ToUpper(zone))
}

// Karpenter.
const (
	clusterKarpenterDescription = "Karpenter configuration of an AWS cluster, where [Karpenter](https://karpenter.sh/) provisions the nodes."
	// clusterKarpenterNote documents the checks of toUpsertClusterRequest and rejectKarpenterDisable.
	clusterKarpenterNote = " New AWS `MANAGED` clusters require it, and it cannot be added or removed after creation."

	karpenterDiskSizeDescription                   = "Root disk size of the nodes Karpenter provisions, in GiB."
	karpenterDiskIopsDescription                   = "Provisioned IOPS of the gp3 root disk of the nodes Karpenter provisions, other than the GPU nodes."
	karpenterDiskThroughputDescription             = "Provisioned throughput of the gp3 root disk of the nodes Karpenter provisions, other than the GPU nodes, in MB/s."
	karpenterDefaultServiceArchitectureDescription = "Default CPU architecture of the services deployed on the cluster: `AMD64` or `ARM64`."
	karpenterNodePoolsDescription                  = "Node pools Karpenter provisions, and the instances they can use."
	karpenterRequirementsDescription               = "Requirements that select the EC2 instances Karpenter can provision."
	karpenterGpuRequirementsDescription            = "Requirements that select the GPU instances of the GPU node pool, for example the `g5` instance family."
	// karpenterRequirementsNote documents the check of toQoveryNodePoolRequirements.
	karpenterRequirementsNote            = " Set one requirement for each key: `InstanceFamily`, `InstanceSize` and `Arch`."
	karpenterRequirementKeyDescription   = "Key of the requirement: `InstanceFamily` (for example `c6i`), `InstanceSize` (for example `xlarge`) or `Arch` (`AMD64` or `ARM64`)."
	karpenterRequirementOperDescription  = "Operator of the requirement: `In`, which matches any of `values`."
	karpenterRequirementValueDescription = "Values of the requirement, for example `[\"c6i\", \"m6i\"]` for `InstanceFamily`."
	// karpenterDiskSizeMinNote is the minimum node disk size of the Qovery API.
	karpenterDiskSizeMinNote = " Qovery requires at least 20 GiB."
)

// karpenterNodePoolDescription holds the descriptions of a Karpenter node pool override and of
// its nested attributes, shared by the resource and data source schemas.
type karpenterNodePoolDescription struct {
	Override               string
	SpotEnabled            string
	Consolidation          string
	ConsolidationEnabled   string
	ConsolidationDays      string
	ConsolidationStartTime string
	ConsolidationDuration  string
	Limits                 string
	LimitsEnabled          string
	LimitsMaxCPU           string
	LimitsMaxMemory        string
	LimitsMaxGPU           string
	DiskSize               string
	DiskIops               string
	DiskThroughput         string
	ConsolidateAfter       string
}

// Resource-only sentences of the Karpenter node pool overrides.
const (
	// karpenterConsolidationOmittedNote matches the budgets the Qovery API sets for a node pool
	// without a consolidation window.
	karpenterConsolidationOmittedNote = " Without it, Karpenter does not replace underutilized nodes."
	// karpenterNodePoolLimitsMinNote is the minimum of the Qovery API for the limits of every node
	// pool but the GPU one.
	karpenterNodePoolLimitsMinNote = " Qovery requires at least 6 vCPU and 6 GiB."
	// karpenterConsolidateAfterNote documents NewConsolidateAfterValidator.
	karpenterConsolidateAfterNote = " At most `24h`, written in the largest whole unit: `1h`, not `60m`."
	// karpenterOptionalNodePoolNote documents the cronjob and GPU node pools, which exist only
	// while their override is declared.
	karpenterOptionalNodePoolNote = " Declaring the block creates the pool, and removing it deletes the pool."
	// karpenterGpuNodePoolNote documents the GPU node pool and warnKarpenterGpuNodePoolRemoval.
	karpenterGpuNodePoolNote = " Declaring the block creates the pool, and removing it deletes the pool and its nodes, which the plan warns about."
)

// karpenterNodePoolDescriptions returns the descriptions of the override of the named node pool:
// "stable", "default", "cronjob" or "GPU". It panics on an unknown pool, which Schema calls in
// every test and at provider start.
func karpenterNodePoolDescriptions(pool string) karpenterNodePoolDescription {
	overrides := map[string]string{
		"stable":  "Settings of the stable node pool, which runs the workloads that need steady availability, such as the Qovery agents.",
		"default": "Settings of the default node pool, which runs the application workloads.",
		"cronjob": "Settings of the cronjob node pool, which runs the cron jobs and lifecycle jobs.",
		"GPU":     "Settings of the GPU node pool, which runs the workloads that request GPUs.",
	}
	override, ok := overrides[pool]
	if !ok {
		panic("karpenterNodePoolDescriptions: unknown node pool " + pool)
	}

	return karpenterNodePoolDescription{
		Override:               override,
		SpotEnabled:            fmt.Sprintf("Whether the %s node pool runs on EC2 Spot instances, which AWS can interrupt with a two-minute notice.", pool),
		Consolidation:          fmt.Sprintf("Window when Karpenter replaces underutilized nodes of the %s node pool with cheaper ones.", pool),
		ConsolidationEnabled:   "Whether the consolidation window is active.",
		ConsolidationDays:      "Days of the week of the window, for example `[\"MONDAY\", \"TUESDAY\"]`.",
		ConsolidationStartTime: "Start time of the window in UTC, as `PThh:mm`, for example `PT02:00`.",
		ConsolidationDuration:  "Duration of the window, as `PThhHmmM`, for example `PT04H00M`.",
		Limits:                 fmt.Sprintf("Limits on the total resources Karpenter provisions for the %s node pool.", pool),
		LimitsEnabled:          "Whether Karpenter enforces the limits.",
		LimitsMaxCPU:           fmt.Sprintf("Maximum total vCPUs of the %s node pool.", pool),
		LimitsMaxMemory:        fmt.Sprintf("Maximum total memory of the %s node pool, in GiB.", pool),
		LimitsMaxGPU:           fmt.Sprintf("Maximum total number of GPUs of the %s node pool.", pool),
		DiskSize:               fmt.Sprintf("Root disk size of the %s nodes, in GiB.", pool),
		DiskIops:               fmt.Sprintf("Provisioned IOPS of the gp3 root disk of the %s nodes.", pool),
		DiskThroughput:         fmt.Sprintf("Provisioned throughput of the gp3 root disk of the %s nodes, in MB/s.", pool),
		ConsolidateAfter:       fmt.Sprintf("Time Karpenter waits before it consolidates an empty or underutilized node of the %s node pool, for example `30s`, `10m` or `1h`.", pool),
	}
}

// KEDA, routing table, labels groups and kubeconfig.
const (
	clusterKedaDescription        = "KEDA configuration of the cluster: [KEDA](https://keda.sh/) enables the event-driven autoscaling of its services, including scale to zero."
	clusterKedaNote               = " `PARTIALLY_MANAGED` clusters do not support it."
	clusterKedaEnabledDescription = "Whether KEDA is installed on the cluster."

	clusterRoutingTableDescription     = "Routes of the cluster VPC, for example to a VPN or a peered VPC."
	clusterRoutingTableNote            = " Omitting it removes every route."
	clusterRouteDescriptionDescription = "Description of the route."
	clusterRouteDestinationDescription = "Destination CIDR block of the route, for example `10.1.0.0/16`."
	clusterRouteTargetDescription      = "Target of the route, for example the ID of a VPC peering connection or of a NAT gateway."

	// clusterLabelsGroupIDsNote is the Qovery API rule on cluster labels groups.
	clusterLabelsGroupIDsNote = " Only AWS `MANAGED` clusters support them."

	clusterKubeconfigDescription    = "Kubeconfig of the cluster, as YAML."
	dataSourceClusterKubeconfigNote = " Null when the API cannot return one, for example before the first deployment or without the cluster admin permission."
	// clusterPartiallyManagedRequiredNote documents the PARTIALLY_MANAGED checks of
	// toUpsertClusterRequest.
	clusterPartiallyManagedRequiredNote = " Required for `PARTIALLY_MANAGED` clusters."
)

// Infrastructure outputs.
const (
	clusterInfrastructureOutputsDescription = "Outputs of the cluster infrastructure, set once the cluster is deployed."
	clusterOutputClusterNameDescription     = "Name of the Kubernetes cluster at the cloud provider."
	clusterOutputClusterArnDescription      = "ARN of the EKS cluster, on AWS."
	clusterOutputClusterSelfLinkDescription = "Self link of the GKE cluster, on GCP."
	clusterOutputOidcIssuerDescription      = "OIDC issuer URL of the cluster, on AWS and Azure, for example to grant IAM roles to service accounts."
	clusterOutputVpcIDDescription           = "ID of the VPC of the cluster, on AWS."
)

// Infrastructure charts parameters of PARTIALLY_MANAGED (EKS Anywhere) clusters.
const (
	clusterInfraChartsDescription                 = "Parameters of the infrastructure Helm charts Qovery installs on a `PARTIALLY_MANAGED` cluster, for ingress, TLS certificates and load balancing."
	clusterNginxParametersDescription             = "Parameters of the NGINX ingress controller."
	clusterNginxReplicaCountDescription           = "Number of replicas of the NGINX ingress controller."
	clusterNginxDefaultSSLCertificateDescription  = "Default TLS certificate, as `<namespace>/<secret name>`, for example `qovery/letsencrypt-acme-qovery-cert`."
	clusterNginxPublishStatusAddressDescription   = "IP address the ingress status reports, which the DNS records resolve to."
	clusterNginxMetalLbLoadBalancerIPsDescription = "IP address MetalLB assigns to the ingress load balancer, for example `192.168.1.100`."
	clusterNginxMetalLbLoadBalancerIPsNote        = " It must be in a MetalLB IP address pool."
	clusterNginxExternalDNSTargetDescription      = "IP address or hostname external-dns sets as the target of the DNS records, for example `192.168.1.100`."
	clusterCertManagerParametersDescription       = "Parameters of cert-manager, which issues the TLS certificates."
	clusterCertManagerNamespaceDescription        = "Kubernetes namespace of cert-manager, for example `cert-manager` or `qovery`."
	clusterMetalLbParametersDescription           = "Parameters of MetalLB, the load balancer that exposes the services of the cluster."
	clusterMetalLbIPAddressPoolsDescription       = "IP address pools of MetalLB, each an IP address or a range, for example `192.168.1.100-192.168.1.200`."
	clusterEksAnywhereParametersDescription       = "Parameters of the EKS Anywhere GitOps integration: the git repository and the path of the cluster YAML file."
	clusterEksAnywhereYAMLFilePathDescription     = "Path of the EKS Anywhere cluster YAML file in the repository, for example `clusters/prod/cluster.yaml`."
	clusterEksAnywhereGitRepositoryDescription    = "Git repository of the EKS Anywhere configuration, which Qovery reads and updates."
	clusterEksAnywhereCommitIDDescription         = "Commit the configuration is pinned to, instead of the latest commit of `branch`."
	clusterEksAnywhereBranchDescription           = "Branch of the repository."
	clusterEksAnywhereProviderDescription         = "Git provider of the repository: `BITBUCKET`, `GITHUB` or `GITLAB`."
	clusterEksAnywhereBackupDescription           = "Backup of the EKS Anywhere cluster."
	clusterEksAnywhereBackupEnabledDescription    = "Whether the cluster is backed up."
	clusterEksAnywhereBackupS3Description         = "S3 bucket that stores the backups."
	clusterEksAnywhereBackupBucketDescription     = "Name of the S3 bucket."
	clusterEksAnywhereBackupRegionDescription     = "AWS region of the S3 bucket."
	clusterEksAnywhereBackupRoleARNDescription    = "ARN of the IAM role Qovery assumes to upload the backups."
	clusterEksAnywhereBackupKeyPrefixDescription  = "Prefix of the keys of the backup objects."
)

// Secret manager accesses.
const (
	clusterSecretManagerAccessesDescription     = "Secret managers the cluster reads external secrets from: AWS Parameter Store, AWS Secrets Manager or GCP Secret Manager."
	secretManagerAccessIDDescription            = "ID of the secret manager access."
	secretManagerAccessNameDescription          = "Name of the secret manager access."
	secretManagerEndpointDescription            = "Endpoint of the secret manager."
	secretManagerEndpointTypeDescription        = "Type of the secret manager: `AWS_PARAMETER_STORE`, `AWS_SECRET_MANAGER` or `GCP_SECRET_MANAGER`."
	secretManagerEndpointRegionDescription      = "Region of the secret manager."
	secretManagerEndpointProjectIDDescription   = "ID of the GCP project of the secret manager."
	secretManagerAuthenticationDescription      = "How the cluster authenticates to the secret manager."
	secretManagerAuthTypeDescription            = "Authentication mode: `AUTOMATICALLY_CONFIGURED`, `AWS_ROLE_ARN`, `AWS_STATIC_CREDENTIALS` or `GCP_JSON_CREDENTIALS`."
	secretManagerAuthRoleARNDescription         = "ARN of the IAM role the cluster assumes."
	secretManagerAuthRegionDescription          = "AWS region of the static credentials."
	secretManagerAuthAccessKeyDescription       = "AWS access key ID."
	secretManagerAuthSecretKeyDescription       = "AWS secret access key."
	secretManagerAuthJSONCredentialsDescription = "JSON key of the GCP service account."
)

// requiredWhenType documents an attribute of a secret manager access that
// toQoverySecretManagerEndpoint or toQoverySecretManagerAuth requires for one type.
func requiredWhenType(value string) string {
	return " Required when `type` is `" + value + "`."
}
