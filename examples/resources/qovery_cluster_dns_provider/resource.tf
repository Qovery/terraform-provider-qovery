# Choose one DNS provider resource per cluster.

# Qovery DNS provider: use the domain returned by the Qovery API/import for this cluster.
resource "qovery_cluster_dns_provider" "qovery" {
  cluster_id    = qovery_cluster.my_cluster.id
  provider_type = "QOVERY"
  domain        = "my-cluster.qovery.dev"
}

# Cloudflare DNS provider.
resource "qovery_cluster_dns_provider" "cloudflare" {
  cluster_id    = qovery_cluster.my_cluster.id
  provider_type = "CLOUDFLARE"
  domain        = "example.com"

  cloudflare = {
    email     = "admin@example.com"
    api_token = var.cloudflare_api_token
    proxied   = false
  }
}

# Route53 DNS provider.
resource "qovery_cluster_dns_provider" "route53" {
  cluster_id    = qovery_cluster.my_cluster.id
  provider_type = "ROUTE53"
  domain        = "example.com"

  route53 = {
    credentials = {
      type                  = "STATIC"
      aws_access_key_id     = var.aws_access_key_id
      aws_secret_access_key = var.aws_secret_access_key
    }

    aws_region     = "eu-west-3"
    hosted_zone_id = "Z0965488LE74BWDEVQDB"
  }
}
