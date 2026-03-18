data "qovery_helm" "my_helm" {
  id = "<helm_id>"
}

# Access the helm service's attributes
# data.qovery_helm.my_helm.name
# data.qovery_helm.my_helm.environment_id
# data.qovery_helm.my_helm.external_host
