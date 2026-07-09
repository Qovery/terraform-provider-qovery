data "qovery_custom_role" "project_admin" {
  organization_id = qovery_organization.my_organization.id
  id              = qovery_custom_role.project_admin.id
}
