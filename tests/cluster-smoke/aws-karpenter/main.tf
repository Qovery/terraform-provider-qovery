terraform {
  required_version = ">= 1.13"
  required_providers {
    qovery = {
      source = "qovery/qovery"
    }
  }
}

provider "qovery" {}

variable "credentials_id" {
  type        = string
  description = "Qovery AWS credentials ID"
}

variable "organization_id" {
  type        = string
  description = "Qovery organization ID"
}

variable "cluster_name" {
  type        = string
  description = "Cluster name (format: tf-smoke-aws-<run_id>)"
}

resource "qovery_cluster" "smoke" {
  credentials_id  = var.credentials_id
  organization_id = var.organization_id
  name            = var.cluster_name
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      spot_enabled                 = true
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize", operator = "In", values = ["small", "medium"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a"] },
          { key = "Arch", operator = "In", values = ["AMD64"] },
        ]
      }
    }
  }
}
