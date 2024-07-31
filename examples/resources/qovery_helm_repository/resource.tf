resource "qovery_helm_repository" "my_helm_repository" {
  # Required
  organization_id       = qovery_organization.my_organization.id
  name                  = "my_helm_repository"
  kind                  = "OCI_DOCKER_HUB"
  url                   = "https://docker.io"
  skip_tls_verification = false

  # Optional
  description = "My Helm repository"
  config = {
    username = "<my_username>"
    password = "<my_password>"
  }


  depends_on = [
    qovery_organization.my_organization
  ]
}
