# Authenticate with a GCP service account key (JSON)
resource "qovery_gcp_credentials" "my_gcp_credentials" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-gcp-credentials"
  gcp_credentials = file("${path.module}/service-account.json")
}

# Authenticate with Workload Identity Federation (keyless)
resource "qovery_gcp_credentials" "my_gcp_wif_credentials" {
  organization_id                     = qovery_organization.my_organization.id
  name                                = "my-gcp-wif-credentials"
  service_account_email               = "qovery@my-project.iam.gserviceaccount.com"
  workload_identity_provider_resource = "projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider"
}
