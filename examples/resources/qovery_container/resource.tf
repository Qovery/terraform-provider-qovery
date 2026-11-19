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

# A container can switch in place to an image Qovery builds from a Dockerfile in a Git
# repository, and back: it keeps its id, volumes and built-in variables. Qovery cannot create
# a container from a Git repository, so create it with an image as above and apply first.
# Then replace registry_id, image_name and tag with source:
#
# resource "qovery_container" "my_container" {
#   environment_id = qovery_environment.my_environment.id
#   name           = "my-container"
#
#   source = {
#     git_repository = {
#       url             = "https://github.com/my-org/my-app.git"
#       branch          = "main"
#       dockerfile_path = "Dockerfile"
#     }
#   }
#
#   build_settings = {
#     timeout_max_sec = 1800
#   }
#
#   # ports, healthchecks and environment_variables as above
# }
#
# On a Git container, registry_id, image_name and tag report the image Qovery last built, and
# auto_deploy is false. To switch back, set registry_id, image_name and tag again and remove
# source and build_settings.
