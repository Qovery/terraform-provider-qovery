data "qovery_gcp_credentials" "my_gcp_credentials" {
  id              = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  organization_id = qovery_organization.my_organization.id
}
