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
  description = "Qovery GCP credentials ID"
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
  credentials_id  = var.credentials_id
  organization_id = var.organization_id
  name            = var.cluster_name
  cloud_provider  = "GCP"
  region          = "europe-west9"
  kubernetes_mode = "MANAGED"
  instance_type   = "AUTO_PILOT"
  state           = "DEPLOYED"
}
