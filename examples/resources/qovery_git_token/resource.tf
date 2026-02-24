# Example: GitHub personal access token
resource "qovery_git_token" "github_token" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-github-token"
  type            = "GITHUB"
  token           = var.github_token
  description     = "GitHub token for accessing private repositories"
}

# Example: GitLab token
resource "qovery_git_token" "gitlab_token" {
  organization_id = qovery_organization.my_organization.id
  name            = "my-gitlab-token"
  type            = "GITLAB"
  token           = var.gitlab_token
  description     = "GitLab token for CI/CD pipelines"
}

# Example: Bitbucket token (requires bitbucket_workspace)
resource "qovery_git_token" "bitbucket_token" {
  organization_id     = qovery_organization.my_organization.id
  name                = "my-bitbucket-token"
  type                = "BITBUCKET"
  token               = var.bitbucket_token
  description         = "Bitbucket token for workspace access"
  bitbucket_workspace = "my-workspace"
}
