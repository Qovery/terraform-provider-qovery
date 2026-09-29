terraform {
  required_providers {
    qovery = {
      source  = "qovery/qovery"
      version = "~> 1.0"
    }
  }
}

# The provider reads the API token from the QOVERY_API_TOKEN environment variable.
provider "qovery" {}
