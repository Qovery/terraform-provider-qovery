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
  description = "Qovery Azure credentials ID"
}

variable "organization_id" {
  type        = string
  description = "Qovery organization ID"
}

variable "cluster_name" {
  type        = string
  description = "Cluster name"
}

resource "qovery_cluster" "smoke" {
  credentials_id    = var.credentials_id
  organization_id   = var.organization_id
  name              = var.cluster_name
  cloud_provider    = "AZURE"
  region            = "westeurope"
  kubernetes_mode   = "MANAGED"
  instance_type     = "Standard_B2s_v2"
  min_running_nodes = 1
  max_running_nodes = 3
  state             = "READY"
}
