resource "qovery_gcp_credentials" "my_gcp_credentials" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-gcp-credentials"
  gcp_credentials = file("${path.module}/service-account.json")
}
