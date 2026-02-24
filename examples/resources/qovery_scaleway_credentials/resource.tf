resource "qovery_scaleway_credentials" "my_scaleway_creds" {
  organization_id          = qovery_organization.my_organization.id
  name                     = "my-scaleway-credentials"
  scaleway_access_key      = var.scaleway_access_key
  scaleway_secret_key      = var.scaleway_secret_key
  scaleway_project_id      = var.scaleway_project_id
  scaleway_organization_id = var.scaleway_organization_id
}
