resource "qovery_deployment_stage" "my_deployment_stage" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyDeploymentStage"
  description    = ""

  depends_on = [
    qovery_environment.my_environment
  ]
}