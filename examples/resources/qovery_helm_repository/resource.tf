resource "qovery_helm_repository" "my_helm_repository" {
  organization_id       = qovery_organization.my_organization.id
  name                  = "podinfo"
  kind                  = "HTTPS"
  url                   = "https://stefanprodan.github.io/podinfo"
  skip_tls_verification = false
}
