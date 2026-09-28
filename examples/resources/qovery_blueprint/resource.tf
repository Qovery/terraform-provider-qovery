resource "qovery_blueprint" "my_postgres" {
  # Required
  environment_id = qovery_environment.my_environment.id
  name           = "my-postgres"
  blueprint      = "AWS/postgres/17"

  # Optional
  icon_uri = "app://qovery-console/postgresql"
  variables = {
    db_name           = "app"
    instance_class    = "db.t3.micro"
    allocated_storage = "20"
  }
  secret_variables = {
    db_password = var.db_password
  }
  spec_overrides = {
    engine_version = "1.13.3"
    timeout        = 3600
  }
  deploy = true
}
