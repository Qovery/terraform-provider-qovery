# Example: Cron Job using a container image
resource "qovery_job" "my_cron_job" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "my-cron-job"

  # Optional
  auto_preview         = true
  auto_deploy          = true
  cpu                  = 500
  memory               = 512
  max_duration_seconds = 300
  max_nb_restart       = 1
  port                 = 5432

  # Cron job schedule (runs every 2 minutes)
  schedule = {
    cronjob = {
      schedule = "*/2 * * * *"
      command = {
        entrypoint = "/bin/sh"
        arguments  = ["-c", "echo 'Job completed'"]
      }
    }
  }

  # Source: pre-built image from a container registry
  source = {
    image = {
      registry_id = qovery_container_registry.my_container_registry.id
      name        = "debian"
      tag         = "stable"
    }
  }

  healthchecks = {}

  environment_variables = [
    {
      key   = "MY_VARIABLE"
      value = "my_value"
    }
  ]

  secrets = [
    {
      key   = "MY_SECRET"
      value = "my_secret_value"
    }
  ]

  advanced_settings_json = jsonencode({
    # Non-exhaustive list. Full list: https://api-doc.qovery.com/#tag/Jobs/operation/getDefaultJobAdvancedSettings
    "deployment.termination_grace_period_seconds" : 120,
    "build.timeout_max_sec" : 120
  })

  depends_on = [
    qovery_environment.my_environment,
  ]
}

# Example: Lifecycle Job using a Docker source (runs on environment start/stop/delete)
resource "qovery_job" "my_lifecycle_job" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "my-lifecycle-job"

  # Optional
  cpu                  = 1000
  memory               = 1024
  max_duration_seconds = 600
  max_nb_restart       = 0

  # Lifecycle schedule: triggers on environment events
  schedule = {
    on_start = {
      entrypoint = "/bin/sh"
      arguments  = ["-c", "echo 'Environment starting'"]
    }
    on_stop = {
      entrypoint = "/bin/sh"
      arguments  = ["-c", "echo 'Environment stopping'"]
    }
    on_delete = {
      entrypoint = "/bin/sh"
      arguments  = ["-c", "echo 'Environment deleting'"]
    }
  }

  # Source: build from a Dockerfile in a git repository
  source = {
    docker = {
      dockerfile_path = "Dockerfile"
      git_repository = {
        url       = "https://github.com/my-org/my-repo.git"
        branch    = "main"
        root_path = "/"
        # git_token_id = qovery_git_token.my_git_token.id  # For private repos
      }
    }
  }

  healthchecks = {}

  # Optional: control deployment order
  # deployment_stage_id = qovery_deployment_stage.my_stage.id

  # Optional: restrict deployments to specific file changes
  deployment_restrictions = [
    {
      mode  = "MATCH"
      type  = "PATH"
      value = "src/jobs/**"
    }
  ]

  # Optional: attach Kubernetes annotations and labels
  # annotations_group_ids = [qovery_annotations_group.my_annotations.id]
  # labels_group_ids      = [qovery_labels_group.my_labels.id]

  depends_on = [
    qovery_environment.my_environment,
  ]
}
