# qovery_project (Resource)

Provides a Qovery project resource. This can be used to create and manage Qovery projects.


## Example

<div class="alert alert-info">
  <i style="font-size:24px" class="fa">&#xf05a;</i> If you're not familiar with Terraform or just want more examples, you can configure everything you need directly from the <a href="https://console.qovery.com">Qovery console</a>. Then, use our <a href="https://hub.qovery.com/docs/using-qovery/configuration/environment/#terraform-exporter">Terraform exporter</a> feature to generate the corresponding Terraform code.
</div><br />

```terraform
resource "qovery_project" "my_project" {
  # Required
  organization_id = qovery_organization.my_organization.id
  name            = "MyProject"

  # Optional
  description = "My project description"
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

  depends_on = [
    qovery_organization.my_organization
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the project.
- `organization_id` (String) Id of the organization.

### Optional

- `description` (String) Description of the project.
- `environment_variable_aliases` (Attributes Set) List of environment variable aliases linked to this project. (see [below for nested schema](#nestedatt--environment_variable_aliases))
- `environment_variables` (Attributes Set) List of environment variables linked to this project. (see [below for nested schema](#nestedatt--environment_variables))
- `secret_aliases` (Attributes Set) List of secret aliases linked to this project. (see [below for nested schema](#nestedatt--secret_aliases))
- `secrets` (Attributes Set) List of secrets linked to this project. (see [below for nested schema](#nestedatt--secrets))

### Read-Only

- `built_in_environment_variables` (Attributes Set) List of built-in environment variables linked to this project. (see [below for nested schema](#nestedatt--built_in_environment_variables))
- `id` (String) Id of the project.

<a id="nestedatt--environment_variable_aliases"></a>
### Nested Schema for `environment_variable_aliases`

Required:

- `key` (String) Name of the environment variable alias.
- `value` (String) Name of the variable to alias.

Optional:

- `description` (String) Description of the environment variable alias.

Read-Only:

- `id` (String) Id of the environment variable alias.


<a id="nestedatt--environment_variables"></a>
### Nested Schema for `environment_variables`

Required:

- `key` (String) Key of the environment variable.
- `value` (String) Value of the environment variable.

Optional:

- `description` (String) Description of the environment variable.

Read-Only:

- `id` (String) Id of the environment variable.


<a id="nestedatt--secret_aliases"></a>
### Nested Schema for `secret_aliases`

Required:

- `key` (String) Name of the secret alias.
- `value` (String) Name of the secret to alias.

Optional:

- `description` (String) Description of the secret alias.

Read-Only:

- `id` (String) Id of the secret alias.


<a id="nestedatt--secrets"></a>
### Nested Schema for `secrets`

Required:

- `key` (String) Key of the secret.
- `value` (String, Sensitive) Value of the secret.

Optional:

- `description` (String) Description of the secret.

Read-Only:

- `id` (String) Id of the secret.


<a id="nestedatt--built_in_environment_variables"></a>
### Nested Schema for `built_in_environment_variables`

Read-Only:

- `description` (String) Description of the environment variable.
- `id` (String) Id of the environment variable.
- `key` (String) Key of the environment variable.
- `value` (String) Value of the environment variable.
## Import
```shell
terraform import qovery_project.my_project "<project_id>"
```