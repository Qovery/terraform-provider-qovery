package qoveryapi

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

func TestNewQoveryTerraformRequestFromDomain(t *testing.T) {
	t.Parallel()

	description := "Test terraform service description"
	timeout := int32(3600)

	testCases := []struct {
		TestName    string
		Request     terraformservice.UpsertRepositoryRequest
		ExpectError bool
	}{
		{
			TestName: "success_minimal",
			Request: terraformservice.UpsertRepositoryRequest{
				Name:       "test-terraform-service",
				AutoDeploy: false,
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/example/repo.git",
					Branch:   "main",
					RootPath: "/",
				},
				TfVarFiles: []string{},
				Variables:  []terraformservice.Variable{},
				Backend: terraformservice.Backend{
					Kubernetes: &terraformservice.KubernetesBackend{},
				},
				Engine: terraformservice.EngineTerraform,
				EngineVersion: terraformservice.EngineVersion{
					ExplicitVersion: "1.5.7",
				},
				JobResources: terraformservice.JobResources{
					CPUMilli:   1000,
					RAMMiB:     1024,
					GPU:        0,
					StorageGiB: 20,
				},
			},
		},
		{
			TestName: "success_with_description_and_timeout",
			Request: terraformservice.UpsertRepositoryRequest{
				Name:        "terraform-with-options",
				Description: &description,
				AutoDeploy:  true,
				TimeoutSec:  &timeout,
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/example/repo.git",
					Branch:   "develop",
					RootPath: "/infra",
				},
				TfVarFiles: []string{"/infra/dev.tfvars"},
				Variables: []terraformservice.Variable{
					{Key: "environment", Value: "dev", Secret: false},
					{Key: "api_key", Value: "secret123", Secret: true},
				},
				Backend: terraformservice.Backend{
					Kubernetes: &terraformservice.KubernetesBackend{},
				},
				Engine: terraformservice.EngineTerraform,
				EngineVersion: terraformservice.EngineVersion{
					ExplicitVersion: "1.6.0",
				},
				JobResources: terraformservice.JobResources{
					CPUMilli:   2000,
					RAMMiB:     2048,
					GPU:        1,
					StorageGiB: 50,
				},
				IconURI:               "custom-icon",
				UseClusterCredentials: true,
			},
		},
		{
			TestName: "success_with_user_provided_backend",
			Request: terraformservice.UpsertRepositoryRequest{
				Name:       "terraform-user-backend",
				AutoDeploy: false,
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/example/repo.git",
					Branch:   "main",
					RootPath: "/",
				},
				Backend: terraformservice.Backend{
					UserProvided: &terraformservice.UserProvidedBackend{},
				},
				Engine: terraformservice.EngineOpenTofu,
				EngineVersion: terraformservice.EngineVersion{
					ExplicitVersion: "1.6.0",
				},
				JobResources: terraformservice.JobResources{
					CPUMilli:   1000,
					RAMMiB:     1024,
					GPU:        0,
					StorageGiB: 20,
				},
			},
		},
		{
			TestName: "success_with_git_token",
			Request: terraformservice.UpsertRepositoryRequest{
				Name:       "terraform-with-token",
				AutoDeploy: false,
				GitRepository: terraformservice.GitRepository{
					URL:        "https://github.com/private/repo.git",
					Branch:     "main",
					RootPath:   "/",
					GitTokenID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				},
				Backend: terraformservice.Backend{
					Kubernetes: &terraformservice.KubernetesBackend{},
				},
				Engine: terraformservice.EngineTerraform,
				EngineVersion: terraformservice.EngineVersion{
					ExplicitVersion: "1.5.7",
				},
				JobResources: terraformservice.JobResources{
					CPUMilli:   1000,
					RAMMiB:     1024,
					GPU:        0,
					StorageGiB: 20,
				},
			},
		},
		{
			TestName: "success_with_action_extra_arguments",
			Request: terraformservice.UpsertRepositoryRequest{
				Name:       "terraform-with-extra-args",
				AutoDeploy: false,
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/example/repo.git",
					Branch:   "main",
					RootPath: "/",
				},
				Backend: terraformservice.Backend{
					Kubernetes: &terraformservice.KubernetesBackend{},
				},
				Engine: terraformservice.EngineTerraform,
				EngineVersion: terraformservice.EngineVersion{
					ExplicitVersion: "1.5.7",
				},
				JobResources: terraformservice.JobResources{
					CPUMilli:   1000,
					RAMMiB:     1024,
					GPU:        0,
					StorageGiB: 20,
				},
				ActionExtraArguments: map[string][]string{
					"apply": {"-lock=false", "-auto-approve"},
					"plan":  {"-detailed-exitcode"},
				},
			},
		},
		{
			TestName: "error_invalid_request_missing_backend",
			Request: terraformservice.UpsertRepositoryRequest{
				Name:       "terraform-no-backend",
				AutoDeploy: false,
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/example/repo.git",
					Branch:   "main",
					RootPath: "/",
				},
				Backend: terraformservice.Backend{}, // No backend specified
				Engine:  terraformservice.EngineTerraform,
				EngineVersion: terraformservice.EngineVersion{
					ExplicitVersion: "1.5.7",
				},
				JobResources: terraformservice.JobResources{
					CPUMilli:   1000,
					RAMMiB:     1024,
					GPU:        0,
					StorageGiB: 20,
				},
			},
			ExpectError: true,
		},
		{
			TestName: "error_invalid_request_missing_name",
			Request: terraformservice.UpsertRepositoryRequest{
				Name:       "", // Empty name
				AutoDeploy: false,
				GitRepository: terraformservice.GitRepository{
					URL:      "https://github.com/example/repo.git",
					Branch:   "main",
					RootPath: "/",
				},
				Backend: terraformservice.Backend{
					Kubernetes: &terraformservice.KubernetesBackend{},
				},
				Engine: terraformservice.EngineTerraform,
				EngineVersion: terraformservice.EngineVersion{
					ExplicitVersion: "1.5.7",
				},
				JobResources: terraformservice.JobResources{
					CPUMilli:   1000,
					RAMMiB:     1024,
					GPU:        0,
					StorageGiB: 20,
				},
			},
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result, err := newQoveryTerraformRequestFromDomain(tc.Request)
			if tc.ExpectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.Request.Name, result.Name)
			assert.Equal(t, tc.Request.AutoDeploy, result.AutoDeploy)

			// Verify description
			if tc.Request.Description != nil {
				assert.Equal(t, *tc.Request.Description, result.Description)
			}

			// Verify timeout
			if tc.Request.TimeoutSec != nil {
				assert.NotNil(t, result.TimeoutSec)
				assert.Equal(t, *tc.Request.TimeoutSec, *result.TimeoutSec)
			}

			// Verify icon
			if tc.Request.IconURI != "" {
				assert.NotNil(t, result.IconUri)
				assert.Equal(t, tc.Request.IconURI, *result.IconUri)
			}

			// Verify action extra arguments
			if len(tc.Request.ActionExtraArguments) > 0 {
				assert.NotNil(t, result.ActionExtraArguments)
			}
		})
	}
}

func TestNewDomainTerraformServiceFromQovery_NilResponse(t *testing.T) {
	t.Parallel()

	result, err := newDomainTerraformServiceFromQovery(nil, "deployment-stage-id", "{}")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "terraform response cannot be nil")
}
