---
page_title: "Managing changes - Qovery Provider"
subcategory: ""
description: |-
  How the Qovery provider plans changes: the configuration is the source of truth, with a few documented exceptions.
---

# Managing changes

The configuration is the source of truth. `terraform plan` compares it with the values the Qovery API returns, and `terraform apply` makes Qovery match it.

## Changes made outside Terraform

The refresh reads each attribute from the Qovery API. A change made from the Qovery Console, the CLI or the API shows up in `terraform plan`, and the next `terraform apply` reverts it. To keep the change, copy it into the configuration.

The API never returns secrets, such as secret values, passwords, secret keys and JSON keys, nor the few other values their attribute description points out. Terraform keeps them as configured, so a change made to them outside Terraform does not show up in the plan.

## Removing an attribute

Removing an attribute from the configuration plans its removal:

- An attribute with a default plans the default. The resource pages list each default.
- A list, a set or a block without a default means none. Removing `labels_group_ids`, for example, detaches every labels group.

## Attributes that keep their value when removed

The Qovery API gives Terraform no way to plan the reset of a few attributes. Removing one of them from the configuration keeps its current value:

- `deployment_stage_id` on `qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm`, `qovery_terraform_service` and `qovery_database`: Qovery attaches every service to a deployment stage and cannot detach it.
- `ports.is_default` on `qovery_application`, `qovery_container` and `qovery_helm`: Qovery always marks one port as the default, and picks one when the configuration does not.
- `git_repository.branch` on `qovery_application` and `source.git_repository.branch` on `qovery_helm`: an omitted branch is the repository's default branch, which only Qovery resolves. Changing the repository URL while the branch is omitted resolves the default branch of the new repository.
- `blueprint_id` on `qovery_helm` and `qovery_terraform_service`: Qovery records it only when it creates the service, so changing it afterwards fails at plan time.
- `advanced_settings_json`, described below.

## Advanced settings

`advanced_settings_json` on `qovery_cluster`, `qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm` and `qovery_terraform_service` holds only the settings you override:

- Terraform tracks the keys the configuration sets. A setting overridden only from the Console stays out of the state, so declaring it later plans an addition even when the remote value already matches.
- A change made outside Terraform to a tracked key, including a reset to its default, shows up in the plan, and `terraform apply` sets it back.
- Removing a key keeps its current value. To reset a setting, set it to its default value.
- `terraform import` records every setting whose value differs from its default. A key the configuration sets to its default value is therefore missing from the imported state: it shows in the first plan after the import, and applying that plan changes nothing in Qovery.

The [Qovery API documentation](https://api-doc.qovery.com) lists the settings of each service with their defaults.

## Importing resources

`terraform import` records the values the API returns, secrets excepted. Run `terraform plan` after an import: it shows the differences between the configuration and the imported values, including the secrets, which the plan sets once.

## Data sources

Data sources report the values the API returns.
