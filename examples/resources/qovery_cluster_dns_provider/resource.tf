# A cluster has a single DNS provider: declare one qovery_cluster_dns_provider per cluster.
resource "qovery_cluster_dns_provider" "my_cluster_dns_provider" {
  cluster_id    = qovery_cluster.my_cluster.id
  provider_type = "CLOUDFLARE"
  domain        = "example.com"

  cloudflare = {
    email     = "admin@example.com"
    api_token = var.cloudflare_api_token
  }
}
