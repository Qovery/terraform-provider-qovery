# AWS (EKS) with Karpenter
resource "qovery_cluster" "my_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_aws_credentials.my_aws_credentials.id
  name            = "my-cluster"
  cloud_provider  = "AWS"
  region          = "us-east-2"

  features = {
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        # Allow as many instance families and sizes as possible: a short list makes node allocation fail more often.
        requirements = [
          {
            key      = "InstanceFamily"
            operator = "In"
            values   = ["c6i", "c7i", "m6i", "m7i", "r6i", "r7i", "t3", "t3a"]
          },
          {
            key      = "InstanceSize"
            operator = "In"
            values   = ["medium", "large", "xlarge", "2xlarge", "4xlarge"]
          },
          {
            key      = "Arch"
            operator = "In"
            values   = ["AMD64"]
          }
        ]

        # Runs the application workloads on Spot instances; the stable node pool stays on on-demand instances.
        default_override = {
          spot_enabled = true
        }
      }
    }
  }
}

# GCP (GKE Autopilot): Autopilot sizes the nodes, so instance_type, disk_size and the node counts do not apply.
resource "qovery_cluster" "my_gcp_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_gcp_credentials.my_gcp_credentials.id
  name            = "my-gcp-cluster"
  cloud_provider  = "GCP"
  region          = "europe-west9"
}

# Azure (AKS): Azure credentials are created from the Qovery Console, the provider cannot create them.
resource "qovery_cluster" "my_azure_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = var.azure_credentials_id
  name            = "my-azure-cluster"
  cloud_provider  = "AZURE"
  region          = "westeurope"
  instance_type   = "Standard_B2s_v2"
}

# Scaleway (Kapsule)
resource "qovery_cluster" "my_scaleway_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_scaleway_credentials.my_scaleway_credentials.id
  name            = "my-scaleway-cluster"
  cloud_provider  = "SCW"
  region          = "pl-waw-1"
  instance_type   = "DEV1-XL"
}

# EKS Anywhere on vSphere: Qovery manages the workloads of an existing on-premise cluster.
resource "qovery_cluster" "my_eks_anywhere_cluster" {
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_eks_anywhere_vsphere_credentials.my_eks_anywhere_vsphere_credentials.id
  name            = "my-eks-anywhere-cluster"
  cloud_provider  = "AWS"
  region          = "on-premise"
  kubernetes_mode = "PARTIALLY_MANAGED"
  kubeconfig      = file("${path.module}/kubeconfig.yaml")

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
}
