# qovery_cluster (Resource)

Provides a Qovery cluster resource. This is used to create and manage Kubernetes clusters on your chosen cloud provider through Qovery.

Qovery supports clusters on **AWS** (EKS), **GCP** (GKE), **Scaleway** (Kapsule), and **Azure** (AKS). Each cloud provider requires its own credentials resource (e.g., `qovery_aws_credentials`). For AWS clusters, you can optionally enable **Karpenter** for automatic node provisioning or deploy on an **existing VPC**. For GCP clusters, you can use **Autopilot** mode or deploy on an **existing VPC**. AWS also supports **PARTIALLY_MANAGED** mode for EKS Anywhere on-premise clusters.


## Example

<div class="alert alert-info">
  <i style="font-size:24px" class="fa">&#xf05a;</i> If you're not familiar with Terraform or just want more examples, you can configure everything you need directly from the <a href="https://console.qovery.com">Qovery console</a>. Then, use our <a href="https://www.qovery.com/docs/terraform-provider/exporter">Terraform exporter</a> feature to generate the corresponding Terraform code.
</div><br />

```terraform
# Common
resource "qovery_organization" "my_organization" {
  name = "My Organization"
  plan = "PROFESSIONAL"
}

#######
# AWS #
#######

# AWS Credentials
resource "qovery_aws_credentials" "aws_creds" {
  organization_id   = qovery_organization.my_organization.id
  name              = "My AWS credentials"
  access_key_id     = var.access_key_id
  secret_access_key = var.secret_access_key
}

# Labels group (EKS clusters only)
resource "qovery_labels_group" "cluster_labels" {
  organization_id = qovery_organization.my_organization.id
  name            = "cluster-labels"

  labels = [
    {
      key                         = "team"
      value                       = "platform"
      propagate_to_cloud_provider = true
    },
  ]
}

# AWS Cluster with Karpenter example
resource "qovery_cluster" "cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_aws_credentials.aws_creds.id
  name            = "test_terraform_provider"
  cloud_provider  = "AWS"
  region          = "us-east-2"

  description = "My cluster description"

  features = {
    vpc_subnet = "10.0.0.0/16"
    static_ip  = "true"
    karpenter = {
      spot_enabled                 = true
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      # set the maximum instance size and familly you can to reduce allocation issue
      qovery_node_pools = {
        requirements = [
          {
            key      = "InstanceSize"
            operator = "In"
            values   = ["small", "medium", "large", "xlarge", "2xlarge", "3xlarge", "4xlarge", "6xlarge", "8xlarge", "9xlarge", "12xlarge", "16xlarge", "18xlarge", "24xlarge", "32xlarge"]
          },
          {
            key      = "InstanceFamily"
            operator = "In"
            values   = ["c5", "c5a", "c5d", "c5n", "c6gd", "c6gn", "c6i", "c6in", "c7g", "c7i", "c7i-flex", "d2", "d3", "i3", "i3en", "i4i", "im4gn", "inf2", "is4gen", "m5", "m5a", "m5ad", "m5d", "m6g", "m6gd", "m6i", "m7g", "m7gd", "m7i", "m7i-flex", "r4", "r5", "r5a", "r5ad", "r5d", "r5dn", "r5n", "r6g", "r6gd", "r6i", "r7i", "t2", "t3", "t3a", "t4g", "x2iedn"]
          },
          {
            key      = "Arch"
            operator = "In"
            values   = ["ARM64", "AMD64"]
          }
        ]
      }
    }
  }

  # Labels groups are only supported on EKS (AWS MANAGED) clusters.
  labels_group_ids = [qovery_labels_group.cluster_labels.id]

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings
    # you can only indicate settings that you need to override
    "aws.vpc.flow_logs_retention_days" : 100,
    "aws.vpc.enable_s3_flow_logs" : true
  })

  state = "DEPLOYED"
}

#######
# GCP #
#######

# GCP Credentials
resource "qovery_gcp_credentials" "gcp_creds" {
  organization_id = qovery_organization.my_organization.id
  name            = "My GCP credentials"
  gcp_credentials = file("${path.module}/service-account.json")
}

resource "qovery_cluster" "gcp_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_gcp_credentials.gcp_creds.id
  name            = "test_terraform_provider"
  cloud_provider  = "GCP"
  region          = "europe-west9"
  state           = "DEPLOYED"

  description       = "My cluster description"
  instance_type     = "AUTO_PILOT"
  min_running_nodes = 3
  max_running_nodes = 200

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings
    # you can only indicate settings that you need to override
    "gcp.vpc.enable_flow_logs" : false,
    "gcp.vpc.flow_logs_sampling" : 0.0,
  })
}

# GCP Cluster with existing VPC
resource "qovery_cluster" "gcp_cluster_custom_vpc" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_gcp_credentials.gcp_creds.id
  name            = "gke-custom-vpc"
  cloud_provider  = "GCP"
  region          = "europe-west1"
  state           = "DEPLOYED"

  instance_type     = "AUTO_PILOT"
  min_running_nodes = 3
  max_running_nodes = 200

  features = {
    gcp_existing_vpc = {
      vpc_name                       = "my-existing-vpc"
      vpc_project_id                 = "my-gcp-project-id"
      subnetwork_name                = "my-subnetwork"
      ip_range_services_name         = "gke-services"
      ip_range_pods_name             = "gke-pods"
      additional_ip_range_pods_names = ["gke-pods-extra-1", "gke-pods-extra-2"]
    }
  }
}

#########
# Azure #
#########

# Azure credentials must be created via the Qovery console (provisioning requires server-side scripts).
# Use data source to reference existing credentials.
data "qovery_azure_credentials" "azure_creds" {
  id              = var.azure_credentials_id
  organization_id = qovery_organization.my_organization.id
}

# Azure AKS Cluster
resource "qovery_cluster" "azure_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = data.qovery_azure_credentials.azure_creds.id
  name            = "my-azure-cluster"
  cloud_provider  = "AZURE"
  region          = "westeurope"
  state           = "DEPLOYED"

  description       = "Azure AKS cluster managed by Qovery"
  instance_type     = "Standard_B2s_v2"
  min_running_nodes = 3
  max_running_nodes = 10
}

############
# Scaleway #
############

resource "qovery_scaleway_credentials" "scw_creds" {
  organization_id = qovery_organization.organization.id
  name            = "qovery-scaleway-tests-creds"

  scaleway_access_key      = var.scaleway_access_key
  scaleway_secret_key      = var.scaleway_secret_key
  scaleway_project_id      = var.scaleway_project_id
  scaleway_organization_id = var.scaleway_organization_id
}

resource "qovery_cluster" "cluster" {
  organization_id = qovery_organization.organization.id
  credentials_id  = qovery_scaleway_credentials.scw_creds.id
  name            = "test_terraform_provider"
  cloud_provider  = "SCW"
  region          = "pl-waw-1"
  state           = "DEPLOYED"

  instance_type     = "DEV1-L"
  min_running_nodes = 3
  max_running_nodes = 10

  description = "test"

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings
    # you can only indicate settings that you need to override
    "load_balancer.size" : "lb-s",
    "scaleway.enable_private_network_migration" : false,
  })
}

################
# EKS Anywhere #
################

resource "qovery_aws_credentials" "eks_anywhere_creds" {
  organization_id   = qovery_organization.my_organization.id
  name              = "My EKS Anywhere credentials"
  access_key_id     = var.access_key_id
  secret_access_key = var.secret_access_key
}

resource "qovery_cluster" "eks_anywhere_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_aws_credentials.eks_anywhere_creds.id
  name            = "my-eks-anywhere-cluster"
  cloud_provider  = "AWS"
  region          = "on-premise"
  kubernetes_mode = "PARTIALLY_MANAGED"

  description = "EKS Anywhere cluster managed by Qovery"

  kubeconfig = file("${path.module}/kubeconfig.yaml")

  infrastructure_charts_parameters = {
    nginx_parameters = {
      replica_count                             = 2
      default_ssl_certificate                   = "qovery/letsencrypt-acme-qovery-cert"
      publish_status_address                    = "192.168.1.100"
      annotation_metal_lb_load_balancer_ips     = "192.168.1.100"
      annotation_external_dns_kubernetes_target = "192.168.1.100"
    }
    cert_manager_parameters = {
      kubernetes_namespace = "qovery"
    }
    metal_lb_parameters = {
      ip_address_pools = ["192.168.1.100-192.168.1.110"]
    }
  }

  state = "DEPLOYED"
}
```

You can find complete examples within these repositories:
* [Deploy an Application and Database within 3 environments](https://github.com/Qovery/terraform-examples/tree/main/examples/deploy-an-application-within-3-environments)
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_provider` (String) Cloud provider where the cluster will be deployed.

  - `AWS` - Amazon Web Services (EKS).
  - `GCP` - Google Cloud Platform (GKE).
  - `SCW` - Scaleway (Kapsule).
  - `AZURE` - Microsoft Azure (AKS).
  - `ON_PREMISE` - On-premise infrastructure.
- `credentials_id` (String) ID of the cloud provider credentials to use for this cluster. Must match the `cloud_provider` type (e.g., use `qovery_aws_credentials.id` for AWS clusters, `qovery_gcp_credentials.id` for GCP clusters).
- `name` (String) Name of the cluster. Must be unique within the organization.
- `organization_id` (String) ID of the Qovery organization in which to create the cluster.
- `region` (String) Cloud provider region where the cluster will be deployed (e.g., `us-east-2` for AWS, `europe-west9` for GCP, `pl-waw-1` for Scaleway, `westeurope` for Azure). For PARTIALLY_MANAGED clusters, use `on-premise`.

### Optional

- `advanced_settings_json` (String) Advanced settings of the cluster as a JSON string. Use `jsonencode()` to set values. The complete list of available settings is in the [Qovery API documentation](https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings). Only include settings you want to override.
- `description` (String) Description of the cluster. Default: `""`.
- `disk_size` (Number) Disk size of the cluster nodes in GB. The default value depends on the cloud provider and instance type.
- `features` (Attributes) Optional cluster features configuration. Use this block to customize VPC settings, enable static IPs, deploy on an existing VPC (AWS or GCP), or enable Karpenter for AWS clusters. (see [below for nested schema](#nestedatt--features))
- `infrastructure_charts_parameters` (Attributes) Infrastructure Helm chart parameters for `PARTIALLY_MANAGED` (EKS Anywhere) clusters. **Required** when `kubernetes_mode` is `PARTIALLY_MANAGED`. These configure the core infrastructure components (ingress, TLS, load balancing) on your on-premise cluster. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters))
- `instance_type` (String) Instance type for the cluster nodes. The available values depend on the cloud provider:

  - **AWS**: EC2 instance types (e.g., `t3a.xlarge`, `m5.large`). Not required when Karpenter is enabled.
  - **GCP**: Machine types or `AUTO_PILOT` for GKE Autopilot mode.
  - **Scaleway**: Node types (e.g., `DEV1-L`, `GP1-S`).
  - **Azure**: VM sizes (e.g., `Standard_B2s_v2`, `Standard_D4s_v3`).
- `kubeconfig` (String, Sensitive) Kubeconfig YAML content for connecting to the cluster. **Required** for `PARTIALLY_MANAGED` (EKS Anywhere) clusters. This is a sensitive value and will not be displayed in plan output. Use `file()` to read from a file.
- `kubernetes_mode` (String) Kubernetes management mode for the cluster. Default: `MANAGED`.

  - `MANAGED` - Fully managed Kubernetes cluster provisioned and managed by Qovery (e.g., AWS EKS, GCP GKE, Azure AKS).
  - `SELF_MANAGED` - Bring your own Kubernetes cluster. Qovery deploys workloads but does not manage infrastructure.
  - `PARTIALLY_MANAGED` - EKS Anywhere / on-premise mode. Qovery manages workloads on a user-provided Kubernetes cluster via kubeconfig. Requires `kubeconfig` and `infrastructure_charts_parameters`.
- `labels_group_ids` (Set of String) List of labels group ids. Labels groups allow you to add Kubernetes labels to the cluster's resources. **Currently supported only for EKS (AWS managed Kubernetes) clusters.** See [Labels & Annotations](https://www.qovery.com/docs/configuration/organization/labels-annotations).
- `max_running_nodes` (Number) Maximum number of nodes the cluster autoscaler can scale up to. Must be `>= 1`. Default: `10`.

~> **Note:** Must be set to `1` for K3S clusters. Do not set this attribute when Karpenter is enabled (Karpenter manages scaling automatically).
- `min_running_nodes` (Number) Minimum number of nodes running for the cluster autoscaler. Must be `>= 1`. Default: `3`.

~> **Note:** Must be set to `1` for K3S clusters. Do not set this attribute when Karpenter is enabled (Karpenter manages scaling automatically).
- `production` (Boolean) Flag to mark this cluster as a production cluster. Production clusters may have different default settings and safeguards. Default: `false`.
- `routing_table` (Attributes Set) Custom routing table entries for the cluster VPC. Use this to define network routes for traffic between the cluster and other networks (e.g., VPN, peering connections). (see [below for nested schema](#nestedatt--routing_table))
- `state` (String) Desired state of the cluster. Default: `DEPLOYED`.

  - `DEPLOYED` - The cluster is running and ready to accept workloads.
  - `STOPPED` - The cluster infrastructure is stopped to save costs. All workloads will be unavailable.

### Read-Only

- `id` (String) Unique identifier of the cluster (UUID format).
- `infrastructure_outputs` (Attributes) Read-only outputs from the underlying Kubernetes infrastructure. These values are populated after the cluster is deployed and can be used to integrate with other infrastructure resources. (see [below for nested schema](#nestedatt--infrastructure_outputs))

<a id="nestedatt--features"></a>
### Nested Schema for `features`

Optional:

- `existing_vpc` (Attributes) AWS existing VPC configuration. Use this block to deploy the Qovery cluster into an existing AWS VPC instead of creating a new one. All EKS subnets are required, while database and cache subnets are optional.

~> **Warning:** This configuration cannot be changed after cluster creation. (see [below for nested schema](#nestedatt--features--existing_vpc))
- `gcp_existing_vpc` (Attributes) GCP existing VPC configuration. Use this block to deploy the Qovery GKE cluster into an existing Google Cloud VPC network instead of creating a new one.

~> **Warning:** This configuration cannot be changed after cluster creation. (see [below for nested schema](#nestedatt--features--gcp_existing_vpc))
- `karpenter` (Attributes) Karpenter configuration for AWS EKS clusters. [Karpenter](https://karpenter.sh/) is a Kubernetes node autoscaler that automatically provisions right-sized compute resources. When Karpenter is enabled, do not set `instance_type`, `min_running_nodes`, or `max_running_nodes` — Karpenter manages node scaling automatically. (see [below for nested schema](#nestedatt--features--karpenter))
- `static_ip` (Boolean) Whether to assign static/elastic IP addresses to the cluster nodes (AWS only). Useful when your services need to be allowlisted by IP. Default: `false`.

~> **Warning:** This value cannot be changed after cluster creation. Changing it will require destroying and recreating the cluster.
- `vpc_subnet` (String) Custom VPC CIDR block for AWS clusters. This defines the IP address range for the entire VPC. Default: `10.0.0.0/16`.

~> **Warning:** This value cannot be changed after cluster creation. Changing it will require destroying and recreating the cluster.

<a id="nestedatt--features--existing_vpc"></a>
### Nested Schema for `features.existing_vpc`

Required:

- `aws_vpc_eks_id` (String) The ID of the existing AWS VPC (e.g., `vpc-0123456789abcdef0`).
- `eks_subnets_zone_a_ids` (List of String) List of subnet IDs in availability zone A for EKS worker nodes. These subnets must have `map_public_ip_on_launch` set to `true`.
- `eks_subnets_zone_b_ids` (List of String) List of subnet IDs in availability zone B for EKS worker nodes. These subnets must have `map_public_ip_on_launch` set to `true`.
- `eks_subnets_zone_c_ids` (List of String) List of subnet IDs in availability zone C for EKS worker nodes. These subnets must have `map_public_ip_on_launch` set to `true`.

Optional:

- `documentdb_subnets_zone_a_ids` (List of String) List of subnet IDs in availability zone A for Amazon DocumentDB. These should be private subnets.
- `documentdb_subnets_zone_b_ids` (List of String) List of subnet IDs in availability zone B for Amazon DocumentDB. These should be private subnets.
- `documentdb_subnets_zone_c_ids` (List of String) List of subnet IDs in availability zone C for Amazon DocumentDB. These should be private subnets.
- `eks_create_nodes_in_private_subnet` (Boolean) Whether to create EKS worker nodes in private subnets. When `true`, nodes are not directly accessible from the internet and route traffic through a NAT Gateway.
- `eks_karpenter_fargate_subnets_zone_a_ids` (List of String) List of private subnet IDs in availability zone A for EKS Fargate (required when using Karpenter). These subnets must be private and connected to the internet through a NAT Gateway.
- `eks_karpenter_fargate_subnets_zone_b_ids` (List of String) List of private subnet IDs in availability zone B for EKS Fargate (required when using Karpenter). These subnets must be private and connected to the internet through a NAT Gateway.
- `eks_karpenter_fargate_subnets_zone_c_ids` (List of String) List of private subnet IDs in availability zone C for EKS Fargate (required when using Karpenter). These subnets must be private and connected to the internet through a NAT Gateway.
- `elasticache_subnets_zone_a_ids` (List of String) List of subnet IDs in availability zone A for Amazon ElastiCache. These should be private subnets.
- `elasticache_subnets_zone_b_ids` (List of String) List of subnet IDs in availability zone B for Amazon ElastiCache. These should be private subnets.
- `elasticache_subnets_zone_c_ids` (List of String) List of subnet IDs in availability zone C for Amazon ElastiCache. These should be private subnets.
- `rds_subnets_zone_a_ids` (List of String) List of subnet IDs in availability zone A for Amazon RDS databases. These should be private subnets.
- `rds_subnets_zone_b_ids` (List of String) List of subnet IDs in availability zone B for Amazon RDS databases. These should be private subnets.
- `rds_subnets_zone_c_ids` (List of String) List of subnet IDs in availability zone C for Amazon RDS databases. These should be private subnets.


<a id="nestedatt--features--gcp_existing_vpc"></a>
### Nested Schema for `features.gcp_existing_vpc`

Required:

- `vpc_name` (String) Name of the existing GCP VPC network to use (e.g., `my-existing-vpc`).

Optional:

- `additional_ip_range_pods_names` (List of String) Additional secondary IP range names for pods. Use this when you need multiple pod IP ranges (e.g., for multi-tenancy or large clusters).
- `ip_range_pods_name` (String) Name of the primary secondary IP range in the subnetwork to use for GKE pods.
- `ip_range_services_name` (String) Name of the secondary IP range in the subnetwork to use for GKE services (ClusterIP range).
- `subnetwork_name` (String) Name of the GCP subnetwork within the VPC to use for the GKE cluster nodes.
- `vpc_project_id` (String) GCP project ID that owns the VPC. If omitted, defaults to the project associated with your GCP credentials. Use this when the VPC is in a different project (Shared VPC pattern).


<a id="nestedatt--features--karpenter"></a>
### Nested Schema for `features.karpenter`

Required:

- `default_service_architecture` (String) Default CPU architecture for services deployed on this cluster. Common values: `AMD64`, `ARM64`. This determines the default node architecture when no specific architecture is requested by a service.
- `disk_size_in_gib` (Number) Root disk size in GiB for nodes provisioned by Karpenter (e.g., `50`).
- `qovery_node_pools` (Attributes) Karpenter node pool configuration. Defines the requirements (instance families, sizes, architectures) and optional resource limits for Qovery-managed node pools. (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools))
- `spot_enabled` (Boolean) Whether to enable EC2 Spot instances for cost savings. Spot instances can be interrupted by AWS with a 2-minute notice, so enable this only for fault-tolerant workloads.

<a id="nestedatt--features--karpenter--qovery_node_pools"></a>
### Nested Schema for `features.karpenter.qovery_node_pools`

Required:

- `requirements` (Attributes List) List of node selection requirements for the Karpenter node pool. Each requirement constrains which EC2 instances Karpenter can provision. You should define at least `InstanceFamily`, `InstanceSize`, and `Arch` requirements. (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--requirements))

Optional:

- `default_override` (Attributes) Override options for the Qovery **default** node pool. The default node pool runs user application workloads. Use this to set resource limits. (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--default_override))
- `stable_override` (Attributes) Override options for the Qovery **stable** node pool. The stable node pool runs services that require consistent availability (e.g., Qovery agents). Use this to configure consolidation windows and resource limits. (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--stable_override))

<a id="nestedatt--features--karpenter--qovery_node_pools--requirements"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.requirements`

Required:

- `key` (String) The requirement key. Valid values:

  - `InstanceFamily` - EC2 instance family (e.g., `c5`, `m5`, `t3a`). Use broad families to reduce allocation issues.
  - `InstanceSize` - EC2 instance size (e.g., `small`, `medium`, `xlarge`, `2xlarge`).
  - `Arch` - CPU architecture (e.g., `AMD64`, `ARM64`).
- `operator` (String) The operator for the requirement. Currently only `In` is supported, meaning the node must match one of the specified values.
- `values` (List of String) List of allowed values for the requirement. For example, for `InstanceFamily`: `["c5", "m5", "t3a"]`, for `Arch`: `["AMD64", "ARM64"]`.


<a id="nestedatt--features--karpenter--qovery_node_pools--default_override"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.default_override`

Optional:

- `limits` (Attributes) Resource limits for the default node pool. Use this to cap the total resources Karpenter can provision for application workloads. (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--default_override--limits))

<a id="nestedatt--features--karpenter--qovery_node_pools--default_override--limits"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.default_override.limits`

Required:

- `enabled` (Boolean) Whether to enforce resource limits on the default node pool.
- `max_cpu_in_vcpu` (Number) Maximum total vCPU cores that Karpenter can provision for the default node pool.
- `max_memory_in_gibibytes` (Number) Maximum total memory in GiB that Karpenter can provision for the default node pool.



<a id="nestedatt--features--karpenter--qovery_node_pools--stable_override"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.stable_override`

Optional:

- `consolidation` (Attributes) Node consolidation schedule for the stable node pool. Consolidation replaces underutilized nodes with more cost-effective alternatives. By default, no consolidation occurs on stable nodes. (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--stable_override--consolidation))
- `limits` (Attributes) Resource limits for the stable node pool. Use this to cap the total resources Karpenter can provision for stable workloads. (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--stable_override--limits))

<a id="nestedatt--features--karpenter--qovery_node_pools--stable_override--consolidation"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.stable_override.consolidation`

Required:

- `days` (List of String) List of days of the week when consolidation should run (e.g., `["Monday", "Tuesday", "Wednesday"]`).
- `duration` (String) Duration of the consolidation window. Must follow the ISO-8601 duration format: `PThhHmmM` (e.g., `PT04H00M` for a 4-hour window).
- `enabled` (Boolean) Whether the consolidation schedule defined here is active. Set to `true` to enable scheduled consolidation.
- `start_time` (String) Start time for the consolidation window. Must follow the ISO-8601 time format: `PThh:mm` (e.g., `PT02:00` for 2:00 AM UTC).


<a id="nestedatt--features--karpenter--qovery_node_pools--stable_override--limits"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.stable_override.limits`

Required:

- `enabled` (Boolean) Whether to enforce resource limits on the stable node pool.
- `max_cpu_in_vcpu` (Number) Maximum total vCPU cores that Karpenter can provision for the stable node pool.
- `max_memory_in_gibibytes` (Number) Maximum total memory in GiB that Karpenter can provision for the stable node pool.






<a id="nestedatt--infrastructure_charts_parameters"></a>
### Nested Schema for `infrastructure_charts_parameters`

Optional:

- `cert_manager_parameters` (Attributes) Configuration for cert-manager, used for automatic TLS certificate provisioning. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters--cert_manager_parameters))
- `eks_anywhere_parameters` (Attributes) Configuration for EKS Anywhere GitOps integration. Use this block to declare the Git repository and YAML path used for EKS Anywhere cluster lifecycle. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters))
- `metal_lb_parameters` (Attributes) Configuration for MetalLB, a bare-metal load balancer for Kubernetes. Required for `PARTIALLY_MANAGED` clusters to expose services externally. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters--metal_lb_parameters))
- `nginx_parameters` (Attributes) Configuration for the Nginx ingress controller deployed on the cluster. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters--nginx_parameters))

<a id="nestedatt--infrastructure_charts_parameters--cert_manager_parameters"></a>
### Nested Schema for `infrastructure_charts_parameters.cert_manager_parameters`

Optional:

- `kubernetes_namespace` (String) Kubernetes namespace where cert-manager is installed (e.g., `cert-manager` or `qovery`).


<a id="nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters"></a>
### Nested Schema for `infrastructure_charts_parameters.eks_anywhere_parameters`

Required:

- `git_repository` (Attributes) Git repository settings used by Qovery to read and update EKS Anywhere configuration. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters--git_repository))
- `yaml_file_path` (String) Path to the EKS Anywhere cluster YAML file in the Git repository (for example: `clusters/prod/cluster.yaml`).

Optional:

- `cluster_backup` (Attributes) Backup settings for EKS Anywhere clusters. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters--cluster_backup))

<a id="nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters--git_repository"></a>
### Nested Schema for `infrastructure_charts_parameters.eks_anywhere_parameters.git_repository`

Required:

- `git_token_id` (String) Qovery Git token ID used to access the repository.
- `url` (String) Git repository URL containing the EKS Anywhere YAML files.

Optional:

- `branch` (String) Repository branch name. Defaults to the repository default branch when omitted.
- `provider` (String) Git provider (`BITBUCKET`, `GITHUB`, `GITLAB`).


<a id="nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters--cluster_backup"></a>
### Nested Schema for `infrastructure_charts_parameters.eks_anywhere_parameters.cluster_backup`

Required:

- `s3` (Attributes) S3 settings used to store backup artifacts. (see [below for nested schema](#nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters--cluster_backup--s3))

Optional:

- `enabled` (Boolean) Enable or disable EKS Anywhere cluster backup.

<a id="nestedatt--infrastructure_charts_parameters--eks_anywhere_parameters--cluster_backup--s3"></a>
### Nested Schema for `infrastructure_charts_parameters.eks_anywhere_parameters.cluster_backup.s3`

Required:

- `bucket` (String) S3 bucket name used to store EKS Anywhere backup artifacts.
- `region` (String) AWS region where the backup bucket is hosted.
- `role_arn` (String) IAM role ARN assumed to upload backup artifacts.

Optional:

- `key_prefix` (String) Optional S3 key prefix used for backup object keys.




<a id="nestedatt--infrastructure_charts_parameters--metal_lb_parameters"></a>
### Nested Schema for `infrastructure_charts_parameters.metal_lb_parameters`

Required:

- `ip_address_pools` (List of String) List of IP address pools for MetalLB. Each entry can be a single IP or an IP range (e.g., `192.168.1.100` or `192.168.1.100-192.168.1.200`). These IPs must be routable on your network.


<a id="nestedatt--infrastructure_charts_parameters--nginx_parameters"></a>
### Nested Schema for `infrastructure_charts_parameters.nginx_parameters`

Optional:

- `annotation_external_dns_kubernetes_target` (String) IP address or hostname used by external-dns for DNS record creation (e.g., `192.168.1.100`).
- `annotation_metal_lb_load_balancer_ips` (String) IP address annotation for MetalLB load balancer allocation (e.g., `192.168.1.100`). Must be within a MetalLB IP address pool.
- `default_ssl_certificate` (String) Default SSL certificate reference in `namespace/secret-name` format (e.g., `qovery/letsencrypt-acme-qovery-cert`).
- `publish_status_address` (String) Public IP address reported in the ingress status. This is the IP that external DNS will resolve to.
- `replica_count` (Number) Number of Nginx ingress controller replicas. Increase for high-availability setups.



<a id="nestedatt--routing_table"></a>
### Nested Schema for `routing_table`

Required:

- `description` (String) Human-readable description of the route's purpose.
- `destination` (String) Destination CIDR block for the route (e.g., `10.1.0.0/16`).
- `target` (String) Target gateway or endpoint for the route (e.g., a VPC peering connection ID or NAT gateway ID).


<a id="nestedatt--infrastructure_outputs"></a>
### Nested Schema for `infrastructure_outputs`

Read-Only:

- `cluster_arn` (String) The Amazon Resource Name (ARN) of the EKS cluster. Only populated for AWS clusters after deployment.
- `cluster_name` (String) The name of the Kubernetes cluster as assigned by the cloud provider. Available after deployment for all providers.
- `cluster_oidc_issuer` (String) The OIDC issuer URL for the cluster. Useful for configuring IAM roles for service accounts (IRSA on AWS, workload identity on Azure). Available for AWS and Azure after deployment.
- `cluster_self_link` (String) The self-link URL of the GKE cluster. Only populated for GCP clusters after deployment.
- `vpc_id` (String) The VPC ID used by the cluster. Only populated for AWS clusters after deployment. Useful for setting up VPC peering or other networking resources.
## Import
```shell
terraform import qovery_cluster.my_cluster "<organization_id>,<cluster_id>"
```