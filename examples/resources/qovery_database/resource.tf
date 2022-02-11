resource "qovery_database" "my_database" {
  environment_id = qovery_environment.my_environment.id
  name = "MyDatabase"
  type = "POSTGRESQL"
  version = "10"
  mode = "CONTAINER"
  accessibility = "PRIVATE"
  cpu = 250
  memory = 256
  storage = 10240

  depends_on = [
    qovery_environment.my_environment
  ]
}