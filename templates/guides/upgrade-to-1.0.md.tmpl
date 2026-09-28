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

1. Upgrade to 0.89.0, the last 0.x release, run `terraform apply`, and make sure `terraform plan` reports no changes. The migration below relies on the state written by that release.
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

## Behaviour changes

These changes need no configuration edit, but they can make `terraform plan` show differences that 0.x hid.

### `qovery_cluster`: refresh reflects the dedicated cronjob node pool

In 0.x, the refresh only kept `features.karpenter.qovery_node_pools.cronjob_override` when the configuration declared it. A cronjob node pool enabled from the Qovery Console stayed invisible to Terraform, and the next `terraform apply` disabled it without the plan showing it. A pool disabled from the Console while the block was declared was re-enabled the same way.

In 1.0 the refresh stores `cronjob_override` exactly when the pool is enabled on the cluster. A pool enabled from the Console shows in `terraform plan` as the block being removed, and applying that plan disables the pool: declare `cronjob_override` to keep it. A pool disabled from the Console shows as the block being added back.

### `qovery_cluster`: the GPU node pool is managed by `gpu_override`

In 0.x, `qovery_cluster` had no attribute for the Karpenter GPU node pool. A GPU node pool created from the Qovery Console stayed invisible to Terraform, and the next `terraform apply`, even one that only changed the description, deleted it without the plan showing it.

In 1.0 the pool is managed by `features.karpenter.qovery_node_pools.gpu_override`, and the refresh stores the block whenever the pool exists on the cluster. A GPU node pool created from the Console shows in `terraform plan` as the block being removed, with a warning, and applying that plan deletes the pool. To keep it, copy its settings into a `gpu_override` block: `terraform plan` shows them in the block being removed, and the plan is empty once the block matches.

### Service attributes that keep their value when removed

A few service attributes keep their current value when you remove them from the configuration, because the Qovery API gives Terraform no way to plan the reset. They are the only exceptions to the rule that the configuration is the source of truth:

- `deployment_stage_id` on `qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm`, `qovery_terraform_service` and `qovery_database`: Qovery attaches every service to a deployment stage and cannot detach it, so removing the attribute keeps the service in its current stage.
- `ports.is_default` on `qovery_application`, `qovery_container` and `qovery_helm`: the API always marks one port as the default, so an omitted value keeps the value the API chose.
- `git_repository.branch` on `qovery_application` and `source.git_repository.branch` on `qovery_helm`: an omitted branch means the repository's default branch, which is only known once the API resolves it, so removing the attribute keeps the current branch. Changing the repository URL while the branch is omitted resolves the new repository's default branch. In 0.x an unrelated change reset it to the default branch.
- `blueprint_id` on `qovery_helm` and `qovery_terraform_service`: the API records it only when the service is created. Removing it keeps the recorded value, and changing it is now a plan error; to use another blueprint, recreate the service.
- `advanced_settings_json`, described below.

### `advanced_settings_json`: Console resets of tracked keys are reflected

On `qovery_cluster`, `qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm` and `qovery_terraform_service`, `advanced_settings_json` is desired state, not a mirror of the remote configuration:

- Refresh only reconciles keys already tracked in the Terraform state. A setting overridden only in the Console is not pulled into the state, so declaring it afterwards plans as an addition even if the remote value already matches.
- A Console change to a tracked key, including a reset to its default value, is now reflected on refresh and planned back to the configured value. In 0.x a reset was invisible.
- Removing a key from the JSON does not reset it remotely: omitted keys keep their current value. To reset a setting, set it to its default value explicitly.
- `terraform import` records every setting whose value differs from the default.

### Deployment waits fail on timeout (since 0.89.0)

When a deployment, a deletion or a cluster operation does not complete before the provider's wait timeout, `terraform apply` fails with a timeout error. Before 0.89.0 the wait returned success and the resource was reported as ready although it had not converged. The timeout is one hour for environment, container, job and Helm deployments and four hours for the other operations. A failed wait leaves the resource in the state: fix the cause from the deployment logs, then apply again.

## Getting help

If `terraform plan` shows a change you did not expect after upgrading, open an issue on the [GitHub repository](https://github.com/qovery/terraform-provider-qovery/issues) with the plan output and the provider version. Never paste a state file or a plan that contains secrets.
