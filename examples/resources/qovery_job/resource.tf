resource "qovery_job" "my_job" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-cron-job"

  schedule = {
    cronjob = {
      schedule = "0 3 * * *"
      command = {
        entrypoint = "/bin/sh"
        arguments  = ["-c", "echo 'Job completed'"]
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

  healthchecks = {}
}
