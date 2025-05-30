# qovery_container (Resource)

Provides a Qovery container resource. This can be used to create and manage Qovery container registry.


## Example

<div class="alert alert-info">
  <i style="font-size:24px" class="fa">&#xf05a;</i> If you're not familiar with Terraform or just want more examples, you can configure everything you need directly from the <a href="https://console.qovery.com">Qovery console</a>. Then, use our <a href="https://hub.qovery.com/docs/using-qovery/configuration/environment/#terraform-exporter">Terraform exporter</a> feature to generate the corresponding Terraform code.
</div><br />

```terraform
resource "qovery_container" "my_container" {
  # Required
  environment_id = qovery_environment.my_environment.id
  registry_id    = qovery_container_registry.my_container_registry.id
  name           = "MyContainer"
  image_name     = "qovery-api"
  tag            = "1.0.0"

  # Optional
  entrypoint            = "/dev/api"
  auto_preview          = "true"
  cpu                   = 500
  memory                = 512
  min_running_instances = 1
  max_running_instances = 1

  healthchecks = {
    readiness_probe = {
      type = {
        http = {
          port = 8000
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 10
      success_threshold     = 1
      failure_threshold     = 3
    }


    liveness_probe = {
      type = {
        http = {
          port = 8000
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 10
      success_threshold     = 1
      failure_threshold     = 3
    }
  }

  environment_variables = [
    {
      key   = "ENV_VAR_KEY"
      value = "ENV_VAR_VALUE"
    }
  ]
  environment_variable_aliases = [
    {
      key = "ENV_VAR_KEY_ALIAS"
      # the value of the alias must be the name of the aliased variable
      # e.g here it is an alias to the above declared environment variable "ENV_VAR_KEY"
      value = "ENV_VAR_KEY"
    }
  ]
  environment_variable_overrides = [
    {
      # the key of the override must be the name of the overridden variable
      # e.g here it is an override on a variable declared at project scope "SOME_PROJECT_VARIABLE"
      key   = "SOME_PROJECT_VARIABLE"
      value = "OVERRIDDEN_VALUE"
    }
  ]
  secrets = [
    {
      key   = "SECRET_KEY"
      value = "SECRET_VALUE"
    }
  ]
  secret_aliases = [
    {
      key = "SECRET_KEY_ALIAS"
      # the value of the alias must be the name of the aliased secret
      # e.g here it is an alias to the above declared secret "SECRET_KEY"
      value = "SECRET_KEY"
    }
  ]
  secret_overrides = [
    {
      # the key of the override must be the name of the overridden secret
      # e.g here it is an override on a secret declared at project scope "SOME_PROJECT_SECRET"
      key   = "SOME_PROJECT_SECRET"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  custom_domains = [
    {
      domain = "example.com"
    }
  ]

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Containers/operation/getDefaultContainerAdvancedSettings
    # you can only indicate settings that you need to override
    "network.ingress.proxy_send_timeout_seconds" : 80,
    "network.ingress.proxy_body_size_mb" : 200,
  })

  depends_on = [
    qovery_environment.my_environment,
    qovery_container_registry.my_container_registry
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) Id of the environment.
- `healthchecks` (Attributes) Configuration for the healthchecks that are going to be executed against your service (see [below for nested schema](#nestedatt--healthchecks))
- `image_name` (String) Name of the container image.
- `name` (String) Name of the container.
- `registry_id` (String) Id of the registry.
- `tag` (String) Tag of the container image.

### Optional

- `advanced_settings_json` (String) Advanced settings.
- `annotations_group_ids` (Set of String) List of annotations group ids
- `arguments` (List of String) List of arguments of this container.
- `auto_deploy` (Boolean) Specify if the container will be automatically updated after receiving a new image tag.
- `auto_preview` (Boolean) Specify if the environment preview option is activated or not for this container.
- `cpu` (Number) CPU of the container in millicores (m) [1000m = 1 CPU].
	- Must be: `>= 10`.
	- Default: `500`.
- `custom_domains` (Attributes Set) List of custom domains linked to this container. (see [below for nested schema](#nestedatt--custom_domains))
- `deployment_stage_id` (String) Id of the deployment stage.
- `entrypoint` (String) Entrypoint of the container.
- `environment_variable_aliases` (Attributes Set) List of environment variable aliases linked to this container. (see [below for nested schema](#nestedatt--environment_variable_aliases))
- `environment_variable_overrides` (Attributes Set) List of environment variable overrides linked to this container. (see [below for nested schema](#nestedatt--environment_variable_overrides))
- `environment_variables` (Attributes Set) List of environment variables linked to this container. (see [below for nested schema](#nestedatt--environment_variables))
- `icon_uri` (String) Icon URI representing the container.
- `labels_group_ids` (Set of String) List of labels group ids
- `max_running_instances` (Number) Maximum number of instances running for the container.
	- Must be: `>= -1`.
	- Default: `1`.
- `memory` (Number) RAM of the container in MB [1024MB = 1GB].
	- Must be: `>= 10`.
	- Default: `512`.
- `min_running_instances` (Number) Minimum number of instances running for the container.
	- Must be: `>= 1`.
	- Default: `1`.
- `ports` (Attributes List) List of ports linked to this container. (see [below for nested schema](#nestedatt--ports))
- `secret_aliases` (Attributes Set) List of secret aliases linked to this container. (see [below for nested schema](#nestedatt--secret_aliases))
- `secret_overrides` (Attributes Set) List of secret overrides linked to this container. (see [below for nested schema](#nestedatt--secret_overrides))
- `secrets` (Attributes Set) List of secrets linked to this container. (see [below for nested schema](#nestedatt--secrets))
- `storage` (Attributes Set) List of storages linked to this container. (see [below for nested schema](#nestedatt--storage))

### Read-Only

- `built_in_environment_variables` (Attributes Set) List of built-in environment variables linked to this container. (see [below for nested schema](#nestedatt--built_in_environment_variables))
- `external_host` (String) The container external FQDN host [NOTE: only if your container is using a publicly accessible port].
- `id` (String) Id of the container.
- `internal_host` (String) The container internal host.

<a id="nestedatt--healthchecks"></a>
### Nested Schema for `healthchecks`

Optional:

- `liveness_probe` (Attributes) Configuration for the liveness probe, in order to know when your service is working correctly. Failing the probe means your service being killed/ask to be restarted. (see [below for nested schema](#nestedatt--healthchecks--liveness_probe))
- `readiness_probe` (Attributes) Configuration for the readiness probe, in order to know when your service is ready to receive traffic. Failing the probe means your service will stop receiving traffic. (see [below for nested schema](#nestedatt--healthchecks--readiness_probe))

<a id="nestedatt--healthchecks--liveness_probe"></a>
### Nested Schema for `healthchecks.liveness_probe`

Required:

- `failure_threshold` (Number) Number of time the an ok probe should fail before declaring it as failed
- `initial_delay_seconds` (Number) Number of seconds to wait before the first execution of the probe to be trigerred
- `period_seconds` (Number) Number of seconds before each execution of the probe
- `success_threshold` (Number) Number of time the probe should success before declaring a failed probe as ok again
- `timeout_seconds` (Number) Number of seconds within which the check need to respond before declaring it as a failure
- `type` (Attributes) Kind of check to run for this probe. There can only be one configured at a time (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type))

<a id="nestedatt--healthchecks--liveness_probe--type"></a>
### Nested Schema for `healthchecks.liveness_probe.type`

Optional:

- `exec` (Attributes) Check that the given command return an exit 0. Binary should be present in the image (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--exec))
- `grpc` (Attributes) Check that the given port respond to GRPC call (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--grpc))
- `http` (Attributes) Check that the given port respond to HTTP call (should return a 2xx response code) (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--http))
- `tcp` (Attributes) Check that the given port accepting connection (see [below for nested schema](#nestedatt--healthchecks--liveness_probe--type--tcp))

<a id="nestedatt--healthchecks--liveness_probe--type--exec"></a>
### Nested Schema for `healthchecks.liveness_probe.type.exec`

Required:

- `command` (List of String) The command and its arguments to exec


<a id="nestedatt--healthchecks--liveness_probe--type--grpc"></a>
### Nested Schema for `healthchecks.liveness_probe.type.grpc`

Required:

- `port` (Number) The port number to try to connect to

Optional:

- `service` (String) The grpc service to connect to. It needs to implement grpc health protocol. https://kubernetes.io/blog/2018/10/01/health-checking-grpc-servers-on-kubernetes/#introducing-grpc-health-probe


<a id="nestedatt--healthchecks--liveness_probe--type--http"></a>
### Nested Schema for `healthchecks.liveness_probe.type.http`

Required:

- `port` (Number) The port number to try to connect to
- `scheme` (String) if the HTTP GET request should be done in HTTP or HTTPS.

Optional:

- `path` (String) The path that the HTTP GET request. By default it is `/`


<a id="nestedatt--healthchecks--liveness_probe--type--tcp"></a>
### Nested Schema for `healthchecks.liveness_probe.type.tcp`

Required:

- `port` (Number) The port number to try to connect to

Optional:

- `host` (String) Optional. If the host need to be different than localhost/pod ip




<a id="nestedatt--healthchecks--readiness_probe"></a>
### Nested Schema for `healthchecks.readiness_probe`

Required:

- `failure_threshold` (Number) Number of time the an ok probe should fail before declaring it as failed
- `initial_delay_seconds` (Number) Number of seconds to wait before the first execution of the probe to be trigerred
- `period_seconds` (Number) Number of seconds before each execution of the probe
- `success_threshold` (Number) Number of time the probe should success before declaring a failed probe as ok again
- `timeout_seconds` (Number) Number of seconds within which the check need to respond before declaring it as a failure
- `type` (Attributes) Kind of check to run for this probe. There can only be one configured at a time (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type))

<a id="nestedatt--healthchecks--readiness_probe--type"></a>
### Nested Schema for `healthchecks.readiness_probe.type`

Optional:

- `exec` (Attributes) Check that the given command return an exit 0. Binary should be present in the image (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--exec))
- `grpc` (Attributes) Check that the given port respond to GRPC call (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--grpc))
- `http` (Attributes) Check that the given port respond to HTTP call (should return a 2xx response code) (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--http))
- `tcp` (Attributes) Check that the given port accepting connection (see [below for nested schema](#nestedatt--healthchecks--readiness_probe--type--tcp))

<a id="nestedatt--healthchecks--readiness_probe--type--exec"></a>
### Nested Schema for `healthchecks.readiness_probe.type.exec`

Required:

- `command` (List of String) The command and its arguments to exec


<a id="nestedatt--healthchecks--readiness_probe--type--grpc"></a>
### Nested Schema for `healthchecks.readiness_probe.type.grpc`

Required:

- `port` (Number) The port number to try to connect to

Optional:

- `service` (String) The grpc service to connect to. It needs to implement grpc health protocol. https://kubernetes.io/blog/2018/10/01/health-checking-grpc-servers-on-kubernetes/#introducing-grpc-health-probe


<a id="nestedatt--healthchecks--readiness_probe--type--http"></a>
### Nested Schema for `healthchecks.readiness_probe.type.http`

Required:

- `port` (Number) The port number to try to connect to
- `scheme` (String) if the HTTP GET request should be done in HTTP or HTTPS.

Optional:

- `path` (String) The path that the HTTP GET request. By default it is `/`


<a id="nestedatt--healthchecks--readiness_probe--type--tcp"></a>
### Nested Schema for `healthchecks.readiness_probe.type.tcp`

Required:

- `port` (Number) The port number to try to connect to

Optional:

- `host` (String) Optional. If the host need to be different than localhost/pod ip





<a id="nestedatt--custom_domains"></a>
### Nested Schema for `custom_domains`

Required:

- `domain` (String) Your custom domain.

Optional:

- `generate_certificate` (Boolean) Qovery will generate and manage the certificate for this domain.
- `use_cdn` (Boolean) Indicates if the custom domain is behind a CDN (i.e Cloudflare).
This will condition the way we are checking CNAME before & during a deployment:
 * If `true` then we only check the domain points to an IP
 * If `false` then we check that the domain resolves to the correct service Load Balancer

Read-Only:

- `id` (String) Id of the custom domain.
- `status` (String) Status of the custom domain.
- `validation_domain` (String) URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.


<a id="nestedatt--environment_variable_aliases"></a>
### Nested Schema for `environment_variable_aliases`

Required:

- `key` (String) Name of the environment variable alias.
- `value` (String) Name of the variable to alias.

Optional:

- `description` (String) Description of the environment variable alias.

Read-Only:

- `id` (String) Id of the environment variable alias.


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

- `internal_port` (Number) Internal port of the container.
	- Must be: `>= 1` and `<= 65535`.
- `is_default` (Boolean) If this port will be used for the root domain
- `publicly_accessible` (Boolean) Specify if the port is exposed to the world or not for this container.

Optional:

- `external_port` (Number) External port of the container.
	- Required if: `ports.publicly_accessible=true`.
	- Must be: `>= 1` and `<= 65535`.
- `name` (String) Name of the port.
- `protocol` (String) Protocol used for the port of the container.
	- Can be: `GRPC`, `HTTP`, `TCP`, `UDP`.
	- Default: `HTTP`.

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


<a id="nestedatt--secret_overrides"></a>
### Nested Schema for `secret_overrides`

Required:

- `key` (String) Name of the secret override.
- `value` (String, Sensitive) Value of the secret override.

Optional:

- `description` (String) Description of the secret override.

Read-Only:

- `id` (String) Id of the secret override.


<a id="nestedatt--secrets"></a>
### Nested Schema for `secrets`

Required:

- `key` (String) Key of the secret.
- `value` (String, Sensitive) Value of the secret.

Optional:

- `description` (String) Description of the secret.

Read-Only:

- `id` (String) Id of the secret.


<a id="nestedatt--storage"></a>
### Nested Schema for `storage`

Required:

- `mount_point` (String) Mount point of the storage for the container.
- `size` (Number) Size of the storage for the container in GB [1024MB = 1GB].
	- Must be: `>= 1`.
- `type` (String) Type of the storage for the container.
	- Can be: `FAST_SSD`.

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
terraform import qovery_container.my_container "<container_id>"
```