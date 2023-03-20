resource "qovery_deployment_stage" "my_deployment_stage" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyDeploymentStage"

  # Optional
  description = ""
  move_after  = qovery_deployment_stage.first_deployment_stage.id
  move_before = qovery_deployment_stage.third_deployment_stage.id

  depends_on = [
    qovery_environment.my_environment
  ]
}