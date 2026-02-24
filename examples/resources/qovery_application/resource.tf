resource "qovery_application" "my_application" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyApplication"
  git_repository = {
    url       = "https://github.com/my-org/my-app.git"
    branch    = "main" # Optional (defaults to main or master)
    root_path = "/"    # Optional (defaults to "/", useful for monorepos)
  }

  # Build configuration
  build_mode      = "DOCKER"     # DOCKER or BUILDPACKS
  dockerfile_path = "Dockerfile" # Required when build_mode = "DOCKER"

  # Optional
  auto_preview          = false
  auto_deploy           = true
  cpu                   = 500
  memory                = 512
  min_running_instances = 1
  max_running_instances = 3
  entrypoint            = "/bin/sh"
  arguments             = ["-c", "start-server"]

  # Port configuration
  ports = [
    {
      internal_port       = 8080
      external_port       = 443
      publicly_accessible = true
      protocol            = "HTTP"
      is_default          = true
      name                = "http"
    },
    {
      internal_port       = 9090
      publicly_accessible = false
      protocol            = "HTTP"
      name                = "metrics"
    }
  ]

  # Healthchecks
  healthchecks = {
    readiness_probe = {
      type = {
        http = {
          port   = 8080
          path   = "/ready"
          scheme = "HTTP"
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
        http = {
          port   = 8080
          path   = "/health"
          scheme = "HTTP"
        }
      }
      initial_delay_seconds = 30
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
  }

  # Environment variables
  environment_variables = [
    {
      key   = "APP_PORT"
      value = "8080"
    }
  ]
  environment_variable_aliases = [
    {
      key = "PORT"
      # The value of the alias must be the name of the aliased variable.
      # Here it creates an alias "PORT" pointing to the "APP_PORT" variable above.
      value = "APP_PORT"
    }
  ]
  environment_variable_overrides = [
    {
      # The key must match a variable defined at a higher scope (project or environment).
      key   = "SOME_PROJECT_VARIABLE"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  # Secrets
  secrets = [
    {
      key   = "SECRET_KEY"
      value = "SECRET_VALUE"
    }
  ]
  secret_aliases = [
    {
      key = "SECRET_KEY_ALIAS"
      # The value of the alias must be the name of the aliased secret.
      value = "SECRET_KEY"
    }
  ]
  secret_overrides = [
    {
      # The key must match a secret defined at a higher scope (project or environment).
      key   = "SOME_PROJECT_SECRET"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  # Custom domains
  custom_domains = [
    {
      domain               = "app.example.com"
      generate_certificate = true
    }
  ]

  # Deployment restrictions (only deploy when specific paths change)
  deployment_restrictions = [
    {
      mode  = "MATCH"
      type  = "PATH"
      value = "src/"
    }
  ]

  # Advanced settings (JSON)
  advanced_settings_json = jsonencode({
    # Non-exhaustive list. Full list: https://api-doc.qovery.com/#tag/Applications/operation/getDefaultApplicationAdvancedSettings
    "network.ingress.proxy_buffer_size_kb" : 8,
    "network.ingress.keepalive_time_seconds" : 1000,
  })

  depends_on = [
    qovery_environment.my_environment
  ]
}
