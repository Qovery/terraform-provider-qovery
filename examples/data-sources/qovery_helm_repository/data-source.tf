data "qovery_helm_repository" "my_helm_repository" {
  id              = qovery_helm_repository.my_helm_repository.id
  organization_id = qovery_organization.my_organization.id
}
