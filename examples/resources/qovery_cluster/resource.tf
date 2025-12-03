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

  depends_on = [
    qovery_organization.my_organization,
    qovery_aws_credentials.eks_anywhere_creds
  ]
}
