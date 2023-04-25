resource "qovery_deployment" "my_deployment" {
  # Required
  environment_id = qovery_environment.my_environment.id
  desired_state  = "RUNNING"
  version        = "random_uuid_to_force_retrigger_terraform_apply"

  depends_on = [
    qovery_application.my_application,
    qovery_database.my_database,
    qovery_container.my_container,
  ]
}