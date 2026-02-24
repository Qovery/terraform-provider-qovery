data "qovery_labels_group" "my_labels_group" {
  id              = qovery_labels_group.my_labels_group.id
  organization_id = qovery_organization.my_organization.id
}
