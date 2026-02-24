data "qovery_container_registry" "my_container_registry" {
  id              = "<container_registry_id>"
  organization_id = "<organization_id>"
}

# Access container registry attributes
# data.qovery_container_registry.my_container_registry.name
# data.qovery_container_registry.my_container_registry.kind
# data.qovery_container_registry.my_container_registry.url
