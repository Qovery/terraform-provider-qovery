resource "qovery_application" "my_application" {
  environment_id = qovery_environment.my_environment.id
  name = "MyApplication"
  git_repository = {
    url = "https://github.com/Qovery/terraform-provider-qovery.git"
    root_path = "/"
  }

  depends_on = [
    qovery_environment.my_environment
  ]
}