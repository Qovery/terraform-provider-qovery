resource "qovery_cluster" "my_cluster" {
  credentials_id  = qovery_aws_credentials.my_aws_creds.id
  organization_id = qovery_organization.my_organization.id
  name            = "test_terraform_provider"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
}
