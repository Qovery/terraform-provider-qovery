resource "qovery_application" "my_application" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyApplication"
  git_repository = {
    url       = "https://github.com/Qovery/terraform-provider-qovery.git"
    branch    = "main" # Optional
    root_path = "/"    # Optional
  }

  # Optional
  build_mode            = "DOCKER"
  dockerfile_path       = "Dockerfile"
  state                 = "RUNNING"
  auto_preview          = "true"
  cpu                   = 500
  memory                = 512
  min_running_instances = 1
  max_running_instances = 1
  entrypoint            = "/bin/sh"
  arguments             = ["arg"]
  environment_variables = [
    {
      key   = "ENV_VAR_KEY"
      value = "ENV_VAR_VALUE"
    }
  ]
  secrets = [
    {
      key   = "SECRET_KEY"
      value = "SECRET_VALUE"
    }
  ]
  custom_domains = [
    {
      domain = "example.com"
    }
  ]

  depends_on = [
    qovery_environment.my_environment
  ]
}
