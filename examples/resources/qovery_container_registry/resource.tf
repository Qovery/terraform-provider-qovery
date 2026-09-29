resource "qovery_container_registry" "my_container_registry" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-docker-hub"
  kind            = "DOCKER_HUB"
  url             = "https://docker.io"

  config = {
    username = "my-docker-hub-user"
    password = var.docker_hub_access_token
  }
}
