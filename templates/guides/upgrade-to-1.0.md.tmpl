---
page_title: "Upgrading to version 1.0 - Qovery Provider"
subcategory: ""
description: |-
  How to move from a 0.x release of the Qovery provider to 1.0: what breaks, what behaves differently, and the steps to migrate safely.
---

# Upgrading to version 1.0

Version 1.0 is the first stable release of the Qovery provider. It removes the last deprecated attribute, makes `terraform plan` reflect changes made from the Qovery Console on more attributes, and marks every secret-bearing attribute as sensitive.

This guide lists every change that can affect an existing configuration and the steps to migrate. The complete list of changes is in the [CHANGELOG](https://github.com/qovery/terraform-provider-qovery/blob/main/CHANGELOG.md).

## Before you start

1. Upgrade to the latest 0.x release (0.88 or later), run `terraform apply`, and make sure `terraform plan` reports no changes. The migration below relies on the state written by that release.
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

### `qovery_git_token` data source: `organization_id` is required

The data source now requires `organization_id`. In 0.x it was a computed attribute that the lookup could never populate, so the data source did not work. Add the argument:

```terraform
data "qovery_git_token" "my_git_token" {
  id              = "<git_token_id>"
  organization_id = "<organization_id>"
}
```

The `token` attribute is now `null` instead of an empty string: the Qovery API never returns the token value.

### Secret-bearing attributes are now sensitive

The following attributes are marked sensitive in 1.0:

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

## Behaviour changes

These changes need no configuration edit, but they can make `terraform plan` show differences that 0.x hid.

### `qovery_cluster`: refresh reflects Console changes on `routing_table` and `labels_group_ids`

In 0.x, `routing_table` and `labels_group_ids` only refreshed from the API when the state already held a value. A route or a labels group added from the Qovery Console stayed invisible to Terraform, so the plan did not reflect the real remote state.

In 1.0 the refresh is authoritative. A route or labels group added, changed or removed from the Console appears as a difference in `terraform plan`, and `terraform apply` reverts it to the configured value. To keep a Console-made change, add it to the configuration before applying. Corrective changes on these attributes redeploy the cluster, like any other change to them.

Two related fixes: `routing_table = []` now deletes the remaining remote routes (0.x skipped the API call when the list was empty), and an omitted `labels_group_ids` no longer hides labels groups attached from the Console. Declare `labels_group_ids` explicitly, with the ids you want or `[]`, to make Terraform authoritative.

### `qovery_cluster`: refresh reflects the dedicated cronjob node pool

In 0.x, the refresh only kept `features.karpenter.qovery_node_pools.cronjob_override` when the configuration declared it. A cronjob node pool enabled from the Qovery Console stayed invisible to Terraform, and the next `terraform apply` disabled it without the plan showing it. A pool disabled from the Console while the block was declared was re-enabled the same way.

In 1.0 the refresh stores `cronjob_override` exactly when the pool is enabled on the cluster. A pool enabled from the Console shows in `terraform plan` as the block being removed, and applying that plan disables the pool: declare `cronjob_override` to keep it. A pool disabled from the Console shows as the block being added back.

### `advanced_settings_json`: Console resets of tracked keys are reflected

On `qovery_cluster`, `qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm` and `qovery_terraform_service`, `advanced_settings_json` is desired state, not a mirror of the remote configuration:

- Refresh only reconciles keys already tracked in the Terraform state. A setting overridden only in the Console is not pulled into the state, so declaring it afterwards plans as an addition even if the remote value already matches.
- A Console change to a tracked key, including a reset to its default value, is now reflected on refresh and planned back to the configured value. In 0.x a reset was invisible.
- Removing a key from the JSON does not reset it remotely: omitted keys keep their current value. To reset a setting, set it to its default value explicitly.
- `terraform import` records every setting whose value differs from the default.

### Deployment waits fail on timeout

When a deployment, a deletion or a cluster operation does not complete before the provider's wait timeout, `terraform apply` now fails with a timeout error. In 0.x the wait returned success and the resource was reported as ready although it had not converged. The timeout is one hour for environment, container, job and Helm deployments and four hours for the other operations. A failed wait leaves the resource in the state: fix the cause from the deployment logs, then apply again.

## Getting help

If `terraform plan` shows a change you did not expect after upgrading, open an issue on the [GitHub repository](https://github.com/qovery/terraform-provider-qovery/issues) with the plan output and the provider version. Never paste a state file or a plan that contains secrets.
