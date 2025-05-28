# qovery_cluster (Resource)

Provides a Qovery cluster resource. This can be used to create and manage Qovery cluster.


## Example

<div class="alert alert-info">
  <i style="font-size:24px" class="fa">&#xf05a;</i> If you're not familiar with Terraform or just want more examples, you can configure everything you need directly from the <a href="https://console.qovery.com">Qovery console</a>. Then, use our <a href="https://hub.qovery.com/docs/using-qovery/configuration/environment/#terraform-exporter">Terraform exporter</a> feature to generate the corresponding Terraform code.
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

# AWS Cluster with Karpenter example
resource "qovery_cluster" "cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_aws_credentials.my_aws_creds.id
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

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings
    # you can only indicate settings that you need to override
    "aws.vpc.flow_logs_retention_days" : 100,
    "aws.vpc.enable_s3_flow_logs" : true
  })

  state = "DEPLOYED"

  depends_on = [
    qovery_organization.my_organization,
    qovery_aws_credentials.my_aws_creds
  ]
}

# AWS Cluster without Karpenter example (soon deprecated)
resource "qovery_cluster" "cluster" {
  # Required
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_aws_credentials.my_aws_creds.id
  name            = "test_terraform_provider"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  instance_type   = "T3A_MEDIUM"

  # Optional
  description       = "My cluster description"
  min_running_nodes = 3
  max_running_nodes = 10
  features = {
    vpc_subnet = "10.0.0.0/16"
  }
  routing_table = [
    {
      description = "RDS database peering"
      destination = "172.30.0.0/16"
      target      = "pcx-06f8f5512c91e389c"
    }
  ]

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings
    # you can only indicate settings that you need to override
    "aws.vpc.flow_logs_retention_days" : 100,
    "aws.vpc.enable_s3_flow_logs" : true
  })

  state = "DEPLOYED"

  depends_on = [
    qovery_organization.my_organization,
    qovery_aws_credentials.my_aws_creds
  ]
}

#######
# GCP #
#######

resource "qovery_cluster" "cluster" {
  organization_id = qovery_organization.my_organization.id
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
  credentials_id  = qovery_aws_credentials.scw_creds.id
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
```

You can find complete examples within these repositories:
* [Deploy an Application and Database within 3 environments](https://github.com/Qovery/terraform-examples/tree/main/examples/deploy-an-application-within-3-environments)
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_provider` (String) Cloud provider of the cluster.
	- Can be: `AWS`, `AZURE`, `GCP`, `ON_PREMISE`, `SCW`.
- `credentials_id` (String) Id of the credentials.
- `name` (String) Name of the cluster.
- `organization_id` (String) Id of the organization.
- `region` (String) Region of the cluster.

### Optional

- `advanced_settings_json` (String) Advanced settings of the cluster.
- `description` (String) Description of the cluster.
	- Default: ``.
- `disk_size` (Number)
- `features` (Attributes) Features of the cluster. (see [below for nested schema](#nestedatt--features))
- `instance_type` (String) Instance type of the cluster. I.e: For Aws `t3a.xlarge`, for Scaleway `DEV-L`, and not set for Karpenter-enabled clusters
- `kubernetes_mode` (String) Kubernetes mode of the cluster.
	- Can be: `MANAGED`, `SELF_MANAGED`.
	- Default: `MANAGED`.
- `max_running_nodes` (Number) Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters; and not set for Karpenter-enabled clusters]
	- Must be: `>= 1`.
	- Default: `10`.
- `min_running_nodes` (Number) Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters, and not set for Karpenter-enabled clusters].
	- Must be: `>= 1`.
	- Default: `3`.
- `production` (Boolean) Specific flag to indicate that this cluster is a production one.
- `routing_table` (Attributes Set) List of routes of the cluster. (see [below for nested schema](#nestedatt--routing_table))
- `state` (String) State of the cluster.
	- Can be: `DEPLOYED`, `STOPPED`.
	- Default: `DEPLOYED`.

### Read-Only

- `id` (String) Id of the cluster.
- `infrastructure_outputs` (Attributes) Outputs related to the underlying Kubernetes infrastructure. These values are only available once the cluster is deployed. (see [below for nested schema](#nestedatt--infrastructure_outputs))

<a id="nestedatt--features"></a>
### Nested Schema for `features`

Optional:

- `existing_vpc` (Attributes) Network configuration if you want to install qovery on an existing VPC (see [below for nested schema](#nestedatt--features--existing_vpc))
- `karpenter` (Attributes) Karpenter parameters if you want to use Karpenter on an EKS cluster (see [below for nested schema](#nestedatt--features--karpenter))
- `static_ip` (Boolean) Static IP (AWS only) [NOTE: can't be updated after creation].
	- Default: `false`.
- `vpc_subnet` (String) Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].
	- Default: `10.0.0.0/16`.

<a id="nestedatt--features--existing_vpc"></a>
### Nested Schema for `features.existing_vpc`

Required:

- `aws_vpc_eks_id` (String) Aws VPC id
- `eks_subnets_zone_a_ids` (List of String) Ids of the subnets for EKS zone a. Must have map_public_ip_on_launch set to true
- `eks_subnets_zone_b_ids` (List of String) Ids of the subnets for EKS zone b. Must have map_public_ip_on_launch set to true
- `eks_subnets_zone_c_ids` (List of String) Ids of the subnets for EKS zone c. Must have map_public_ip_on_launch set to true

Optional:

- `documentdb_subnets_zone_a_ids` (List of String) Ids of the subnets for document db
- `documentdb_subnets_zone_b_ids` (List of String) Ids of the subnets for document db
- `documentdb_subnets_zone_c_ids` (List of String) Ids of the subnets for document db
- `eks_karpenter_fargate_subnets_zone_a_ids` (List of String) Ids of the subnets for EKS fargate zone a. Must have to be private and connected to internet through a NAT Gateway
- `eks_karpenter_fargate_subnets_zone_b_ids` (List of String) Ids of the subnets for EKS fargate zone b. Must have to be private and connected to internet through a NAT Gateway
- `eks_karpenter_fargate_subnets_zone_c_ids` (List of String) Ids of the subnets for EKS fargate zone c. Must have to be private and connected to internet through a NAT Gateway
- `elasticache_subnets_zone_a_ids` (List of String) Ids of the subnets for elasticache
- `elasticache_subnets_zone_b_ids` (List of String) Ids of the subnets for elasticache
- `elasticache_subnets_zone_c_ids` (List of String) Ids of the subnets for elasticache
- `rds_subnets_zone_a_ids` (List of String) Ids of the subnets for RDS
- `rds_subnets_zone_b_ids` (List of String) Ids of the subnets for RDS
- `rds_subnets_zone_c_ids` (List of String) Ids of the subnets for RDS


<a id="nestedatt--features--karpenter"></a>
### Nested Schema for `features.karpenter`

Required:

- `default_service_architecture` (String) The default architecture of service
- `disk_size_in_gib` (Number)
- `qovery_node_pools` (Attributes) Karpenter node pool configuration (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools))
- `spot_enabled` (Boolean) Enable spot instances

<a id="nestedatt--features--karpenter--qovery_node_pools"></a>
### Nested Schema for `features.karpenter.qovery_node_pools`

Required:

- `requirements` (Attributes List) List of requirements for the node pool (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--requirements))

Optional:

- `default_override` (Attributes) Defines some overriden options for Qovery default node pool (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--default_override))
- `stable_override` (Attributes) Defines some overriden options for Qovery stable node pool (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--stable_override))

<a id="nestedatt--features--karpenter--qovery_node_pools--requirements"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.requirements`

Required:

- `key` (String) The key of the requirement (e.g., InstanceFamily, InstanceSize, Arch)
- `operator` (String) The operator for the requirement (e.g., In)
- `values` (List of String) List of values for the requirement


<a id="nestedatt--features--karpenter--qovery_node_pools--default_override"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.default_override`

Optional:

- `limits` (Attributes) Specifies the limits to apply on the default node pool (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--default_override--limits))

<a id="nestedatt--features--karpenter--qovery_node_pools--default_override--limits"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.default_override.limits`

Required:

- `enabled` (Boolean) Enabled the limit
- `max_cpu_in_vcpu` (Number)
- `max_memory_in_gibibytes` (Number)



<a id="nestedatt--features--karpenter--qovery_node_pools--stable_override"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.stable_override`

Optional:

- `consolidation` (Attributes) Specifies the period to consolidate nodes (by default, no consolidation happens) (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--stable_override--consolidation))
- `limits` (Attributes) Specifies the limits to apply on the stable node pool (see [below for nested schema](#nestedatt--features--karpenter--qovery_node_pools--stable_override--limits))

<a id="nestedatt--features--karpenter--qovery_node_pools--stable_override--consolidation"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.stable_override.consolidation`

Required:

- `days` (List of String)
- `duration` (String)
- `enabled` (Boolean)
- `start_time` (String)


<a id="nestedatt--features--karpenter--qovery_node_pools--stable_override--limits"></a>
### Nested Schema for `features.karpenter.qovery_node_pools.stable_override.limits`

Required:

- `enabled` (Boolean) Enabled the limit
- `max_cpu_in_vcpu` (Number)
- `max_memory_in_gibibytes` (Number)






<a id="nestedatt--routing_table"></a>
### Nested Schema for `routing_table`

Required:

- `description` (String) Description of the route.
- `destination` (String) Destination of the route.
- `target` (String) Target of the route.


<a id="nestedatt--infrastructure_outputs"></a>
### Nested Schema for `infrastructure_outputs`

Read-Only:

- `cluster_arn` (String) The ARN of the AWS cluster. Only available for AWS after deployment.
- `cluster_name` (String) The name of the Kubernetes cluster. Available after deployment for all providers.
- `cluster_oidc_issuer` (String) The OIDC issuer URL for the cluster. Available for AWS and Azure after deployment.
- `cluster_self_link` (String) The self-link of the GCP cluster. Only available for GCP after deployment.
- `vpc_id` (String) The VPC ID used by the cluster. Only available for AWS after deployment.
## Import
```shell
terraform import qovery_cluster.my_cluster "<organization_id>,<cluster_id>"
```