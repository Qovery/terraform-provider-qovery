resource "qovery_git_token" "my_git_token" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-github-token"
  type            = "GITHUB"
  token           = var.github_token
  description     = "Access to the private repositories of my-org"
}
