resource "qovery_cluster" "my_cluster" {
  # Required
  organization_id = qovery_organization.my_organization.id
  credentials_id  = qovery_aws_credentials.my_aws_creds.id
  name            = "test_terraform_provider"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  instance_type   = "T3A_MEDIUM"

  # Optional
  description       = "My cluster description"
  min_running_nodes = 3
  max_running_nodes = 10
  features = {
    vpc_subnet = "10.0.0.0/16"
  }
  routing_table = [
    {
      description = "RDS database peering"
      destination = "172.30.0.0/16"
      target      = "pcx-06f8f5512c91e389c"
    }
  ]

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings
    # you can only indicate settings that you need to override
    "aws.vpc.flow_logs_retention_days" : 100,
    "aws.vpc.enable_s3_flow_logs" : true
  })

  state = "DEPLOYED"

  depends_on = [
    qovery_organization.my_organization,
    qovery_aws_credentials.my_aws_creds
  ]
}
