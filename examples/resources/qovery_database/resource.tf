# mode = "MANAGED" uses the managed database service of the cloud provider, such as AWS RDS, and requires instance_type.
resource "qovery_database" "my_database" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-database"
  type           = "POSTGRESQL"
  version        = "17"
  mode           = "CONTAINER"
  accessibility  = "PRIVATE"
  storage        = 20
}
