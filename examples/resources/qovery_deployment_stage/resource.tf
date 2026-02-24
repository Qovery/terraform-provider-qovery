resource "qovery_deployment_stage" "my_deployment_stage" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyDeploymentStage"

  # Optional
  description = "Deploy backend services after databases are ready"

  # Position this stage relative to other stages using is_after / is_before
  is_after  = qovery_deployment_stage.first_deployment_stage.id
  is_before = qovery_deployment_stage.third_deployment_stage.id

  depends_on = [
    qovery_environment.my_environment
  ]
}
