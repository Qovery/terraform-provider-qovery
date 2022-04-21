resource "qovery_environment" "my_environment" {
  project_id = qovery_project.my_project.id
  name       = "MyEnvironment"

  depends_on = [
    qovery_project.my_project
  ]
}

resource "qovery_environment" "my_environment_with_environment_variables" {
  project_id = qovery_project.my_project.id
  name       = "MyEnvironment"
  environment_variables = [
    {
      "key" : "key",
      "value" : "value"
    }
  ]

  depends_on = [
    qovery_organization.my_project
  ]
}