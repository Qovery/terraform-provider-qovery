# Changelog

All notable changes to the Qovery Terraform provider are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Each entry
references the pull request or the `QOV-XXXX` ticket that introduced it. Changes released
before 0.87.0 are not listed here; see the
[GitHub releases](https://github.com/qovery/terraform-provider-qovery/releases) and the
commit history.

Breaking changes are listed first in each release, under **Breaking changes**, and are
explained in the matching upgrade guide.

## [Unreleased]

The next release is **1.0.0**, the first stable release of the provider. Read the
[1.0 upgrade guide](docs/guides/upgrade-to-1.0.md) before upgrading from a 0.x release.

### Breaking changes

- **`qovery_cluster`**: the global `features.karpenter.spot_enabled` is removed from the
  resource and the data source. A Karpenter node pool without `spot_enabled` now runs on
  on-demand instances: the per-node-pool `spot_enabled` defaults to `false`, and the provider
  sends an explicit value for every node pool instead of letting the API apply the global
  flag. A node pool that runs on spot instances without being declared shows up in
  `terraform plan`, with a warning, instead of being moved by the next apply without notice.
  The data source always reports the stable and default node pools. Pin your node pools as
  described in the upgrade guide before upgrading. (QOV-2301)
- **`qovery_cluster`**: `routing_table` and `labels_group_ids` are managed as a whole. The
  refresh always reads the API, so a route or labels group added, changed or removed from the
  Console shows up in `terraform plan` even when the attribute is omitted, and the next apply
  reverts it. Omitting the attribute now means no route and no labels group: removing it from
  the configuration deletes the routes or detaches the labels groups, and `routing_table = []`
  deletes the remaining routes. Import and the data source report the routes and labels groups
  attached to the cluster. Add the routes and labels groups you manage from the Console to
  the configuration before upgrading. (QOV-2029)
- **`qovery_application`, `qovery_container`, `qovery_job`, `qovery_database`**:
  `labels_group_ids` and `annotations_group_ids` are managed as a whole. The refresh always
  reads the API, so a group attached or detached from the Console shows up in `terraform plan`
  even when the attribute is omitted, and the next apply reverts it. 0.x ignored such a group
  and detached it on the next update without showing it in the plan. Omitting the attribute
  means no group: removing it from the configuration plans the detach. Import records the
  attached groups. In the four data sources both attributes are now read-only and report the
  groups attached to the service; remove them from data source configurations. Add the groups
  you attach from the Console to the configuration before upgrading. (QOV-2326)
- **`qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm`,
  `qovery_environment`, `qovery_project`**: the `description` of environment variables,
  aliases, overrides, files, secrets and secret files is read from the API. A description set
  from the Console now shows up in `terraform plan`, and the next apply clears it when the
  configuration omits it. 0.x kept any description the state did not hold out of the state.
  Import and the data sources report the descriptions. Secret values, and the mount path of
  secret files where the API does not return it, are still taken from the state. Add the
  descriptions you set from the Console to the configuration before upgrading. (QOV-2326)
- **`qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm`, `qovery_database`**:
  omitted service attributes plan the Qovery default instead of keeping the last value:
  `icon_uri` (the icon of the service type), `auto_deploy` (`true`; `false` on `qovery_helm`,
  the value 0.x sent), `auto_preview` (`false`), `ephemeral_storage` (`0`, none),
  `ports.name` (`p<internal_port>`), `ports.protocol` (`HTTP`), the `custom_domains` flags
  `generate_certificate` and `use_cdn` (`false`), and on `qovery_job`
  `schedule.lifecycle_type` (`GENERIC` for a lifecycle job) and
  `source.docker.git_repository.root_path` (`/`, and `""` is now rejected). The refresh reads
  the API, so a value changed from the Console shows up in `terraform plan` and the next apply
  reverts it, and removing one of these attributes from the configuration plans the reset. 0.x
  kept such values out of the plan. Declare the values you set from the Console before
  upgrading. (QOV-2327)
- **`qovery_application`, `qovery_container`, `qovery_job`, `qovery_helm`**: `arguments` on
  applications and containers, the four `schedule` entrypoints of `qovery_job` and both
  `git_token_id` of `qovery_helm` are Optional only: omitting one means none, so an argument
  list, entrypoint or git token set from the Console now shows up in `terraform plan` as a
  removal. The state upgrade turns the empty `arguments` list 0.x stored into null, so an
  unchanged configuration plans nothing. (QOV-2327)
- **`qovery_terraform_service`**: the refresh reads the value of non-secret `variables` from
  the API, so a value changed from the Console shows up in `terraform plan`. 0.x kept the
  state value. Secret values are still taken from the state, because the API does not return
  them. (QOV-2327)
- **`qovery_database`**: a `MANAGED` database without `instance_type` now fails at plan time
  instead of at apply, and a `CONTAINER` database that sets it gets a warning, because the API
  ignores it there. (QOV-2327)
- **`qovery_helm`, `qovery_terraform_service`, `qovery_job`**: changing `blueprint_id`, or the
  `schedule.lifecycle_type` of a job, after creation is a plan error. The API records a
  blueprint only on create and rejects a lifecycle type change, so 0.x planned the change and
  then failed at apply. A job created with a lifecycle type other than `GENERIC` must declare
  it, otherwise the plan fails on the reset to the default. (QOV-2327)
- **`qovery_blueprint`**: the refresh reads the API instead of keeping state values. It
  reports every variable whose value differs from its catalog default, so a variable set from
  the Console that the configuration omits shows up in `terraform plan` as a removal, and the
  next apply resets it to the default. 0.90.0 only tracked the declared variables and never reset the others. `blueprint`
  is derived from the deployed tag, so a major version changed from the Console shows up as
  well. A deploy that failed outside Terraform no longer hides the settings it saved: only a
  failed Terraform apply keeps the last applied values, so that the next apply retries it.
  Declare the variables you set from the Console before upgrading. (QOV-2337)
- **`qovery_blueprint`**: changing `icon_uri` after creation is a plan error. The Qovery API
  applies the icon only when the blueprint is created, so 0.90.0 reported the new icon while
  the service kept the old one. The refresh now reads the icon from the service the blueprint
  materialized, so an icon changed from the Console fails the plan until the configuration
  sets the same value. To change the icon, change it from the Console, then set the same value
  in the configuration. (QOV-2337)

### Added

- `qovery_cluster`: `features.karpenter.qovery_node_pools.gpu_override` manages the Karpenter
  GPU node pool, and the data source reports it. Declaring the block creates the pool and
  removing it deletes the pool; the plan warns when it removes the block. (QOV-2318)
- `qovery_blueprint` supports `terraform import` by blueprint ID. The data source reports
  `icon_uri`. (QOV-2337)

### Changed

- `advanced_settings_json` on `qovery_cluster`, `qovery_application`, `qovery_container`,
  `qovery_job`, `qovery_helm` and `qovery_terraform_service` now reflects a Console-side
  change to a tracked key, including a reset to its default value, on refresh and plans it
  back to the configured value. The refresh semantics are documented on the attribute.
  (QOV-2028)

### Fixed

- `qovery_cluster`: a configuration that sets `spot_enabled` on only some Karpenter node pools
  no longer moves the other node pools to spot instances on the next unrelated apply. A spot
  change made from the Console on a declared node pool now shows in `terraform plan`.
  (QOV-2301)
- `qovery_cluster`: a dedicated cronjob node pool enabled or disabled from the Console now shows
  in `terraform plan`. The refresh used to ignore it, so the next apply reverted the change
  without the plan showing it. (QOV-2301)
- `qovery_cluster`: an apply no longer deletes a GPU node pool created from the Console while
  the plan shows nothing. The pool now shows in `terraform plan` as `gpu_override` being
  removed; declare the block to keep it. (QOV-2318)
- `qovery_application`: changing or removing the `description` of a secret, secret alias or
  secret override now reaches the API. The update skipped a change that touched only the
  description and otherwise resent the previous description, so the apply failed with an
  inconsistent result. (QOV-2326)
- `qovery_helm`: a port declared without `protocol` no longer fails the apply; it defaults
  to `HTTP`. `values_override.set = {}` with `set_string` unset no longer fails with an
  inconsistent result. (QOV-2327)
- `qovery_application`, `qovery_helm`: an unrelated change no longer resets
  `git_repository.branch` to the repository's default branch when the configuration omits
  it; the current branch is kept, as documented in the upgrade guide. (QOV-2327)
- `qovery_database`: an update no longer clears the description set from the Console. The
  resource does not manage the description, so the update now resends the current one.
  (QOV-2331)

## [0.90.0] - 2026-09-28

### Added

- `qovery_blueprint` resource and data source: a service instantiated from the Qovery service
  catalog, such as a managed database, materialized as a Terraform or Helm service.
  (QOV-2241, #630)

## [0.89.0] - 2026-09-23

### Fixed

- `qovery_git_token` data source: `organization_id` is now a required argument. It used to be
  a computed attribute that the lookup could never populate, so the data source could not
  resolve a token. The `token` attribute is now `null` instead of an empty string, because the
  Qovery API never returns it. (QOV-2031, #627)
- Waiting for a deployment, a deletion or a cluster operation now fails with a timeout error
  when the operation does not complete in time. Previously the wait returned success and
  `terraform apply` reported a resource as ready although it had not converged.
  (QOV-2299, #626)

### Security

- Secret-bearing attributes are now sensitive: `qovery_database.password` (resource and data
  source), `qovery_container_registry.config.{password,secret_access_key,scaleway_secret_key}`
  and `qovery_helm_repository.config.{password,secret_access_key,scaleway_secret_key}`.
  Terraform hides them in plan output and refuses an `output` that exposes one of them unless
  the output is declared `sensitive = true`. (QOV-2298, #625)
- The Helm values sent to the API are no longer written to the provider log. (QOV-2298, #625)

## [0.88.0] - 2026-09-21

### Fixed

- A service whose creation partially succeeded (the API created it but a later step of the
  create failed) is now recorded in the Terraform state instead of being orphaned, so the
  next apply updates it rather than creating a duplicate. Applies to `qovery_application`,
  `qovery_container`, `qovery_job`, `qovery_helm`, `qovery_terraform_service` and
  `qovery_deployment_stage`. (#624)

## [0.87.4] - 2026-09-16

### Fixed

- `qovery_terraform_service`: variable and `tfvar` path validation errors now name the
  offending variable or path. (QOV-2213)

### Security

- Bumped `google.golang.org/grpc` to 1.83.2 and `golang.org/x/net` to 0.59.0 to fix
  reported vulnerabilities. (#622)

## [0.87.2] - 2026-08-28

### Fixed

- Updated the Qovery API client so the new ongoing service status introduced by the API is
  recognised while waiting for a deployment. (#618)

## [0.87.1] - 2026-08-26

### Fixed

- Cancelling a run while the provider waits for an environment deployment now aborts the
  wait with a context error instead of hanging and keeping the state locked. (#617)

## [0.87.0] - 2026-08-20

### Added

- `qovery_cluster`: spot instances are configured per Karpenter node pool with `spot_enabled`
  on `features.karpenter.qovery_node_pools.stable_override`, `default_override` and
  `cronjob_override`. Declaring `cronjob_override` enables the dedicated cronjob node pool.
  The `qovery_cluster` data source exposes the same per-pool values. (QOV-2158, #616)
- The `AGENTIC_WORKFLOW` environment variable scope is accepted when parsing API responses.
  (QOV-2168, #616)

### Deprecated

- `qovery_cluster` `features.karpenter.spot_enabled` (resource and data source). The global
  flag is replaced by the per-node-pool values above and is removed in 1.0. (QOV-2158, #616)

### Changed

- Go toolchain upgraded to 1.26.6 (QOV-2149, #615) and `google.golang.org/grpc` to 1.82.1
  (#614).

[Unreleased]: https://github.com/qovery/terraform-provider-qovery/compare/v0.89.0...HEAD
[0.89.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.88.0...v0.89.0
[0.88.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.4...v0.88.0
[0.87.4]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.2...v0.87.4
[0.87.2]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.1...v0.87.2
[0.87.1]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.0...v0.87.1
[0.87.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.86.1...v0.87.0
