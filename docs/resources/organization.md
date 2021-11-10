---
page_title: "qovery_organization Resource - terraform-provider-qovery"
subcategory: ""
description: |- The organization resource allows you to manage organizations on Qovery
---

# Resource `qovery_organization`

## Example Usage

```terraform
resource "qovery_organization" "terratestorg" {
  name = "terratestorg"
  plan = "FREE"
}
```

### Name

The name of your organization

### Plan

The plan of your organization