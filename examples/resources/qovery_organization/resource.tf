# Qovery organizations cannot be created or deleted via Terraform: import an existing one.
import {
  to = qovery_organization.my_organization
  id = "<organization_id>"
}

resource "qovery_organization" "my_organization" {
  name        = "my-organization"
  plan        = "TEAM"
  description = "Production organization for our SaaS platform"
}
