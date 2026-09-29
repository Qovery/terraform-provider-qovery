resource "qovery_project" "my_project" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-project"
  description     = "Backend services for our SaaS platform"

  # Inherited by every environment of the project.
  environment_variables = [
    {
      key   = "COMPANY_NAME"
      value = "my-company"
    }
  ]
}
