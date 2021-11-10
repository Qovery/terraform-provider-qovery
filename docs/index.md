---
page_title: "Provider: Qovery"
subcategory: ""
description: |- Terraform provider for interacting with Qovery API.
---

# Qovery Provider

The Qovery provider is created to allow you to interact via Qovery platform using Terraform. It allows to perform CRUD
operations on Qovery API resources.

## Example Usage

Do not keep your authentication token in HCL for production environments, use Terraform environment variables.

```terraform
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

resource "qovery_organization" "test-orga" {
  name = "test-orga"
  plan = "FREE"
}

resource "qovery_aws_credentials" "test-orga-creds" {
  name = "test-orga-creds"
  organization_id = qovery_organization.test-orga.id
  access_key_id = "XYZ"
  secret_access_key = "XYZ"
}

resource "qovery_cluster" "test-orga-cluster" {
  organization_id = qovery_organization.test-orga.id
  name = "test-orga-cluster"
  region = "eu-west-3"
  cloud_provider = "AWS"
  credentials_id = qovery_aws_credentials.test-orga-creds.id
}
```

### Provider

- **token** (String, Required) API token used to interact with Qovery API

## Adding New Resources

### Add model

Add new model struct in the `qovery/models.go` file, e.g.

```
type Organization struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Plan types.String `tfsdk:"plan"`
}
```

### Register New Resource

Register new resource in `qovery/provider.go`:

```go
// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"qovery_organization":    resourceOrganizationType{},
		"qovery_aws_credentials": resourceAwsCredentialsType{},
		"qovery_cluster":         resourceClusterType{},
		"my_new_resource":         resourceMyNewResourceType{},
	}, nil
}
```

### Create Resource

Copy one of the ```qovery/resource_XYZ.go``` files and rename to `qovery/resource_my_new_resource.go`

Inside the file, declare your resource schema, e.g.:

```go
// Organization Resource schema
func (r resourceOrganizationType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				Computed: false,
			},
			"plan": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}
```

and update all the CRUD (Create, Read, Update, Delete) methods to use the correct Qovery API endpoints