# OCI Docker Hub repository with authentication
resource "qovery_helm_repository" "my_helm_repository" {
  organization_id       = qovery_organization.my_organization.id
  name                  = "my-docker-hub-helm"
  kind                  = "OCI_DOCKER_HUB"
  url                   = "https://docker.io"
  skip_tls_verification = false

  description = "Docker Hub OCI Helm repository"

  config = {
    username = "<my_username>"
    password = "<my_password>"
  }
}
