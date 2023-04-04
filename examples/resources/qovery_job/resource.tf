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
  secrets = [
    {
      key   = "MY_TERRAFORM_CONTAINER_SECRET"
      value = "MY_TERRAFORM_CONTAINER_SECRET_VALUE"
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

  depends_on = [
    qovery_environment.my_environment,
  ]
}
