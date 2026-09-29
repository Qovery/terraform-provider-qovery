resource "qovery_custom_role" "my_custom_role" {
  organization_id = qovery_organization.my_organization.id
  name            = "developer"
  description     = "Manages the non-production environments of the main project"

  cluster_permissions = [
    {
      cluster_id = qovery_cluster.my_cluster.id
      permission = "ENV_CREATOR"
    }
  ]

  # List every environment type, or set is_admin = true instead of permissions.
  project_permissions = [
    {
      project_id = qovery_project.my_project.id
      permissions = [
        { environment_type = "DEVELOPMENT", permission = "MANAGER" },
        { environment_type = "PREVIEW", permission = "MANAGER" },
        { environment_type = "STAGING", permission = "DEPLOYER" },
        { environment_type = "PRODUCTION", permission = "VIEWER" },
      ]
    }
  ]
}
