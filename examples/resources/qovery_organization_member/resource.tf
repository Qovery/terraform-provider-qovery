resource "qovery_organization_member" "my_organization_member" {
  organization_id = qovery_organization.my_organization.id
  email           = "dev@example.com"
  role_id         = qovery_custom_role.my_custom_role.id
}
