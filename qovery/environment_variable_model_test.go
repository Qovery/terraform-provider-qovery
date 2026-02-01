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

// TestBugBuiltInEnvVarHashMismatch reproduces the bug where built_in_environment_variables
// causes "Provider produced inconsistent result after apply" error.
//
// The bug occurs because:
//  1. User creates a job without specifying description for env vars (description=null in plan)
//  2. During plan, Terraform creates set elements with description=null
//  3. After apply, the provider calls convertDomainVariablesToEnvironmentVariableList
//     which passes nil for state parameter
//  4. This causes description to be populated from API response (e.g., "Currently deployed image tag")
//  5. The set element hash changes because null != "Currently deployed image tag"
//  6. Terraform cannot correlate planned element with actual element -> crash
func TestBugBuiltInEnvVarHashMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Simulate what Terraform plans: env var with null description
	// (user didn't specify description in their .tf file)
	plannedEnvVar := EnvironmentVariable{
		Id:          basetypes.NewStringValue("019c0a91-f392-73de-bb54-952f5d3bd444"),
		Key:         basetypes.NewStringValue("QOVERY_JOB_Z96FAE300_TAG"),
		Value:       basetypes.NewStringValue("01292026-2"),
		Description: basetypes.NewStringNull(), // User didn't specify description
	}
	plannedList := EnvironmentVariableList{plannedEnvVar}
	plannedSet := plannedList.toTerraformSet(ctx)

	// Simulate API response: same env var but with description from server
	apiResponse := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-f392-73de-bb54-952f5d3bd444"),
			Key:         "QOVERY_JOB_Z96FAE300_TAG",
			Value:       "01292026-2",
			Description: "Currently deployed image tag", // API returns a description
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	// BUG: convertDomainVariablesToEnvironmentVariableList passes nil for state
	// This is what happens in resource_job_model.go:380
	actualList := convertDomainVariablesToEnvironmentVariableList(
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)
	actualSet := actualList.toTerraformSet(ctx)

	// Get set elements to compare
	plannedElements := plannedSet.Elements()
	actualElements := actualSet.Elements()

	t.Logf("Planned set elements: %v", plannedElements)
	t.Logf("Actual set elements: %v", actualElements)

	// The bug: these should be equal but they're not because description differs
	// Planned has description=null, actual has description="Currently deployed image tag"
	assert.Equal(t, len(plannedElements), len(actualElements), "should have same number of elements")

	// This assertion will FAIL - demonstrating the bug
	// The set elements have different hashes due to description field mismatch
	if !plannedSet.Equal(actualSet) {
		t.Log("BUG REPRODUCED: Planned set != Actual set")
		t.Log("This causes 'Provider produced inconsistent result after apply' error")
		t.Log("")
		t.Logf("Planned element description: null=%v", plannedEnvVar.Description.IsNull())
		t.Logf("Actual element description: null=%v, value=%q",
			actualList[0].Description.IsNull(),
			actualList[0].Description.ValueString())
	}

	// Verify the root cause: description field mismatch
	assert.True(t, plannedEnvVar.Description.IsNull(), "planned description should be null")
	assert.False(t, actualList[0].Description.IsNull(), "BUG: actual description is NOT null when it should preserve null from state")
	assert.Equal(t, "Currently deployed image tag", actualList[0].Description.ValueString())

	// Now show the FIX: using convertDomainVariablesToEnvironmentVariableListWithNullableInitialState
	// This is what's used for regular environment_variables (line 381) but NOT for built_in (line 380)
	fixedList := convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(
		ctx,
		plannedSet, // Pass the planned state
		apiResponse,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)
	fixedSet := fixedList.toTerraformSet(ctx)

	t.Logf("Fixed set elements: %v", fixedSet.Elements())

	// With the fix, the description is preserved as null from state
	assert.True(t, fixedList[0].Description.IsNull(), "FIX: description should be null (preserved from state)")

	// And the sets should be equal
	assert.True(t, plannedSet.Equal(fixedSet), "FIX: planned and fixed sets should be equal")
}

// TestBugStateDriftNotDetected demonstrates the second bug where state drift
// for built-in environment variables is not detected because the provider
// doesn't compare against previous state.
func TestBugStateDriftNotDetected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Initial state: env var with specific value
	initialEnvVar := EnvironmentVariable{
		Id:          basetypes.NewStringValue("019c0a91-f392-73de-bb54-952f5d3bd444"),
		Key:         basetypes.NewStringValue("DATABASE_URL_CANADA"),
		Value:       basetypes.NewStringValue("postgres://host1:5432/db"),
		Description: basetypes.NewStringNull(),
	}
	initialList := EnvironmentVariableList{initialEnvVar}
	initialSet := initialList.toTerraformSet(ctx)

	// Server-side change: someone flipped the value via UI
	// (This is what the user reported - values were flipped)
	apiResponseWithFlippedValue := variable.Variables{
		{
			ID:          uuid.MustParse("019c0a91-f392-73de-bb54-952f5d3bd444"),
			Key:         "DATABASE_URL_CANADA",
			Value:       "postgres://host2:5432/db_readonly", // FLIPPED! Wrong value
			Description: "",
			Scope:       variable.ScopeBuiltIn,
			Type:        "BUILT_IN",
		},
	}

	// When provider reads state, it should detect the drift
	actualList := convertDomainVariablesToEnvironmentVariableList(
		apiResponseWithFlippedValue,
		variable.ScopeBuiltIn,
		"BUILT_IN",
	)
	actualSet := actualList.toTerraformSet(ctx)

	// The VALUE drift IS detected (different values)
	assert.False(t, initialSet.Equal(actualSet), "Value drift should be detected")

	// But for built-in vars, the provider completely regenerates from API
	// without comparing to prior state, which can cause issues with
	// the description field as shown in TestBugBuiltInEnvVarHashMismatch
	t.Log("Value drift is detected, but the lack of state preservation")
	t.Log("for built_in_environment_variables causes the hash mismatch bug")
	t.Log("when description fields differ between plan and apply phases.")
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
// that were not in state also get null description (consistent behavior).
// This is because the map lookup returns a zero-value EnvironmentVariable
// whose Description.IsNull() is true.
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

	// New var also gets null description because map lookup returns zero-value
	// whose Description.IsNull() is true
	assert.NotNil(t, newVar)
	assert.True(t, newVar.Description.IsNull(), "new var gets null description (zero-value behavior)")
}

// TestBuiltInEnvVar_MultipleVars verifies correct handling of multiple variables.
// The function ONLY preserves null descriptions - non-null descriptions are
// overwritten by API values. This is by design to fix the hash mismatch bug.
func TestBuiltInEnvVar_MultipleVars(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State has multiple variables - only null descriptions will be preserved
	stateList := EnvironmentVariableList{
		{
			Id:          basetypes.NewStringValue("019c0a91-0000-0000-0000-000000000001"),
			Key:         basetypes.NewStringValue("QOVERY_VAR_NULL_DESC"),
			Value:       basetypes.NewStringValue("value1"),
			Description: basetypes.NewStringNull(), // null - WILL be preserved
		},
		{
			Id:          basetypes.NewStringValue("019c0a91-0000-0000-0000-000000000002"),
			Key:         basetypes.NewStringValue("QOVERY_VAR_WITH_DESC"),
			Value:       basetypes.NewStringValue("value2"),
			Description: basetypes.NewStringValue("user-provided"), // not null - will be overwritten
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

	// Null description IS preserved (this is the fix)
	assert.True(t, varMap["QOVERY_VAR_NULL_DESC"].Description.IsNull(),
		"null description should be preserved from state")

	// Non-null description is overwritten by API (this is expected behavior)
	assert.Equal(t, "API description 2", varMap["QOVERY_VAR_WITH_DESC"].Description.ValueString(),
		"non-null description is overwritten by API value")
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

// TestBuiltInEnvVar_EmptyStateSet verifies behavior when state set is empty
// (e.g., during initial create before any state exists).
// Variables not in state get null description due to zero-value map lookup behavior.
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
	// With no prior state, map lookup returns zero-value whose Description.IsNull() is true
	// So description becomes null (not from API)
	assert.True(t, result[0].Description.IsNull(),
		"description should be null when var not found in state")
	assert.Equal(t, "value", result[0].Value.ValueString())
	assert.Equal(t, "QOVERY_VAR", result[0].Key.ValueString())
}
