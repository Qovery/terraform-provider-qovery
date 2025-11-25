# qovery_terraform_service (Resource)

Provides a Qovery Terraform service resource. This can be used to create and manage Qovery terraform services.


## Example

<div class="alert alert-info">
  <i style="font-size:24px" class="fa">&#xf05a;</i> If you're not familiar with Terraform or just want more examples, you can configure everything you need directly from the <a href="https://console.qovery.com">Qovery console</a>. Then, use our <a href="https://hub.qovery.com/docs/using-qovery/configuration/environment/#terraform-exporter">Terraform exporter</a> feature to generate the corresponding Terraform code.
</div><br />

```terraform
resource "qovery_terraform_service" "my_terraform_service" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "my-terraform-service"
  description    = "Terraform service managed by Terraform provider"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform_service_engine_testing.git"
    branch    = "main"
    root_path = "/s3_terraform_unique"
    # git_token_id = qovery_git_token.my_git_token.id  # Optional, for private repos
  }

  tfvar_files = []

  # Optional: Variables
  variable = [
    {
      key    = "AWS_REGION"
      value  = "us-east-1"
      secret = false
    },
    {
      key    = "DATABASE_PASSWORD"
      value  = "supersecret"
      secret = true
    }
  ]

  # Backend configuration - choose ONE
  backend = {
    kubernetes = {}
    # OR
    # user_provided = {}
  }

  # Engine configuration
  engine = "TERRAFORM" # or "OPEN_TOFU"

  engine_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  # Job resources
  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    gpu         = 0
    storage_gib = 20
  }

  # Optional settings
  timeout_sec             = 1800
  icon_uri                = "app://qovery-console/terraform"
  use_cluster_credentials = false

  # Optional: Extra arguments for Terraform actions
  # action_extra_arguments = {
  #   plan    = ["-parallelism=10"]
  #   apply   = ["-parallelism=10"]
  #   destroy = ["-auto-approve"]
  # }

  # Optional: Advanced settings
  advanced_settings_json = jsonencode({
    # Non-exhaustive list, the complete list is available in Qovery API doc
    # You can only indicate settings that you need to override
    "deployment.termination_grace_period_seconds" : 120,
    "build.timeout_max_sec" : 1800
  })

  depends_on = [
    qovery_environment.my_environment,
  ]
}
```

You can find complete examples within these repositories:
# Basic Example

Create a Terraform service with Kubernetes backend:

```terraform
resource "qovery_terraform_service" "my_terraform_service" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-terraform-service"
  description    = "Terraform service managed by Terraform provider"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  backend = {
    kubernetes = {}
  }

  engine = "TERRAFORM"

  engine_version = {
    explicit_version = "1.5.0"
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    storage_gib = 20
  }
}

With User-Provided Backend
Use your own backend configuration defined in Terraform code:

resource "qovery_terraform_service" "my_terraform_service" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-terraform-service"
  description    = "Terraform service with custom backend"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/Qovery/terraform-examples.git"
    branch    = "main"
    root_path = "/"
  }

  tfvar_files = []

  backend = {
    user_provided = {}
  }

  engine = "TERRAFORM"

  engine_version = {
    explicit_version = "1.5.0"
  }

  job_resources = {
    cpu_milli   = 1000
    ram_mib     = 1024
    storage_gib = 20
  }
}

Full Example with Variables
Complete example with variables, tfvar files, and advanced settings:

resource "qovery_terraform_service" "my_terraform_service" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-terraform-service"
  description    = "Full-featured Terraform service"
  auto_deploy    = true

  git_repository = {
    url          = "https://github.com/Qovery/terraform-examples.git"
    branch       = "main"
    root_path    = "/infrastructure"
    git_token_id = qovery_git_token.my_git_token.id
  }

  tfvar_files = [
    "/infrastructure/production.tfvars",
    "/infrastructure/common.tfvars"
  ]

  variable = [
    {
      key    = "AWS_REGION"
      value  = "us-east-1"
      secret = false
    },
    {
      key    = "DATABASE_PASSWORD"
      value  = "supersecret"
      secret = true
    }
  ]

  backend = {
    kubernetes = {}
  }

  engine = "TERRAFORM"

  engine_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  job_resources = {
    cpu_milli   = 2000
    ram_mib     = 2048
    gpu         = 0
    storage_gib = 50
  }

  timeout_sec             = 3600
  use_cluster_credentials = true

  action_extra_arguments = {
    plan    = ["-parallelism=10"]
    apply   = ["-parallelism=10", "-auto-approve"]
    destroy = ["-auto-approve"]
  }

  advanced_settings_json = jsonencode({
    "deployment.termination_grace_period_seconds" : 180,
    "build.timeout_max_sec" : 3600,
    "security.read_only_root_filesystem" : true
  })

  depends_on = [
    qovery_environment.my_environment,
    qovery_git_token.my_git_token
  ]
}


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `auto_deploy` (Boolean) Specify if the terraform service will be automatically updated on every new commit.
- `backend` (Attributes) Terraform backend configuration. Exactly one backend type must be specified. (see [below for nested schema](#nestedatt--backend))
- `engine` (String) Terraform engine to use (TERRAFORM or OPEN_TOFU).
- `engine_version` (Attributes) Terraform/OpenTofu engine version configuration. (see [below for nested schema](#nestedatt--engine_version))
- `environment_id` (String) Id of the environment.
- `git_repository` (Attributes) Terraform service git repository configuration. (see [below for nested schema](#nestedatt--git_repository))
- `job_resources` (Attributes) Resource allocation for the Terraform job. (see [below for nested schema](#nestedatt--job_resources))
- `name` (String) Name of the terraform service.
- `tfvar_files` (List of String) List of .tfvars file paths relative to the root path.

### Optional

- `action_extra_arguments` (Map of List of String) Extra CLI arguments for specific Terraform actions (plan, apply, destroy).
- `advanced_settings_json` (String) Advanced settings in JSON format.
- `description` (String) Description of the terraform service.
- `icon_uri` (String) Icon URI representing the terraform service.
- `timeout_sec` (Number) Timeout in seconds for Terraform operations.
	- Must be: `>= 0`.
	- Default: `1800`.
- `use_cluster_credentials` (Boolean) Use cluster credentials for cloud provider authentication.
- `variable` (Attributes Set) Terraform variables. (see [below for nested schema](#nestedatt--variable))

### Read-Only

- `created_at` (String) Creation date of the terraform service.
- `id` (String) Id of the terraform service.
- `updated_at` (String) Last update date of the terraform service.

<a id="nestedatt--backend"></a>
### Nested Schema for `backend`

Optional:

- `kubernetes` (Attributes) Use Kubernetes backend for state management. (see [below for nested schema](#nestedatt--backend--kubernetes))
- `user_provided` (Attributes) Use user-provided backend configuration (configured in Terraform code). (see [below for nested schema](#nestedatt--backend--user_provided))

<a id="nestedatt--backend--kubernetes"></a>
### Nested Schema for `backend.kubernetes`


<a id="nestedatt--backend--user_provided"></a>
### Nested Schema for `backend.user_provided`



<a id="nestedatt--engine_version"></a>
### Nested Schema for `engine_version`

Required:

- `explicit_version` (String) Explicit version to use for the Terraform/OpenTofu binary.

Optional:

- `read_from_terraform_block` (Boolean) Whether to read the version from the terraform block in the code.


<a id="nestedatt--git_repository"></a>
### Nested Schema for `git_repository`

Required:

- `url` (String) Git repository URL.

Optional:

- `branch` (String) Git branch.
- `git_token_id` (String) Git token ID for private repositories.
- `root_path` (String) Git root path.


<a id="nestedatt--job_resources"></a>
### Nested Schema for `job_resources`

Optional:

- `cpu_milli` (Number) CPU of the terraform job in millicores (m) [1000m = 1 CPU].
	- Must be: `>= 10`.
	- Default: `1000`.
- `gpu` (Number) Number of GPUs for the terraform job.
	- Must be: `>= 0`.
	- Default: `0`.
- `ram_mib` (Number) RAM of the terraform job in MiB [1024 MiB = 1GiB].
	- Must be: `>= 1`.
	- Default: `1024`.
- `storage_gib` (Number) Storage of the terraform job in GiB [1 GiB = 1024 MiB]. WARNING: Cannot be reduced after creation.
	- Must be: `>= 1`.
	- Default: `20`.


<a id="nestedatt--variable"></a>
### Nested Schema for `variable`

Required:

- `key` (String) Variable key.
- `value` (String) Variable value.

Optional:

- `secret` (Boolean) Is this variable a secret.
## Import
```shell
terraform import qovery_terraform_service.my_terraform_service "<terraform_service_id>"
```