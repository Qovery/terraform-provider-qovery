resource "qovery_aws_credentials" "my_aws_creds" {
  organization_id   = qovery_organization.my_organization.id
  name              = "my_aws_creds"
  access_key_id     = "<your-aws-access-key-id>"
  secret_access_key = "<your-aws-secret-access-key>"

  depends_on = [
    qovery_organization.my_organization
  ]
}