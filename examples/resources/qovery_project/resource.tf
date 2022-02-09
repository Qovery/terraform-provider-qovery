resource "qovery_project" "my_project" {
  organization_id = qovery_organization.my_organization.id
  name = "MyProject"
}