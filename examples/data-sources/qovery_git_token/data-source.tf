data "qovery_git_token" "my_git_token" {
  id              = "<git_token_id>"
  organization_id = "<organization_id>"
}

# Access the git token's attributes
# data.qovery_git_token.my_git_token.name
# data.qovery_git_token.my_git_token.type
# data.qovery_git_token.my_git_token.description
