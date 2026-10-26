//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// errorPaths returns the attribute paths of the errors in diags, sorted. An error without a path
// counts as "<unscoped>", so a test cannot pass while one is reported.
func errorPaths(diags diag.Diagnostics) []string {
	paths := []string{}
	for _, d := range diags.Errors() {
		if withPath, ok := d.(diag.DiagnosticWithPath); ok {
			paths = append(paths, withPath.Path().String())
		} else {
			paths = append(paths, "<unscoped>")
		}
	}
	sort.Strings(paths)
	return paths
}

func TestValidateContainerRegistryConfig(t *testing.T) {
	t.Parallel()

	set := types.StringValue("value")
	unknown := types.StringUnknown()
	testCases := []struct {
		TestName       string
		Kind           string
		Config         ContainerRegistryConfig
		ExpectedErrors []string
	}{
		{
			TestName:       "error_ecr_without_config",
			Kind:           "ECR",
			ExpectedErrors: []string{"config.region", "config.role_arn"},
		},
		{
			TestName:       "error_ecr_region_only",
			Kind:           "ECR",
			Config:         ContainerRegistryConfig{Region: set},
			ExpectedErrors: []string{"config.role_arn"},
		},
		{
			TestName: "success_ecr_with_role_arn",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: set, RoleArn: set},
		},
		{
			TestName:       "error_ecr_role_arn_with_static_keys",
			Kind:           "ECR",
			Config:         ContainerRegistryConfig{Region: set, AccessKeyID: set, SecretAccessKey: set, RoleArn: set},
			ExpectedErrors: []string{"config.role_arn"},
		},
		{
			TestName:       "error_ecr_role_arn_with_access_key_id",
			Kind:           "ECR",
			Config:         ContainerRegistryConfig{Region: set, AccessKeyID: set, RoleArn: set},
			ExpectedErrors: []string{"config.role_arn", "config.secret_access_key"},
		},
		{
			// The unknown role_arn may turn out null.
			TestName: "success_ecr_unknown_role_arn_with_static_keys",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: set, AccessKeyID: set, SecretAccessKey: set, RoleArn: unknown},
		},
		{
			TestName: "success_ecr_unknown_role_arn",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: set, RoleArn: unknown},
		},
		{
			TestName: "success_ecr_role_arn_with_unknown_access_key_id",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: set, AccessKeyID: unknown, RoleArn: set},
		},
		{
			TestName:       "error_ecr_without_access_key_id",
			Kind:           "ECR",
			Config:         ContainerRegistryConfig{Region: set, SecretAccessKey: set},
			ExpectedErrors: []string{"config.access_key_id"},
		},
		{
			TestName: "success_ecr_unknown_secret_access_key",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: set, AccessKeyID: set, SecretAccessKey: unknown},
		},
		{
			TestName:       "error_ecr_without_secret_access_key",
			Kind:           "ECR",
			Config:         ContainerRegistryConfig{Region: set, AccessKeyID: set},
			ExpectedErrors: []string{"config.secret_access_key"},
		},
		{
			TestName: "success_ecr_with_static_keys",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: set, AccessKeyID: set, SecretAccessKey: set},
		},
		{
			TestName: "success_ecr_with_unknown_keys",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: unknown, AccessKeyID: unknown, SecretAccessKey: unknown},
		},
		{
			// q-core checks that the key is present, not its value.
			TestName: "success_ecr_with_blank_keys",
			Kind:     "ECR",
			Config:   ContainerRegistryConfig{Region: types.StringValue(""), AccessKeyID: types.StringValue(""), SecretAccessKey: types.StringValue("")},
		},
		{
			TestName:       "error_scaleway_without_project_id",
			Kind:           "SCALEWAY_CR",
			Config:         ContainerRegistryConfig{Region: set, ScalewayAccessKey: set, ScalewaySecretKey: set},
			ExpectedErrors: []string{"config.scaleway_project_id"},
		},
		{
			TestName: "success_scaleway",
			Kind:     "SCALEWAY_CR",
			Config:   ContainerRegistryConfig{Region: set, ScalewayAccessKey: set, ScalewaySecretKey: set, ScalewayProjectId: set},
		},
		{
			TestName: "success_public_ecr_anonymous",
			Kind:     "PUBLIC_ECR",
		},
		{
			TestName: "success_docker_hub_anonymous",
			Kind:     "DOCKER_HUB",
		},
		{
			// Container registries accept a username without a password: q-core only skips the login.
			TestName: "success_generic_cr_username_only",
			Kind:     "GENERIC_CR",
			Config:   ContainerRegistryConfig{Username: set},
		},
		{
			TestName: "success_github_cr_anonymous",
			Kind:     "GITHUB_CR",
		},
		{
			TestName:       "error_github_cr_password_without_username",
			Kind:           "GITHUB_CR",
			Config:         ContainerRegistryConfig{Password: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName:       "error_github_cr_password_with_blank_username",
			Kind:           "GITHUB_CR",
			Config:         ContainerRegistryConfig{Username: types.StringValue(" "), Password: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName: "success_github_cr_blank_password_without_username",
			Kind:     "GITHUB_CR",
			Config:   ContainerRegistryConfig{Password: types.StringValue("")},
		},
		{
			TestName: "success_github_cr_unknown_password_without_username",
			Kind:     "GITHUB_CR",
			Config:   ContainerRegistryConfig{Password: unknown},
		},
		{
			TestName: "success_github_enterprise_cr_password_only",
			Kind:     "GITHUB_ENTERPRISE_CR",
			Config:   ContainerRegistryConfig{Password: set},
		},
		{
			TestName: "success_generic_cr_write_only_password",
			Kind:     "GENERIC_CR",
			Config:   ContainerRegistryConfig{Username: set, PasswordWo: set, PasswordWoVersion: set},
		},
		{
			TestName:       "error_generic_cr_password_and_write_only_password",
			Kind:           "GENERIC_CR",
			Config:         ContainerRegistryConfig{Username: set, Password: set, PasswordWo: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.password_wo"},
		},
		{
			TestName:       "error_generic_cr_write_only_password_without_version",
			Kind:           "GENERIC_CR",
			Config:         ContainerRegistryConfig{Username: set, PasswordWo: set},
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			TestName:       "error_generic_cr_version_without_write_only_password",
			Kind:           "GENERIC_CR",
			Config:         ContainerRegistryConfig{Username: set, Password: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			// The pair is checked whatever the kind: the version alone means nothing anywhere.
			TestName:       "error_ecr_version_without_write_only_password",
			Kind:           "ECR",
			Config:         ContainerRegistryConfig{Region: set, RoleArn: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			TestName:       "error_github_cr_write_only_password_without_username",
			Kind:           "GITHUB_CR",
			Config:         ContainerRegistryConfig{PasswordWo: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName: "success_github_cr_write_only_password_with_username",
			Kind:     "GITHUB_CR",
			Config:   ContainerRegistryConfig{Username: set, PasswordWo: set, PasswordWoVersion: set},
		},
		{
			TestName: "success_github_cr_unknown_write_only_password_without_username",
			Kind:     "GITHUB_CR",
			Config:   ContainerRegistryConfig{PasswordWo: unknown, PasswordWoVersion: set},
		},
		{
			TestName:       "error_gcp_without_config",
			Kind:           "GCP_ARTIFACT_REGISTRY",
			ExpectedErrors: []string{"config.json_credentials", "config.region"},
		},
		{
			TestName: "success_gcp_json_key",
			Kind:     "GCP_ARTIFACT_REGISTRY",
			Config:   ContainerRegistryConfig{Region: set, JsonCredentials: set},
		},
		{
			TestName:       "error_gcp_blank_json_key",
			Kind:           "GCP_ARTIFACT_REGISTRY",
			Config:         ContainerRegistryConfig{Region: set, JsonCredentials: types.StringValue("")},
			ExpectedErrors: []string{"config.json_credentials"},
		},
		{
			TestName: "success_gcp_workload_identity_federation",
			Kind:     "GCP_ARTIFACT_REGISTRY",
			Config: ContainerRegistryConfig{
				Region:                           set,
				GcpCredentialsType:               types.StringValue("workload_identity_federation"),
				ProjectId:                        set,
				ServiceAccountEmail:              set,
				WorkloadIdentityProviderResource: unknown,
			},
		},
		{
			TestName: "error_gcp_workload_identity_federation_missing_keys",
			Kind:     "GCP_ARTIFACT_REGISTRY",
			Config: ContainerRegistryConfig{
				Region:             set,
				GcpCredentialsType: types.StringValue("workload_identity_federation"),
				ProjectId:          types.StringValue(""),
			},
			ExpectedErrors: []string{"config.project_id", "config.service_account_email", "config.workload_identity_provider_resource"},
		},
		{
			TestName: "success_gcp_unknown_credentials_type",
			Kind:     "GCP_ARTIFACT_REGISTRY",
			Config:   ContainerRegistryConfig{Region: set, GcpCredentialsType: unknown},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			diags := validateContainerRegistryConfig(tc.Kind, tc.Config)
			expected := tc.ExpectedErrors
			if expected == nil {
				expected = []string{}
			}
			assert.Equal(t, expected, errorPaths(diags), "diagnostics: %v", diags)
		})
	}
}

func TestValidateHelmRepositoryConfig(t *testing.T) {
	t.Parallel()

	set := types.StringValue("value")
	unknown := types.StringUnknown()
	testCases := []struct {
		TestName       string
		Kind           string
		Config         HelmRepositoryConfig
		ExpectedErrors []string
	}{
		{
			TestName: "success_https_anonymous",
			Kind:     "HTTPS",
		},
		{
			TestName: "success_https_username_and_password",
			Kind:     "HTTPS",
			Config:   HelmRepositoryConfig{Username: set, Password: set},
		},
		{
			TestName:       "error_https_username_without_password",
			Kind:           "HTTPS",
			Config:         HelmRepositoryConfig{Username: set},
			ExpectedErrors: []string{"config.password"},
		},
		{
			TestName:       "error_oci_generic_cr_password_without_username",
			Kind:           "OCI_GENERIC_CR",
			Config:         HelmRepositoryConfig{Password: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName: "success_oci_docker_hub_unknown_password",
			Kind:     "OCI_DOCKER_HUB",
			Config:   HelmRepositoryConfig{Username: set, Password: types.StringUnknown()},
		},
		{
			// The unknown username may turn out null, which is anonymous access.
			TestName: "success_https_unknown_username_without_password",
			Kind:     "HTTPS",
			Config:   HelmRepositoryConfig{Username: types.StringUnknown()},
		},
		{
			TestName: "success_https_unknown_password_without_username",
			Kind:     "HTTPS",
			Config:   HelmRepositoryConfig{Password: types.StringUnknown()},
		},
		{
			TestName: "success_oci_github_cr_unknown_username_without_password",
			Kind:     "OCI_GITHUB_CR",
			Config:   HelmRepositoryConfig{Username: types.StringUnknown()},
		},
		{
			TestName:       "error_oci_github_cr_password_without_username",
			Kind:           "OCI_GITHUB_CR",
			Config:         HelmRepositoryConfig{Password: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName:       "error_oci_github_cr_blank_username",
			Kind:           "OCI_GITHUB_CR",
			Config:         HelmRepositoryConfig{Username: types.StringValue(""), Password: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName:       "error_oci_ecr_without_config",
			Kind:           "OCI_ECR",
			ExpectedErrors: []string{"config.region", "config.role_arn"},
		},
		{
			TestName:       "error_oci_ecr_region_only",
			Kind:           "OCI_ECR",
			Config:         HelmRepositoryConfig{Region: set},
			ExpectedErrors: []string{"config.role_arn"},
		},
		{
			TestName: "success_oci_ecr_with_static_keys",
			Kind:     "OCI_ECR",
			Config:   HelmRepositoryConfig{Region: set, AccessKeyID: set, SecretAccessKey: set},
		},
		{
			TestName: "success_oci_ecr_with_role_arn",
			Kind:     "OCI_ECR",
			Config:   HelmRepositoryConfig{Region: set, RoleArn: set},
		},
		{
			TestName:       "error_oci_ecr_role_arn_with_static_keys",
			Kind:           "OCI_ECR",
			Config:         HelmRepositoryConfig{Region: set, AccessKeyID: set, SecretAccessKey: set, RoleArn: set},
			ExpectedErrors: []string{"config.role_arn"},
		},
		{
			TestName: "success_oci_ecr_unknown_role_arn",
			Kind:     "OCI_ECR",
			Config:   HelmRepositoryConfig{Region: set, RoleArn: unknown},
		},
		{
			TestName:       "error_oci_scaleway_cr_without_project_id",
			Kind:           "OCI_SCALEWAY_CR",
			Config:         HelmRepositoryConfig{Region: set, ScalewayAccessKey: set, ScalewaySecretKey: set},
			ExpectedErrors: []string{"config.scaleway_project_id"},
		},
		{
			TestName: "success_https_username_and_write_only_password",
			Kind:     "HTTPS",
			Config:   HelmRepositoryConfig{Username: set, PasswordWo: set, PasswordWoVersion: set},
		},
		{
			TestName:       "error_https_write_only_password_without_username",
			Kind:           "HTTPS",
			Config:         HelmRepositoryConfig{PasswordWo: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName:       "error_https_username_with_version_only",
			Kind:           "HTTPS",
			Config:         HelmRepositoryConfig{Username: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.password", "config.password_wo_version"},
		},
		{
			TestName:       "error_https_password_and_write_only_password",
			Kind:           "HTTPS",
			Config:         HelmRepositoryConfig{Username: set, Password: set, PasswordWo: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.password_wo"},
		},
		{
			TestName:       "error_https_write_only_password_without_version",
			Kind:           "HTTPS",
			Config:         HelmRepositoryConfig{Username: set, PasswordWo: set},
			ExpectedErrors: []string{"config.password_wo_version"},
		},
		{
			// The ephemeral value is unknown until the ephemeral resource opens.
			TestName: "success_https_username_and_unknown_write_only_password",
			Kind:     "HTTPS",
			Config:   HelmRepositoryConfig{Username: set, PasswordWo: unknown, PasswordWoVersion: set},
		},
		{
			TestName:       "error_oci_github_cr_write_only_password_with_blank_username",
			Kind:           "OCI_GITHUB_CR",
			Config:         HelmRepositoryConfig{Username: types.StringValue(""), PasswordWo: set, PasswordWoVersion: set},
			ExpectedErrors: []string{"config.username"},
		},
		{
			TestName: "success_oci_public_ecr_ignores_config",
			Kind:     "OCI_PUBLIC_ECR",
			Config:   HelmRepositoryConfig{Username: set},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			diags := validateHelmRepositoryConfig(tc.Kind, tc.Config)
			expected := tc.ExpectedErrors
			if expected == nil {
				expected = []string{}
			}
			assert.Equal(t, expected, errorPaths(diags), "diagnostics: %v", diags)
		})
	}
}

func TestRequireUsernameWithPassword_NamesTheWriteOnlyPassword(t *testing.T) {
	t.Parallel()

	diags := validateContainerRegistryConfig("GITHUB_CR", ContainerRegistryConfig{PasswordWo: types.StringValue("token"), PasswordWoVersion: types.StringValue("1")})

	assert.Equal(t, "config.username is required when config.password_wo is set and kind is GITHUB_CR.", diags.Errors()[0].Detail())
}

// credentialsConfig builds the config of r with kind and config set and every other attribute null.
func credentialsConfig(t *testing.T, r resource.Resource, kind tftypes.Value, config func(tftypes.Object) tftypes.Value) tfsdk.Config {
	t.Helper()
	s := schemaFor(t, r)
	objType := s.Type().TerraformType(context.Background()).(tftypes.Object)
	values := map[string]tftypes.Value{}
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["kind"] = kind
	values["config"] = config(objType.AttributeTypes["config"].(tftypes.Object))
	return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, values)}
}

func TestContainerRegistryResource_ValidateConfig(t *testing.T) {
	t.Parallel()

	nullBlock := func(o tftypes.Object) tftypes.Value { return tftypes.NewValue(o, nil) }
	unknownBlock := func(o tftypes.Object) tftypes.Value { return tftypes.NewValue(o, tftypes.UnknownValue) }
	ecrConfig := func(keys map[string]string) func(tftypes.Object) tftypes.Value {
		return func(o tftypes.Object) tftypes.Value {
			values := map[string]tftypes.Value{}
			for name, attrType := range o.AttributeTypes {
				values[name] = tftypes.NewValue(attrType, nil)
			}
			for name, value := range keys {
				values[name] = tftypes.NewValue(tftypes.String, value)
			}
			return tftypes.NewValue(o, values)
		}
	}

	testCases := []struct {
		TestName       string
		Kind           tftypes.Value
		Config         func(tftypes.Object) tftypes.Value
		ExpectedErrors []string
	}{
		{
			TestName:       "error_ecr_null_config",
			Kind:           tftypes.NewValue(tftypes.String, "ECR"),
			Config:         nullBlock,
			ExpectedErrors: []string{"config.region", "config.role_arn"},
		},
		{
			TestName:       "error_ecr_region_only",
			Kind:           tftypes.NewValue(tftypes.String, "ECR"),
			Config:         ecrConfig(map[string]string{"region": "eu-west-3"}),
			ExpectedErrors: []string{"config.role_arn"},
		},
		{
			TestName: "success_ecr_role_arn",
			Kind:     tftypes.NewValue(tftypes.String, "ECR"),
			Config:   ecrConfig(map[string]string{"region": "eu-west-3", "role_arn": "arn:aws:iam::123456789012:role/qovery"}),
		},
		{
			TestName: "success_ecr_unknown_config",
			Kind:     tftypes.NewValue(tftypes.String, "ECR"),
			Config:   unknownBlock,
		},
		{
			TestName: "success_unknown_kind",
			Kind:     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			Config:   nullBlock,
		},
		{
			// The write-only rules do not depend on the kind.
			TestName:       "error_unknown_kind_password_and_write_only_password",
			Kind:           tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			Config:         ecrConfig(map[string]string{"password": "plain", "password_wo": "secret", "password_wo_version": "1"}),
			ExpectedErrors: []string{"config.password_wo"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			r := containerRegistryResource{}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: credentialsConfig(t, r, tc.Kind, tc.Config)}, resp)
			expected := tc.ExpectedErrors
			if expected == nil {
				expected = []string{}
			}
			assert.Equal(t, expected, errorPaths(resp.Diagnostics), "diagnostics: %v", resp.Diagnostics)
		})
	}
}

func TestHelmRepositoryResource_ValidateConfig(t *testing.T) {
	t.Parallel()

	r := helmRepositoryResource{}
	resp := &resource.ValidateConfigResponse{}
	config := credentialsConfig(t, r, tftypes.NewValue(tftypes.String, "OCI_ECR"), func(o tftypes.Object) tftypes.Value {
		return tftypes.NewValue(o, nil)
	})
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: config}, resp)
	assert.Equal(t, []string{"config.region", "config.role_arn"}, errorPaths(resp.Diagnostics))
}
