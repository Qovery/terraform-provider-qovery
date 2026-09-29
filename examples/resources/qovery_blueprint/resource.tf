resource "qovery_blueprint" "my_blueprint" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-redis"
  blueprint      = "HELM/redis/8"

  variables = {
    memory_limit = "1Gi"
  }
  secret_variables = {
    password = var.redis_password
  }
}
