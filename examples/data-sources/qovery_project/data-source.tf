# Retrieve an existing project by its ID
data "qovery_project" "my_project" {
  id = "<project_id>"
}

# Use project attributes in other resources
resource "qovery_environment" "example" {
  project_id = data.qovery_project.my_project.id
  name       = "production"
}
