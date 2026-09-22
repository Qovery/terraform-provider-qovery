data "qovery_database" "my_database" {
  id = "<database_id>"
}

# Access database connection attributes
# data.qovery_database.my_database.internal_host
# data.qovery_database.my_database.external_host
# data.qovery_database.my_database.port
# data.qovery_database.my_database.login
# data.qovery_database.my_database.password

# `password` is a sensitive attribute: an output exposing it must be marked sensitive.
output "database_password" {
  value     = data.qovery_database.my_database.password
  sensitive = true
}
