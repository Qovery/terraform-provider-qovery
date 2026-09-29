//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The data source shares the TerraformService model with the resource, so a model field without
// a data source attribute, at any nesting level, fails every read at runtime, never at build time.
func TestTerraformServiceDataSourceSchemaMatchesTheModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	schemaResp := &datasource.SchemaResponse{}
	terraformServiceDataSource{}.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	assertModelMatchesObjectType(t, "", reflect.TypeFor[TerraformService](), schemaResp.Schema.Type())
}

// assertModelMatchesObjectType compares the tfsdk fields of a model struct with the attributes of
// an object type, recursing into nested structs, the way the framework matches them on Get and Set.
func assertModelMatchesObjectType(t *testing.T, path string, model reflect.Type, objectType attr.Type) {
	t.Helper()

	object, ok := objectType.(basetypes.ObjectType)
	if !ok {
		t.Errorf("%s: model is a struct but the schema type is %s", path, objectType)
		return
	}

	fields := make(map[string]reflect.Type, model.NumField())
	for i := range model.NumField() {
		if tag := model.Field(i).Tag.Get("tfsdk"); tag != "" && tag != "-" {
			fields[tag] = model.Field(i).Type
		}
	}

	names := make([]string, 0, len(fields)+len(object.AttrTypes))
	for name := range fields {
		names = append(names, name)
	}
	for name := range object.AttrTypes {
		if _, ok := fields[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		fieldType, inModel := fields[name]
		attributeType, inSchema := object.AttrTypes[name]
		if !assert.Truef(t, inModel && inSchema, "%s%s: in model %t, in schema %t", path, name, inModel, inSchema) {
			continue
		}
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct && !fieldType.Implements(reflect.TypeFor[attr.Value]()) {
			assertModelMatchesObjectType(t, path+name+".", fieldType, attributeType)
		}
	}
}
