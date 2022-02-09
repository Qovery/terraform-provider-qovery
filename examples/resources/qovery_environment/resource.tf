resource "qovery_environment" "my_environment" {
  project_id = qovery_project.my_project.id
  name = "MyEnvironment"
}