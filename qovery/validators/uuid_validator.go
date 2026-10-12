package validators

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ validator.String = UUIDValidator{}
	_ validator.Set    = UUIDValidator{}
)

// UUIDValidator validates that a string, or every string of a set, is a UUID. It is meant for the
// attributes that reference a Qovery entity (organization_id, cluster_id, ...): it parses with
// uuid.Parse, as the services do before calling the API, so plan rejects only what apply rejects.
// Null and unknown values pass: a reference to another resource is unknown until it is created.
type UUIDValidator struct{}

// NewUUIDValidator returns a validator for an attribute that references a Qovery entity.
func NewUUIDValidator() UUIDValidator {
	return UUIDValidator{}
}

// Description returns a plain text description of the validator's behavior.
func (v UUIDValidator) Description(_ context.Context) string {
	return "value must be a UUID"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v UUIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString checks that a string attribute is a UUID.
func (v UUIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	resp.Diagnostics.Append(validateUUID(req.Path, req.Path, req.ConfigValue)...)
}

// ValidateSet checks that every element of a set of strings is a UUID.
func (v UUIDValidator) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	for _, element := range req.ConfigValue.Elements() {
		if str, ok := element.(types.String); ok {
			resp.Diagnostics.Append(validateUUID(req.Path.AtSetValue(element), req.Path, str)...)
		}
	}
}

// validateUUID reports an invalid value at valuePath, naming the attribute at attributePath.
func validateUUID(valuePath, attributePath path.Path, str types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	if str.IsNull() || str.IsUnknown() {
		return diags
	}
	if _, err := uuid.Parse(str.ValueString()); err != nil {
		diags.AddAttributeError(
			valuePath,
			"Invalid UUID",
			fmt.Sprintf("%s must be the UUID of a Qovery entity, such as the id attribute of the resource it references, got: %q.", attributePath, str.ValueString()),
		)
	}
	return diags
}
