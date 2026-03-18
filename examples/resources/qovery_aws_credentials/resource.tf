# AWS credentials using IAM access keys
resource "qovery_aws_credentials" "my_aws_creds" {
  organization_id   = qovery_organization.my_organization.id
  name              = "my-aws-credentials"
  access_key_id     = var.aws_access_key_id
  secret_access_key = var.aws_secret_access_key
}

# AWS credentials using IAM role (cross-account access)
resource "qovery_aws_credentials" "my_aws_role_creds" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-aws-role-credentials"
  role_arn        = var.aws_role_arn
}
