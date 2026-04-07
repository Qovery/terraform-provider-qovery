# qovery_application (Resource)

Provides a Qovery application resource. This can be used to create and manage Qovery applications.

An application is a service built from source code in a git repository. Qovery builds the application using either Docker (with a Dockerfile) or Buildpacks, then deploys it to your cluster.


## Example

<div class="alert alert-info">
  <i style="font-size:24px" class="fa">&#xf05a;</i> If you're not familiar with Terraform or just want more examples, you can configure everything you need directly from the <a href="https://console.qovery.com">Qovery console</a>. Then, use our <a href="https://www.qovery.com/docs/terraform-provider/exporter">Terraform exporter</a> feature to generate the corresponding Terraform code.
</div><br />

```terraform
resource "qovery_application" "my_application" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyApplication"
  git_repository = {
    url       = "https://github.com/my-org/my-app.git"
    branch    = "main" # Optional (defaults to main or master)
    root_path = "/"    # Optional (defaults to "/", useful for monorepos)
  }

  # Build configuration
  build_mode      = "DOCKER"     # DOCKER or BUILDPACKS
  dockerfile_path = "Dockerfile" # Required when build_mode = "DOCKER"

  # Optional
  auto_preview          = false
  auto_deploy           = true
  cpu                   = 500
  memory                = 512
  min_running_instances = 1
  max_running_instances = 3
  entrypoint            = "/bin/sh"
  arguments             = ["-c", "start-server"]

  # Port configuration
  ports = [
    {
      internal_port       = 8080
      external_port       = 443
      publicly_accessible = true
      protocol            = "HTTP"
      is_default          = true
      name                = "http"
    },
    {
      internal_port       = 9090
      publicly_accessible = false
      protocol            = "HTTP"
      name                = "metrics"
    }
  ]

  # Healthchecks
  healthchecks = {
    readiness_probe = {
      type = {
        http = {
          port   = 8080
          path   = "/ready"
          scheme = "HTTP"
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }

    liveness_probe = {
      type = {
        http = {
          port   = 8080
          path   = "/health"
          scheme = "HTTP"
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
  }

  # Environment variables
  environment_variables = [
    {
      key   = "APP_PORT"
      value = "8080"
    }
  ]
  environment_variable_aliases = [
    {
      key = "PORT"
      # The value of the alias must be the name of the aliased variable.
      # Here it creates an alias "PORT" pointing to the "APP_PORT" variable above.
      value = "APP_PORT"
    }
  ]
  environment_variable_overrides = [
    {
      # The key must match a variable defined at a higher scope (project or environment).
      key   = "SOME_PROJECT_VARIABLE"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  # Environment variable files (mounted as files in the container)
  environment_variable_files = [
    {
      key        = "APP_CONFIG"
      value      = "config-content"
      mount_path = "/etc/app/config.yaml"
    }
  ]

  # Secrets
  secrets = [
    {
      key   = "SECRET_KEY"
      value = "SECRET_VALUE"
    }
  ]
  secret_aliases = [
    {
      key = "SECRET_KEY_ALIAS"
      # The value of the alias must be the name of the aliased secret.
      value = "SECRET_KEY"
    }
  ]
  secret_overrides = [
    {
      # The key must match a secret defined at a higher scope (project or environment).
      key   = "SOME_PROJECT_SECRET"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  # Secret files (mounted as files, value is encrypted)
  secret_files = [
    {
      key        = "API_KEY"
      value      = "secret-value"
      mount_path = "/usr/local/secrets/api-key"
    }
  ]

  # Custom domains
  custom_domains = [
    {
      domain               = "app.example.com"
      generate_certificate = true
    }
  ]

  # Deployment restrictions (only deploy when specific paths change)
  deployment_restrictions = [
    {
      mode  = "MATCH"
      type  = "PATH"
      value = "src/"
    }
  ]

  # Advanced settings (JSON)
  advanced_settings_json = jsonencode({
    # Non-exhaustive list. Full list: https://api-doc.qovery.com/#tag/Applications/operation/getDefaultApplicationAdvancedSettings
    "network.ingress.proxy_buffer_size_kb" : 8,
    "network.ingress.keepalive_time_seconds" : 1000,
  })

  depends_on = [
    qovery_environment.my_environment
  ]
}
```

You can find complete examples within these repositories:
* [Deploy an Application and Database within 3 environments](https://github.com/Qovery/terraform-examples/tree/main/examples/deploy-an-application-within-3-environments)
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) Id of the environment. Changing this forces the application to be re-created.
- `git_repository` (Attributes) Git repository configuration for the application source code. (see [below for nested schema](#nestedatt--git_repository))
- `healthchecks` (Attributes) Configuration for the healthchecks that are going to be executed against your service. At least one of `readiness_probe` or `liveness_probe` should be configured for production workloads. (see [below for nested schema](#nestedatt--healthchecks))
- `name` (String) Name of the application.

### Optional

- `advanced_settings_json` (String) Advanced settings as JSON. Use `jsonencode()` to set values. Only include settings you want to override. Full list available in [Qovery API documentation](https://api-doc.qovery.com/#tag/Applications/operation/getDefaultApplicationAdvancedSettings).
- `annotations_group_ids` (Set of String) List of annotations group ids. Annotations groups allow you to add Kubernetes annotations to the application's pods.
- `arguments` (List of String) List of arguments of this application. Overrides the Docker image's default `CMD`.
- `auto_deploy` (Boolean) Specify if the application will be automatically redeployed after receiving a new commit on the configured branch.
- `auto_preview` (Boolean) Specify if the environment preview option is activated or not for this application. When enabled, Qovery creates a preview environment for each pull request. Default: `false`.
- `build_mode` (String) Build mode of the application.
  - `DOCKER`: Build using a Dockerfile in the repository. Requires `dockerfile_path` to be set.
  - `BUILDPACKS`: Build using Cloud Native Buildpacks (auto-detects language and framework).

Default: `DOCKER`.
- `cpu` (Number) CPU of the application in millicores (m) [1000m = 1 CPU].
- `custom_domains` (Attributes Set) List of custom domains linked to this application. You must configure a CNAME record on your DNS provider pointing to the `validation_domain` value. (see [below for nested schema](#nestedatt--custom_domains))
- `deployment_restrictions` (Attributes Set) List of deployment restrictions. Deployment restrictions allow you to control when an application is deployed based on file path changes in the git repository. (see [below for nested schema](#nestedatt--deployment_restrictions))
- `deployment_stage_id` (String) Id of the deployment stage. Deployment stages allow you to control the order in which services are deployed within an environment.
- `docker_target_build_stage` (String) The target build stage in a multi-stage Dockerfile to build. Only applicable when `build_mode = "DOCKER"` and using a multi-stage Dockerfile.
- `dockerfile_path` (String) Path to the Dockerfile relative to the `git_repository.root_path`. Required when `build_mode = "DOCKER"`. Example: `Dockerfile` or `docker/Dockerfile.prod`.
- `entrypoint` (String) Entrypoint of the application. Overrides the Docker image's default `ENTRYPOINT`.
- `environment_variable_aliases` (Attributes Set) List of environment variable aliases linked to this application. (see [below for nested schema](#nestedatt--environment_variable_aliases))
- `environment_variable_files` (Attributes Set) List of environment variable files linked to this application. (see [below for nested schema](#nestedatt--environment_variable_files))
- `environment_variable_overrides` (Attributes Set) List of environment variable overrides linked to this application. (see [below for nested schema](#nestedatt--environment_variable_overrides))
- `environment_variables` (Attributes Set) List of environment variables linked to this application. (see [below for nested schema](#nestedatt--environment_variables))
- `icon_uri` (String) Icon URI representing the application. Used in the Qovery console UI.
- `is_skipped` (Boolean) If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.
- `labels_group_ids` (Set of String) List of labels group ids. Labels groups allow you to add Kubernetes labels to the application's pods.
- `max_running_instances` (Number) Maximum number of instances running for the application.
- `memory` (Number) RAM of the application in MB [1024MB = 1GB].
- `min_running_instances` (Number) Minimum number of instances running for the application.
- `ports` (Attributes List) List of ports linked to this application. At least one port must be set as `publicly_accessible = true` with an `external_port` for the application to be reachable from the internet. (see [below for nested schema](#nestedatt--ports))
- `secret_aliases` (Attributes Set) List of secret aliases linked to this application. (see [below for nested schema](#nestedatt--secret_aliases))
- `secret_files` (Attributes Set) List of secret files linked to this application. (see [below for nested schema](#nestedatt--secret_files))
- `secret_overrides` (Attributes Set) List of secret overrides linked to this application. (see [below for nested schema](#nestedatt--secret_overrides))
- `secrets` (Attributes Set) List of secrets linked to this application. (see [below for nested schema](#nestedatt--secrets))
- `storage` (Attributes Set) List of persistent storage volumes linked to this application. Data stored in these volumes persists across application restarts. (see [below for nested schema](#nestedatt--storage))

### Read-Only

- `built_in_environment_variables` (Attributes List) List of built-in environment variables linked to this application. Built-in variables are automatically generated by Qovery and include host information, port mappings, and other service metadata. These are read-only and cannot be modified. (see [below for nested schema](#nestedatt--built_in_environment_variables))
- `external_host` (String) The application external FQDN host. Only available if your application has at least one publicly accessible port.
- `id` (String) Id of the application.
- `internal_host` (String) The application internal host. Use this to communicate between services within the same environment.

<a id="nestedatt--git_repository"></a>
### Nested Schema for `git_repository`

Required:

- `url` (String) URL of the git repository (e.g. `https://github.com/my-org/my-app.git`).

Optional:

- `branch` (String) Branch of the git repository to use for builds. Defaults to `main` or `master` (depending on repository).
- `git_token_id` (String) The git token ID to be used for authenticating with the git provider. Required for private repositories. Reference a `qovery_git_token` resource.
- `root_path` (String) Root path of the application within the repository. Useful for monorepos where the application code is in a subdirectory. Defaults to `/`.


<a id="nestedatt--healthchecks"></a>
### Nested Schema for `healthchecks`

Optional:

- `liveness_probe` (Attributes) Configuration for the liveness probe, used to determine when your service is working correctly. If the liveness probe fails, the service container is killed and restarted. (see [below for nested schema](#nestedatt--healthchecks--liveness_probe))
- `readiness_probe` (Attributes) Configuration for the readiness probe, used to determine when your service is ready to receive traffic. If the readiness probe fails, the service is temporarily removed from the load balancer until it passes again. (see [below for nested schema](#nestedatt--healthchecks--readiness_probe))

<a id="nestedatt--healthchecks--liveness_probe"></a>
### Nested Schema for `healthchecks.liveness_probe`

Required:

- `failure_threshold` (Number) Number of consecutive failures required to declare the probe as failed.
- `initial_delay_seconds` (Number) Number of seconds to wait after the container starts before the first probe is executed. Use this to give your application time to initialize.
- `period_seconds` (Number) How often (in seconds) to perform the probe after the initial delay.
- `success_threshold` (Number) Minimum consecutive successes for the probe to be considered successful after a failure.
- `timeout_seconds` (Number) Number of seconds after which the probe times out. If the probe does not respond within this time, it is considered failed.
- `type` (Attributes) Kind of check to run for this probe. Exactly one of `tcp`, `http`, `grpc`, or `exec` must be configured. (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type))

<a id="nestedatt--healthchecks--liveness_probe--type"></a>
### Nested Schema for `healthchecks.liveness_probe.type`

Optional:

- `exec` (Attributes) Exec probe: runs a command inside the container. The probe succeeds if the command exits with status code 0. The command binary must be present in the container image. (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--exec))
- `grpc` (Attributes) gRPC probe: checks that the given port responds to gRPC health check requests. The service must implement the [gRPC Health Checking Protocol](https://kubernetes.io/blog/2018/10/01/health-checking-grpc-servers-on-kubernetes/#introducing-grpc-health-probe). (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--grpc))
- `http` (Attributes) HTTP probe: sends an HTTP GET request and expects a 2xx response code. (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--http))
- `tcp` (Attributes) TCP probe: checks that a TCP connection can be established on the given port. (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--tcp))

<a id="nestedatt--healthchecks--liveness_probe--type--exec"></a>
### Nested Schema for `healthchecks.liveness_probe.type.exec`

Required:

- `command` (List of String) The command and its arguments to execute (e.g. `["cat", "/tmp/healthy"]`).


<a id="nestedatt--healthchecks--liveness_probe--type--grpc"></a>
### Nested Schema for `healthchecks.liveness_probe.type.grpc`

Required:

- `port` (Number) The port number to try to connect to.

Optional:

- `service` (String) The gRPC service name to health-check. If not specified, the overall server health is checked.


<a id="nestedatt--healthchecks--liveness_probe--type--http"></a>
### Nested Schema for `healthchecks.liveness_probe.type.http`

Required:

- `port` (Number) The port number to try to connect to.
- `scheme` (String) Scheme to use for the HTTP request. Must be `HTTP` or `HTTPS`.

Optional:

- `path` (String) The path for the HTTP GET request (e.g. `/health`, `/ready`). Defaults to `/`.


<a id="nestedatt--healthchecks--liveness_probe--type--tcp"></a>
### Nested Schema for `healthchecks.liveness_probe.type.tcp`

Required:

- `port` (Number) The port number to try to connect to.

Optional:

- `host` (String) Optional host to connect to. Defaults to the pod IP if not specified.




<a id="nestedatt--healthchecks--readiness_probe"></a>
### Nested Schema for `healthchecks.readiness_probe`

Required:

- `failure_threshold` (Number) Number of consecutive failures required to declare the probe as failed.
- `initial_delay_seconds` (Number) Number of seconds to wait after the container starts before the first probe is executed. Use this to give your application time to initialize.
- `period_seconds` (Number) How often (in seconds) to perform the probe after the initial delay.
- `success_threshold` (Number) Minimum consecutive successes for the probe to be considered successful after a failure.
- `timeout_seconds` (Number) Number of seconds after which the probe times out. If the probe does not respond within this time, it is considered failed.
- `type` (Attributes) Kind of check to run for this probe. Exactly one of `tcp`, `http`, `grpc`, or `exec` must be configured. (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type))

<a id="nestedatt--healthchecks--readiness_probe--type"></a>
### Nested Schema for `healthchecks.readiness_probe.type`

Optional:

- `exec` (Attributes) Exec probe: runs a command inside the container. The probe succeeds if the command exits with status code 0. The command binary must be present in the container image. (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--exec))
- `grpc` (Attributes) gRPC probe: checks that the given port responds to gRPC health check requests. The service must implement the [gRPC Health Checking Protocol](https://kubernetes.io/blog/2018/10/01/health-checking-grpc-servers-on-kubernetes/#introducing-grpc-health-probe). (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--grpc))
- `http` (Attributes) HTTP probe: sends an HTTP GET request and expects a 2xx response code. (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--http))
- `tcp` (Attributes) TCP probe: checks that a TCP connection can be established on the given port. (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--tcp))

<a id="nestedatt--healthchecks--readiness_probe--type--exec"></a>
### Nested Schema for `healthchecks.readiness_probe.type.exec`

Required:

- `command` (List of String) The command and its arguments to execute (e.g. `["cat", "/tmp/healthy"]`).


<a id="nestedatt--healthchecks--readiness_probe--type--grpc"></a>
### Nested Schema for `healthchecks.readiness_probe.type.grpc`

Required:

- `port` (Number) The port number to try to connect to.

Optional:

- `service` (String) The gRPC service name to health-check. If not specified, the overall server health is checked.


<a id="nestedatt--healthchecks--readiness_probe--type--http"></a>
### Nested Schema for `healthchecks.readiness_probe.type.http`

Required:

- `port` (Number) The port number to try to connect to.
- `scheme` (String) Scheme to use for the HTTP request. Must be `HTTP` or `HTTPS`.

Optional:

- `path` (String) The path for the HTTP GET request (e.g. `/health`, `/ready`). Defaults to `/`.


<a id="nestedatt--healthchecks--readiness_probe--type--tcp"></a>
### Nested Schema for `healthchecks.readiness_probe.type.tcp`

Required:

- `port` (Number) The port number to try to connect to.

Optional:

- `host` (String) Optional host to connect to. Defaults to the pod IP if not specified.





<a id="nestedatt--custom_domains"></a>
### Nested Schema for `custom_domains`

Required:

- `domain` (String) Your custom domain (e.g. `app.example.com`).

Optional:

- `generate_certificate` (Boolean) Qovery will generate and manage a TLS/SSL certificate for this domain using Let's Encrypt.
- `use_cdn` (Boolean) Indicates if the custom domain is behind a CDN (e.g. Cloudflare). This affects how Qovery validates the CNAME during deployment:
  - If `true`: Qovery only checks that the domain points to an IP.
  - If `false`: Qovery checks that the domain resolves to the correct service Load Balancer.

Read-Only:

- `id` (String) Id of the custom domain.
- `status` (String) Status of the custom domain.
- `validation_domain` (String) URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.


<a id="nestedatt--deployment_restrictions"></a>
### Nested Schema for `deployment_restrictions`

Required:

- `mode` (String) Restriction mode. `MATCH`: deploy only when changes match the value. `EXCLUDE`: deploy only when changes do NOT match the value.
- `type` (String) Type of deployment restriction. Currently only `PATH` is supported.
- `value` (String) Value of the deployment restriction (e.g. a file path pattern like `src/` or `services/api/`).

Read-Only:

- `id` (String) Id of the deployment restriction.


<a id="nestedatt--environment_variable_aliases"></a>
### Nested Schema for `environment_variable_aliases`

Required:

- `key` (String) Name of the environment variable alias.
- `value` (String) Name of the variable to alias.

Optional:

- `description` (String) Description of the environment variable alias.

Read-Only:

- `id` (String) Id of the environment variable alias.


<a id="nestedatt--environment_variable_files"></a>
### Nested Schema for `environment_variable_files`

Required:

- `key` (String) Key of the environment variable file.
- `mount_path` (String) Mount path of the environment variable file.
- `value` (String) Value of the environment variable file.

Optional:

- `description` (String) Description of the environment variable file.

Read-Only:

- `id` (String) Id of the environment variable file.


<a id="nestedatt--environment_variable_overrides"></a>
### Nested Schema for `environment_variable_overrides`

Required:

- `key` (String) Name of the environment variable override.
- `value` (String) Value of the environment variable override.

Optional:

- `description` (String) Description of the environment variable override.

Read-Only:

- `id` (String) Id of the environment variable override.


<a id="nestedatt--environment_variables"></a>
### Nested Schema for `environment_variables`

Required:

- `key` (String) Key of the environment variable.
- `value` (String) Value of the environment variable.

Optional:

- `description` (String) Description of the environment variable.

Read-Only:

- `id` (String) Id of the environment variable.


<a id="nestedatt--ports"></a>
### Nested Schema for `ports`

Required:

- `internal_port` (Number) Internal port of the application. Must be between 1 and 65535.
- `publicly_accessible` (Boolean) Specify if the port is exposed to the world or not for this application.

Optional:

- `external_port` (Number) External port of the application. Required if `ports.publicly_accessible = true`. Must be between 1 and 65535.
- `is_default` (Boolean) If this port will be used for the root domain. The API may override this value based on port configuration (e.g., when only one publicly accessible port exists, it will be set as default).
- `name` (String) Name of the port.
- `protocol` (String) Protocol used for the port of the application.

Read-Only:

- `id` (String) Id of the port.


<a id="nestedatt--secret_aliases"></a>
### Nested Schema for `secret_aliases`

Required:

- `key` (String) Name of the secret alias.
- `value` (String) Name of the secret to alias.

Optional:

- `description` (String) Description of the secret alias.

Read-Only:

- `id` (String) Id of the secret alias.


<a id="nestedatt--secret_files"></a>
### Nested Schema for `secret_files`

Required:

- `key` (String) Key of the secret file.
- `mount_path` (String) Mount path of the secret file.
- `value` (String, Sensitive) Value of the secret file.

Optional:

- `description` (String) Description of the secret file.

Read-Only:

- `id` (String) Id of the secret file.


<a id="nestedatt--secret_overrides"></a>
### Nested Schema for `secret_overrides`

Required:

- `key` (String) Name of the secret override.
- `value` (String, Sensitive) Value of the secret override. The value is write-only and will not be displayed in plan outputs.

Optional:

- `description` (String) Description of the secret override.

Read-Only:

- `id` (String) Id of the secret override.


<a id="nestedatt--secrets"></a>
### Nested Schema for `secrets`

Required:

- `key` (String) Key of the secret.
- `value` (String, Sensitive) Value of the secret. The value is write-only and will not be displayed in plan outputs.

Optional:

- `description` (String) Description of the secret.

Read-Only:

- `id` (String) Id of the secret.


<a id="nestedatt--storage"></a>
### Nested Schema for `storage`

Required:

- `mount_point` (String) Mount point of the storage for the application.
- `size` (Number) Size of the storage for the application in GB [1024MB = 1GB].
- `type` (String) Type of the storage for the application.

Read-Only:

- `id` (String) Id of the storage.


<a id="nestedatt--built_in_environment_variables"></a>
### Nested Schema for `built_in_environment_variables`

Read-Only:

- `description` (String) Description of the environment variable.
- `id` (String) Id of the environment variable.
- `key` (String) Key of the environment variable.
- `value` (String) Value of the environment variable.
## Import
```shell
terraform import qovery_application.my_application "<application_id>"
```