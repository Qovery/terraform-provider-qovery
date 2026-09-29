resource "qovery_helm" "my_helm" {
  environment_id               = qovery_environment.my_environment.id
  name                         = "podinfo"
  description                  = "podinfo chart deployed from its Helm repository"
  allow_cluster_wide_resources = false

  source = {
    helm_repository = {
      helm_repository_id = qovery_helm_repository.my_helm_repository.id
      chart_name         = "podinfo"
      chart_version      = "6.15.0"
    }
  }

  values_override = {
    set = {
      "replicaCount" = "2"
    }
  }
}
