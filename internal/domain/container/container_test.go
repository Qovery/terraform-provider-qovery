//go:build unit && !integration
// +build unit,!integration

package container_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestNewContainer(t *testing.T) {
	t.Parallel()

	validContainerID := uuid.NewString()
	validEnvironmentID := uuid.NewString()
	validRegistryID := uuid.NewString()
	stateDeployed := status.StateDeployed.String()
	stateStopped := status.StateStopped.String()

	tests := []struct {
		name          string
		params        container.NewContainerParams
		expectError   bool
		expectedError error
	}{
		{
			name: "success with minimal valid params",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError: false,
		},
		{
			name: "success with deployed state",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				State:               &stateDeployed,
			},
			expectError: false,
		},
		{
			name: "success with stopped state",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				State:               &stateStopped,
			},
			expectError: false,
		},
		{
			name: "success with entrypoint and arguments",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				Entrypoint:          stringPtr("/bin/sh"),
				Arguments:           []string{"-c", "echo hello"},
			},
			expectError: false,
		},
		{
			name: "success with auto preview and auto deploy",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				AutoPreview:         true,
				AutoDeploy:          boolPtr(true),
			},
			expectError: false,
		},
		{
			name: "success with deployment stage id",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				DeploymentStageID:   uuid.NewString(),
			},
			expectError: false,
		},
		{
			name: "success with annotations and labels group ids",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				AnnotationsGroupIds: []string{uuid.NewString()},
				LabelsGroupIds:      []string{uuid.NewString()},
			},
			expectError: false,
		},
		{
			name: "fail with invalid container id",
			params: container.NewContainerParams{
				ContainerID:         "invalid-uuid",
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidContainerIDParam,
		},
		{
			name: "fail with empty container id",
			params: container.NewContainerParams{
				ContainerID:         "",
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidContainerIDParam,
		},
		{
			name: "fail with invalid environment id",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       "invalid-uuid",
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidEnvironmentIDParam,
		},
		{
			name: "fail with empty environment id",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       "",
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidEnvironmentIDParam,
		},
		{
			name: "fail with invalid registry id",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          "invalid-uuid",
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidRegistryIDParam,
		},
		{
			name: "fail with empty registry id",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          "",
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidRegistryIDParam,
		},
		{
			name: "fail with empty name",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidNameParam,
		},
		{
			name: "fail with empty image name",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidImageNameParam,
		},
		{
			name: "fail with empty tag",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
			},
			expectError:   true,
			expectedError: container.ErrInvalidTagParam,
		},
		{
			name: "fail with invalid state",
			params: container.NewContainerParams{
				ContainerID:         validContainerID,
				EnvironmentID:       validEnvironmentID,
				RegistryID:          validRegistryID,
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				State:               stringPtr("INVALID_STATE"),
			},
			expectError:   true,
			expectedError: container.ErrInvalidStateParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := container.NewContainer(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedError != nil {
					assert.ErrorContains(t, err, tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.params.Name, result.Name)
				assert.Equal(t, tt.params.ImageName, result.ImageName)
				assert.Equal(t, tt.params.Tag, result.Tag)
				assert.True(t, result.IsValid())
			}
		})
	}
}

func TestContainer_SetState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		inputState    status.State
		expectedState status.State
		expectError   bool
	}{
		{
			name:          "set deployed state",
			inputState:    status.StateDeployed,
			expectedState: status.StateDeployed,
			expectError:   false,
		},
		{
			name:          "set stopped state",
			inputState:    status.StateStopped,
			expectedState: status.StateStopped,
			expectError:   false,
		},
		{
			name:          "ready state converts to stopped",
			inputState:    status.StateReady,
			expectedState: status.StateStopped,
			expectError:   false,
		},
		{
			name:        "invalid state",
			inputState:  status.State("INVALID"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &container.Container{}
			err := c.SetState(tt.inputState)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedState, c.State)
			}
		})
	}
}

func TestContainer_SetEnvironmentVariables(t *testing.T) {
	t.Parallel()

	validVarID := uuid.New()
	builtInVarID := uuid.New()

	tests := []struct {
		name                        string
		inputVars                   variable.Variables
		expectedEnvVarsCount        int
		expectedBuiltInVarsCount    int
		expectError                 bool
	}{
		{
			name:                        "empty variables",
			inputVars:                   variable.Variables{},
			expectedEnvVarsCount:        0,
			expectedBuiltInVarsCount:    0,
			expectError:                 false,
		},
		{
			name: "regular environment variables only",
			inputVars: variable.Variables{
				{ID: validVarID, Key: "MY_VAR", Value: "my_value", Scope: variable.ScopeContainer},
			},
			expectedEnvVarsCount:     1,
			expectedBuiltInVarsCount: 0,
			expectError:              false,
		},
		{
			name: "built-in variables only",
			inputVars: variable.Variables{
				{ID: builtInVarID, Key: "QOVERY_VAR", Value: "qovery_value", Scope: variable.ScopeBuiltIn},
			},
			expectedEnvVarsCount:     0,
			expectedBuiltInVarsCount: 1,
			expectError:              false,
		},
		{
			name: "mixed variables",
			inputVars: variable.Variables{
				{ID: validVarID, Key: "MY_VAR", Value: "my_value", Scope: variable.ScopeContainer},
				{ID: builtInVarID, Key: "QOVERY_VAR", Value: "qovery_value", Scope: variable.ScopeBuiltIn},
			},
			expectedEnvVarsCount:     1,
			expectedBuiltInVarsCount: 1,
			expectError:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &container.Container{
				ID: uuid.New(),
			}
			err := c.SetEnvironmentVariables(tt.inputVars)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, c.EnvironmentVariables, tt.expectedEnvVarsCount)
				assert.Len(t, c.BuiltInEnvironmentVariables, tt.expectedBuiltInVarsCount)
			}
		})
	}
}

func TestContainer_SetSecrets(t *testing.T) {
	t.Parallel()

	validSecretID := uuid.New()
	builtInSecretID := uuid.New()

	tests := []struct {
		name                 string
		inputSecrets         secret.Secrets
		expectedSecretsCount int
		expectError          bool
	}{
		{
			name:                 "empty secrets",
			inputSecrets:         secret.Secrets{},
			expectedSecretsCount: 0,
			expectError:          false,
		},
		{
			name: "regular secrets only",
			inputSecrets: secret.Secrets{
				{ID: validSecretID, Key: "MY_SECRET", Scope: variable.ScopeContainer},
			},
			expectedSecretsCount: 1,
			expectError:          false,
		},
		{
			name: "built-in secrets are filtered out",
			inputSecrets: secret.Secrets{
				{ID: builtInSecretID, Key: "QOVERY_SECRET", Scope: variable.ScopeBuiltIn},
			},
			expectedSecretsCount: 0,
			expectError:          false,
		},
		{
			name: "mixed secrets",
			inputSecrets: secret.Secrets{
				{ID: validSecretID, Key: "MY_SECRET", Scope: variable.ScopeContainer},
				{ID: builtInSecretID, Key: "QOVERY_SECRET", Scope: variable.ScopeBuiltIn},
			},
			expectedSecretsCount: 1,
			expectError:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &container.Container{}
			err := c.SetSecrets(tt.inputSecrets)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, c.Secrets, tt.expectedSecretsCount)
			}
		})
	}
}

func TestContainer_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		container   container.Container
		expectError bool
	}{
		{
			name: "valid container",
			container: container.Container{
				ID:                  uuid.New(),
				EnvironmentID:       uuid.New(),
				RegistryID:          uuid.New(),
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				Storages:            storage.Storages{},
				Ports:               port.Ports{},
			},
			expectError: false,
		},
		{
			name: "invalid container with zero id",
			container: container.Container{
				ID:                  uuid.UUID{},
				EnvironmentID:       uuid.New(),
				RegistryID:          uuid.New(),
				Name:                "test-container",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				Storages:            storage.Storages{},
				Ports:               port.Ports{},
			},
			expectError: true,
		},
		{
			name: "invalid container with empty name",
			container: container.Container{
				ID:                  uuid.New(),
				EnvironmentID:       uuid.New(),
				RegistryID:          uuid.New(),
				Name:                "",
				IconUri:             "app://qovery-console/container",
				ImageName:           "nginx",
				Tag:                 "latest",
				CPU:                 500,
				Memory:              512,
				MinRunningInstances: 1,
				MaxRunningInstances: 1,
				Storages:            storage.Storages{},
				Ports:               port.Ports{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.container.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, tt.container.IsValid())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.container.IsValid())
			}
		})
	}
}

func TestContainerConstants(t *testing.T) {
	t.Parallel()

	// Verify default values are set correctly
	assert.Equal(t, status.StateDeployed, container.DefaultState)
	assert.Equal(t, int32(500), int32(container.DefaultCPU))
	assert.Equal(t, int32(10), int32(container.MinCPU))
	assert.Equal(t, int32(512), int32(container.DefaultMemory))
	assert.Equal(t, int32(10), int32(container.MinMemory))
	assert.Equal(t, int32(1), int32(container.DefaultMinRunningInstances))
	assert.Equal(t, int32(1), int32(container.MinMinRunningInstances))
	assert.Equal(t, int32(1), int32(container.DefaultMaxRunningInstances))
	assert.Equal(t, int32(-1), int32(container.MinMaxRunningInstances))
	assert.Equal(t, int32(1), int32(container.MinStorageSize))
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}
