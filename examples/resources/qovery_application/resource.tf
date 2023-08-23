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
  auto_preview          = "true"
  cpu                   = 500
  memory                = 512
  min_running_instances = 1
  max_running_instances = 1
  entrypoint            = "/bin/sh"
  arguments             = ["arg"]
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
      # the key of the override must be the name of the aliased variable
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
      # the key of the override must be the name of the aliased secret
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
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Applications/operation/getDefaultApplicationAdvancedSettings
    # you can only indicate settings that you need to override
    "network.ingress.proxy_buffer_size_kb" : 8,
    "network.ingress.keepalive_time_seconds" : 1000,
  })

  depends_on = [
    qovery_environment.my_environment
  ]
}
