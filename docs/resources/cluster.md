---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "qovery_cluster Resource - terraform-provider-qovery"
subcategory: ""
description: |-
  Provides a Qovery cluster resource. This can be used to create and manage Qovery cluster.
---

# qovery_cluster (Resource)

Provides a Qovery cluster resource. This can be used to create and manage Qovery cluster.

## Example Usage

```terraform
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
  state = "RUNNING"

  depends_on = [
    qovery_organization.my_organization,
    qovery_aws_credentials.my_aws_creds
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_provider` (String) Cloud provider of the cluster.
	- Can be: `AWS`, `DIGITAL_OCEAN`, `SCALEWAY`.
- `credentials_id` (String) Id of the credentials.
- `instance_type` (String) Instance type of the cluster.
	- AWS: `M5_2XLARGE`, `M5_4XLARGE`, `M5_LARGE`, `M5_XLARGE`, `T2_LARGE`, `T2_XLARGE`, `T3A_2XLARGE`, `T3A_LARGE`, `T3A_MEDIUM`, `T3A_SMALL`, `T3A_XLARGE`, `T3_2XLARGE`, `T3_LARGE`, `T3_MEDIUM`, `T3_SMALL`, `T3_XLARGE`.
	- DIGITAL_OCEAN: `S_1VCPU_1GB`, `S_2VCPU_2GB`, `S_2VCPU_4GB`, `S_4VCPU_8GB`, `S_8VCPU_16GB`.
	- SCALEWAY: `DEV1_L`, `DEV1_M`, `DEV1_XL`, `GP1_L`, `GP1_M`, `GP1_S`, `GP1_XL`, `GP1_XS`.
- `name` (String) Name of the cluster.
- `organization_id` (String) Id of the organization.
- `region` (String) Region of the cluster.

### Optional

- `description` (String) Description of the cluster.
	- Default: ``.
- `features` (Attributes) Features of the cluster. (see [below for nested schema](#nestedatt--features))
- `kubernetes_mode` (String) Kubernetes mode of the cluster.
	- Can be: `K3S`, `MANAGED`.
	- Default: `MANAGED`.
- `max_running_nodes` (Number) Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters]
	- Must be: `>= 1`.
	- Default: `10`.
- `min_running_nodes` (Number) Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters].
	- Must be: `>= 1`.
	- Default: `3`.
- `routing_table` (Attributes Set) List of routes of the cluster. (see [below for nested schema](#nestedatt--routing_table))
- `state` (String) State of the cluster.
	- Can be: `RUNNING`, `STOPPED`.
	- Default: `RUNNING`.

### Read-Only

- `id` (String) Id of the cluster.

<a id="nestedatt--features"></a>
### Nested Schema for `features`

Optional:

- `vpc_subnet` (String) Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].
	- Default: `10.0.0.0/16`.


<a id="nestedatt--routing_table"></a>
### Nested Schema for `routing_table`

Required:

- `description` (String) Description of the route.
- `destination` (String) Destination of the route.
- `target` (String) Target of the route.

## Import

Import is supported using the following syntax:

```shell
terraform import qovery_cluster.my_cluster "<organization_id>,<cluster_id>"
```
