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
```

# With User-Provided Backend

Use your own backend configuration defined in Terraform code:

```terraform
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
```

# Full Example with Variables

Complete example with variables, tfvar files, and advanced settings:

```terraform
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
```
