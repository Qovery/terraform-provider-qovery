# Terraform 1.0.3+ uses the Terraform Registry:

terraform {
  required_providers {
    qovery = {
      source = "qovery/qovery"
    }
  }
}

# Configure the Qovery provider
provider "qovery" {
  token = "<your-qovery-token>"
}
