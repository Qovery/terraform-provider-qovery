resource "qovery_container" "my_container" {
  # Required
  environment_id = qovery_environment.my_environment.id
  registry_id    = qovery_container_registry.my_container_registry.id
  name           = "MyContainer"
  image_name     = "nginx"
  tag            = "1.25-alpine"

  # Optional
  entrypoint            = "/docker-entrypoint.sh"
  auto_preview          = true
  auto_deploy           = true
  cpu                   = 500
  memory                = 512
  min_running_instances = 1
  max_running_instances = 3

  # Port configuration
  ports = [
    {
      internal_port       = 80
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

  # Persistent storage
  storage = [
    {
      type        = "FAST_SSD"
      size        = 10
      mount_point = "/data"
    }
  ]

  # Healthchecks
  healthchecks = {
    readiness_probe = {
      type = {
        http = {
          port   = 80
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

    liveness_probe = {
      type = {
        tcp = {
          port = 80
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
      key   = "NGINX_PORT"
      value = "80"
    }
  ]
  environment_variable_aliases = [
    {
      key = "PORT"
      # The value of the alias must be the name of the aliased variable.
      # Here it creates an alias "PORT" pointing to the "NGINX_PORT" variable above.
      value = "NGINX_PORT"
    }
  ]
  environment_variable_overrides = [
    {
      # The key must match a variable defined at a higher scope (project or environment).
      key   = "SOME_PROJECT_VARIABLE"
      value = "OVERRIDDEN_VALUE"
    }
  ]

  # Environment variable files (mounted as files in the container)
  environment_variable_files = [
    {
      key        = "APP_CONFIG"
      value      = "config-content"
      mount_path = "/etc/app/config.yaml"
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

  # Secret files (mounted as files, value is encrypted)
  secret_files = [
    {
      key        = "API_KEY"
      value      = "secret-value"
      mount_path = "/usr/local/secrets/api-key"
    }
  ]

  # Custom domains
  custom_domains = [
    {
      domain               = "app.example.com"
      generate_certificate = true
    }
  ]

  # Advanced settings (JSON)
  advanced_settings_json = jsonencode({
    # Non-exhaustive list. Full list: https://api-doc.qovery.com/#tag/Containers/operation/getDefaultContainerAdvancedSettings
    "network.ingress.proxy_send_timeout_seconds" : 80,
    "network.ingress.proxy_body_size_mb" : 200,
  })

  depends_on = [
    qovery_environment.my_environment,
    qovery_container_registry.my_container_registry
  ]
}
