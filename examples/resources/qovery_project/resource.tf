resource "qovery_project" "my_project" {
  # Required
  organization_id = qovery_organization.my_organization.id
  name            = "MyProject"

  # Optional
  description = "My project description"
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
    qovery_organization.my_organization
  ]
}
