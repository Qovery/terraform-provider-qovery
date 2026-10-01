# Import uses the blueprint ID, which the Helm or Terraform service it materialized reports as blueprint_id.
# The API returns neither the values of secret_variables nor spec_overrides: they stay null in the state after an import,
# so the first apply writes the values of the configuration and redeploys the service.
# The API does not record deploy either: an import stores its default, true, so a configuration that sets deploy = false
# shows that change in the first plan. Changing only deploy does not redeploy the service.
terraform import qovery_blueprint.my_blueprint "<blueprint_id>"
