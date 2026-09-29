resource "qovery_application" "my_application" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-application"

  git_repository = {
    url    = "https://github.com/my-org/my-app.git"
    branch = "main"
  }
  dockerfile_path = "Dockerfile"

  ports = [
    {
      internal_port       = 8080
      external_port       = 443
      publicly_accessible = true
    }
  ]

  healthchecks = {
    readiness_probe = {
      type = {
        http = {
          port   = 8080
          scheme = "HTTP"
          path   = "/ready"
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
    liveness_probe = {
      type = {
        tcp = {
          port = 8080
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
  }

  environment_variables = [
    {
      key   = "LOG_LEVEL"
      value = "info"
    }
  ]

  secrets = [
    {
      key   = "API_KEY"
      value = var.api_key
    }
  ]
}
