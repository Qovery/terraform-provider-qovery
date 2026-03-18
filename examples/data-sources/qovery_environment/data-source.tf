# Retrieve an existing environment by its ID
data "qovery_environment" "my_environment" {
  id = "<environment_id>"
}

# Use environment attributes in other resources
resource "qovery_deployment" "example" {
  environment_id = data.qovery_environment.my_environment.id
  desired_state  = "RUNNING"
}
