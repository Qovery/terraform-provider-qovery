resource "qovery_project" "my_project" {
  organization_id = qovery_organization.my_organization.id
  name            = "MyProject"

  depends_on = [
    qovery_organization.my_organization
  ]
}

resource "qovery_project" "my_project_with_environment_variables" {
  organization_id = qovery_organization.my_organization.id
  name            = "MyProject"
  environment_variables = [
    {
      "key" : "key",
      "value" : "value"
    }
  ]

  depends_on = [
    qovery_organization.my_organization
  ]
}