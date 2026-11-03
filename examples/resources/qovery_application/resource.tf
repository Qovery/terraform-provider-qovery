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

  # Extract one JSON value instead of the whole external secret.
  # For {"accounts":[{"password":"example"}]}, accounts.0.password selects "example".
  # Use dots between members and numeric indexes for arrays.
  # Escape literal dots and backslashes in the selector; double those backslashes in HCL:
  # "database\\.password" selects the member "database.password".
  # "database\\\\password" selects the member "database\password".
  # Other query characters are literal member names; arbitrary GJSON queries are not evaluated.
  # Omit or remove json_target_key to read the whole secret.
  # Also supported on containers, jobs, Helm services, Terraform services and environments,
  # but not on external_secret_files.
  external_secrets = [
    {
      key                      = "DATABASE_PASSWORD"
      reference                = "/production/database"
      secret_manager_access_id = var.secret_manager_access_id
      json_target_key          = "accounts.0.password"
    }
  ]

}
