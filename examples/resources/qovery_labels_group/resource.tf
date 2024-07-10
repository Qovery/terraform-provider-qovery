resource "qovery_labels_group" "labels_group1" {
  organization_id = qovery_organization.my_organization.id
  name            = "MyLabelsGroup"
  labels = [
    {
      key                         = "key1"
      value                       = "value1"
      propagate_to_cloud_provider = false
    },
    {
      key                         = "key2"
      value                       = "value2"
      propagate_to_cloud_provider = true
    }
  ]
}