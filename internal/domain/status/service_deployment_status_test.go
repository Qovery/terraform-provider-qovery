package status_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// TestNewServiceDeploymentStatusFromString validate that the kinds qovery.ServiceDeploymentStatusEnumEnum defined in Qovery's API Client are valid.
// This is useful to make sure the status.ServiceDeploymentStatus stays up to date.
func TestNewServiceDeploymentStatusFromString(t *testing.T) {
	t.Parallel()

	assert.Len(t, status.AllowedServiceDeploymentStatusValues, len(qovery.AllowedServiceDeploymentStatusEnumEnumValues))
	for _, sds := range qovery.AllowedServiceDeploymentStatusEnumEnumValues {
		sdsStr := string(sds)
		t.Run(sdsStr, func(t *testing.T) {
			s, err := status.NewServiceDeploymentStatusFromString(sdsStr)
			assert.NoError(t, err)
			assert.Equal(t, s.String(), sdsStr)
		})
	}
}
