# IAM role that Qovery assumes (STS). For static IAM access keys, set access_key_id and secret_access_key instead.
resource "qovery_aws_credentials" "my_aws_credentials" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-aws-credentials"
  role_arn        = var.aws_role_arn
}
