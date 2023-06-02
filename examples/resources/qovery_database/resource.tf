resource "qovery_database" "my_container_database" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyContainerDatabase"
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

resource "qovery_database" "my_managed_database" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyManagedDatabase"
  type           = "POSTGRESQL"
  version        = "10"
  mode           = "MANAGED"

  # Instance type to be set for managed databases
  instance_type = "db.t3.micro"

  # Optional
  accessibility = "PRIVATE"
  storage       = 10

  depends_on = [
    qovery_environment.my_environment
  ]
}