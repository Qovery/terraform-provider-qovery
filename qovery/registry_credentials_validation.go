package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// The checks below mirror the config checks q-core runs before it tries the credentials, on create
// and on edit, so plan rejects only a config that apply rejects, except an ECR config that sets
// two authentication methods (see validateECRConfig). A key the kind does not use is ignored, as
// q-core does, and a check passes while a key it reads is unknown.

const gcpWorkloadIdentityFederation = "workload_identity_federation"

// registryPasswordWoPath is the write-only password of a container registry or a helm repository.
var registryPasswordWoPath = path.Root("config").AtName("password_wo")

// credentialKey is a key of the config block of a container registry or a helm repository.
type credentialKey struct {
	name  string
	value types.String
}

// readCredentialsConfig reads the kind and the config block of a container registry or a helm
// repository into config. It returns false when the block is not known yet. The kind is empty
// while it is not known, so only the checks that do not depend on it run.
func readCredentialsConfig(ctx context.Context, cfg tfsdk.Config, config any, diags *diag.Diagnostics) (string, bool) {
	var kind types.String
	var block types.Object
	diags.Append(cfg.GetAttribute(ctx, path.Root("kind"), &kind)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("config"), &block)...)
	if diags.HasError() || block.IsUnknown() {
		return "", false
	}
	// A null block leaves config at its zero value, where every key is null.
	if !block.IsNull() {
		diags.Append(block.As(ctx, config, basetypes.ObjectAsOptions{})...)
	}
	if kind.IsNull() || kind.IsUnknown() {
		return "", !diags.HasError()
	}
	return kind.ValueString(), !diags.HasError()
}

// registryPasswordWoFromConfig reads config.password_wo from the config. It fails when the plan
// holds a config.password_wo_version but no password, and the config no longer backs it with
// config.password_wo, which ignore_changes on config can cause: the API replaces the whole config
// on every write, so the request would clear the password.
func registryPasswordWoFromConfig(ctx context.Context, config tfsdk.Config, plannedPassword, plannedVersion types.String, diags *diag.Diagnostics) types.String {
	passwordWo := writeOnlyFromConfig(ctx, config, registryPasswordWoPath, diags)
	if plannedPassword.IsNull() && isKnownAndSet(plannedVersion) && passwordWo.IsNull() {
		diags.AddAttributeError(
			registryPasswordWoPath,
			"Missing write-only password",
			"The plan keeps config.password_wo_version, but the configuration sets no config.password_wo, so Qovery would store the registry without its password. "+
				"An ignore_changes rule on config can cause this.",
		)
	}
	return passwordWo
}

// validateContainerRegistryConfig checks the password against its write-only form, and that
// config sets the keys q-core requires for kind.
func validateContainerRegistryConfig(kind string, config ContainerRegistryConfig) diag.Diagnostics {
	diags := validateWriteOnly(path.Root("config"), "password", config.Password, config.PasswordWo, config.PasswordWoVersion)
	password := passwordKey(config.Password, config.PasswordWo)
	switch registry.Kind(kind) {
	case registry.KindECR:
		diags.Append(validateECRConfig(kind, config.Region, config.AccessKeyID, config.SecretAccessKey, config.RoleArn)...)
	case registry.KindScalewayCR:
		diags.Append(requireCredentials(kind, scalewayCredentialKeys(config.Region, config.ScalewayAccessKey, config.ScalewaySecretKey, config.ScalewayProjectId)...)...)
	case registry.KindGithubCr:
		diags.Append(requireUsernameWithPassword(kind, config.Username, password)...)
	case registry.KindGcpArtifactRegistry:
		diags.Append(validateGcpArtifactRegistryConfig(kind, config)...)
	}
	return diags
}

// validateHelmRepositoryConfig checks the password against its write-only form, and that config
// sets the keys q-core requires for kind.
func validateHelmRepositoryConfig(kind string, config HelmRepositoryConfig) diag.Diagnostics {
	diags := validateWriteOnly(path.Root("config"), "password", config.Password, config.PasswordWo, config.PasswordWoVersion)
	password := passwordKey(config.Password, config.PasswordWo)
	switch helmRepository.Kind(kind) {
	case helmRepository.KindECR:
		diags.Append(validateECRConfig(kind, config.Region, config.AccessKeyID, config.SecretAccessKey, config.RoleArn)...)
	case helmRepository.KindScalewayCR:
		diags.Append(requireCredentials(kind, scalewayCredentialKeys(config.Region, config.ScalewayAccessKey, config.ScalewaySecretKey, config.ScalewayProjectId)...)...)
	case helmRepository.KindHttps, helmRepository.KindDockerHub, helmRepository.KindGitlabCr, helmRepository.KindGenericCR:
		diags.Append(requireKeysTogether(kind, credentialKey{"username", config.Username}, password)...)
	case helmRepository.KindGithubCr:
		together := requireKeysTogether(kind, credentialKey{"username", config.Username}, password)
		diags.Append(together...)
		if !together.HasError() {
			diags.Append(requireUsernameWithPassword(kind, config.Username, password)...)
		}
	}
	return diags
}

// passwordKey returns the password the config sets, in either form: password_wo when password is
// null, so the kind checks see the write-only password too.
func passwordKey(password, passwordWo types.String) credentialKey {
	if password.IsNull() && !passwordWo.IsNull() {
		return credentialKey{"password_wo", passwordWo}
	}
	return credentialKey{"password", password}
}

// validateECRConfig requires the region and one authentication method: role_arn, or the static
// keys as a pair. q-core accepts both methods at once, but keeps one: the role on a container
// registry, the keys on a helm repository. The API then returns no value for the key of the other
// method, which would not match the config after apply, so both methods are rejected.
func validateECRConfig(kind string, region, accessKeyID, secretAccessKey, roleArn types.String) diag.Diagnostics {
	diags := requireCredentials(kind, credentialKey{"region", region})
	diags.Append(requireKeysTogether(kind, credentialKey{"access_key_id", accessKeyID}, credentialKey{"secret_access_key", secretAccessKey})...)
	// An unknown value may turn out null, so these checks wait until the values are known.
	if roleArn.IsUnknown() || accessKeyID.IsUnknown() || secretAccessKey.IsUnknown() {
		return diags
	}
	switch hasStaticKeys := !accessKeyID.IsNull() || !secretAccessKey.IsNull(); {
	case !roleArn.IsNull() && hasStaticKeys:
		diags.AddAttributeError(
			path.Root("config").AtName("role_arn"),
			"Conflicting credentials",
			fmt.Sprintf("config.role_arn cannot be set with config.access_key_id or config.secret_access_key when kind is %s: set one authentication method.", kind),
		)
	case roleArn.IsNull() && !hasStaticKeys:
		diags.AddAttributeError(
			path.Root("config").AtName("role_arn"),
			"Missing credentials",
			fmt.Sprintf("Either config.role_arn or both config.access_key_id and config.secret_access_key are required when kind is %s.", kind),
		)
	}
	return diags
}

func scalewayCredentialKeys(region, accessKey, secretKey, projectID types.String) []credentialKey {
	return []credentialKey{{"region", region}, {"scaleway_access_key", accessKey}, {"scaleway_secret_key", secretKey}, {"scaleway_project_id", projectID}}
}

// validateGcpArtifactRegistryConfig requires the region and the keys of the authentication mode:
// the Workload Identity Federation keys when gcp_credentials_type selects it, else json_credentials.
// q-core rejects a blank value for these keys, not only a missing one.
func validateGcpArtifactRegistryConfig(kind string, config ContainerRegistryConfig) diag.Diagnostics {
	diags := requireCredentials(kind, credentialKey{"region", config.Region})
	if config.GcpCredentialsType.IsUnknown() {
		return diags
	}
	if config.GcpCredentialsType.ValueString() == gcpWorkloadIdentityFederation {
		for _, key := range []credentialKey{
			{"project_id", config.ProjectId},
			{"service_account_email", config.ServiceAccountEmail},
			{"workload_identity_provider_resource", config.WorkloadIdentityProviderResource},
		} {
			if isBlank(key.value) {
				addMissingCredential(&diags, key.name, fmt.Sprintf("when config.gcp_credentials_type is %s", gcpWorkloadIdentityFederation))
			}
		}
		return diags
	}
	if isBlank(config.JsonCredentials) {
		addMissingCredential(&diags, "json_credentials", fmt.Sprintf("when kind is %s, unless config.gcp_credentials_type is %s", kind, gcpWorkloadIdentityFederation))
	}
	return diags
}

// requireCredentials adds an error for each key that is not set. q-core checks that the key is
// present only, so a blank value passes.
func requireCredentials(kind string, keys ...credentialKey) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, key := range keys {
		if key.value.IsNull() {
			addMissingCredential(&diags, key.name, "when kind is "+kind)
		}
	}
	return diags
}

// requireKeysTogether rejects one key set without the other. q-core checks the username and the
// password of a helm repository this way, not those of a container registry. An unknown value may
// turn out null, so the check waits until both are known.
func requireKeysTogether(kind string, first, second credentialKey) diag.Diagnostics {
	var diags diag.Diagnostics
	if first.value.IsUnknown() || second.value.IsUnknown() {
		return diags
	}
	switch {
	case !first.value.IsNull() && second.value.IsNull():
		addMissingCredential(&diags, second.name, fmt.Sprintf("when config.%s is set and kind is %s", first.name, kind))
	case first.value.IsNull() && !second.value.IsNull():
		addMissingCredential(&diags, first.name, fmt.Sprintf("when config.%s is set and kind is %s", second.name, kind))
	}
	return diags
}

// requireUsernameWithPassword rejects a non-blank password without a non-blank username, which
// q-core checks for GitHub registries before it calls GitHub.
func requireUsernameWithPassword(kind string, username types.String, password credentialKey) diag.Diagnostics {
	var diags diag.Diagnostics
	if !password.value.IsUnknown() && !isBlank(password.value) && isBlank(username) {
		addMissingCredential(&diags, "username", fmt.Sprintf("when config.%s is set and kind is %s", password.name, kind))
	}
	return diags
}

// isBlank reports whether a known value is null, empty or only whitespace.
func isBlank(value types.String) bool {
	return !value.IsUnknown() && strings.TrimSpace(value.ValueString()) == ""
}

func addMissingCredential(diags *diag.Diagnostics, key, condition string) {
	diags.AddAttributeError(
		path.Root("config").AtName(key),
		"Missing credentials",
		fmt.Sprintf("config.%s is required %s.", key, condition),
	)
}
