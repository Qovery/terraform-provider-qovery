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

# A container can also run an image Qovery builds from a Dockerfile in a Git repository. Until the
# first build, registry_id, image_name and tag report a placeholder; then the image Qovery last
# built. auto_deploy defaults to false on a Git container: set it to true to deploy on every push.
resource "qovery_container" "my_git_container" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-git-container"

  source = {
    git_repository = {
      url             = "https://github.com/my-org/my-app.git"
      branch          = "main"
      dockerfile_path = "Dockerfile"
    }
  }

  build_settings = {
    timeout_max_sec = 1800
  }

  deployment_restrictions = [
    {
      mode  = "EXCLUDE"
      type  = "PATH"
      value = "docs/"
    }
  ]

  ports = [
    {
      internal_port       = 8080
      publicly_accessible = false
    }
  ]

  healthchecks = {
    readiness_probe = {
      type = {
        tcp = {
          port = 8080
        }
      }
      initial_delay_seconds = 10
      period_seconds        = 10
      timeout_seconds       = 5
      success_threshold     = 1
      failure_threshold     = 3
    }
  }
}

# A container switches in place between the two, keeping its id, volumes and built-in variables:
# replace registry_id, image_name and tag with source and build_settings, or the reverse.
