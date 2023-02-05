resource "qovery_container_registry" "my_container_registry" {
  # Required
  organization_id = qovery_organization.my_organization.id
  name            = "my_aws_creds"
  kind            = "DOCKER_HUB"
  url             = "https://docker.io"

  # Optional
  config = {
    username = "<my_username>"
    password = "<my_password>"
  }
  description = "My Docker Hub Registry"

  depends_on = [
    qovery_organization.my_organization
  ]
}
