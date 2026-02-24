# Container mode database (runs in a container on your cluster)
# Suitable for development and staging environments
resource "qovery_database" "my_container_database" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyContainerPostgres"
  type           = "POSTGRESQL"
  version        = "16"
  mode           = "CONTAINER"

  # Optional (only applicable for CONTAINER mode)
  accessibility = "PRIVATE"
  cpu           = 500
  memory        = 512
  storage       = 20

  depends_on = [
    qovery_environment.my_environment
  ]
}

# Managed mode database (uses cloud provider's managed service, e.g. AWS RDS)
# Recommended for production environments
resource "qovery_database" "my_managed_database" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "MyManagedPostgres"
  type           = "POSTGRESQL"
  version        = "16"
  mode           = "MANAGED"

  # Instance type is required for MANAGED mode (cpu/memory are ignored)
  instance_type = "db.t3.micro"

  # Optional
  accessibility = "PRIVATE"
  storage       = 20

  depends_on = [
    qovery_environment.my_environment
  ]
}

# MySQL container database
resource "qovery_database" "my_mysql" {
  environment_id = qovery_environment.my_environment.id
  name           = "MyMySQLDatabase"
  type           = "MYSQL"
  version        = "8.0"
  mode           = "CONTAINER"
  accessibility  = "PRIVATE"
  storage        = 10

  depends_on = [
    qovery_environment.my_environment
  ]
}

# Redis container database (in-memory data store)
resource "qovery_database" "my_redis" {
  environment_id = qovery_environment.my_environment.id
  name           = "MyRedis"
  type           = "REDIS"
  version        = "7.0"
  mode           = "CONTAINER"
  accessibility  = "PRIVATE"
  storage        = 10

  depends_on = [
    qovery_environment.my_environment
  ]
}

# MongoDB container database
resource "qovery_database" "my_mongodb" {
  environment_id = qovery_environment.my_environment.id
  name           = "MyMongoDB"
  type           = "MONGODB"
  version        = "6.0"
  mode           = "CONTAINER"
  accessibility  = "PRIVATE"
  storage        = 20

  depends_on = [
    qovery_environment.my_environment
  ]
}
