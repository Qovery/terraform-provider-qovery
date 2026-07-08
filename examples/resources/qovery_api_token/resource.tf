# The token value is only returned at creation time and is stored in the Terraform state.
# Use an encrypted remote state with restricted access.
resource "qovery_api_token" "my_api_token" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-api-token"
  description     = "API token for the delegated terraform workspace"
  role_id         = var.role_id # built-in or custom role id
}

# The API has no update endpoint: every attribute change forces a replacement.
# Rotate a token explicitly with:
#   terraform apply -replace=qovery_api_token.my_api_token
