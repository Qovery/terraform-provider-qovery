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

### Fixed

- `qovery_cluster`: a transient 5xx from the API while reading a cluster is now retried with
  backoff, instead of failing the `terraform plan` or `apply` with `Could not read cluster ...
  500 Internal Server Error`. (QOV-2356)

## [1.0.1] - 2026-10-01

### Fixed

- `qovery_cluster`: creating a `SELF_MANAGED` cluster works again. Since 1.0.0 the create
  deployed the new cluster, which the API refuses with `You cannot deploy a self-managed
  cluster`, and the cluster was created in Qovery but missing from the Terraform state. A
  create no longer forces a deploy: a new managed cluster is still deployed. (QOV-2355)

## [1.0.0] - 2026-10-01

1.0.0 is the first stable release of the provider. Read the
[1.0 upgrade guide](docs/guides/upgrade-to-1.0.md) before upgrading from a 0.x release.

### Breaking changes

- **`qovery_cluster`**: the global `features.karpenter.spot_enabled` is removed from the
  resource and the data source. A Karpenter node pool without `spot_enabled` now runs on
  on-demand instances: the per-node-pool `spot_enabled` defaults to `false`, and the provider
  sends an explicit value for every node pool instead of letting the API apply the global
  flag. A node pool that runs on spot instances without being declared shows up in
  `terraform plan`, with a warning, instead of being moved by the next apply without notice.
  The data source always reports the stable and default node pools. Pin your node pools on 1.0,
  before the first apply, as described in the upgrade guide. (QOV-2301)
- **`qovery_cluster`**: `routing_table` and `labels_group_ids` are managed as a whole. The
  refresh always reads the API, so a route or labels group added, changed or removed from the
  Console shows up in `terraform plan` even when the attribute is omitted, and the next apply
  reverts it. Omitting the attribute now means no route and no labels group: removing it from
  the configuration deletes the routes or detaches the labels groups, and `routing_table = []`
  deletes the remaining routes. Import and the data source report the routes and labels groups
  attached to the cluster. Add the routes and labels groups you manage from the Console to
  the configuration before upgrading. (QOV-2029)
- **`qovery_cluster`**: omitted `instance_type`, `disk_size`, `min_running_nodes` and
  `max_running_nodes` plan the Qovery default instead of keeping the last value, on the
  clusters whose node group Qovery sizes: `MANAGED` clusters on AWS without Karpenter
  (`t3.xlarge`), Scaleway (`DEV1-L`) and Azure (`Standard_DS2_v2`), with a 40 GB disk and 3 to
  10 nodes. The refresh reads the API, so a value changed from the Console shows up in
  `terraform plan` and the next apply reverts it, and removing one of these attributes plans
  the reset. Karpenter, GCP, self-managed and partially managed clusters ignore the four
  attributes: they keep the value the API reports, and the plan warns when the configuration
  sets one. Declare the node sizing you set from the Console before upgrading, and always
  declare `instance_type` when the cluster's instance type has another CPU architecture than
  the default. (QOV-2328)
- **`qovery_cluster`**: omitting `features` or `keda` plans their defaults instead of keeping
  the last value: the default VPC subnet, no static IP, no existing VPC, no Karpenter, no GKE
  KMS key, and KEDA disabled. A feature or KEDA changed from the Console shows up in
  `terraform plan`, and removing a block, or one of its attributes, plans the reset. A plan that
  removes `features.karpenter`, or turns `features.static_ip` from `true` to `false` on a
  deployed AWS, GCP or Azure cluster, now fails at plan time: the API rejects both, so 0.x
  failed at apply. Declare the features and KEDA you set from the Console before upgrading.
  (QOV-2328)
- **`qovery_cluster`**: `features.gke_kms_key` is Optional only, and setting, changing or
  removing it after creation is a plan error instead of a cluster replacement: the API only
  takes the key when it creates the cluster. (QOV-2328)
- **`qovery_cluster` data source**: `instance_type`, `disk_size`, `min_running_nodes` and
  `max_running_nodes` are read-only and report the API values, including the `KARPENTER` and
  `AUTO_PILOT` instance types and the placeholder node counts of Karpenter, GCP and
  self-managed clusters; remove them from data source configurations. (QOV-2328)
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
- **`qovery_job`**: `external_host` and `internal_host` are removed from the resource and the
  data source. Both were always null, because the Qovery API reports no host for a job. Remove
  any reference to them from the configuration; the state upgrade drops them, so an unchanged
  configuration plans nothing. (QOV-2342)
- **`qovery_blueprint`**: the refresh reads the API instead of keeping state values. It reports
  every variable whose value differs from its catalog default, so a variable set from the
  Console that the configuration omits shows up in `terraform plan` as a removal, and the next
  apply resets it to the default. 0.x only tracked the declared variables and never reset
  the others. `blueprint` is derived from the deployed tag, so a major version changed from the
  Console shows up as well. A deploy that failed outside Terraform no longer hides the settings
  it saved: only a failed Terraform apply keeps the last applied values, so that the next apply
  retries it. Declare the variables you set from the Console before upgrading. (QOV-2337)
- **`qovery_blueprint`**: changing `icon_uri` after creation is a plan error. The Qovery API
  applies the icon only when the blueprint is created, so 0.x reported the new icon while
  the service kept the old one. The refresh now reads the icon from the service the blueprint
  materialized, so an icon changed from the Console fails the plan until the configuration
  sets the same value. To change the icon, change it from the Console, then set the same value
  in the configuration. (QOV-2337)
- **`qovery_container_registry`, `qovery_helm_repository`, `qovery_custom_role`,
  `qovery_project`**: `description` defaults to `""`, the value the API stores when a request
  omits it. Removing the description from the configuration clears it, and a description set
  from the Console shows up in `terraform plan`. 0.x kept the last value. (QOV-2333)
- **`qovery_organization`, `qovery_git_token`**: `description` is optional only and omitting it
  means no description. Removing it from the configuration clears it, and a description set from
  the Console shows up in `terraform plan` as a removal. `qovery_git_token.bitbucket_workspace` is
  optional only too, and a `BITBUCKET` token without a workspace now fails at plan time instead of
  at apply. The state upgrade turns the empty description 0.x could store into null, so an
  unchanged configuration plans nothing. (QOV-2333)
- **`qovery_custom_role`**: the refresh records every permission that differs from the defaults
  (`VIEWER` on a cluster, `NO_ACCESS` on a project) on a cluster or project the configuration does
  not list. A permission granted from the Console now shows up in `terraform plan` as an entry to
  remove; 0.x reset it on the next apply without showing it. (QOV-2333)
- **`qovery_container_registry`, `qovery_helm_repository`**: the refresh reads the non-secret
  keys of `config` from the API: `region`, `access_key_id`, `username`, `scaleway_access_key`,
  `scaleway_project_id` and the GCP Workload Identity Federation keys, for the kinds that store
  them. A key changed from the Console shows up in `terraform plan`, and import records them.
  0.x kept the whole block from the state. Secrets are still taken from the state, because the
  API never returns them. (QOV-2333)
- **`qovery_aws_credentials`, `qovery_scaleway_credentials`, `qovery_gcp_credentials`,
  `qovery_eks_anywhere_vsphere_credentials`**: the refresh reads the identifiers from the API:
  `access_key_id`, `role_arn`, the Scaleway access key, project and organization, the GCP
  `service_account_email` and `workload_identity_provider_resource`, and `vsphere_user`. An
  identifier changed from the Console shows up in `terraform plan`, and import records them. 0.x
  kept them from the state. Secrets are still taken from the state. (QOV-2333)

### Added

- `qovery_cluster`: `features.karpenter.qovery_node_pools.gpu_override` manages the Karpenter
  GPU node pool, and the data source reports it. Declaring the block creates the pool and
  removing it deletes the pool; the plan warns when it removes the block. (QOV-2318)
- `qovery_cluster`: `consolidate_after` on every Karpenter node pool override, `consolidation`
  and `limits` on `cronjob_override`, and `disk_iops` and `disk_throughput` on
  `features.karpenter`. The data source reports them. Write `consolidate_after` in the largest
  whole unit, the form Qovery returns: `1h`, not `60m`. (QOV-2322)
- `qovery_blueprint` supports `terraform import` by blueprint ID. The data source reports
  `icon_uri`. (QOV-2337)
- The `qovery_cluster` data source reads `kubeconfig` for every cluster the API has one for,
  not only for `PARTIALLY_MANAGED` clusters. (QOV-2328)
- A "Managing changes" guide explains how the provider plans changes and lists the attributes
  that keep their value when removed. (QOV-2319)

### Changed

- `advanced_settings_json` on `qovery_cluster`, `qovery_application`, `qovery_container`,
  `qovery_job`, `qovery_helm` and `qovery_terraform_service` now reflects a Console-side
  change to a tracked key, including a reset to its default value, on refresh and plans it
  back to the configured value. The refresh semantics are documented in the "Managing
  changes" guide. (QOV-2028)
- The Registry documentation is rewritten for 1.0: shorter attribute descriptions shared by
  each resource and its data source, and minimal examples. (QOV-2319)
- `qovery_cluster`: the database and cache subnet lists of `features.existing_vpc` are
  Optional only, so omitting one means none. The state upgrade turns the empty lists 0.x
  stored into null, so an unchanged configuration plans nothing.
  `existing_vpc.eks_create_nodes_in_private_subnet` and `gcp_existing_vpc.private_nodes`
  default to `false`, the value the API stores when they are omitted. (QOV-2328)

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
- `qovery_cluster`: an apply no longer clears the Karpenter `consolidate_after`, the cronjob
  node pool consolidation and limits, or the Karpenter node disk IOPS and throughput set from
  the Console while the plan shows nothing. They now show in `terraform plan`; declare them to
  keep them. (QOV-2322)
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
- `qovery_organization`: an update no longer clears the organization's logo, website and
  icon URLs, and no longer fails with `Organization contact emails cannot be empty` when the
  organization has admin emails. The update resends the current values of the fields the
  provider does not manage. (QOV-2329)
- `qovery_job`: an update of a cron job no longer resets the timezone set from the Console to
  `Etc/UTC`. The provider does not manage the timezone and now sends back the one the job runs
  at. (QOV-2330)
- `qovery_cluster`: updating a cluster with KEDA enabled no longer resets the KEDA availability
  and resource profiles set from the Console to `NORMAL`. Terraform does not manage the
  profiles; the update now resends their current values. (QOV-2332)
- `qovery_terraform_service` data source: a read no longer fails with `Struct defines fields
  not found in object`, which every read hit since 0.60.0. The data source now reports
  `terraform_action` and `backend.blueprint`, and the `value` of a secret variable is `null`
  instead of the `SECRET_VALUE_UNCHANGED` placeholder the API returns in place of secret
  values. (QOV-2334)
- `qovery_deployment` data source: a read no longer fails with `invalid deployment desired
  state`. Qovery stores no deployment object, so the data source reads nothing and echoes its
  arguments; `environment_id` and `desired_state` are always `null`. (QOV-2334)
- `qovery_cluster`: a Scaleway, Azure or self-managed cluster whose configuration declares
  `features` can be created. The provider sent the VPC subnet feature, which the API only
  accepts when it creates an AWS `MANAGED` cluster, so the create failed with a 400 on the
  custom subnet feature. (QOV-2328)
- `qovery_cluster`: `features.gcp_existing_vpc.additional_ip_range_pods_names = []` no longer
  fails the apply with an inconsistent result. (QOV-2328)
- `qovery_helm_repository`: an `OCI_SCALEWAY_CR` repository now sends
  `config.scaleway_project_id`, which the API requires. The provider left it out, so creating
  such a repository failed with a 400. (QOV-2333)
- `qovery_cluster_dns_provider`: `terraform validate` and `terraform plan` accept a
  `cloudflare.api_token` or a `route53.credentials.aws_secret_access_key` that comes from a
  variable or another resource. The provider treated a value not yet known as missing, and
  failed with "must be set". (QOV-2319)
- `qovery_database`: changing `type` or `mode` of an existing database now fails at plan
  time. The API cannot change them, so the update left them unchanged and the same change
  showed in every later plan. To use another type or mode, recreate the database. (QOV-2343)
- `qovery_terraform_service`: a `timeout_seconds` below 60 now fails at plan time. The
  provider accepted it and the API rejected it at apply time with a 400. (QOV-2345)
- `qovery_scaleway_credentials`: the error for a malformed import identifier now gives the
  format `organization_id,scaleway_credentials_id`, the order the import reads. It gave the two
  parts in the reverse order. (QOV-2344)
- Validation errors for a minimum value now say `must be at least <minimum>`. They said
  `must be greater than <minimum>`, although the minimum itself is accepted, for example
  `timeout_seconds = 60` on `qovery_terraform_service`. (QOV-2348)
- `qovery_helm`: the first plan after `terraform import` no longer changes
  `values_override.set`, `set_string` and `set_json` from `{}` to `null` when the
  configuration omits them. An import now reads an empty map as `null`, so a configuration
  that declares `{}` shows that change once instead; applying it changes nothing. (QOV-2354)
- `qovery_cluster`: saving the Karpenter instances from the Console no longer shows in
  `terraform plan`. The Console rewrites the requirements of the node pools and of
  `gpu_override` in its own order, without duplicate values, which the plan showed as a
  difference that changed nothing. A value added or removed from the Console still shows.
  (QOV-2353)
- `qovery_application`, `qovery_container`, `qovery_job`: saving the service settings from the
  Console no longer plans `entrypoint = "" -> null`. The Console writes an empty entrypoint,
  which means the image's entrypoint, like an omitted one. (QOV-2353)
- `qovery_cluster`: the plan warns when it removes `cronjob_override`, which disables the
  Karpenter cronjob node pool, as it does for `gpu_override`. A cronjob node pool enabled from
  the Console showed in the plan as the block being removed, without a warning. (QOV-2353)

## [0.91.0] - 2026-09-29

### Added

- `qovery_application`, `qovery_job` and `qovery_terraform_service`: the `build_settings` block
  manages the build timeout, CPU, RAM, ephemeral storage, BuildKit cache and git submodules,
  and the data sources report it. Removing the block resets the build settings to their
  defaults. A plan that sets both the block and `build.*` keys in `advanced_settings_json`
  fails; the keys keep working when the block is not set. (QOV-2233, #631)

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

[Unreleased]: https://github.com/qovery/terraform-provider-qovery/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/qovery/terraform-provider-qovery/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.91.0...v1.0.0
[0.91.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.90.0...v0.91.0
[0.90.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.89.0...v0.90.0
[0.89.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.88.0...v0.89.0
[0.88.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.4...v0.88.0
[0.87.4]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.2...v0.87.4
[0.87.2]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.1...v0.87.2
[0.87.1]: https://github.com/qovery/terraform-provider-qovery/compare/v0.87.0...v0.87.1
[0.87.0]: https://github.com/qovery/terraform-provider-qovery/compare/v0.86.1...v0.87.0
