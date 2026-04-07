# Example: Helm chart from a Helm repository
resource "qovery_helm" "my_helm" {
  # Required
  environment_id               = qovery_environment.my_environment.id
  name                         = "my-helm-chart"
  description                  = "Helm chart deployed via Terraform"
  allow_cluster_wide_resources = false

  # Source: Helm chart from a Helm repository
  source = {
    helm_repository = {
      helm_repository_id = qovery_helm_repository.my_helm_repo.id
      chart_name         = "nginx"
      chart_version      = "1.0.0"
    }
  }

  # Helm values overrides
  values_override = {
    # Override values using --set syntax
    "set" = {
      "replicaCount"         = "3"
      "service.type"         = "ClusterIP"
      "resources.limits.cpu" = "500m"
    }
    # Override values using --set-string syntax (always treated as strings)
    "set_string" = {
      "image.tag" = "latest"
    }
    # Override values using --set-json syntax
    "set_json" = {
      "tolerations" = "[{\"key\": \"dedicated\", \"operator\": \"Equal\", \"value\": \"helm\", \"effect\": \"NoSchedule\"}]"
    }
    # Override values from YAML files
    file = {
      raw = {
        "custom-values" = {
          content = <<-EOT
            ingress:
              enabled: true
              hosts:
                - host: my-app.example.com
                  paths:
                    - path: /
          EOT
        }
      }
    }
  }

  # Optional
  auto_preview = true
  auto_deploy  = true
  timeout_sec  = 600

  # Optional: custom Helm CLI arguments
  # arguments = ["--wait", "--atomic", "--debug"]

  environment_variables = [
    {
      key   = "MY_HELM_VARIABLE"
      value = "my_value"
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

  secrets = [
    {
      key   = "MY_HELM_SECRET"
      value = "my_secret_value"
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

  # Optional: custom domains
  # custom_domains = [
  #   {
  #     domain               = "my-app.example.com"
  #     generate_certificate = true
  #   }
  # ]

  # Optional: control deployment order
  # deployment_stage_id = qovery_deployment_stage.my_stage.id

  deployment_restrictions = [
    {
      mode  = "MATCH"
      type  = "PATH"
      value = "helm/**"
    }
  ]

  advanced_settings_json = jsonencode({
    # Non-exhaustive list. Full list: https://api-doc.qovery.com/#tag/Helms/operation/getDefaultHelmAdvancedSettings
  })

  depends_on = [
    qovery_environment.my_environment,
  ]
}

# Example: Helm chart from a git repository
resource "qovery_helm" "my_helm_from_git" {
  environment_id               = qovery_environment.my_environment.id
  name                         = "my-helm-from-git"
  description                  = "Helm chart from a git repository"
  allow_cluster_wide_resources = false

  source = {
    git_repository = {
      url       = "https://github.com/my-org/my-helm-charts.git"
      branch    = "main"
      root_path = "/charts/my-chart"
      # git_token_id = qovery_git_token.my_git_token.id  # For private repos
    }
  }

  values_override = {
    "set" = {
      "replicaCount" = "2"
    }
  }

  depends_on = [
    qovery_environment.my_environment,
  ]
}
