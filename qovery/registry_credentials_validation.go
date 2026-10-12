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
// and on edit, so plan rejects only a config that apply rejects. A key the kind does not use is
// ignored, as q-core does, and an unknown key counts as set.

const gcpWorkloadIdentityFederation = "workload_identity_federation"

// credentialKey is a key of the config block of a container registry or a helm repository.
type credentialKey struct {
	name  string
	value types.String
}

// readCredentialsConfig reads the kind and the config block of a container registry or a helm
// repository into config. It returns false when the kind or the block is not known yet.
func readCredentialsConfig(ctx context.Context, cfg tfsdk.Config, config any, diags *diag.Diagnostics) (string, bool) {
	var kind types.String
	var block types.Object
	diags.Append(cfg.GetAttribute(ctx, path.Root("kind"), &kind)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("config"), &block)...)
	if diags.HasError() || kind.IsNull() || kind.IsUnknown() || block.IsUnknown() {
		return "", false
	}
	// A null block leaves config at its zero value, where every key is null.
	if !block.IsNull() {
		diags.Append(block.As(ctx, config, basetypes.ObjectAsOptions{})...)
	}
	return kind.ValueString(), !diags.HasError()
}

// validateContainerRegistryConfig checks that config sets the keys q-core requires for kind.
func validateContainerRegistryConfig(kind string, config ContainerRegistryConfig) diag.Diagnostics {
	switch registry.Kind(kind) {
	case registry.KindECR:
		return requireCredentials(kind, awsCredentialKeys(config.Region, config.AccessKeyID, config.SecretAccessKey)...)
	case registry.KindScalewayCR:
		return requireCredentials(kind, scalewayCredentialKeys(config.Region, config.ScalewayAccessKey, config.ScalewaySecretKey, config.ScalewayProjectId)...)
	case registry.KindGithubCr:
		return requireUsernameWithPassword(kind, config.Username, config.Password)
	case registry.KindGcpArtifactRegistry:
		return validateGcpArtifactRegistryConfig(kind, config)
	}
	return nil
}

// validateHelmRepositoryConfig checks that config sets the keys q-core requires for kind.
func validateHelmRepositoryConfig(kind string, config HelmRepositoryConfig) diag.Diagnostics {
	switch helmRepository.Kind(kind) {
	case helmRepository.KindECR:
		return requireCredentials(kind, awsCredentialKeys(config.Region, config.AccessKeyID, config.SecretAccessKey)...)
	case helmRepository.KindScalewayCR:
		return requireCredentials(kind, scalewayCredentialKeys(config.Region, config.ScalewayAccessKey, config.ScalewaySecretKey, config.ScalewayProjectId)...)
	case helmRepository.KindHttps, helmRepository.KindDockerHub, helmRepository.KindGitlabCr, helmRepository.KindGenericCR:
		return requireUsernameAndPasswordTogether(kind, config.Username, config.Password)
	case helmRepository.KindGithubCr:
		if diags := requireUsernameAndPasswordTogether(kind, config.Username, config.Password); diags.HasError() {
			return diags
		}
		return requireUsernameWithPassword(kind, config.Username, config.Password)
	}
	return nil
}

// The provider never sends role_arn, so an ECR registry needs static keys.
func awsCredentialKeys(region, accessKeyID, secretAccessKey types.String) []credentialKey {
	return []credentialKey{{"region", region}, {"access_key_id", accessKeyID}, {"secret_access_key", secretAccessKey}}
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

// requireUsernameAndPasswordTogether rejects a username without a password and the reverse: q-core
// checks this for helm repositories, not for container registries.
func requireUsernameAndPasswordTogether(kind string, username, password types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	switch {
	case !username.IsNull() && password.IsNull():
		addMissingCredential(&diags, "password", "when config.username is set and kind is "+kind)
	case username.IsNull() && !password.IsNull():
		addMissingCredential(&diags, "username", "when config.password is set and kind is "+kind)
	}
	return diags
}

// requireUsernameWithPassword rejects a non-blank password without a non-blank username, which
// q-core checks for GitHub registries before it calls GitHub.
func requireUsernameWithPassword(kind string, username, password types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	if !password.IsUnknown() && !isBlank(password) && isBlank(username) {
		addMissingCredential(&diags, "username", "when config.password is set and kind is "+kind)
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
