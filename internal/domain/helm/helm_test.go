//go:build unit && !integration
// +build unit,!integration

package helm_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestNewHelm(t *testing.T) {
	t.Parallel()

	validHelmID := uuid.NewString()
	validEnvironmentID := uuid.NewString()
	stateDeployed := status.StateDeployed.String()
	stateStopped := status.StateStopped.String()
	branch := "main"

	tests := []struct {
		name          string
		params        helm.NewHelmParams
		expectError   bool
		expectedError error
	}{
		{
			name: "success with git repository source",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with helm repository source",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceHelmRepository: &helm.NewHelmSourceHelmRepository{
						RepositoryId: uuid.NewString(),
						ChartName:    "nginx",
						ChartVersion: "1.0.0",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with deployed state",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				State:         &stateDeployed,
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with stopped state",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				State:         &stateStopped,
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with description",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Description:   stringPtr("Test helm service"),
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with timeout",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				TimeoutSec:    int32Ptr(900),
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with auto preview and auto deploy",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				AutoPreview:   true,
				AutoDeploy:    true,
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with allow cluster wide resources",
			params: helm.NewHelmParams{
				HelmID:                    validHelmID,
				EnvironmentID:             validEnvironmentID,
				Name:                      "test-helm",
				AllowClusterWideResources: true,
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with arguments",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Arguments:     []string{"--wait", "--timeout", "600s"},
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "success with ports",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
				Ports: []helm.NewHelmPortParams{
					{
						Name:         "http",
						InternalPort: 8080,
						ExternalPort: int32Ptr(80),
						ServiceName:  "my-service",
						Protocol:     "HTTP",
						IsDefault:    true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "success with values override set",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{
					Set: [][]string{{"key1", "value1"}, {"key2", "value2"}},
				},
			},
			expectError: false,
		},
		{
			name: "success with deployment stage id",
			params: helm.NewHelmParams{
				HelmID:            validHelmID,
				EnvironmentID:     validEnvironmentID,
				Name:              "test-helm",
				DeploymentStageID: uuid.NewString(),
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError: false,
		},
		{
			name: "fail with invalid helm id",
			params: helm.NewHelmParams{
				HelmID:        "invalid-uuid",
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError:   true,
			expectedError: helm.ErrInvalidHelmIDParam,
		},
		{
			name: "fail with empty helm id",
			params: helm.NewHelmParams{
				HelmID:        "",
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError:   true,
			expectedError: helm.ErrInvalidHelmIDParam,
		},
		{
			name: "fail with invalid environment id",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: "invalid-uuid",
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError:   true,
			expectedError: helm.ErrInvalidHelmEnvironmentIDParam,
		},
		{
			name: "fail with empty environment id",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: "",
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError:   true,
			expectedError: helm.ErrInvalidHelmEnvironmentIDParam,
		},
		{
			name: "fail with empty name",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError:   true,
			expectedError: helm.ErrInvalidHelmNameParam,
		},
		{
			name: "fail with invalid state",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				State:         stringPtr("INVALID_STATE"),
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
			},
			expectError:   true,
			expectedError: helm.ErrInvalidHelmStateParam,
		},
		{
			name: "fail with invalid port protocol",
			params: helm.NewHelmParams{
				HelmID:        validHelmID,
				EnvironmentID: validEnvironmentID,
				Name:          "test-helm",
				Source: helm.NewHelmSourceParams{
					HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
						Url:      "https://github.com/example/charts.git",
						Branch:   &branch,
						RootPath: "/charts/myapp",
					},
				},
				ValuesOverride: helm.NewHelmValuesOverrideParams{},
				Ports: []helm.NewHelmPortParams{
					{
						Name:         "http",
						InternalPort: 8080,
						ServiceName:  "my-service",
						Protocol:     "INVALID_PROTOCOL",
						IsDefault:    true,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := helm.NewHelm(tt.params)

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
				assert.True(t, result.IsValid())
			}
		})
	}
}

func TestProtocol_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		protocol    helm.Protocol
		expectError bool
	}{
		{
			name:        "valid HTTP protocol",
			protocol:    helm.ProtocolHttp,
			expectError: false,
		},
		{
			name:        "valid GRPC protocol",
			protocol:    helm.ProtocolGrpc,
			expectError: false,
		},
		{
			name:        "invalid protocol",
			protocol:    helm.Protocol("INVALID"),
			expectError: true,
		},
		{
			name:        "empty protocol",
			protocol:    helm.Protocol(""),
			expectError: true,
		},
		{
			name:        "lowercase http",
			protocol:    helm.Protocol("http"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.protocol.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, tt.protocol.IsValid())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.protocol.IsValid())
			}
		})
	}
}

func TestNewProtocolFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectError    bool
		expectedResult helm.Protocol
	}{
		{
			name:           "valid HTTP",
			input:          "HTTP",
			expectError:    false,
			expectedResult: helm.ProtocolHttp,
		},
		{
			name:           "valid GRPC",
			input:          "GRPC",
			expectError:    false,
			expectedResult: helm.ProtocolGrpc,
		},
		{
			name:        "invalid protocol",
			input:       "INVALID",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "lowercase",
			input:       "http",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := helm.NewProtocolFromString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult, *result)
			}
		})
	}
}

func TestProtocol_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		protocol       helm.Protocol
		expectedString string
	}{
		{
			name:           "HTTP string",
			protocol:       helm.ProtocolHttp,
			expectedString: "HTTP",
		},
		{
			name:           "GRPC string",
			protocol:       helm.ProtocolGrpc,
			expectedString: "GRPC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.protocol.String()
			assert.Equal(t, tt.expectedString, result)
		})
	}
}

func TestHelm_SetState(t *testing.T) {
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
			h := &helm.Helm{}
			err := h.SetState(tt.inputState)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedState, h.State)
			}
		})
	}
}

func TestHelm_SetEnvironmentVariables(t *testing.T) {
	t.Parallel()

	// Create a fixed helm ID for constructing host keys
	helmID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	// Host key format: QOVERY_HELM_Z{first 8 chars uppercase}_HOST_EXTERNAL/INTERNAL
	hostExternalKey := "QOVERY_HELM_Z12345678_HOST_EXTERNAL"
	hostInternalKey := "QOVERY_HELM_Z12345678_HOST_INTERNAL"

	validVarID := uuid.New()
	builtInVarID := uuid.New()
	hostVarID := uuid.New()

	tests := []struct {
		name                     string
		inputVars                variable.Variables
		expectedEnvVarsCount     int
		expectedBuiltInVarsCount int
		expectError              bool
	}{
		{
			name:                     "empty variables",
			inputVars:                variable.Variables{},
			expectedEnvVarsCount:     0,
			expectedBuiltInVarsCount: 0,
			expectError:              false,
		},
		{
			name: "variables with required host keys",
			inputVars: variable.Variables{
				{ID: validVarID, Key: "MY_VAR", Value: "my_value", Scope: variable.ScopeHelm},
				{ID: hostVarID, Key: hostExternalKey, Value: "external.host.com", Scope: variable.ScopeBuiltIn},
			},
			expectedEnvVarsCount:     1,
			expectedBuiltInVarsCount: 1,
			expectError:              false,
		},
		{
			name: "built-in variables with host key",
			inputVars: variable.Variables{
				{ID: builtInVarID, Key: "QOVERY_VAR", Value: "qovery_value", Scope: variable.ScopeBuiltIn},
				{ID: hostVarID, Key: hostInternalKey, Value: "internal.host.com", Scope: variable.ScopeBuiltIn},
			},
			expectedEnvVarsCount:     0,
			expectedBuiltInVarsCount: 2,
			expectError:              false,
		},
		{
			name: "variables without host keys fails",
			inputVars: variable.Variables{
				{ID: validVarID, Key: "MY_VAR", Value: "my_value", Scope: variable.ScopeHelm},
			},
			expectError: true, // SetHosts will fail without host env vars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &helm.Helm{
				ID: helmID,
			}
			err := h.SetEnvironmentVariables(tt.inputVars)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, h.EnvironmentVariables, tt.expectedEnvVarsCount)
				assert.Len(t, h.BuiltInEnvironmentVariables, tt.expectedBuiltInVarsCount)
			}
		})
	}
}

func TestHelm_SetSecrets(t *testing.T) {
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
				{ID: validSecretID, Key: "MY_SECRET", Scope: variable.ScopeHelm},
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
				{ID: validSecretID, Key: "MY_SECRET", Scope: variable.ScopeHelm},
				{ID: builtInSecretID, Key: "QOVERY_SECRET", Scope: variable.ScopeBuiltIn},
			},
			expectedSecretsCount: 1,
			expectError:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &helm.Helm{}
			err := h.SetSecrets(tt.inputSecrets)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, h.Secrets, tt.expectedSecretsCount)
			}
		})
	}
}

func TestNewHelmSource(t *testing.T) {
	t.Parallel()

	branch := "main"

	tests := []struct {
		name        string
		params      helm.NewHelmSourceParams
		expectError bool
		expectGit   bool
		expectHelm  bool
	}{
		{
			name: "git repository source",
			params: helm.NewHelmSourceParams{
				HelmSourceGitRepository: &helm.NewHelmSourceGitRepository{
					Url:      "https://github.com/example/charts.git",
					Branch:   &branch,
					RootPath: "/charts",
				},
			},
			expectError: false,
			expectGit:   true,
			expectHelm:  false,
		},
		{
			name: "helm repository source",
			params: helm.NewHelmSourceParams{
				HelmSourceHelmRepository: &helm.NewHelmSourceHelmRepository{
					RepositoryId: uuid.NewString(),
					ChartName:    "nginx",
					ChartVersion: "1.0.0",
				},
			},
			expectError: false,
			expectGit:   false,
			expectHelm:  true,
		},
		{
			name:        "empty source",
			params:      helm.NewHelmSourceParams{},
			expectError: false,
			expectGit:   false,
			expectHelm:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := helm.NewHelmSource(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectGit {
					assert.NotNil(t, result.GitRepository)
				} else {
					assert.Nil(t, result.GitRepository)
				}
				if tt.expectHelm {
					assert.NotNil(t, result.HelmRepository)
				} else {
					assert.Nil(t, result.HelmRepository)
				}
			}
		})
	}
}

func TestNewHelmValuesOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		params      helm.NewHelmValuesOverrideParams
		expectError bool
	}{
		{
			name:        "empty values override",
			params:      helm.NewHelmValuesOverrideParams{},
			expectError: false,
		},
		{
			name: "with set values",
			params: helm.NewHelmValuesOverrideParams{
				Set: [][]string{{"key1", "value1"}, {"key2", "value2"}},
			},
			expectError: false,
		},
		{
			name: "with set string values",
			params: helm.NewHelmValuesOverrideParams{
				SetString: [][]string{{"key1", "value1"}},
			},
			expectError: false,
		},
		{
			name: "with set json values",
			params: helm.NewHelmValuesOverrideParams{
				SetJson: [][]string{{"key1", `{"nested": "value"}`}},
			},
			expectError: false,
		},
		{
			name: "with file raw values",
			params: helm.NewHelmValuesOverrideParams{
				File: &helm.ValuesOverrideFile{
					Raw: &helm.Raw{
						Values: []helm.RawValue{
							{Name: "values.yaml", Content: "key: value"},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := helm.NewHelmValuesOverride(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestHelm_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		helm        helm.Helm
		expectError bool
	}{
		{
			name: "valid helm",
			helm: helm.Helm{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-helm",
			},
			expectError: false,
		},
		{
			name: "invalid helm with empty name",
			helm: helm.Helm{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "",
			},
			expectError: true,
		},
		{
			name: "invalid helm with zero id",
			helm: helm.Helm{
				ID:            uuid.UUID{},
				EnvironmentID: uuid.New(),
				Name:          "test-helm",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.helm.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, tt.helm.IsValid())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.helm.IsValid())
			}
		})
	}
}

func TestAllowedProtocols(t *testing.T) {
	t.Parallel()

	// Verify all allowed values are present
	assert.Len(t, helm.AllowedProtocols, 2)
	assert.Contains(t, helm.AllowedProtocols, helm.ProtocolHttp)
	assert.Contains(t, helm.AllowedProtocols, helm.ProtocolGrpc)
}

func TestDefaultTimeoutSec(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int64(600), helm.DefaultTimeoutSec)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}
