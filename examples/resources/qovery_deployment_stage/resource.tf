resource "qovery_deployment_stage" "my_deployment_stage" {
  environment_id = qovery_environment.my_environment.id
  name           = "backend"
  description    = "Deploy backend services after databases are ready"
}
