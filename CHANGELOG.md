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

### Added

- `qovery_cluster`: `features.karpenter.qovery_node_pools.gpu_override` manages the Karpenter
  GPU node pool, and the data source reports it. Declaring the block creates the pool and
  removing it deletes the pool; the plan warns when it removes the block. (QOV-2318)

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
