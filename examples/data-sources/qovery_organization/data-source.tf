# Retrieve an existing organization by its ID
data "qovery_organization" "my_organization" {
  id = "<organization_id>"
}

# Use organization attributes in other resources
resource "qovery_project" "example" {
  organization_id = data.qovery_organization.my_organization.id
  name            = "MyProject"
}
