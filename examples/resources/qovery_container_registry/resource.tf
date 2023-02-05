resource "qovery_container_registry" "my_container_registry" {
  # Required
  organization_id = qovery_organization.my_organization.id
  name            = "my_aws_registry"
  kind            = "ECR"
  url             = "https://my-ecr-url.com"

  # Optional for DockerHub
  config = {
    region = "eu-west-3"
    access_key_id = "<my_access_key_id>"
    secret_access_key = "<my_access_key>"
  }

  # Optional
  description = "My AWS Registry"

  depends_on = [
    qovery_organization.my_organization
  ]
}
