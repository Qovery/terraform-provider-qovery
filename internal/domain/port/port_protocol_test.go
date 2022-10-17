package port_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
)

// TestNewProtocolFromString validate that the kinds qovery.PortProtocolEnum defined in Qovery's API Client are valid.
// This is useful to make sure the port.Protocol stays up to date.
func TestNewProtocolFromString(t *testing.T) {
	t.Parallel()

	assert.Len(t, port.AllowedProtocolValues, len(qovery.AllowedPortProtocolEnumEnumValues))
	for _, portType := range qovery.AllowedPortProtocolEnumEnumValues {
		portTypeStr := string(portType)
		t.Run(portTypeStr, func(t *testing.T) {
			st, err := port.NewProtocolFromString(portTypeStr)
			assert.NoError(t, err)
			assert.Equal(t, st.String(), portTypeStr)
		})
	}
}
