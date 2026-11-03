//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalSecretsSchemas_JSONTargetKey(t *testing.T) {
	t.Parallel()
	resources := map[string]resource.Resource{
		"application": &applicationResource{}, "container": &containerResource{},
		"job": &jobResource{}, "helm": &helmResource{},
		"terraform_service": &terraformServiceResource{}, "environment": &environmentResource{},
	}
	for name, r := range resources {
		t.Run("resource_"+name, func(t *testing.T) {
			t.Parallel()
			var resp resource.SchemaResponse
			r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
			require.False(t, resp.Diagnostics.HasError())
			secrets := resp.Schema.Attributes["external_secrets"].(schema.SetNestedAttribute)
			require.Contains(t, secrets.NestedObject.Attributes, "json_target_key")
			selector := secrets.NestedObject.Attributes["json_target_key"]
			assert.Equal(t, types.StringType, selector.GetType())
			assert.True(t, selector.IsOptional())
			assert.False(t, selector.IsComputed())
			files := resp.Schema.Attributes["external_secret_files"].(schema.SetNestedAttribute)
			assert.NotContains(t, files.NestedObject.Attributes, "json_target_key")
		})
	}
	dataSources := map[string]datasource.DataSource{
		"application": &applicationDataSource{}, "container": &containerDataSource{},
		"job": &jobDataSource{}, "helm": &helmDataSource{},
		"terraform_service": &terraformServiceDataSource{}, "environment": &environmentDataSource{},
	}
	for name, ds := range dataSources {
		t.Run("data_source_"+name, func(t *testing.T) {
			t.Parallel()
			var resp datasource.SchemaResponse
			ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
			require.False(t, resp.Diagnostics.HasError())
			secrets := resp.Schema.Attributes["external_secrets"].(dsschema.SetNestedAttribute)
			require.Contains(t, secrets.NestedObject.Attributes, "json_target_key")
			selector := secrets.NestedObject.Attributes["json_target_key"]
			assert.Equal(t, types.StringType, selector.GetType())
			assert.True(t, selector.IsComputed())
		})
	}
}
