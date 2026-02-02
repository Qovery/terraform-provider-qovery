//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	"github.com/stretchr/testify/assert"
)

// TestBuiltInEnvVar_HashStabilityWithoutState verifies that when converting
// environment variables without passing state, the description from API is used,
// which causes set hash differences when state had null descriptions.
//
// This test demonstrates why convertDomainVariablesToEnvironmentVariableListWithNullableInitialState
// should be used instead of convertDomainVariablesToEnvironmentVariableList for built-in variables.
func TestBuiltInEnvVar_HashStabilityWithoutState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State has env var with null description (user didn't specify it in config)
	stateEnvVar := EnvironmentVariable{
		Id:          basetypes.NewStringValue("019c0a91-f392-73de-bb54-952f5d3bd444"),
		Key:         basetypes.NewStringValue("QOVERY_JOB_Z96FAE300_TAG"),
		Value:       basetypes.NewStringValue("01292026-2"),
		Description: basetypes.NewStringNull(),
	}
	stateList := EnvironmentVariableList{stateEnvVar}
	stateSet := stateList.toTerraformSet(ctx)

	// API returns the same env var with a description
	apiResponse := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-f392-73de-bb54-952f5d3bd444"),
			Key:         "QOVERY_JOB_Z96FAE300_TAG",
			Value:       "01292026-2",
			Description: "Currently deployed image tag",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	// Without state: description comes from API
	actualList := convertDomainVariablesToEnvironmentVariableList(
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)
	actualSet := actualList.toTerraformSet(ctx)

	stateElements := stateSet.Elements()
	actualElements := actualSet.Elements()

	t.Logf("State set elements: %v", stateElements)
	t.Logf("Actual set elements: %v", actualElements)

	assert.Equal(t, len(stateElements), len(actualElements), "should have same number of elements")

	// Without state preservation, sets differ due to description field
	if !stateSet.Equal(actualSet) {
		t.Log("Sets differ: state has null description, actual has API description")
		t.Logf("State description: null=%v", stateEnvVar.Description.IsNull())
		t.Logf("Actual description: null=%v, value=%q",
			actualList[0].Description.IsNull(),
			actualList[0].Description.ValueString())
	}

	// Verify behavior without state preservation
	assert.True(t, stateEnvVar.Description.IsNull(), "state description should be null")
	assert.False(t, actualList[0].Description.IsNull(), "without state: description comes from API")
	assert.Equal(t, "Currently deployed image tag", actualList[0].Description.ValueString())

	// With state: null description is preserved
	resultWithState := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(
		ctx,
		stateSet,
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)
	resultWithStateSet := resultWithState.toTerraformSet(ctx)

	t.Logf("Result with state: %v", resultWithStateSet.Elements())

	// With state preservation, null description is kept
	assert.True(t, resultWithState[0].Description.IsNull(), "with state: description preserved as null")

	// Sets are now equal
	assert.True(t, stateSet.Equal(resultWithStateSet), "with state: sets should be equal")
}

// TestBuiltInEnvVar_ValueDriftDetection verifies that value changes from the API
// are properly detected as drift, while description handling depends on state preservation.
func TestBuiltInEnvVar_ValueDriftDetection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State with specific value
	stateEnvVar := EnvironmentVariable{
		Id:          basetypes.NewStringValue("019c0a91-f392-73de-bb54-952f5d3bd444"),
		Key:         basetypes.NewStringValue("DATABASE_URL"),
		Value:       basetypes.NewStringValue("postgres://host1:5432/db"),
		Description: basetypes.NewStringNull(),
	}
	stateList := EnvironmentVariableList{stateEnvVar}
	stateSet := stateList.toTerraformSet(ctx)

	// API returns different value (changed externally)
	apiResponse := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-f392-73de-bb54-952f5d3bd444"),
			Key:         "DATABASE_URL",
			Value:       "postgres://host2:5432/db_readonly",
			Description: "",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	actualList := convertDomainVariablesToEnvironmentVariableList(
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)
	actualSet := actualList.toTerraformSet(ctx)

	// Value drift is detected
	assert.False(t, stateSet.Equal(actualSet), "value drift should be detected")
	assert.Equal(t, "postgres://host2:5432/db_readonly", actualList[0].Value.ValueString())
}

// TestBuiltInEnvVar_EmptySet verifies behavior when API returns no built-in variables.
func TestBuiltInEnvVar_EmptySet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State has one variable
	stateEnvVar := EnvironmentVariable{
		Id:          basetypes.NewStringValue("019c0a91-f392-73de-bb54-952f5d3bd444"),
		Key:         basetypes.NewStringValue("QOVERY_JOB_TAG"),
		Value:       basetypes.NewStringValue("v1.0.0"),
		Description: basetypes.NewStringNull(),
	}
	stateList := EnvironmentVariableList{stateEnvVar}
	stateSet := stateList.toTerraformSet(ctx)

	// API returns empty list
	apiResponse := variable.Variables{}

	result := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(
		ctx,
		stateSet,
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)

	assert.Empty(t, result, "result should be empty when API returns no variables")
}

// TestBuiltInEnvVar_NewVarNotInState verifies that new variables from API
// that were not in state get null description due to zero-value map lookup behavior.
func TestBuiltInEnvVar_NewVarNotInState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State has one variable
	stateEnvVar := EnvironmentVariable{
		Id:          basetypes.NewStringValue("019c0a91-0000-0000-0000-000000000001"),
		Key:         basetypes.NewStringValue("QOVERY_EXISTING_VAR"),
		Value:       basetypes.NewStringValue("existing"),
		Description: basetypes.NewStringNull(),
	}
	stateList := EnvironmentVariableList{stateEnvVar}
	stateSet := stateList.toTerraformSet(ctx)

	// API returns the existing var plus a NEW var
	apiResponse := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-0000-0000-0000-000000000001"),
			Key:         "QOVERY_EXISTING_VAR",
			Value:       "existing",
			Description: "API description for existing",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
		{
			ID:          uuid.MustParse("019c0a91-0000-0000-0000-000000000002"),
			Key:         "QOVERY_NEW_VAR",
			Value:       "new_value",
			Description: "This is a new variable from API",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	result := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(
		ctx,
		stateSet,
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)

	assert.Len(t, result, 2, "should have 2 variables")

	// Find the variables by key
	var existingVar, newVar *EnvironmentVariable
	for i := range result {
		if result[i].Key.ValueString() == "QOVERY_EXISTING_VAR" {
			existingVar = &result[i]
		}
		if result[i].Key.ValueString() == "QOVERY_NEW_VAR" {
			newVar = &result[i]
		}
	}

	// Existing var should preserve null description from state
	assert.NotNil(t, existingVar)
	assert.True(t, existingVar.Description.IsNull(), "existing var should preserve null description from state")

	// New var gets null description (zero-value map lookup behavior)
	assert.NotNil(t, newVar)
	assert.True(t, newVar.Description.IsNull(), "new var gets null description")
}

// TestBuiltInEnvVar_MultipleVars verifies correct handling of multiple variables.
// All descriptions are preserved from state to prevent hash mismatch.
func TestBuiltInEnvVar_MultipleVars(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State has multiple variables with different description states
	stateList := EnvironmentVariableList{
		{
			Id:          basetypes.NewStringValue("019c0a91-0000-0000-0000-000000000001"),
			Key:         basetypes.NewStringValue("QOVERY_VAR_NULL_DESC"),
			Value:       basetypes.NewStringValue("value1"),
			Description: basetypes.NewStringNull(),
		},
		{
			Id:          basetypes.NewStringValue("019c0a91-0000-0000-0000-000000000002"),
			Key:         basetypes.NewStringValue("QOVERY_VAR_WITH_DESC"),
			Value:       basetypes.NewStringValue("value2"),
			Description: basetypes.NewStringValue("user-provided"),
		},
	}
	stateSet := stateList.toTerraformSet(ctx)

	// API returns all vars with descriptions
	apiResponse := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-0000-0000-0000-000000000001"),
			Key:         "QOVERY_VAR_NULL_DESC",
			Value:       "value1",
			Description: "API description 1",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
		{
			ID:          uuid.MustParse("019c0a91-0000-0000-0000-000000000002"),
			Key:         "QOVERY_VAR_WITH_DESC",
			Value:       "value2",
			Description: "API description 2",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	result := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(
		ctx,
		stateSet,
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)

	assert.Len(t, result, 2)

	// Find variables by key
	varMap := make(map[string]EnvironmentVariable)
	for _, v := range result {
		varMap[v.Key.ValueString()] = v
	}

	// Null description is preserved from state
	assert.True(t, varMap["QOVERY_VAR_NULL_DESC"].Description.IsNull(),
		"null description should be preserved from state")

	// Non-null description is also preserved from state (not updated from API)
	assert.Equal(t, "user-provided", varMap["QOVERY_VAR_WITH_DESC"].Description.ValueString(),
		"non-null description should be preserved from state")
}

// TestBuiltInEnvVar_ValueChangeDetected verifies that value changes from API
// are properly reflected in the result (drift detection works).
func TestBuiltInEnvVar_ValueChangeDetected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State with specific value
	stateList := EnvironmentVariableList{
		{
			Id:          basetypes.NewStringValue("019c0a91-0000-0000-0000-000000000001"),
			Key:         basetypes.NewStringValue("QOVERY_VAR"),
			Value:       basetypes.NewStringValue("old_value"),
			Description: basetypes.NewStringNull(),
		},
	}
	stateSet := stateList.toTerraformSet(ctx)

	// API returns updated value (someone changed it via UI)
	apiResponse := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-0000-0000-0000-000000000001"),
			Key:         "QOVERY_VAR",
			Value:       "new_value_from_api",
			Description: "API description",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	result := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(
		ctx,
		stateSet,
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)
	resultSet := result.toTerraformSet(ctx)

	assert.Len(t, result, 1)

	// Value comes from API (drift detection)
	assert.Equal(t, "new_value_from_api", result[0].Value.ValueString(),
		"value should come from API")

	// Description is null (preserved from state)
	assert.True(t, result[0].Description.IsNull(),
		"null description should be preserved")

	// Sets are NOT equal because value changed - this is correct (drift detected)
	assert.False(t, stateSet.Equal(resultSet), "sets should differ due to value change")
}

// TestBuiltInEnvVar_EmptyStateSet verifies behavior when state set is empty,
// such as during initial resource creation.
func TestBuiltInEnvVar_EmptyStateSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// No prior state (empty set)
	emptyList := EnvironmentVariableList{}
	emptySet := emptyList.toTerraformSet(ctx)

	apiResponse := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-0000-0000-0000-000000000001"),
			Key:         "QOVERY_VAR",
			Value:       "value",
			Description: "API description",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	result := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(
		ctx,
		emptySet,
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)

	assert.Len(t, result, 1)
	// Variables not in state get null description
	assert.True(t, result[0].Description.IsNull(),
		"description should be null when var not in state")
	assert.Equal(t, "value", result[0].Value.ValueString())
	assert.Equal(t, "QOVERY_VAR", result[0].Key.ValueString())
}
