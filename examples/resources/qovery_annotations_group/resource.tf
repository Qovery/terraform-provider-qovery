resource "qovery_annotations_group" "my_annotations_group" {
  organization_id = qovery_organization.my_organization.id
  name            = "prometheus-scraping"

  annotations = {
    "prometheus.io/scrape" = "true"
    "prometheus.io/port"   = "8080"
  }

  scopes = ["PODS"]
}
