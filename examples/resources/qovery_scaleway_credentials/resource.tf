resource "qovery_scaleway_credentials" "my_scaleway_creds" {
  # Required
  organization_id     = qovery_organization.my_organization.id
  name                = "my_scaleway_creds"
  scaleway_access_key = "<your-scaleway-access-key>"
  scaleway_secret_key = "<your-scaleway-secret-key>"
  scaleway_project_id = "<your-scaleway-project-id>"

  depends_on = [
    qovery_organization.my_organization
  ]
}