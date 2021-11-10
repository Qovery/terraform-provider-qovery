terraform {
  required_providers {
    qovery = {
      source = "qovery.com/api/qovery"
    }
  }
  required_version = "~> 1.0.3"
}

provider "qovery" {
  token = "XYZ"
}

resource "qovery_organization" "myorg" {
  name = "myorg"
  plan = "FREE"
}

resource "qovery_aws_credentials" "myawscreds" {
  name = "myawscreds"
  organization_id = qovery_organization.myorg.id
  access_key_id = "XYZ"
  secret_access_key = "XYZ"
}

resource "qovery_cluster" "mycluster" {
  organization_id = qovery_organization.myorg.id
  name = "mycluster"
  region = "eu-west-3"
  cloud_provider = "AWS"
  credentials_id = qovery_aws_credentials.myawscreds.id
}
