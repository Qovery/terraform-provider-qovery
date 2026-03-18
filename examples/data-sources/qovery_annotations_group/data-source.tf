data "qovery_annotations_group" "my_annotations_group" {
  id              = qovery_annotations_group.my_annotations_group.id
  organization_id = qovery_organization.my_organization.id
}
