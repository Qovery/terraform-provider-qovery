resource "qovery_annotations_group" "my_annotations_group" {
  organization_id = qovery_organization.my_organization.id
  name            = "MyAnnotationsGroup"

  annotations = {
    "prometheus.io/scrape" = "true"
    "prometheus.io/port"   = "8080"
  }

  # Annotations will be applied to pods and deployments
  scopes = ["PODS", "DEPLOYMENTS"]
}
