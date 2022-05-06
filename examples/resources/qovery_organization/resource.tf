resource "qovery_organization" "my_organization" {
  # Required
  name = "MyOrganization"
  plan = "FREE"

  # Optional
  description = "My organization description"
}