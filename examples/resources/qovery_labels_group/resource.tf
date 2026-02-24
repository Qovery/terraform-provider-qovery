resource "qovery_labels_group" "my_labels_group" {
  organization_id = qovery_organization.my_organization.id
  name            = "MyLabelsGroup"

  labels = [
    {
      key                         = "team"
      value                       = "backend"
      propagate_to_cloud_provider = true
    },
    {
      key                         = "environment"
      value                       = "production"
      propagate_to_cloud_provider = true
    },
    {
      key                         = "managed-by"
      value                       = "qovery"
      propagate_to_cloud_provider = false
    }
  ]
}
