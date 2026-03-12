package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PortExternalPortValidator validates that external_port is not set
// when publicly_accessible is false, because the API does not persist
// external_port for non-public ports.
type PortExternalPortValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (v PortExternalPortValidator) Description(_ context.Context) string {
	return "external_port should not be set when publicly_accessible is false"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v PortExternalPortValidator) MarkdownDescription(_ context.Context) string {
	return "external_port should not be set when `publicly_accessible` is `false`"
}

// ValidateObject checks that external_port is not set when publicly_accessible is false.
func (v PortExternalPortValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	publiclyAccessible, ok := attrs["publicly_accessible"].(types.Bool)
	if !ok || publiclyAccessible.IsNull() || publiclyAccessible.IsUnknown() {
		return
	}

	externalPort, ok := attrs["external_port"].(types.Int64)
	if !ok {
		return
	}

	if !publiclyAccessible.ValueBool() && !externalPort.IsNull() && !externalPort.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("external_port"),
			"Invalid Port Configuration",
			"external_port should not be set when publicly_accessible is false — the API does not persist external_port for non-public ports. "+
				"Please remove external_port from this port configuration.",
		)
	}
}
