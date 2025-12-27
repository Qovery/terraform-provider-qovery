//go:build unit && !integration
// +build unit,!integration

package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInvalidResourceIDParam",
			err:      ErrInvalidResourceIDParam,
			expected: "invalid resource id param",
		},
		{
			name:     "ErrUnexpectedState",
			err:      ErrUnexpectedState,
			expected: "unexpected state",
		},
		{
			name:     "ErrFailedToGetStatus",
			err:      ErrFailedToGetStatus,
			expected: "failed to get status",
		},
		{
			name:     "ErrFailedToUpdateState",
			err:      ErrFailedToUpdateState,
			expected: "failed to update state",
		},
		{
			name:     "ErrFailedToDeploy",
			err:      ErrFailedToDeploy,
			expected: "failed to deploy",
		},
		{
			name:     "ErrFailedToRedeploy",
			err:      ErrFailedToRedeploy,
			expected: "failed to redeploy",
		},
		{
			name:     "ErrFailedToStop",
			err:      ErrFailedToStop,
			expected: "failed to stop",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.NotNil(t, tc.err)
			assert.Equal(t, tc.expected, tc.err.Error())
		})
	}
}

func TestErrorVariables_AreDistinct(t *testing.T) {
	t.Parallel()

	errors := []error{
		ErrInvalidResourceIDParam,
		ErrUnexpectedState,
		ErrFailedToGetStatus,
		ErrFailedToUpdateState,
		ErrFailedToDeploy,
		ErrFailedToRedeploy,
		ErrFailedToStop,
	}

	// Ensure all errors are distinct
	seen := make(map[string]bool)
	for _, err := range errors {
		msg := err.Error()
		assert.False(t, seen[msg], "duplicate error message: %s", msg)
		seen[msg] = true
	}
}
