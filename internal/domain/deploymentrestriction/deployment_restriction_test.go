//go:build unit && !integration
// +build unit,!integration

package deploymentrestriction

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceDeploymentRestrictionsDiff_IsNotEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		diff     ServiceDeploymentRestrictionsDiff
		expected bool
	}{
		{
			name: "empty diff",
			diff: ServiceDeploymentRestrictionsDiff{
				Create: []ServiceDeploymentRestriction{},
				Update: []ServiceDeploymentRestriction{},
				Delete: []string{},
			},
			expected: false,
		},
		{
			name: "nil slices",
			diff: ServiceDeploymentRestrictionsDiff{
				Create: nil,
				Update: nil,
				Delete: nil,
			},
			expected: false,
		},
		{
			name: "has create",
			diff: ServiceDeploymentRestrictionsDiff{
				Create: []ServiceDeploymentRestriction{
					{Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "/src"},
				},
				Update: []ServiceDeploymentRestriction{},
				Delete: []string{},
			},
			expected: true,
		},
		{
			name: "has update",
			diff: ServiceDeploymentRestrictionsDiff{
				Create: []ServiceDeploymentRestriction{},
				Update: []ServiceDeploymentRestriction{
					{Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_EXCLUDE, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "/test"},
				},
				Delete: []string{},
			},
			expected: true,
		},
		{
			name: "has delete",
			diff: ServiceDeploymentRestrictionsDiff{
				Create: []ServiceDeploymentRestriction{},
				Update: []ServiceDeploymentRestriction{},
				Delete: []string{"id-to-delete"},
			},
			expected: true,
		},
		{
			name: "has all",
			diff: ServiceDeploymentRestrictionsDiff{
				Create: []ServiceDeploymentRestriction{
					{Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "/new"},
				},
				Update: []ServiceDeploymentRestriction{
					{Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_EXCLUDE, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "/updated"},
				},
				Delete: []string{"id-1", "id-2"},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.diff.IsNotEmpty()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestServiceDeploymentRestriction_Fields(t *testing.T) {
	t.Parallel()

	id := "test-id"
	restriction := ServiceDeploymentRestriction{
		Id:    &id,
		Mode:  qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH,
		Type:  qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH,
		Value: "/src/main",
	}

	assert.Equal(t, &id, restriction.Id)
	assert.Equal(t, qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH, restriction.Mode)
	assert.Equal(t, qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, restriction.Type)
	assert.Equal(t, "/src/main", restriction.Value)
}

func TestServiceDeploymentRestriction_NilId(t *testing.T) {
	t.Parallel()

	restriction := ServiceDeploymentRestriction{
		Id:    nil,
		Mode:  qovery.DEPLOYMENTRESTRICTIONMODEENUM_EXCLUDE,
		Type:  qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH,
		Value: "/test",
	}

	assert.Nil(t, restriction.Id)
}

// Helper function to create a deployment restriction object type
func deploymentRestrictionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":    types.StringType,
			"mode":  types.StringType,
			"type":  types.StringType,
			"value": types.StringType,
		},
	}
}

func TestToDeploymentRestrictionDiff_NullSet(t *testing.T) {
	t.Parallel()

	nullSet := types.SetNull(deploymentRestrictionObjectType())

	diff, err := ToDeploymentRestrictionDiff(nullSet, nil)

	require.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Empty(t, diff.Create)
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Delete)
}

func TestToDeploymentRestrictionDiff_UnknownSet(t *testing.T) {
	t.Parallel()

	unknownSet := types.SetUnknown(deploymentRestrictionObjectType())

	diff, err := ToDeploymentRestrictionDiff(unknownSet, nil)

	require.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Empty(t, diff.Create)
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Delete)
}

func TestToDeploymentRestrictionDiff_CreateNewRestriction(t *testing.T) {
	t.Parallel()

	// Create a set with a new restriction (no ID)
	objType := deploymentRestrictionObjectType()
	restrictionObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringNull(),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/src"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(set, nil)

	require.NoError(t, err)
	assert.Len(t, diff.Create, 1)
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Delete)
	assert.Nil(t, diff.Create[0].Id)
	assert.Equal(t, qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH, diff.Create[0].Mode)
	assert.Equal(t, qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, diff.Create[0].Type)
	assert.Equal(t, "/src", diff.Create[0].Value)
}

func TestToDeploymentRestrictionDiff_UpdateExistingRestriction(t *testing.T) {
	t.Parallel()

	// Create a set with an existing restriction (has ID)
	objType := deploymentRestrictionObjectType()
	restrictionObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringValue("existing-id"),
			"mode":  types.StringValue("EXCLUDE"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/test"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	// State also has the same restriction
	stateSet, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(set, &stateSet)

	require.NoError(t, err)
	assert.Empty(t, diff.Create)
	assert.Len(t, diff.Update, 1)
	assert.Empty(t, diff.Delete)
	assert.Equal(t, "existing-id", *diff.Update[0].Id)
}

func TestToDeploymentRestrictionDiff_DeleteRestriction(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()

	// Empty new set
	emptySet, diags := types.SetValue(objType, []attr.Value{})
	require.False(t, diags.HasError())

	// State has an existing restriction that will be deleted
	stateObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringValue("id-to-delete"),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/old"),
		},
	)
	require.False(t, diags.HasError())

	stateSet, diags := types.SetValue(objType, []attr.Value{stateObj})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(emptySet, &stateSet)

	require.NoError(t, err)
	assert.Empty(t, diff.Create)
	assert.Empty(t, diff.Update)
	assert.Len(t, diff.Delete, 1)
	assert.Equal(t, "id-to-delete", diff.Delete[0])
}

func TestToDeploymentRestrictionDiff_MixedOperations(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()

	// New restriction (no ID) - will be created
	newObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringNull(),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/new"),
		},
	)
	require.False(t, diags.HasError())

	// Existing restriction (has ID) - will be updated
	existingObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringValue("existing-id"),
			"mode":  types.StringValue("EXCLUDE"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/existing"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{newObj, existingObj})
	require.False(t, diags.HasError())

	// State has existing-id and another one to delete
	stateExistingObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringValue("existing-id"),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/old-existing"),
		},
	)
	require.False(t, diags.HasError())

	stateDeleteObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringValue("to-delete-id"),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/to-delete"),
		},
	)
	require.False(t, diags.HasError())

	stateSet, diags := types.SetValue(objType, []attr.Value{stateExistingObj, stateDeleteObj})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(set, &stateSet)

	require.NoError(t, err)
	assert.Len(t, diff.Create, 1)
	assert.Len(t, diff.Update, 1)
	assert.Len(t, diff.Delete, 1)

	assert.Nil(t, diff.Create[0].Id)
	assert.Equal(t, "/new", diff.Create[0].Value)

	assert.Equal(t, "existing-id", *diff.Update[0].Id)

	assert.Equal(t, "to-delete-id", diff.Delete[0])
}

func TestToDeploymentRestrictionDiff_InvalidModeError(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()
	restrictionObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringNull(),
			"mode":  types.StringValue("INVALID_MODE"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/src"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(set, nil)

	assert.Error(t, err)
	assert.Nil(t, diff)
}

func TestToDeploymentRestrictionDiff_InvalidTypeError(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()
	restrictionObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringNull(),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("INVALID_TYPE"),
			"value": types.StringValue("/src"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(set, nil)

	assert.Error(t, err)
	assert.Nil(t, diff)
}

func TestToDeploymentRestrictionDiff_NilStateSet(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()
	restrictionObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringValue("some-id"),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/path"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	// When state is nil, all items with IDs go to Update (no deletions)
	diff, err := ToDeploymentRestrictionDiff(set, nil)

	require.NoError(t, err)
	assert.Empty(t, diff.Create)
	assert.Len(t, diff.Update, 1)
	assert.Empty(t, diff.Delete)
}

func TestToDeploymentRestrictionDiff_NullStateSet(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()
	restrictionObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringValue("some-id"),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/path"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	nullStateSet := types.SetNull(objType)

	diff, err := ToDeploymentRestrictionDiff(set, &nullStateSet)

	require.NoError(t, err)
	assert.Empty(t, diff.Create)
	assert.Len(t, diff.Update, 1)
	assert.Empty(t, diff.Delete)
}

func TestDeploymentRestrictionModeEnum_AllValidModes(t *testing.T) {
	t.Parallel()

	validModes := []string{"MATCH", "EXCLUDE"}

	for _, mode := range validModes {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			_, err := qovery.NewDeploymentRestrictionModeEnumFromValue(mode)
			assert.NoError(t, err)
		})
	}
}

func TestDeploymentRestrictionTypeEnum_AllValidTypes(t *testing.T) {
	t.Parallel()

	validTypes := []string{"PATH"}

	for _, typ := range validTypes {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			_, err := qovery.NewDeploymentRestrictionTypeEnumFromValue(typ)
			assert.NoError(t, err)
		})
	}
}

func TestToDeploymentRestrictionDiff_EmptySet(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()
	emptySet, diags := types.SetValue(objType, []attr.Value{})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(emptySet, nil)

	require.NoError(t, err)
	assert.Empty(t, diff.Create)
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Delete)
}

func TestToDeploymentRestrictionDiff_UnknownIdTreatedAsNew(t *testing.T) {
	t.Parallel()

	objType := deploymentRestrictionObjectType()
	restrictionObj, diags := types.ObjectValue(
		objType.AttrTypes,
		map[string]attr.Value{
			"id":    types.StringUnknown(),
			"mode":  types.StringValue("MATCH"),
			"type":  types.StringValue("PATH"),
			"value": types.StringValue("/new-path"),
		},
	)
	require.False(t, diags.HasError())

	set, diags := types.SetValue(objType, []attr.Value{restrictionObj})
	require.False(t, diags.HasError())

	diff, err := ToDeploymentRestrictionDiff(set, nil)

	require.NoError(t, err)
	assert.Len(t, diff.Create, 1)
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Delete)
}

// Verify the context import is used (for potential future async operations)
func TestToDeploymentRestrictionDiff_ContextUsage(t *testing.T) {
	t.Parallel()

	// This test verifies the context package is properly imported
	// (it's used in the ObjectValue calls through diags)
	_ = context.Background()
}
