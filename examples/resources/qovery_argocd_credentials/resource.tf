resource "qovery_argocd_credentials" "my_argocd_credentials" {
  cluster_id   = qovery_cluster.my_cluster.id
  argocd_url   = "https://argocd.example.com"
  argocd_token = var.argocd_token
}
