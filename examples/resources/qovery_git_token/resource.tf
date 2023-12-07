resource "qovery_git_token" "my_git_token" {
  # Required
  organization_id = qovery_organization.my_organization.id
  name            = "my-git-token"
  type            = "GITHUB"
  token           = "my-git-provider-token"

  # Optional
  description = "Github token"

  # Only necessary for BITBUCKET git tokens
  bitbucket_workspace = "workspace-bitbucket"
}