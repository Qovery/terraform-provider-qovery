resource "qovery_environment" "my_environment" {
  project_id = qovery_project.my_project.id
  cluster_id = qovery_cluster.my_cluster.id
  name       = "production"
  mode       = "PRODUCTION"

  # Inherited by every service of the environment.
  environment_variables = [
    {
      key   = "APP_ENV"
      value = "production"
    }
  ]
}
