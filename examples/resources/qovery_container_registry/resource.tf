# Docker Hub registry
resource "qovery_container_registry" "docker_hub" {
  organization_id = qovery_organization.my_organization.id
  name            = "My Docker Hub"
  kind            = "DOCKER_HUB"
  url             = "https://docker.io"
  config = {
    username = "<my_username>"
    password = "<my_password_or_access_token>"
  }
  description = "Docker Hub Registry"

  depends_on = [
    qovery_organization.my_organization
  ]
}

# AWS ECR (Elastic Container Registry)
resource "qovery_container_registry" "ecr" {
  organization_id = qovery_organization.my_organization.id
  name            = "My AWS ECR"
  kind            = "ECR"
  url             = "https://<account_id>.dkr.ecr.<region>.amazonaws.com"
  config = {
    access_key_id     = "<aws_access_key_id>"
    secret_access_key = "<aws_secret_access_key>"
    region            = "us-east-1"
  }
  description = "AWS ECR Registry"

  depends_on = [
    qovery_organization.my_organization
  ]
}

# GitHub Container Registry
resource "qovery_container_registry" "github_cr" {
  organization_id = qovery_organization.my_organization.id
  name            = "My GitHub CR"
  kind            = "GITHUB_CR"
  url             = "https://ghcr.io"
  config = {
    username = "<github_username>"
    password = "<github_personal_access_token>"
  }
  description = "GitHub Container Registry"

  depends_on = [
    qovery_organization.my_organization
  ]
}

# GCP Artifact Registry
resource "qovery_container_registry" "gcp_artifact_registry" {
  organization_id = qovery_organization.my_organization.id
  name            = "My GCP Artifact Registry"
  kind            = "GCP_ARTIFACT_REGISTRY"
  url             = "https://<region>-docker.pkg.dev"
  config = {
    region           = "<region>"
    json_credentials = "<gcp_service_account_json_key>"
  }
  description = "GCP Artifact Registry"

  depends_on = [
    qovery_organization.my_organization
  ]
}

# GCP Artifact Registry using Workload Identity Federation (keyless)
resource "qovery_container_registry" "gcp_artifact_registry_wif" {
  organization_id = qovery_organization.my_organization.id
  name            = "My GCP Artifact Registry (WIF)"
  kind            = "GCP_ARTIFACT_REGISTRY"
  url             = "https://<region>-docker.pkg.dev"
  config = {
    region                              = "<region>"
    gcp_credentials_type                = "workload_identity_federation"
    project_id                          = "<gcp_project_id>"
    service_account_email               = "<service_account>@<gcp_project_id>.iam.gserviceaccount.com"
    workload_identity_provider_resource = "projects/<project_number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>"
  }
  description = "GCP Artifact Registry (Workload Identity Federation)"

  depends_on = [
    qovery_organization.my_organization
  ]
}
