// go:build unit && !integration
//go:build unit && !integration
// +build unit,!integration

package terraformservice_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

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

	validProviderVersion := terraformservice.ProviderVersion{
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
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "test-service",
				Description:     "Test description",
				AutoDeploy:      true,
				GitRepository:   validGitRepo,
				TfVarFiles:      []string{"/terraform/prod.tfvars"},
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: false,
		},
		{
			name: "missing name",
			service: terraformservice.TerraformService{
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "",
				Description:     "Test description",
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "name without ASCII letters",
			service: terraformservice.TerraformService{
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "123-456",
				Description:     "Test description",
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "missing description",
			service: terraformservice.TerraformService{
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "test-service",
				Description:     "",
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "both backends specified",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   "Test description",
				GitRepository: validGitRepo,
				Backend: terraformservice.Backend{
					Kubernetes:   &terraformservice.KubernetesBackend{},
					UserProvided: &terraformservice.UserProvidedBackend{},
				},
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "no backend specified",
			service: terraformservice.TerraformService{
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "test-service",
				Description:     "Test description",
				GitRepository:   validGitRepo,
				Backend:         terraformservice.Backend{},
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "invalid tfvar path",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   "Test description",
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/org/repo",
					Branch:   "main",
					RootPath: "/terraform",
				},
				TfVarFiles:      []string{"/invalid/prod.tfvars"},
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "directory traversal in tfvar path",
			service: terraformservice.TerraformService{
				ID:            uuid.New(),
				EnvironmentID: uuid.New(),
				Name:          "test-service",
				Description:   "Test description",
				GitRepository: validGitRepo,
				TfVarFiles:    []string{"/../prod.tfvars"},
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
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
				Description:   "Test description",
				GitRepository: validGitRepo,
				Backend:       validBackend,
				Engine:        terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
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
			name: "invalid engine",
			service: terraformservice.TerraformService{
				ID:              uuid.New(),
				EnvironmentID:   uuid.New(),
				Name:            "test-service",
				Description:     "Test description",
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          "INVALID",
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
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

	validProviderVersion := terraformservice.ProviderVersion{
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
				Name:            "test-service",
				Description:     "Test description",
				AutoDeploy:      true,
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: terraformservice.UpsertRepositoryRequest{
				Name:            "",
				Description:     "Test description",
				GitRepository:   validGitRepo,
				Backend:         validBackend,
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
			},
			expectError: true,
		},
		{
			name: "invalid backend",
			request: terraformservice.UpsertRepositoryRequest{
				Name:            "test-service",
				Description:     "Test description",
				GitRepository:   validGitRepo,
				Backend:         terraformservice.Backend{},
				Engine:          terraformservice.EngineTerraform,
				ProviderVersion: validProviderVersion,
				JobResources:    validJobResources,
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
