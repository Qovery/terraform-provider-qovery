resource "qovery_job" "my_job" {
  environment_id = "<my-environment-id>"
  name           = "<my-job-name>"

  schedule = {
    on_start  = {}
    on_stop   = {}
    on_delete = {}
    cronjob = {
      schedule = "* * * * *"
      command = {
        entrypoint = "<my-job-entrypoint>"
        arguments  = ["<my-job-argument-1>", "<my-job-argument-2>"]
      }
    }
  }

  source = {
    image = {
      registry_id = "<my-Qovery-container-registry-id>"
      name        = "<my-image-name>"
      tag         = "<my-image-tag>"
    }
    # or
    docker = {
      git_repository  = "<my-git-repository-url>"
      dockerfile_path = "<my-dockerfile-path-in-the-repo>"
    }
  }
}
