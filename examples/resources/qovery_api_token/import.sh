# Import requires both the organization ID and api token ID, separated by a comma.
# The token value cannot be retrieved from the API: it stays null in the state after an import.
terraform import qovery_api_token.my_api_token "<organization_id>,<api_token_id>"
