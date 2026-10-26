package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Write-only arguments, Terraform 1.11 and later: a secret attribute <name> has a write-only form
// <name>_wo, which reaches the API but never the plan or the state, and a <name>_wo_version kept
// in the state. The contract is trigger-only: changing the version plans an update, which sends
// the current value, as does any other update of the resource, and changing the value alone
// plans nothing.
//
// The framework nulls every attribute marked WriteOnly in the plan and in the state it returns to
// Terraform. Create and Update therefore read the value from the config and hand it to the API
// request only, never to a model field. The converters copy <name>_wo_version from prior like any
// other attribute: from the plan after a write, from the state on refresh, null on import.

// writeOnlyFromConfig reads the write-only attribute at p from the config. It is null when the
// config does not set it, including when a parent of p is null.
func writeOnlyFromConfig(ctx context.Context, config tfsdk.Config, p path.Path, diags *diag.Diagnostics) types.String {
	var value types.String
	diags.Append(config.GetAttribute(ctx, p, &value)...)
	return value
}

// writeOnlyOrPlain returns the secret to send to the API: the write-only value when config sets
// it, else the plain value from the plan.
func writeOnlyOrPlain(plain, writeOnly types.String) *string {
	if isKnownAndSet(writeOnly) {
		return writeOnly.ValueStringPointer()
	}
	return ToStringPointer(plain)
}

// validateWriteOnly checks the attribute name under parent against its write-only form: the plain
// and write-only forms conflict, the write-only form requires its version, and a version requires
// the write-only form. An unknown value may turn out null, so each check waits until the values
// it reads are known.
func validateWriteOnly(parent path.Path, name string, plain, writeOnly, version types.String) diag.Diagnostics {
	plainPath := parent.AtName(name)
	writeOnlyPath := parent.AtName(name + "_wo")
	versionPath := parent.AtName(name + "_wo_version")

	var diags diag.Diagnostics
	if isKnownAndSet(plain) && isKnownAndSet(writeOnly) {
		diags.AddAttributeError(
			writeOnlyPath,
			"Conflicting secret",
			fmt.Sprintf("%s cannot be set with %s: set one of them.", writeOnlyPath, plainPath),
		)
	}
	if isKnownAndSet(writeOnly) && version.IsNull() {
		diags.AddAttributeError(
			versionPath,
			"Missing write-only version",
			fmt.Sprintf("%s is required when %s is set: change it to send a new value.", versionPath, writeOnlyPath),
		)
	}
	if isKnownAndSet(version) && writeOnly.IsNull() {
		diags.AddAttributeError(
			versionPath,
			"Missing write-only value",
			fmt.Sprintf("%s can only be set with %s.", versionPath, writeOnlyPath),
		)
	}
	return diags
}

// isKnownAndSet reports whether a value is known and not null.
func isKnownAndSet(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

// writeOnlyDescription describes <name>_wo, the write-only form of the secret attribute name.
func writeOnlyDescription(name string) string {
	return fmt.Sprintf("Write-only form of `%[1]s`, which Terraform 1.11 and later send to Qovery without storing it in the plan or the state. "+
		"Changing the value alone does not trigger an update: change `%[1]s_wo_version` to send it.", name)
}

// writeOnlyVersionDescription describes <name>_wo_version, the change token of <name>_wo.
func writeOnlyVersionDescription(name string) string {
	return fmt.Sprintf("Version of `%[1]s_wo`, any string, required with it and kept in the state. "+
		"Changing it sends the current value of `%[1]s_wo`, as does any other update of the resource.", name)
}
