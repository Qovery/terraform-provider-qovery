resource "qovery_job" "my_job" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "test-job"

  # Optional
  auto_preview         = "true"
  cpu                  = 500
  memory               = 512
  max_duration_seconds = 23
  max_nb_restart       = 1
  port                 = 5432
  environment_variables = [
    {
      key   = "MY_TERRAFORM_CONTAINER_VARIABLE"
      value = "MY_TERRAFORM_CONTAINER_VARIABLE_VALUE"
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
      key   = "MY_TERRAFORM_CONTAINER_SECRET"
      value = "MY_TERRAFORM_CONTAINER_SECRET_VALUE"
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

  schedule = {
    on_start  = {}
    on_stop   = {}
    on_delete = {}
    cronjob = {
      schedule = "*/2 * * * *"
      command = {
        entrypoint = ""
        arguments  = ["echo", "'DONE'"]
      }
    }
  }
  source = {
    image = {
      registry_id = qovery_container_registry.my_container_registry.id
      name        = "debian"
      tag         = "stable"
    }
  }


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

  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Jobs/operation/getDefaultJobAdvancedSettings
    # you can only indicate settings that you need to override
    "deployment.termination_grace_period_seconds" : 120,
    "build.timeout_max_sec" : 120
  })

  depends_on = [
    qovery_environment.my_environment,
  ]
}
