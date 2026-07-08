data "qovery_api_token" "my_api_token" {
  id              = "<api_token_id>"
  organization_id = "<organization_id>"
}

# Access the api token's metadata
# data.qovery_api_token.my_api_token.name
# data.qovery_api_token.my_api_token.description
# data.qovery_api_token.my_api_token.role_id
#
# The token secret value is NOT available: the API only returns it at creation time.
