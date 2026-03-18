data "qovery_application" "my_application" {
  id = "<application_id>"
}

# Access application attributes
# data.qovery_application.my_application.name
# data.qovery_application.my_application.internal_host
# data.qovery_application.my_application.external_host
# data.qovery_application.my_application.git_repository
