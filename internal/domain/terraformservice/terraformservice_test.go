// go:build unit && !integration
//go:build unit && !integration

package terraformservice_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

func TestTerraformAction_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		action      terraformservice.TerraformAction
		expectError bool
	}{
		{
			name:        "DEFAULT is valid",
			action:      terraformservice.TerraformActionDefault,
			expectError: false,
		},
		{
			name:        "PLAN is valid",
			action:      terraformservice.TerraformActionPlan,
			expectError: false,
		},
		{
			name:        "NOOP is valid",
			action:      terraformservice.TerraformActionNoop,
			expectError: false,
		},
		{
			name:        "INVALID is rejected",
			action:      terraformservice.TerraformAction("INVALID"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTerraformService_Validate(t *testing.T) {
	t.Parallel()

	validGitRepo := terraformservice.GitRepository{
		URL:      "https://github.com/org/repo",
		Branch:   "main",
		RootPath: "/",
	}

	validBackend := terraformservice.Backend{
		Kubernetes: &terraformservice.KubernetesBackend{},
	}

	validEngineVersion := terraformservice.EngineVersion{
		ExplicitVersion:        "1.5.7",
		ReadFromTerraformBlock: false,
	}

	validJobResources := terraformservice.JobResources{
		CPUMilli:   1000,
		RAMMiB:     1024,
		GPU:        0,
		StorageGiB: 20,
	}

	tests := []struct {
		name        string
		service     terraformservice.TerraformService
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid terraform service",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				AutoDeploy:    true,
				GitRepository: validGitRepo,
				TfVarFiles:    []string{"/terraform/prod.tfvars"},
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: false,
		},
		{
			name: "missing name",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
		{
			name: "name without ASCII letters",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "123-456",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
		{
			name: "missing description",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   nil,
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: false,
		},
		{
			name: "both backends specified",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend: terraformservice.Backend{
					Kubernetes:   &terraformservice.KubernetesBackend{},
					UserProvided: &terraformservice.UserProvidedBackend{},
				},
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
		{
			name: "no backend specified",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend:       terraformservice.Backend{},
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
		{
			name: "invalid tfvar path",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/org/repo",
					Branch:   "main",
					RootPath: "/terraform",
				},
				TfVarFiles:    []string{"/invalid/prod.tfvars"},
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
		{
			name: "directory traversal in tfvar path",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				TfVarFiles:    []string{"/../prod.tfvars"},
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
		{
			name: "CPU below minimum",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources: terraformservice.JobResources{
					CPUMilli:   5,
					RAMMiB:     1024,
					GPU:        0,
					StorageGiB: 20,
				},
			},
			expectError: true,
		},
		{
			name: "valid terraform action",
			service: terraformservice.TerraformService{
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "test-service",
				Description:     stringPtr("Test description"),
				AutoDeploy:      true,
				TerraformAction: terraformservice.TerraformActionPlan,
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				EngineVersion:   validEngineVersion,
				JobResources:    validJobResources,
			},
			expectError: false,
		},
		{
			name: "invalid terraform action",
			service: terraformservice.TerraformService{
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "test-service",
				Description:     stringPtr("Test description"),
				AutoDeploy:      true,
				TerraformAction: terraformservice.TerraformAction("INVALID"),
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				EngineVersion:   validEngineVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "invalid engine",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        "INVALID",
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitRepository_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repo        terraformservice.GitRepository
		expectError bool
	}{
		{
			name: "valid git repository",
			repo: terraformservice.GitRepository{
				URL:      "https://github.com/org/repo",
				Branch:   "main",
				RootPath: "/terraform",
			},
			expectError: false,
		},
		{
			name: "missing URL",
			repo: terraformservice.GitRepository{
				URL:      "",
				Branch:   "main",
				RootPath: "/",
			},
			expectError: true,
		},
		{
			name: "directory traversal in root_path",
			repo: terraformservice.GitRepository{
				URL:      "https://github.com/org/repo",
				Branch:   "main",
				RootPath: "../../../etc",
			},
			expectError: true,
		},
		{
			name: "tilde in root_path",
			repo: terraformservice.GitRepository{
				URL:      "https://github.com/org/repo",
				Branch:   "main",
				RootPath: "~/terraform",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBackend_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		backend     terraformservice.Backend
		expectError bool
	}{
		{
			name: "kubernetes backend",
			backend: terraformservice.Backend{
				Kubernetes: &terraformservice.KubernetesBackend{},
			},
			expectError: false,
		},
		{
			name: "user provided backend",
			backend: terraformservice.Backend{
				UserProvided: &terraformservice.UserProvidedBackend{},
			},
			expectError: false,
		},
		{
			name: "both backends",
			backend: terraformservice.Backend{
				Kubernetes:   &terraformservice.KubernetesBackend{},
				UserProvided: &terraformservice.UserProvidedBackend{},
			},
			expectError: true,
		},
		{
			name:        "no backend",
			backend:     terraformservice.Backend{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.backend.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJobResources_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		resources   terraformservice.JobResources
		expectError bool
	}{
		{
			name: "valid resources",
			resources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 20,
			},
			expectError: false,
		},
		{
			name: "CPU below minimum",
			resources: terraformservice.JobResources{
				CPUMilli:   5,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 20,
			},
			expectError: true,
		},
		{
			name: "RAM below minimum",
			resources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     0,
				GPU:        0,
				StorageGiB: 20,
			},
			expectError: true,
		},
		{
			name: "storage below minimum",
			resources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resources.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpsertRepositoryRequest_Validate(t *testing.T) {
	t.Parallel()

	validGitRepo := terraformservice.GitRepository{
		URL:      "https://github.com/org/repo",
		Branch:   "main",
		RootPath: "/",
	}

	validBackend := terraformservice.Backend{
		Kubernetes: &terraformservice.KubernetesBackend{},
	}

	validEngineVersion := terraformservice.EngineVersion{
		ExplicitVersion:        "1.5.7",
		ReadFromTerraformBlock: false,
	}

	validJobResources := terraformservice.JobResources{
		CPUMilli:   1000,
		RAMMiB:     1024,
		GPU:        0,
		StorageGiB: 20,
	}

	tests := []struct {
		name        string
		request     terraformservice.UpsertRepositoryRequest
		expectError bool
	}{
		{
			name: "valid request",
			request: terraformservice.UpsertRepositoryRequest{
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				AutoDeploy:    true,
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: terraformservice.UpsertRepositoryRequest{
				Name:          "",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
		{
			name: "invalid backend",
			request: terraformservice.UpsertRepositoryRequest{
				Name:          "test-service",
				Description:   stringPtr("Test description"),
				GitRepository: validGitRepo,
				Backend:       terraformservice.Backend{},
				Engine:        terraformservice.EngineTerraform,
				EngineVersion: validEngineVersion,
				JobResources:  validJobResources,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestVariable_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		variable    terraformservice.Variable
		expectedErr string
	}{
		{
			name:     "valid variable",
			variable: terraformservice.Variable{Key: "record_name", Value: "example.com"},
		},
		{
			name:        "missing key",
			variable:    terraformservice.Variable{Key: "", Value: "example.com"},
			expectedErr: "variable key is required",
		},
		{
			name:        "missing value names the offending key",
			variable:    terraformservice.Variable{Key: "record_name", Value: ""},
			expectedErr: `variable "record_name": value is required`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.variable.Validate()
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}

func TestTerraformService_Validate_VariableErrorsIdentifyTheVariable(t *testing.T) {
	t.Parallel()

	newService := func(variables []terraformservice.Variable) terraformservice.TerraformService {
		return terraformservice.TerraformService{
			ID:            uuid.New(),
			EnvironmentID: uuid.New(),
			Name:          "test-service",
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/org/repo",
				Branch:   "main",
				RootPath: "/",
			},
			Variables:     variables,
			Backend:       terraformservice.Backend{Kubernetes: &terraformservice.KubernetesBackend{}},
			Engine:        terraformservice.EngineTerraform,
			EngineVersion: terraformservice.EngineVersion{ExplicitVersion: "1.5.7"},
			JobResources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 20,
			},
		}
	}

	tests := []struct {
		name        string
		variables   []terraformservice.Variable
		expectedErr string
	}{
		{
			name: "empty value reports the key, not just the position",
			variables: []terraformservice.Variable{
				{Key: "domain", Value: "example.com"},
				{Key: "record_name", Value: ""},
			},
			expectedErr: `invalid variable param: variable "record_name": value is required`,
		},
		{
			name: "empty key falls back to the index",
			variables: []terraformservice.Variable{
				{Key: "domain", Value: "example.com"},
				{Key: "", Value: "some-value"},
			},
			expectedErr: "invalid variable param: variable at index 1: variable key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, newService(tt.variables).Validate(), tt.expectedErr)
		})
	}
}

func TestUpsertRepositoryRequest_Validate_VariableErrorsIdentifyTheVariable(t *testing.T) {
	t.Parallel()

	newRequest := func(variables []terraformservice.Variable) terraformservice.UpsertRepositoryRequest {
		return terraformservice.UpsertRepositoryRequest{
			Name: "test-service",
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/org/repo",
				Branch:   "main",
				RootPath: "/",
			},
			Variables:     variables,
			Backend:       terraformservice.Backend{Kubernetes: &terraformservice.KubernetesBackend{}},
			Engine:        terraformservice.EngineTerraform,
			EngineVersion: terraformservice.EngineVersion{ExplicitVersion: "1.5.7"},
			JobResources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 20,
			},
		}
	}

	tests := []struct {
		name        string
		variables   []terraformservice.Variable
		expectedErr string
	}{
		{
			name: "empty value reports the key",
			variables: []terraformservice.Variable{
				{Key: "domain", Value: "example.com"},
				{Key: "record_name", Value: ""},
			},
			expectedErr: `invalid terraform service upsert request: variable "record_name": value is required`,
		},
		{
			name: "empty key falls back to the index",
			variables: []terraformservice.Variable{
				{Key: "", Value: "some-value"},
			},
			expectedErr: "invalid terraform service upsert request: variable at index 0: variable key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, newRequest(tt.variables).Validate(), tt.expectedErr)
		})
	}
}

func TestTerraformService_Validate_TfVarPathErrorsIdentifyThePath(t *testing.T) {
	t.Parallel()

	newService := func(rootPath string, tfVarFiles []string) terraformservice.TerraformService {
		return terraformservice.TerraformService{
			ID:            uuid.New(),
			EnvironmentID: uuid.New(),
			Name:          "test-service",
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/org/repo",
				Branch:   "main",
				RootPath: rootPath,
			},
			TfVarFiles:    tfVarFiles,
			Backend:       terraformservice.Backend{Kubernetes: &terraformservice.KubernetesBackend{}},
			Engine:        terraformservice.EngineTerraform,
			EngineVersion: terraformservice.EngineVersion{ExplicitVersion: "1.5.7"},
			JobResources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 20,
			},
		}
	}

	tests := []struct {
		name        string
		rootPath    string
		tfVarFiles  []string
		expectedErr string
	}{
		{
			name:        "path outside root_path",
			rootPath:    "/terraform",
			tfVarFiles:  []string{"/terraform/prod.tfvars", "/invalid/prod.tfvars"},
			expectedErr: `invalid tfvar path: must start with root_path: tfvar path "/invalid/prod.tfvars" must start with root_path "/terraform"`,
		},
		{
			name:        "directory traversal reports the offending path",
			rootPath:    "/",
			tfVarFiles:  []string{"/prod.tfvars", "/../secrets.tfvars"},
			expectedErr: `invalid tfvar path: must start with root_path: tfvar path "/../secrets.tfvars" cannot contain directory traversal sequences (.., ~)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, newService(tt.rootPath, tt.tfVarFiles).Validate(), tt.expectedErr)
		})
	}
}
