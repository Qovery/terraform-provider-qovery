package qovery

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// credentialIdentifierFromAPI reads a non-secret identifier of cloud provider credentials, such as
// an AWS access key ID. The API value wins, so an identifier changed outside Terraform shows up in
// the plan. q-core trims the values it stores, so prior (the plan on apply, the state on refresh)
// is kept when it only differs by surrounding whitespace, such as a trailing newline read with
// file().
func credentialIdentifierFromAPI(prior types.String, apiVal *string) types.String {
	if apiVal != nil && !prior.IsNull() && !prior.IsUnknown() && strings.TrimSpace(prior.ValueString()) == *apiVal {
		return prior
	}
	return optionalStringFromAPI(prior, apiVal)
}
