# A CONTAINER database, which runs on the cluster. For a managed database, such as Amazon RDS,
# declare a separate qovery_blueprint resource instead, for example with blueprint = "AWS/postgres/17".
resource "qovery_database" "my_database" {
  environment_id = qovery_environment.my_environment.id
  name           = "my-database"
  type           = "POSTGRESQL"
  version        = "17"
  mode           = "CONTAINER"
  accessibility  = "PRIVATE"
  storage        = 20
}
