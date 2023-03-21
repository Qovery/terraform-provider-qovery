resource "qovery_database" "my_database" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyDatabase"
  type           = "POSTGRESQL"
  version        = "10"
  mode           = "CONTAINER"

  # Optional
  accessibility = "PRIVATE"
  cpu           = 250
  memory        = 256
  storage       = 10

  depends_on = [
    qovery_environment.my_environment
  ]
}