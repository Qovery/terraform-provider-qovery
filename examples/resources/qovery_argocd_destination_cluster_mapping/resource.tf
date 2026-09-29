# https://kubernetes.default.svc is the in-cluster destination of the ArgoCD instance:
# the cluster ArgoCD runs on, here mapped to that same Qovery cluster.
resource "qovery_argocd_destination_cluster_mapping" "my_argocd_destination_cluster_mapping" {
  organization_id    = qovery_organization.my_organization.id
  agent_cluster_id   = qovery_cluster.my_cluster.id
  argocd_cluster_url = "https://kubernetes.default.svc"
  cluster_id         = qovery_cluster.my_cluster.id
}
