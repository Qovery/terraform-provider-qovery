package qovery

import "github.com/hashicorp/terraform-plugin-framework/types"

// storedDescriptionDefault is the description q-core stores for a container registry, helm
// repository, custom role or project whose request omits it. Those schemas use it as the
// Default, so removing the description from the configuration plans its reset, and a
// description set outside Terraform shows up in the plan (see "HCL Config Is the Source of
// Truth" in AGENTS.md).
const storedDescriptionDefault = ""

// storedDescriptionFromAPI reads the description of a resource whose schema defaults it to
// storedDescriptionDefault. A null from the API reads as that default, which is what the plan
// holds when the configuration omits the description.
func storedDescriptionFromAPI(apiVal *string) types.String {
	if apiVal == nil {
		return types.StringValue(storedDescriptionDefault)
	}
	return types.StringValue(*apiVal)
}

// upgradeOptionalDescriptionFrom0x turns the empty description that 0.x could store into null.
// description was Optional + Computed in 0.x and kept the "" the API returned for a description
// cleared from the Console; since 1.0 it is Optional only, so keeping "" would plan a
// "\"\" -> null" change on a resource whose configuration omits it.
func upgradeOptionalDescriptionFrom0x(description types.String) types.String {
	if !description.IsNull() && !description.IsUnknown() && description.ValueString() == "" {
		return types.StringNull()
	}
	return description
}
