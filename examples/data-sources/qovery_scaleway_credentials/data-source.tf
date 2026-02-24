data "qovery_scaleway_credentials" "my_scaleway_creds" {
  id              = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  organization_id = qovery_organization.my_organization.id
}
