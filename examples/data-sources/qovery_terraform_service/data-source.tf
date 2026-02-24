data "qovery_terraform_service" "my_terraform_service" {
  id = "<terraform_service_id>"
}

# Access the terraform service's attributes
# data.qovery_terraform_service.my_terraform_service.name
# data.qovery_terraform_service.my_terraform_service.engine
# data.qovery_terraform_service.my_terraform_service.environment_id
