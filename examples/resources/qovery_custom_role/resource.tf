resource "qovery_custom_role" "project_admin" {
  organization_id = qovery_organization.my_organization.id
  name            = "project-admin"
  description     = "Admin on the main project, can create environments on the main cluster"

  cluster_permissions = [
    {
      cluster_id = qovery_cluster.my_cluster.id
      permission = "ENV_CREATOR"
    }
  ]

  project_permissions = [
    {
      project_id = qovery_project.my_project.id
      is_admin   = true
    },
    {
      project_id = qovery_project.my_other_project.id
      permissions = [
        { environment_type = "DEVELOPMENT", permission = "MANAGER" },
        { environment_type = "PREVIEW", permission = "MANAGER" },
        { environment_type = "STAGING", permission = "DEPLOYER" },
        { environment_type = "PRODUCTION", permission = "VIEWER" },
      ]
    }
  ]
}
