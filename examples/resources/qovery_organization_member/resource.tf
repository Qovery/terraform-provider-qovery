resource "qovery_organization_member" "dev" {
  organization_id = qovery_organization.my_organization.id
  email           = "dev@company.com"
  role_id         = qovery_custom_role.project_admin.id
}
