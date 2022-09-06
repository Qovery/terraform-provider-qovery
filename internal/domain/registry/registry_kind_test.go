package registry_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// TestNewKindFromString validate that the kinds qovery.ContainerRegistryKindEnum defined in Qovery's API Client are valid.
// This is useful to make sure the registry.Kind stays up to date.
func TestNewKindFromString(t *testing.T) {
	t.Parallel()

	assert.Len(t, registry.AllowedKindValues, len(qovery.AllowedContainerRegistryKindEnumEnumValues))
	for _, registryKind := range qovery.AllowedContainerRegistryKindEnumEnumValues {
		registryKindStr := string(registryKind)
		t.Run(registryKindStr, func(t *testing.T) {
			kind, err := registry.NewKindFromString(registryKindStr)
			assert.NoError(t, err)
			assert.Equal(t, kind.String(), registryKindStr)
		})
	}
}
