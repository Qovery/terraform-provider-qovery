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

  # Environment variable files (mounted as files in services)
  environment_variable_files = [
    {
      key        = "APP_CONFIG"
      value      = "config-content"
      mount_path = "/etc/app/config.yaml"
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

  # Secret files (mounted as files in services, value is encrypted)
  secret_files = [
    {
      key        = "API_KEY"
      value      = "secret-value"
      mount_path = "/usr/local/secrets/api-key"
    }
  ]

  depends_on = [
    qovery_organization.my_organization
  ]
}
