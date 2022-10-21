package environment_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
)

// TestNewModeFromString validate that the modes qovery.EnvironmentModeEnum defined in Qovery's API Client are valid.
// This is useful to make sure the environment.Mode stays up to date.
func TestNewModeFromString(t *testing.T) {
	t.Parallel()

	assert.Len(t, environment.AllowedModeValues, len(qovery.AllowedEnvironmentModeEnumEnumValues))
	for _, qoveryMode := range qovery.AllowedEnvironmentModeEnumEnumValues {
		qoveryModeStr := string(qoveryMode)
		t.Run(qoveryModeStr, func(t *testing.T) {
			mode, err := environment.NewModeFromString(qoveryModeStr)
			assert.NoError(t, err)
			assert.Equal(t, mode.String(), qoveryModeStr)
		})
	}
}
