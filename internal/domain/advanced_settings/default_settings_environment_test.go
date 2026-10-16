//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
)

func TestComputeDefaultServiceAdvancedSettingsUrl(t *testing.T) {
	t.Parallel()

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: "https://api.qovery.com"}}
	service := NewServiceAdvancedSettingsService(cfg)

	withEnvironment, err := service.computeDefaultServiceAdvancedSettingsUrl(domain.APPLICATION, "11111111-2222-3333-4444-555555555555")
	require.NoError(t, err)
	assert.Equal(t, "https://api.qovery.com/defaultApplicationAdvancedSettings?environmentId=11111111-2222-3333-4444-555555555555", *withEnvironment)

	withoutEnvironment, err := service.computeDefaultServiceAdvancedSettingsUrl(domain.APPLICATION, "")
	require.NoError(t, err)
	assert.Equal(t, "https://api.qovery.com/defaultApplicationAdvancedSettings", *withoutEnvironment)
}
