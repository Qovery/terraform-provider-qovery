resource "qovery_environment" "my_environment" {
  # Required
  project_id = qovery_project.my_project.id
  name       = "production"

  # Optional
  # cluster_id cannot be changed after creation (forces replacement)
  cluster_id = qovery_cluster.my_cluster.id
  mode       = "PRODUCTION"

  # Environment-level variables are inherited by all services
  environment_variables = [
    {
      key   = "ENV_VAR_KEY"
      value = "ENV_VAR_VALUE"
    }
  ]

  # Aliases create alternative names for existing variables
  environment_variable_aliases = [
    {
      key = "ENV_VAR_KEY_ALIAS"
      # Must match the key of an existing environment variable
      value = "ENV_VAR_KEY"
    }
  ]

  # Overrides replace values of variables inherited from the project level
  environment_variable_overrides = [
    {
      # Must match the key of a project-level variable
      key   = "SOME_PROJECT_VARIABLE"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  # Secrets are encrypted and not visible after creation
  secrets = [
    {
      key   = "SECRET_KEY"
      value = "SECRET_VALUE"
    }
  ]

  # Aliases create alternative names for existing secrets
  secret_aliases = [
    {
      key = "SECRET_KEY_ALIAS"
      # Must match the key of an existing secret
      value = "SECRET_KEY"
    }
  ]

  # Overrides replace values of secrets inherited from the project level
  secret_overrides = [
    {
      # Must match the key of a project-level secret
      key   = "SOME_PROJECT_SECRET"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  depends_on = [
    qovery_project.my_project
  ]
}
