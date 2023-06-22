resource "qovery_deployment_stage" "my_deployment_stage" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyDeploymentStage"

  # Optional
  description = ""
  is_after    = qovery_deployment_stage.first_deployment_stage.id
  is_before   = qovery_deployment_stage.third_deployment_stage.id

  depends_on = [
    qovery_environment.my_environment
  ]
}