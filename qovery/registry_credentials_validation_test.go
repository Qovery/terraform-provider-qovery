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

// errorPaths returns the attribute paths of the errors in diags, sorted.
func errorPaths(diags diag.Diagnostics) []string {
	paths := []string{}
	for _, d := range diags.Errors() {
		if withPath, ok := d.(diag.DiagnosticWithPath); ok {
			paths = append(paths, withPath.Path().String())
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
			ExpectedErrors: []string{"config.access_key_id", "config.region", "config.secret_access_key"},
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
			ExpectedErrors: []string{"config.access_key_id", "config.region", "config.secret_access_key"},
		},
		{
			TestName:       "error_oci_scaleway_cr_without_project_id",
			Kind:           "OCI_SCALEWAY_CR",
			Config:         HelmRepositoryConfig{Region: set, ScalewayAccessKey: set, ScalewaySecretKey: set},
			ExpectedErrors: []string{"config.scaleway_project_id"},
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
	regionOnly := func(o tftypes.Object) tftypes.Value {
		values := map[string]tftypes.Value{}
		for name, attrType := range o.AttributeTypes {
			values[name] = tftypes.NewValue(attrType, nil)
		}
		values["region"] = tftypes.NewValue(tftypes.String, "eu-west-3")
		return tftypes.NewValue(o, values)
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
			ExpectedErrors: []string{"config.access_key_id", "config.region", "config.secret_access_key"},
		},
		{
			TestName:       "error_ecr_region_only",
			Kind:           tftypes.NewValue(tftypes.String, "ECR"),
			Config:         regionOnly,
			ExpectedErrors: []string{"config.access_key_id", "config.secret_access_key"},
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
	assert.Equal(t, []string{"config.access_key_id", "config.region", "config.secret_access_key"}, errorPaths(resp.Diagnostics))
}
