resource "qovery_argocd_destination_cluster_mapping" "example" {
  organization_id    = var.qovery_organization_id
  agent_cluster_id   = qovery_cluster.agent_cluster.id
  argocd_cluster_url = "https://kubernetes.default.svc"
  cluster_id         = qovery_cluster.target_cluster.id
}
