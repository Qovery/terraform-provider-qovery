resource "qovery_helm" "my_helm" {
  # Required
  environment_id               = qovery_environment.my_environment.id
  name                         = "test-helm"
  allow_cluster_wide_resources = false

  source = {
    helm_repository = {
      helm_repository_id = "5a4a2dd6-02e1-4e3a-a3cc-8ebb97e135a9"
      chart_name         = "httpbin"
      chart_version      = "1.0.0"
    }
  }

  values_override = {
    "set" = {
      "key1" = 6600
      "key2" = "values1"
    }
    "set_string" = {
      "s-key1" = "value1"
      "s-key2" = "value2"
    }
    "set_json" = {
      "j-key1" = "{}"
      "j-key2" = "{}"
    }
    file = {
      raw = {
        file1 = {
          content = "--- \n ssss"
        }
        file2 = {
          content = "a \n eee"
        }
      }
    }
  }



  # Optional
  auto_preview = "true"


  environment_variables = [
    {
      key   = "MY_TERRAFORM_HELM_VARIABLE"
      value = "MY_TERRAFORM_HELM_VARIABLE_VALUE"
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
      key   = "MY_TERRAFORM_HELM_SECRET"
      value = "MY_TERRAFORM_HELM_SECRET_VALUE"
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


  advanced_settings_json = jsonencode({
    # non exhaustive list, the complete list is available in Qovery API doc: https://api-doc.qovery.com/#tag/Helms/operation/getDefaultHelmAdvancedSettings
    # you can only indicate settings that you need to override
  })

  depends_on = [
    qovery_environment.my_environment,
  ]
}
