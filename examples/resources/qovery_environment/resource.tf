resource "qovery_environment" "my_environment" {
  # Required
  project_id = qovery_project.my_project.id
  name       = "MyEnvironment"

  # Optional
  cluster_id = qovery_cluster.my_cluster.id
  mode       = "DEVELOPMENT"
  environment_variables = [
    {
      key   = "ENV_VAR_KEY"
      value = "ENV_VAR_VALUE"
    }
  ]
  secrets = [
    {
      key   = "SECRET_KEY"
      value = "SECRET_VALUE"
    }
  ]

  depends_on = [
    qovery_project.my_project
  ]
}