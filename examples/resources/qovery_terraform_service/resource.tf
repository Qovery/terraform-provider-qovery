resource "qovery_terraform_service" "my_terraform_service" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-terraform-service"
  auto_deploy    = true

  git_repository = {
    url       = "https://github.com/my-org/terraform-infra.git"
    branch    = "main"
    root_path = "/environments/production"
  }

  engine = "TERRAFORM"
  engine_version = {
    explicit_version = "1.14"
  }

  # State stored by Qovery in the cluster. Use user_provided = {} to keep the backend declared in the Terraform code.
  backend = {
    kubernetes = {}
  }

  tfvars_files  = []
  job_resources = {}

  variables = [
    {
      key   = "region"
      value = "us-east-1"
    },
    {
      key       = "database_password"
      value     = var.database_password
      is_secret = true
    }
  ]
}
