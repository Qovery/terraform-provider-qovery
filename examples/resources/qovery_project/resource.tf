resource "qovery_project" "my_project" {
  # Required
  organization_id = qovery_organization.my_organization.id
  name            = "MyProject"

  # Optional
  description = "Backend services for our SaaS platform"

  # Project-level environment variables are inherited by all environments
  environment_variables = [
    {
      key   = "ENV_VAR_KEY"
      value = "ENV_VAR_VALUE"
    }
  ]

  # Aliases create alternative names for existing environment variables
  environment_variable_aliases = [
    {
      key = "ENV_VAR_KEY_ALIAS"
      # Must match the key of an existing environment variable
      value = "ENV_VAR_KEY"
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

  depends_on = [
    qovery_organization.my_organization
  ]
}
