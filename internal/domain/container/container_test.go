//go:build unit && !integration

package container_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
)

func newValidContainer() container.Container {
	return container.Container{
		ID:                  uuid.New(),
		EnvironmentID:       uuid.New(),
		RegistryID:          uuid.New(),
		Name:                "my-container",
		IconUri:             "app://icon",
		ImageName:           "nginx",
		Tag:                 "latest",
		CPU:                 500,
		Memory:              512,
		MinRunningInstances: 1,
		MaxRunningInstances: 1,
	}
}

func TestContainer_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, newValidContainer().Validate())
	})

	// Scale-to-zero: min_running_instances = 0 is allowed (KEDA). The `required`
	// validator tag rejected the int32 zero value, so this guards the gte=0 fix.
	t.Run("min_running_instances zero is valid", func(t *testing.T) {
		t.Parallel()
		c := newValidContainer()
		c.MinRunningInstances = 0
		assert.NoError(t, c.Validate())
	})

	t.Run("invalid autoscaling fails", func(t *testing.T) {
		t.Parallel()
		c := newValidContainer()
		c.Autoscaling = &autoscaling.AutoscalingPolicy{} // no scalers
		assert.Error(t, c.Validate())
	})
}
