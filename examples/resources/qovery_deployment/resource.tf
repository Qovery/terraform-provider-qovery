resource "qovery_deployment" "my_deployment" {
  # Required
  environment_id = qovery_environment.my_environment.id
  desired_state  = "RUNNING"

  # Optional - use a random UUID to force redeployment on every apply
  version = "random_uuid_to_force_retrigger_terraform_apply"

  # Ensure all services are created before deploying the environment
  depends_on = [
    qovery_application.my_application,
    qovery_database.my_database,
    qovery_container.my_container,
  ]
}
