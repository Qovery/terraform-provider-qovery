package status_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// TestNewStateFromString validate that the kinds qovery.StateEnumEnum defined in Qovery's API Client are valid.
// This is useful to make sure the status.State stays up to date.
func TestNewStateFromString(t *testing.T) {
	t.Parallel()

	assert.Len(t, status.AllowedStateValues, len(qovery.AllowedStateEnumEnumValues))
	for _, state := range qovery.AllowedStateEnumEnumValues {
		stateStr := string(state)
		t.Run(stateStr, func(t *testing.T) {
			s, err := status.NewStateFromString(stateStr)
			assert.NoError(t, err)
			assert.Equal(t, s.String(), stateStr)
		})
	}
}
