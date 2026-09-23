terraform {
  required_providers {
    qovery = {
      source  = "qovery/qovery"
      version = "~> 1.0"
    }
  }
}

# Configure the Qovery provider
provider "qovery" {
  token = "<your-qovery-token>"
}
