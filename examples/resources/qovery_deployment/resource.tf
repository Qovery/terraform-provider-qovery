resource "qovery_deployment" "my_deployment" {
  environment_id = qovery_environment.my_environment.id
  desired_state  = "RUNNING"

  # To deploy again without changing desired_state, set `version` to a new UUID.

  # Deploy the environment once its services exist.
  depends_on = [
    qovery_application.my_application,
    qovery_container.my_container,
    qovery_database.my_database,
  ]
}
