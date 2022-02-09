# Terraform 0.13+ uses the Terraform Registry:

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
