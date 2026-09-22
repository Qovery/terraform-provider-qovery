//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sensitiveFlagsOfResource flattens a resource schema into a map of attribute
// path to its Sensitive flag. Attributes nested in a single-nested object are
// keyed as "<parent>.<child>".
func sensitiveFlagsOfResource(t *testing.T, r resource.Resource) map[string]bool {
	t.Helper()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "resource schema returned diagnostics")

	flags := make(map[string]bool)
	for name, attribute := range resp.Schema.Attributes {
		flags[name] = attribute.IsSensitive()

		nested, ok := attribute.(rsschema.NestedAttribute)
		if !ok {
			continue
		}
		for childName, child := range nested.GetNestedObject().GetAttributes() {
			flags[name+"."+childName] = child.IsSensitive()
		}
	}
	return flags
}

// sensitiveFlagsOfDataSource is the data source counterpart of
// sensitiveFlagsOfResource.
func sensitiveFlagsOfDataSource(t *testing.T, d datasource.DataSource) map[string]bool {
	t.Helper()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "data source schema returned diagnostics")

	flags := make(map[string]bool)
	for name, attribute := range resp.Schema.Attributes {
		flags[name] = attribute.IsSensitive()

		nested, ok := attribute.(dsschema.NestedAttribute)
		if !ok {
			continue
		}
		for childName, child := range nested.GetNestedObject().GetAttributes() {
			flags[name+"."+childName] = child.IsSensitive()
		}
	}
	return flags
}

// TestSchemaSensitiveAttributes locks in which credential attributes are
// redacted from plan output. Secret keys and passwords must be Sensitive;
// the matching identifiers (access key IDs, project IDs, usernames, regions)
// must not be, so they stay readable in plans and outputs.
func TestSchemaSensitiveAttributes(t *testing.T) {
	t.Parallel()

	schemas := map[string]map[string]bool{
		"qovery_database":           sensitiveFlagsOfResource(t, newDatabaseResource()),
		"qovery_container_registry": sensitiveFlagsOfResource(t, newContainerRegistryResource()),
		"qovery_helm_repository":    sensitiveFlagsOfResource(t, newHelmRepositoryResource()),
		"qovery_aws_credentials":    sensitiveFlagsOfResource(t, newAwsCredentialsResource()),
		"data.qovery_database":      sensitiveFlagsOfDataSource(t, newDatabaseDataSource()),
	}

	testCases := []struct {
		TestName        string
		Schema          string
		Attribute       string
		ExpectSensitive bool
	}{
		// Secrets: must never appear in plan output.
		{"database_password", "qovery_database", "password", true},
		{"database_data_source_password", "data.qovery_database", "password", true},
		{"container_registry_secret_access_key", "qovery_container_registry", "config.secret_access_key", true},
		{"container_registry_scaleway_secret_key", "qovery_container_registry", "config.scaleway_secret_key", true},
		{"container_registry_password", "qovery_container_registry", "config.password", true},
		{"container_registry_json_credentials", "qovery_container_registry", "config.json_credentials", true},
		{"helm_repository_secret_access_key", "qovery_helm_repository", "config.secret_access_key", true},
		{"helm_repository_scaleway_secret_key", "qovery_helm_repository", "config.scaleway_secret_key", true},
		{"helm_repository_password", "qovery_helm_repository", "config.password", true},

		// Identifiers: stay visible, matching qovery_aws_credentials.access_key_id.
		{"aws_credentials_access_key_id", "qovery_aws_credentials", "access_key_id", false},
		{"container_registry_access_key_id", "qovery_container_registry", "config.access_key_id", false},
		{"container_registry_scaleway_access_key", "qovery_container_registry", "config.scaleway_access_key", false},
		{"container_registry_scaleway_project_id", "qovery_container_registry", "config.scaleway_project_id", false},
		{"container_registry_username", "qovery_container_registry", "config.username", false},
		{"container_registry_region", "qovery_container_registry", "config.region", false},
		{"helm_repository_access_key_id", "qovery_helm_repository", "config.access_key_id", false},
		{"helm_repository_scaleway_access_key", "qovery_helm_repository", "config.scaleway_access_key", false},
		{"helm_repository_scaleway_project_id", "qovery_helm_repository", "config.scaleway_project_id", false},
		{"helm_repository_username", "qovery_helm_repository", "config.username", false},
		{"helm_repository_region", "qovery_helm_repository", "config.region", false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			flags, ok := schemas[tc.Schema]
			require.True(t, ok, "unknown schema %s", tc.Schema)

			got, ok := flags[tc.Attribute]
			require.True(t, ok, "%s has no attribute %s", tc.Schema, tc.Attribute)
			assert.Equal(t, tc.ExpectSensitive, got,
				"%s.%s: expected Sensitive=%t", tc.Schema, tc.Attribute, tc.ExpectSensitive)
		})
	}
}
