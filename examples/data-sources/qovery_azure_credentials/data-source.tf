# Azure credentials must be created via the Qovery console.
# The provisioning process requires server-side scripts that cannot be run via Terraform.
# Use this data source to reference existing Azure credentials in your Terraform configuration.

data "qovery_azure_credentials" "my_azure_creds" {
  id              = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  organization_id = qovery_organization.my_org.id
}
