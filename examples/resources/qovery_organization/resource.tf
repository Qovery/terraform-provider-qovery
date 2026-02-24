# Qovery organizations cannot be created or deleted via Terraform.
# Use `terraform import` to bring an existing organization under management.
resource "qovery_organization" "my_organization" {
  # Required
  name = "MyOrganization"
  plan = "TEAM"

  # Optional
  description = "Production organization for our SaaS platform"
}
