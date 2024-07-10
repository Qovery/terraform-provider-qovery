resource "qovery_annotations_group" "annotations_group1" {
  organization_id = qovery_organization.my_organization.id
  name            = "MyAnnotationsGroup"
  annotations = {
    "key1" = "value1"
    "key2" = "value2"
  }
  scopes = ["PODS", "DEPLOYMENTS"]
}