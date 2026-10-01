resource "qovery_labels_group" "my_labels_group" {
  organization_id = qovery_organization.my_organization.id
  name            = "team-backend"

  labels = [
    {
      key                         = "team"
      value                       = "backend"
      propagate_to_cloud_provider = true
    }
  ]
}
