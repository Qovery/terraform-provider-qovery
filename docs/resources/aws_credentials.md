---
page_title: "qovery_aws_credentials Resource - terraform-provider-qovery"
subcategory: ""
description: |- The cluster resource allows you to manage clusters on Qovery
---

# Resource `qovery_aws_credentials`

## Example Usage

```terraform
resource "qovery_aws_credentials" "test-orga-creds" {
  name = "test-orga-creds"
  organization_id = qovery_organization.test-orga.id
  access_key_id = "XYZ"
  secret_access_key = "XYZ"
}
```

### Name

The name of your credentials

### Organization

Organization to use for credentials

### Access Key Id

AWS Access Key ID

### Secret Access Key

AWS Secret Access Key