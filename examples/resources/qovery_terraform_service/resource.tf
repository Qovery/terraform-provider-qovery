resource "qovery_terraform_service" "my_terraform_service" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "my-terraform-service"
  description    = "Terraform service managed by Qovery"
  auto_deploy    = true

  # Git repository containing Terraform files
  git_repository = {
    url       = "https://github.com/my-org/terraform-infra.git"
    branch    = "main"
    root_path = "/environments/production"
    # git_token_id = qovery_git_token.my_git_token.id  # For private repos
  }

  # List of .tfvars files relative to the root path
  tfvars_files = ["terraform.tfvars"]

  # Terraform input variables
  variables = [
    {
      key       = "AWS_REGION"
      value     = "us-east-1"
      is_secret = false
    },
    {
      key       = "DATABASE_PASSWORD"
      value     = "supersecret"
      is_secret = true
    }
  ]

  # Backend configuration - choose exactly one
  backend = {
    kubernetes = {} # Qovery-managed Kubernetes backend
    # user_provided = {}   # Use backend configured in your Terraform code
  }

  # Engine configuration
  engine = "TERRAFORM" # Can be "TERRAFORM" or "OPEN_TOFU"

  engine_version = {
    explicit_version          = "1.5.0"
    read_from_terraform_block = false
  }

  # Job resources (compute allocation for Terraform runs)
  job_resources = {
    cpu_milli   = 1000 # 1 CPU
    ram_mib     = 1024 # 1 GiB
    gpu         = 0
    storage_gib = 20 # WARNING: Cannot be reduced after creation
  }

  # Optional settings
  timeout_seconds         = 1800 # 30 minutes
  icon_uri                = "app://qovery-console/terraform"
  use_cluster_credentials = false

  # Optional: Extra CLI arguments for Terraform actions
  # action_extra_arguments = {
  #   plan    = ["-parallelism=10"]
  #   apply   = ["-parallelism=10", "-auto-approve"]
  #   destroy = ["-auto-approve"]
  # }

  # Optional: control deployment order
  # deployment_stage_id = qovery_deployment_stage.my_stage.id

  # Optional: Advanced settings
  advanced_settings_json = jsonencode({
    # Non-exhaustive list. See Qovery API documentation for all available settings.
    "deployment.termination_grace_period_seconds" : 120,
    "build.timeout_max_sec" : 1800
  })

  depends_on = [
    qovery_environment.my_environment,
  ]
}
