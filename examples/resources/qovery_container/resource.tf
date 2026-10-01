resource "qovery_container" "my_container" {
  environment_id = qovery_environment.my_environment.id
  registry_id    = qovery_container_registry.my_container_registry.id
  name           = "my-container"
  image_name     = "nginx"
  tag            = "1.27-alpine"

  ports = [
    {
      internal_port       = 80
      external_port       = 443
      publicly_accessible = true
    }
  ]

  healthchecks = {
    readiness_probe = {
      type = {
        http = {
          port   = 80
          scheme = "HTTP"
        }
      }
      initial_delay_seconds = 10
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
    liveness_probe = {
      type = {
        tcp = {
          port = 80
        }
      }
      initial_delay_seconds = 10
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
  }

  environment_variables = [
    {
      key   = "NGINX_ENTRYPOINT_QUIET_LOGS"
      value = "1"
    }
  ]
}
