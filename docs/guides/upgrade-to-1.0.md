---
page_title: "Upgrading to version 1.0 - Qovery Provider"
subcategory: ""
description: |-
  How to move from a 0.x release of the Qovery provider to 1.0: what breaks, what behaves differently, and the steps to migrate safely.
---

# Upgrading to version 1.0

Version 1.0 is the first stable release of the Qovery provider. It removes the last deprecated attribute and makes `terraform plan` reflect changes made from the Qovery Console on more attributes. Three changes shipped ahead of it in 0.89.0: the `qovery_git_token` data source fix, sensitive secret-bearing attributes and deployment waits that fail on timeout. They are kept below, marked *since 0.89.0*, for upgrades from 0.88 or earlier.

This guide lists every change that can affect an existing configuration and the steps to migrate. The complete list of changes is in the [CHANGELOG](https://github.com/qovery/terraform-provider-qovery/blob/main/CHANGELOG.md).

## Before you start

1. Upgrade to 0.90.0, the last 0.x release, run `terraform apply`, and make sure `terraform plan` reports no changes. The migration below relies on the state written by that release.
2. Back up your state: `terraform state pull > pre-1.0.tfstate`.
3. Provider 1.0 is tested against Terraform 1.15. Earlier Terraform versions are expected to work but are not tested.

## Pin the provider version

Pin the major version so a future release cannot introduce a breaking change without an explicit upgrade:

```terraform
terraform {
  required_providers {
    qovery = {
      source  = "qovery/qovery"
      version = "~> 1.0"
    }
  }
}
```

Run `terraform init -upgrade` only after completing the migration steps below.

## Breaking changes

### `qovery_cluster`: the global `features.karpenter.spot_enabled` is removed

Since 0.87.0, spot instances are configured per Karpenter node pool with `spot_enabled` on `features.karpenter.qovery_node_pools.stable_override`, `default_override` and `cronjob_override`. The global `features.karpenter.spot_enabled` was deprecated at the same time and is removed in 1.0, from both the resource and the data source.

In 1.0, the configuration decides where every node pool runs:

- A node pool without `spot_enabled` runs on **on-demand** instances. The per node pool `spot_enabled` defaults to `false`, and an override block left out of the configuration means `false` for that pool. In 0.x such a pool inherited the global flag.
- The provider sends an explicit value for the stable and default node pools on every apply, and for the cronjob node pool while `cronjob_override` exists.
- A node pool that runs on spot instances without `spot_enabled = true` in the configuration shows up in `terraform plan`, for example because it inherited the global flag or was changed from the Console. An override block the configuration does not declare is shown being removed; a declared block without `spot_enabled` shows `spot_enabled` changing from `true` to `false`. Applying that plan moves the pool to on-demand instances, and the plan prints a warning naming each such node pool.

**Recommended: pin every node pool on the last 0.x release first.** For each node pool, write the value it effectively has today: its own `spot_enabled` if it has one, otherwise the global value. If the global argument is not in your configuration, read its computed value with `terraform state show qovery_cluster.<name>`. Only declare `cronjob_override` if the block already exists in your configuration: declaring it enables a dedicated cronjob node pool.

Before:

```terraform
features = {
  karpenter = {
    spot_enabled                 = true
    disk_size_in_gib             = 50
    default_service_architecture = "AMD64"
    qovery_node_pools = {
      # requirements unchanged
      stable_override = {
        spot_enabled = false
      }
    }
  }
}
```

After, still on 0.x:

```terraform
features = {
  karpenter = {
    spot_enabled                 = true
    disk_size_in_gib             = 50
    default_service_architecture = "AMD64"
    qovery_node_pools = {
      # requirements unchanged
      stable_override = {
        spot_enabled = false
      }
      default_override = {
        spot_enabled = true # the value this pool inherited from the global flag
      }
    }
  }
}
```

Run `terraform apply`, then `terraform plan`: it must report no changes.

**Then remove the global argument and upgrade.** Delete `spot_enabled` from `features.karpenter`, set the provider version to `~> 1.0`, run `terraform init -upgrade`, then `terraform plan`. The plan must report no changes. The removed attribute is dropped from the state automatically; no `terraform state` command is needed.

```terraform
features = {
  karpenter = {
    disk_size_in_gib             = 50
    default_service_architecture = "AMD64"
    qovery_node_pools = {
      # requirements unchanged
      stable_override = {
        spot_enabled = false
      }
      default_override = {
        spot_enabled = true
      }
    }
  }
}
```

**If you upgrade without pinning first,** the first 1.0 `terraform plan` lists every node pool that runs on spot instances without `spot_enabled = true` in the configuration, as described above. Do not apply that plan unless you want those node pools on on-demand instances. Add `spot_enabled = true` to them, then plan again: it must report no changes. A pipeline that applies without a human reviewing the plan moves those node pools to on-demand instances.

Things to check while migrating:

- A `stable_override` or `default_override` block declared only for `limits` or `consolidation`, without `spot_enabled`, now runs its node pool on on-demand instances. If the pool inherited spot instances from the global flag, add `spot_enabled = true` to the block.
- `terraform plan -refresh=false` compares against the previous state only, so it does not show node pools that run on spot instances without being declared.
- `terraform import` cannot tell whether an override block holding only `spot_enabled = false` was declared, because that is the default. The first plan after an import proposes adding the block back; applying it changes nothing on the cluster.

The `qovery_cluster` data source no longer exposes `features.karpenter.spot_enabled`. It now always reports `stable_override` and `default_override` with the `spot_enabled` value each node pool runs with, plus `cronjob_override` while the cronjob node pool is enabled. Read `features.karpenter.qovery_node_pools.<pool>_override.spot_enabled` instead.

### `qovery_git_token` data source: `organization_id` is required (since 0.89.0)

Since 0.89.0 the data source requires `organization_id`. Before that it was a computed attribute that the lookup could never populate, so the data source did not work. Add the argument:

```terraform
data "qovery_git_token" "my_git_token" {
  id              = "<git_token_id>"
  organization_id = "<organization_id>"
}
```

The `token` attribute is now `null` instead of an empty string: the Qovery API never returns the token value.

### Secret-bearing attributes are sensitive (since 0.89.0)

The following attributes are marked sensitive since 0.89.0:

| Resource or data source | Attributes |
|---|---|
| `qovery_database` (resource and data source) | `password` |
| `qovery_container_registry` | `config.password`, `config.secret_access_key`, `config.scaleway_secret_key` |
| `qovery_helm_repository` | `config.password`, `config.secret_access_key`, `config.scaleway_secret_key` |

Terraform hides them in plan and apply output and **rejects an `output` that exposes one of them** unless the output is declared `sensitive = true`. If your configuration has such an output, `terraform plan` fails with `Output refers to sensitive values`. Mark the output:

```terraform
output "database_password" {
  value     = qovery_database.my_database.password
  sensitive = true
}
```

This applies to module outputs too: a module that passes one of these values up must mark its own output sensitive. The values stored in the state file are unchanged; only their display changes.

### `qovery_cluster`: `routing_table` and `labels_group_ids` are managed as a whole

In 0.x, `routing_table` and `labels_group_ids` only refreshed from the API when the state already held a value. A route or a labels group added from the Qovery Console stayed invisible to Terraform while the attribute was omitted, and removing the attribute from the configuration left the remote routes and labels groups in place.

In 1.0 the configuration is the source of truth for both attributes:

- The refresh always reads the API. A route or labels group added, changed or removed from the Console shows up as a difference in `terraform plan`, and `terraform apply` reverts it to the configured value.
- Omitting the attribute means no route and no labels group. Removing it from the configuration plans the removal, and the apply deletes the routes or detaches the labels groups.
- `routing_table = []` now deletes the remaining remote routes. 0.x skipped the API call when the list was empty.
- `terraform import` records the routes and labels groups attached to the cluster.
- The `qovery_cluster` data source reports the routes and labels groups the API holds. In 0.x it reported none unless its configuration set them.

Corrective changes on these attributes redeploy the cluster, like any other change to them.

If you manage routes or labels groups from the Console, add them to the configuration before the first apply on 1.0, otherwise that apply removes them:

```terraform
resource "qovery_cluster" "my_cluster" {
  # ...
  routing_table = [
    { description = "vpn", destination = "172.30.0.0/16", target = "vgw-0123456789abcdef0" },
  ]
  labels_group_ids = [qovery_labels_group.team.id]
}
```

A cluster that declares neither attribute and has nothing attached from the Console plans no change after the upgrade. A configuration that declares `routing_table = []` shows it being added on the first plan; applying it does not change the cluster's configuration.

### `qovery_cluster`: node sizing plans the Qovery default

In 0.x, `instance_type`, `disk_size`, `min_running_nodes` and `max_running_nodes` kept their last value when the configuration omitted them. A value changed from the Qovery Console never showed up in `terraform plan`, and removing the attribute from the configuration left the remote value in place.

In 1.0 these attributes have a default on the clusters whose node group Qovery sizes, `MANAGED` clusters on AWS without Karpenter, on Scaleway and on Azure. The defaults are the values Qovery uses when a request omits them:

| attribute | AWS without Karpenter | Scaleway | Azure |
|---|---|---|---|
| `instance_type` | `t3.xlarge` | `DEV1-L` | `Standard_DS2_v2` |
| `disk_size` | `40` | `40` | `40` |
| `min_running_nodes` | `3` | `3` | `3` |
| `max_running_nodes` | `10` | `10` | `10` |

- The refresh reads the API. A value changed from the Console shows up as a difference in `terraform plan`, and `terraform apply` reverts it to the configured value or to the default.
- Removing one of these attributes from the configuration plans the reset to the default.
- `terraform import` records the remote values.

Qovery does not use these attributes on Karpenter, GCP, `SELF_MANAGED` and `PARTIALLY_MANAGED` clusters: Karpenter, GKE Autopilot or the cluster owner sizes the nodes. There the provider keeps the value the API reports, such as `KARPENTER` or `AUTO_PILOT` for the instance type, and the plan warns when the configuration sets one of them. Remove them from such configurations.

If you set the node sizing from the Console, declare it before the first apply on 1.0, otherwise that apply resets it:

```terraform
resource "qovery_cluster" "my_cluster" {
  # ...
  instance_type     = "GP1-S"
  disk_size         = 100
  min_running_nodes = 3
  max_running_nodes = 6
}
```

Qovery cannot change the CPU architecture of an existing cluster's nodes. On a cluster whose instance type has another architecture than the default, for example an ARM instance type on AWS, always declare `instance_type`: the default would change the architecture, and the apply would fail.

A cluster whose remote values already equal the defaults plans no change after the upgrade. This covers every cluster that was created by Terraform with these attributes omitted and not changed from the Console since.

### `qovery_cluster`: omitted `features` and `keda` plan their defaults

In 0.x, the `features` and `keda` blocks kept their last value when the configuration omitted them: a feature or KEDA changed from the Qovery Console stayed invisible to Terraform, and removing a block left the remote configuration in place.

In 1.0, omitting `features` means the default of every feature: `vpc_subnet = "10.0.0.0/16"`, `static_ip = false`, no reserved NAT gateway IP, and no `existing_vpc`, `gcp_existing_vpc`, `karpenter` or `gke_kms_key`. Omitting `keda` means `enabled = false`.

- The refresh reads the API. A feature or KEDA changed from the Console shows up as a difference in `terraform plan`, and `terraform apply` reverts it. KEDA enabled from the Console on a cluster whose configuration omits `keda` shows up as its disable.
- Removing a block, or one of its attributes, from the configuration plans the reset to the default.
- `features.vpc_subnet` still cannot change after creation: on a cluster created with another subnet, omitting `features` plans the replacement of the cluster. Declare `vpc_subnet` to keep it.
- The Qovery API cannot disable Karpenter on an existing cluster, nor enable or disable static IPs once an AWS, GCP or Azure cluster has been deployed. A plan that removes `features.karpenter`, or turns `features.static_ip` from `true` to `false` on such a cluster, now fails at plan time and names the value to declare. 0.x failed at apply instead.

If you set features or KEDA from the Console, declare them before the first apply on 1.0:

```terraform
resource "qovery_cluster" "my_cluster" {
  # ...
  features = {
    vpc_subnet = "10.42.0.0/16"
    static_ip  = true
  }
  keda = {
    enabled = true
  }
}
```

A cluster whose features and KEDA already equal the defaults plans no change after the upgrade.

### `qovery_cluster`: `features.gke_kms_key` cannot change after creation

In 0.x, changing `features.gke_kms_key` planned the replacement of the cluster, and removing it from the configuration kept the recorded key.

In 1.0 `gke_kms_key` is optional only: omitting it means no KMS key, and the refresh reports the key the API holds. Setting, changing or removing the key on an existing cluster fails at plan time, because the Qovery API only takes it when it creates the cluster. To use another key, destroy the cluster explicitly, e.g. `terraform destroy -target=qovery_cluster.my_cluster`, then apply. `terraform apply -replace` cannot be used, because the plan fails before the replacement.

### `qovery_cluster` data source: node sizing is read-only

The `instance_type`, `disk_size`, `min_running_nodes` and `max_running_nodes` of the `qovery_cluster` data source are read-only and report the values the Qovery API holds, including `KARPENTER` or `AUTO_PILOT` as the instance type and the placeholder node counts of Karpenter, GCP and self-managed clusters, such as `2147483647`. Remove them from data source configurations, otherwise Terraform rejects the configuration.

### Services: `labels_group_ids` and `annotations_group_ids` are managed as a whole

This applies to `qovery_application`, `qovery_container`, `qovery_job` and `qovery_database`.

In 0.x, `labels_group_ids` and `annotations_group_ids` only refreshed from the API when the state already held a value. A group attached from the Qovery Console to a service whose configuration omits the attribute never reached the state, and the next update of the service sent an empty list and detached it without showing anything in the plan.

In 1.0 the configuration is the source of truth for both attributes:

- The refresh always reads the API. A group attached or detached from the Console shows up as a difference in `terraform plan`, and `terraform apply` reverts it to the configured value.
- Omitting the attribute means no group. Removing it from the configuration plans the detach.
- `terraform import` records the groups attached to the service.
- In the `qovery_application`, `qovery_container`, `qovery_job` and `qovery_database` data sources, both attributes are now read-only and report the groups attached to the service. In 0.x they reported nothing unless the data source configuration set them. Remove them from data source configurations, otherwise Terraform rejects the configuration.

If you attach groups from the Console, add them to the configuration before the first apply on 1.0, otherwise that apply detaches them:

```terraform
resource "qovery_container" "my_container" {
  # ...
  labels_group_ids      = [qovery_labels_group.team.id]
  annotations_group_ids = [qovery_annotations_group.team.id]
}
```

A service with no group attached from the Console plans no change after the upgrade.

### Variable, secret and file descriptions are read from the API

This applies to the `description` of `environment_variables`, `environment_variable_aliases`, `environment_variable_overrides`, `environment_variable_files`, `secrets`, `secret_aliases`, `secret_overrides` and `secret_files` on `qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm`, `qovery_environment` and `qovery_project`, wherever the resource has the attribute.

In 0.x the refresh kept a variable's description out of the state whenever the state held none for that variable, so a description set from the Qovery Console never showed up in `terraform plan`. The data sources reported no description, except on application files.

In 1.0 the refresh reads the description from the API:

- A description set or changed from the Console shows up as a difference in `terraform plan`. When the configuration omits the description, `terraform apply` clears it.
- `terraform import` and the data sources report the descriptions the API holds.

The API never returns secret values and does not always return the mount path of secret files, so those are still taken from the state.

If you set descriptions from the Console, add them to the configuration before the first apply on 1.0, otherwise that apply clears them:

```terraform
resource "qovery_application" "my_app" {
  # ...
  environment_variables = [
    { key = "LOG_LEVEL", value = "info", description = "Set from the Console" },
  ]
}
```

### Services: omitted attributes plan the Qovery default

In 0.x, the attributes below kept their last value when the configuration omitted them. A value changed from the Qovery Console never showed up in `terraform plan`, and removing the attribute from the configuration left the remote value in place.

In 1.0 each of them has a default, the value Qovery uses when a request omits it:

| attribute | resources | default |
|---|---|---|
| `icon_uri` | `qovery_application`, `qovery_container`, `qovery_helm`, `qovery_database` | `app://qovery-console/application`, `…/container`, `…/helm`, `…/database` |
| `icon_uri` | `qovery_job` | `app://qovery-console/cron-job` for a cron job, `app://qovery-console/lifecycle-job` for a lifecycle job |
| `auto_deploy` | `qovery_application`, `qovery_container`, `qovery_job` | `true` |
| `auto_deploy` | `qovery_helm` | `false`, the value every 0.x release sent |
| `auto_preview` | `qovery_container`, `qovery_job`, `qovery_helm` (already `false` on `qovery_application`) | `false` |
| `ephemeral_storage` | `qovery_application`, `qovery_container`, `qovery_job` | `0`: no ephemeral storage, the platform default applies |
| `ports.name` | `qovery_application`, `qovery_container` | `p<internal_port>`, for example `p8080` |
| `ports.protocol` | `qovery_container`, `qovery_helm` (already `HTTP` on `qovery_application`) | `HTTP` |
| `custom_domains.generate_certificate` | `qovery_application`, `qovery_container` | `false`, the value 0.x sent |
| `custom_domains.use_cdn` | `qovery_application`, `qovery_container`, `qovery_helm` | `false` |
| `schedule.lifecycle_type` | `qovery_job` | `GENERIC` for a lifecycle job, none for a cron job |
| `source.docker.git_repository.root_path` | `qovery_job` | `/` |

- The refresh always reads the API. A value changed from the Console shows up as a difference in `terraform plan`, and `terraform apply` reverts it to the configured value or to the default.
- Removing one of these attributes from the configuration plans the reset to the default.
- `terraform import` records the remote values.
- An empty job `root_path` is now rejected at plan time: Qovery builds `""` and `/` the same way, and the refresh reads a stored `""` as `/`. Use `/` or omit the attribute.
- Qovery cannot change the lifecycle type of an existing job. A lifecycle job created with `TERRAFORM` or `CLOUDFORMATION` whose configuration omits `schedule.lifecycle_type` now fails at plan time, and the error names the type to declare. Any other change of the type fails at plan time too, where 0.x failed at apply.

If you set any of these from the Console, declare them before the first apply on 1.0, otherwise that apply resets them:

```terraform
resource "qovery_container" "my_container" {
  # ...
  auto_deploy       = false
  ephemeral_storage = 4
  ports = [
    { name = "web", internal_port = 8080, external_port = 443, publicly_accessible = true, protocol = "GRPC" },
  ]
}
```

A service whose remote values already equal the defaults plans no change after the upgrade. This covers every service that was created by Terraform with these attributes omitted and not changed from the Console since.

### Services: `arguments`, job entrypoints and Helm git tokens mean none when omitted

This applies to `arguments` on `qovery_application` and `qovery_container`, to `schedule.on_start.entrypoint`, `schedule.on_stop.entrypoint`, `schedule.on_delete.entrypoint` and `schedule.cronjob.command.entrypoint` on `qovery_job`, and to `source.git_repository.git_token_id` and `values_override.file.git_repository.git_token_id` on `qovery_helm`.

In 0.x these attributes were computed: a value set from the Console stayed invisible to Terraform until an unrelated change cleared it.

In 1.0 they are optional only. Omitting one means no argument, the image's entrypoint or no git token: a value set from the Console shows up in `terraform plan` as a removal, and removing the attribute from the configuration plans its removal. The state upgrade turns the empty `arguments` list that 0.x stored into null, so an unchanged configuration plans nothing. A configuration that declares `arguments = []` shows it being added on the first plan; applying it does not change the service.

If you set any of these from the Console, declare them before the first apply on 1.0. A Helm chart in a private repository keeps its access only if its `git_token_id` is declared.

### `qovery_terraform_service`: variable values are read from the API

In 0.x the refresh kept the state value of every declared variable, so a value changed from the Qovery Console never showed up in `terraform plan`.

In 1.0 the refresh reads the value of non-secret `variables` from the API. A value changed from the Console shows up as a difference, and `terraform apply` reverts it. The API does not return secret values, so a variable with `is_secret = true` still takes its value from the state.

### `qovery_database`: `instance_type` is checked at plan time

A `MANAGED` database requires `instance_type`: a configuration that omits it now fails at plan time instead of at apply. A `CONTAINER` database that sets it gets a plan warning, because the Qovery API ignores the value and reports the type it derives from the cluster. Remove `instance_type` from `CONTAINER` databases.

### `qovery_blueprint`: the refresh reads the API

`qovery_blueprint` was released in 0.90.0. There, the refresh:

- tracked only the declared `variables`, so a variable set from the Qovery Console stayed invisible and the next apply never reset it;
- kept `blueprint` from the state, so a major version changed from the Console showed only as a `tag` difference;
- kept `name`, `tag` and `variables` from the state whenever the last deploy had failed, including a deploy started from the Console;
- took `icon_uri` from the configuration. A change after creation was applied without error, but the Qovery API kept the old icon.

In 1.0:

- The refresh reports every variable whose value differs from its catalog default. A variable set from the Console that the configuration omits shows up in `terraform plan` as a removal, and `terraform apply` resets it to the default. Removing a variable from the configuration plans the same reset. Secret values are still taken from the state, because the API does not return them.
- `blueprint` is derived from the deployed tag. A major version changed from the Console shows up as a difference, and the spelling of the configuration is kept when it names the same version.
- The last applied values are kept only while an apply made by Terraform waits for its retry.
- `icon_uri` is read from the service the blueprint materialized, and changing it after creation is a plan error. To change the icon, change it from the Console, then set the same value in the configuration.

Before the first apply on 1.0, declare the variables you set from the Console, and set `icon_uri` to the icon of the service if you changed it from the Console.

### Organization-level resources: `description` plans its removal

This applies to `qovery_container_registry`, `qovery_helm_repository`, `qovery_custom_role`, `qovery_project`, `qovery_organization` and `qovery_git_token`.

In 0.x `description` kept its last value when the configuration omitted it. Removing it from the configuration left the description in place, and a description changed from the Qovery Console never showed up in `terraform plan`.

In 1.0:

- On `qovery_container_registry`, `qovery_helm_repository`, `qovery_custom_role` and `qovery_project`, `description` defaults to `""`, which is what Qovery stores when a request omits it. Removing it from the configuration plans its reset to `""`.
- On `qovery_organization` and `qovery_git_token`, `description` is optional only, because Qovery stores no description when a request omits it. Removing it from the configuration plans its removal. The state upgrade turns the empty description that 0.x could store into null, so an unchanged configuration plans nothing.
- The refresh reads the description from the API. A description set or changed from the Console shows up as a difference, and `terraform apply` reverts it. `terraform import` records it.

`qovery_git_token.bitbucket_workspace` follows the same rule: it is optional only, and a `BITBUCKET` token without a workspace now fails at plan time instead of at apply.

If you set descriptions from the Console, declare them before the first apply on 1.0, otherwise that apply clears them.

### `qovery_custom_role`: permissions granted from the Console show up in the plan

A custom role applies the default permissions (`VIEWER` on a cluster, `NO_ACCESS` on a project) to every cluster and project that `cluster_permissions` and `project_permissions` do not list. In 0.x the refresh ignored those clusters and projects: a permission granted on one of them from the Qovery Console stayed invisible, and the next apply reset it to the default without the plan showing it.

In 1.0 the refresh records every permission that differs from the defaults, listed or not. A permission granted from the Console on a cluster or project that the configuration does not list shows up in `terraform plan` as an entry to remove, and `terraform apply` resets it. Declare it to keep it:

```terraform
resource "qovery_custom_role" "developer" {
  # ...
  cluster_permissions = [
    { cluster_id = qovery_cluster.production.id, permission = "ENV_CREATOR" },
  ]
}
```

A role without Console-only permissions plans no change after the upgrade.

### Registry `config` blocks and credential identifiers are read from the API

This applies to the `config` block of `qovery_container_registry` and `qovery_helm_repository`, and to `qovery_aws_credentials`, `qovery_scaleway_credentials`, `qovery_gcp_credentials` and `qovery_eks_anywhere_vsphere_credentials`.

In 0.x the refresh kept these values from the state, so a change made from the Qovery Console never showed up in `terraform plan`, and `terraform import` recorded none of them.

In 1.0 the refresh reads the non-secret values that the API returns:

- the `config` keys `region`, `access_key_id`, `username`, `scaleway_access_key`, `scaleway_project_id`, `gcp_credentials_type`, `project_id`, `service_account_email`, `workload_identity_provider_resource` and `token_lifetime_seconds`, for the registry kinds that store them;
- the credential identifiers `access_key_id`, `role_arn`, `scaleway_access_key`, `scaleway_project_id`, `scaleway_organization_id`, `service_account_email`, `workload_identity_provider_resource` and `vsphere_user`.

A value changed from the Console shows up as a difference, and `terraform apply` reverts it. `terraform import` records these values. The API never returns secrets (secret access keys, secret keys, passwords, JSON keys), so they are still taken from the state, and an imported resource plans them once.

A few values keep the state value because the API cannot report them:

- the keys a registry kind does not store, such as `region` on a `PUBLIC_ECR` registry or on an `HTTPS` helm repository;
- every key of `DOCR` and `AZURE_CR` registries and of `OCI_PUBLIC_ECR` helm repositories;
- `project_id` on a GCP registry that uses a JSON key, because Qovery reads it from the key;
- `scaleway_project_id` on an `OCI_SCALEWAY_CR` helm repository, which Qovery requires but does not store.

`token_lifetime_seconds` reads as unset while it holds the Qovery default of 14400. A Scaleway `region` is kept as written when it names the region Qovery stores, such as `fr-par-1` for `fr-par`. A credential identifier is kept when it only differs by surrounding whitespace, which Qovery trims.

A registry or credential whose remote values match the configuration plans no change after the upgrade.

## Behaviour changes

These changes need no configuration edit, but they can make `terraform plan` show differences that 0.x hid.

### `qovery_cluster`: refresh reflects the dedicated cronjob node pool

In 0.x, the refresh only kept `features.karpenter.qovery_node_pools.cronjob_override` when the configuration declared it. A cronjob node pool enabled from the Qovery Console stayed invisible to Terraform, and the next `terraform apply` disabled it without the plan showing it. A pool disabled from the Console while the block was declared was re-enabled the same way.

In 1.0 the refresh stores `cronjob_override` exactly when the pool is enabled on the cluster. A pool enabled from the Console shows in `terraform plan` as the block being removed, and applying that plan disables the pool: declare `cronjob_override` to keep it. A pool disabled from the Console shows as the block being added back.

### `qovery_cluster`: existing VPC subnet lists mean none when omitted

The database and cache subnet lists of `features.existing_vpc` (`rds_subnets_zone_*_ids`, `documentdb_subnets_zone_*_ids` and `elasticache_subnets_zone_*_ids`) are optional only: omitting one means no subnet, and the refresh reports the subnets the API holds. The state upgrade turns the empty lists that 0.x stored for the omitted ones into null, so an unchanged configuration plans nothing. The existing VPC configuration still cannot change after creation.

`existing_vpc.eks_create_nodes_in_private_subnet` and `gcp_existing_vpc.private_nodes` default to `false`, the value Qovery stores when they are omitted.

### `qovery_cluster`: the GPU node pool is managed by `gpu_override`

In 0.x, `qovery_cluster` had no attribute for the Karpenter GPU node pool. A GPU node pool created from the Qovery Console stayed invisible to Terraform, and the next `terraform apply`, even one that only changed the description, deleted it without the plan showing it.

In 1.0 the pool is managed by `features.karpenter.qovery_node_pools.gpu_override`, and the refresh stores the block whenever the pool exists on the cluster. A GPU node pool created from the Console shows in `terraform plan` as the block being removed, with a warning, and applying that plan deletes the pool. To keep it, copy its settings into a `gpu_override` block: `terraform plan` shows them in the block being removed, and the plan is empty once the block matches.

### `qovery_cluster`: Karpenter consolidation and disk settings are managed

In 0.x, `qovery_cluster` had no attribute for these Karpenter settings, which the Qovery Console sets: `consolidate_after` on the node pools, the consolidation schedule and limits of the cronjob node pool, and the disk IOPS and throughput of the Karpenter nodes. The next `terraform apply`, even one that only changed the description, cleared them without the plan showing it.

In 1.0 they are managed by `consolidate_after` on `stable_override`, `default_override`, `cronjob_override` and `gpu_override`, by `consolidation` and `limits` on `cronjob_override`, and by `disk_iops` and `disk_throughput` on `features.karpenter`. The refresh reads them from the API, so a value set from the Console shows in `terraform plan` as being removed, and applying that plan clears it. A `stable_override` or `default_override` block the configuration does not declare shows up in the plan as a block being removed when the Console set its `consolidate_after`. To keep a value, copy it into the configuration.

Write `consolidate_after` in the largest whole unit. Qovery stores it in seconds and returns `1h` for `60m`, so the provider rejects `60m` at plan time.

### Service attributes that keep their value when removed

A few attributes keep their current value when you remove them from the configuration, because the Qovery API gives Terraform no way to plan the reset: `deployment_stage_id`, `ports.is_default`, the git `branch` of `qovery_application` and `qovery_helm`, `blueprint_id` and `advanced_settings_json`. They are the only exceptions to the rule that the configuration is the source of truth, and the [Managing changes](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/managing-changes#attributes-that-keep-their-value-when-removed) guide explains each one.

Two of them behave differently from 0.x. An omitted git branch keeps the current branch, where an unrelated change in 0.x reset it to the default branch. Changing `blueprint_id` after creation is a plan error, where 0.x applied it without effect.

### `advanced_settings_json`: Console resets of tracked keys are reflected

A change made from the Qovery Console to a key the configuration sets in `advanced_settings_json`, including a reset to its default value, now shows up in `terraform plan`, and `terraform apply` sets it back. In 0.x a reset was invisible. The [Managing changes](https://registry.terraform.io/providers/qovery/qovery/latest/docs/guides/managing-changes#advanced-settings) guide describes how the provider tracks advanced settings.

### Deployment waits fail on timeout (since 0.89.0)

When a deployment, a deletion or a cluster operation does not complete before the provider's wait timeout, `terraform apply` fails with a timeout error. Before 0.89.0 the wait returned success and the resource was reported as ready although it had not converged. The timeout is one hour for environment, container, job and Helm deployments and four hours for the other operations. A failed wait leaves the resource in the state: fix the cause from the deployment logs, then apply again.

## Getting help

If `terraform plan` shows a change you did not expect after upgrading, open an issue on the [GitHub repository](https://github.com/qovery/terraform-provider-qovery/issues) with the plan output and the provider version. Never paste a state file or a plan that contains secrets.
