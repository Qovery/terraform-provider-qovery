data "qovery_cluster" "my_cluster" {
  id              = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  organization_id = qovery_organization.my_organization.id
}
