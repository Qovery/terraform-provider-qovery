data "qovery_organization_member" "dev" {
  organization_id = qovery_organization.my_organization.id
  email           = "dev@company.com"
}
