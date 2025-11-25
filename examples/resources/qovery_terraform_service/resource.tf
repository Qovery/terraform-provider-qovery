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
