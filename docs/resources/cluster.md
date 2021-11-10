---
page_title: "qovery_cluster Resource - terraform-provider-qovery"
subcategory: ""
description: |- The cluster resource allows you to manage clusters on Qovery
---

# Resource `qovery_cluster`

## Example Usage

```terraform
resource "qovery_cluster" "test-orga-cluster" {
  organization_id = qovery_organization.test-orga.id
  name = "test-orga-cluster"
  region = "eu-west-3"
  cloud_provider = "AWS"
  credentials_id = qovery_aws_credentials.test-orga-creds.id
}
```

### Name

The name of your cluster

### Cloud Provider

The cloud provider you want to use

### Region

The region on cloud provider you want to use

### Credentials

`qovery_credentials` you want to use for the cluster

### Organization

`qovery_organization` you want to use for the cluster