package variable_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// TestNewScopeFromString validate that the scopes qovery.APIVariableScopeEnum defined in Qovery's API Client are valid.
// This is useful to make sure the variable.Scope stays up to date.
func TestNewScopeFromString(t *testing.T) {
	t.Parallel()

	assert.Len(t, variable.AllowedScopeValues, len(qovery.AllowedAPIVariableScopeEnumEnumValues))
	for _, variableScope := range qovery.AllowedAPIVariableScopeEnumEnumValues {
		variableScopeStr := string(variableScope)
		t.Run(variableScopeStr, func(t *testing.T) {
			scope, err := variable.NewScopeFromString(variableScopeStr)
			assert.NoError(t, err)
			assert.Equal(t, scope.String(), variableScopeStr)
		})
	}
}
