//go:build unit && !integration
// +build unit,!integration

package helmRepository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
)

func TestNewHelmRepository(t *testing.T) {
	t.Parallel()

	validRepositoryID := uuid.NewString()
	validOrganizationID := uuid.NewString()
	validURL := "https://charts.example.com"
	description := "Test helm repository"

	tests := []struct {
		name          string
		params        helmRepository.NewHelmRepositoryParams
		expectError   bool
		expectedError error
	}{
		{
			name: "success with valid params HTTPS kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "HTTPS",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with description",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "HTTPS",
				URL:            validURL,
				Description:    &description,
			},
			expectError: false,
		},
		{
			name: "success with skip tls verification",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:       validRepositoryID,
				OrganizationID:     validOrganizationID,
				Name:               "test-repo",
				Kind:               "HTTPS",
				URL:                validURL,
				SkiTlsVerification: boolPtr(true),
			},
			expectError: false,
		},
		{
			name: "success with OCI_ECR kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_ECR",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with OCI_DOCR kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_DOCR",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with OCI_SCALEWAY_CR kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_SCALEWAY_CR",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with OCI_DOCKER_HUB kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_DOCKER_HUB",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with OCI_GITHUB_CR kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_GITHUB_CR",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with OCI_GITLAB_CR kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_GITLAB_CR",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with OCI_PUBLIC_ECR kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_PUBLIC_ECR",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "success with OCI_GENERIC_CR kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "OCI_GENERIC_CR",
				URL:            validURL,
			},
			expectError: false,
		},
		{
			name: "fail with invalid repository id",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   "invalid-uuid",
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "HTTPS",
				URL:            validURL,
			},
			expectError:   true,
			expectedError: helmRepository.ErrInvalidRepositoryIdParam,
		},
		{
			name: "fail with empty repository id",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   "",
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "HTTPS",
				URL:            validURL,
			},
			expectError:   true,
			expectedError: helmRepository.ErrInvalidRepositoryIdParam,
		},
		{
			name: "fail with invalid organization id",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: "invalid-uuid",
				Name:           "test-repo",
				Kind:           "HTTPS",
				URL:            validURL,
			},
			expectError:   true,
			expectedError: helmRepository.ErrInvalidRepositoryOrganizationIDParam,
		},
		{
			name: "fail with empty organization id",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: "",
				Name:           "test-repo",
				Kind:           "HTTPS",
				URL:            validURL,
			},
			expectError:   true,
			expectedError: helmRepository.ErrInvalidRepositoryOrganizationIDParam,
		},
		{
			name: "fail with empty name",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "",
				Kind:           "HTTPS",
				URL:            validURL,
			},
			expectError:   true,
			expectedError: helmRepository.ErrInvalidRepositoryNameParam,
		},
		{
			name: "fail with invalid kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "INVALID_KIND",
				URL:            validURL,
			},
			expectError:   true,
			expectedError: helmRepository.ErrInvalidKindParam,
		},
		{
			name: "fail with empty kind",
			params: helmRepository.NewHelmRepositoryParams{
				RepositoryId:   validRepositoryID,
				OrganizationID: validOrganizationID,
				Name:           "test-repo",
				Kind:           "",
				URL:            validURL,
			},
			expectError:   true,
			expectedError: helmRepository.ErrInvalidKindParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := helmRepository.NewHelmRepository(tt.params)

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
			}
		})
	}
}

func TestKind_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		kind        helmRepository.Kind
		expectError bool
	}{
		{
			name:        "valid HTTPS kind",
			kind:        helmRepository.KindHttps,
			expectError: false,
		},
		{
			name:        "valid OCI_ECR kind",
			kind:        helmRepository.KindECR,
			expectError: false,
		},
		{
			name:        "valid OCI_DOCR kind",
			kind:        helmRepository.KindDocker,
			expectError: false,
		},
		{
			name:        "valid OCI_SCALEWAY_CR kind",
			kind:        helmRepository.KindScalewayCR,
			expectError: false,
		},
		{
			name:        "valid OCI_DOCKER_HUB kind",
			kind:        helmRepository.KindDockerHub,
			expectError: false,
		},
		{
			name:        "valid OCI_GITHUB_CR kind",
			kind:        helmRepository.KindGithubCr,
			expectError: false,
		},
		{
			name:        "valid OCI_GITLAB_CR kind",
			kind:        helmRepository.KindGitlabCr,
			expectError: false,
		},
		{
			name:        "valid OCI_PUBLIC_ECR kind",
			kind:        helmRepository.KindPublicECR,
			expectError: false,
		},
		{
			name:        "valid OCI_GENERIC_CR kind",
			kind:        helmRepository.KindGenericCR,
			expectError: false,
		},
		{
			name:        "invalid kind",
			kind:        helmRepository.Kind("INVALID"),
			expectError: true,
		},
		{
			name:        "empty kind",
			kind:        helmRepository.Kind(""),
			expectError: true,
		},
		{
			name:        "lowercase https",
			kind:        helmRepository.Kind("https"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.kind.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, tt.kind.IsValid())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.kind.IsValid())
			}
		})
	}
}

func TestNewKindFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectError    bool
		expectedResult helmRepository.Kind
	}{
		{
			name:           "valid HTTPS",
			input:          "HTTPS",
			expectError:    false,
			expectedResult: helmRepository.KindHttps,
		},
		{
			name:           "valid OCI_ECR",
			input:          "OCI_ECR",
			expectError:    false,
			expectedResult: helmRepository.KindECR,
		},
		{
			name:           "valid OCI_DOCR",
			input:          "OCI_DOCR",
			expectError:    false,
			expectedResult: helmRepository.KindDocker,
		},
		{
			name:           "valid OCI_SCALEWAY_CR",
			input:          "OCI_SCALEWAY_CR",
			expectError:    false,
			expectedResult: helmRepository.KindScalewayCR,
		},
		{
			name:           "valid OCI_DOCKER_HUB",
			input:          "OCI_DOCKER_HUB",
			expectError:    false,
			expectedResult: helmRepository.KindDockerHub,
		},
		{
			name:           "valid OCI_GITHUB_CR",
			input:          "OCI_GITHUB_CR",
			expectError:    false,
			expectedResult: helmRepository.KindGithubCr,
		},
		{
			name:           "valid OCI_GITLAB_CR",
			input:          "OCI_GITLAB_CR",
			expectError:    false,
			expectedResult: helmRepository.KindGitlabCr,
		},
		{
			name:           "valid OCI_PUBLIC_ECR",
			input:          "OCI_PUBLIC_ECR",
			expectError:    false,
			expectedResult: helmRepository.KindPublicECR,
		},
		{
			name:           "valid OCI_GENERIC_CR",
			input:          "OCI_GENERIC_CR",
			expectError:    false,
			expectedResult: helmRepository.KindGenericCR,
		},
		{
			name:        "invalid type",
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
			input:       "https",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := helmRepository.NewKindFromString(tt.input)

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

func TestKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		kind           helmRepository.Kind
		expectedString string
	}{
		{
			name:           "HTTPS string",
			kind:           helmRepository.KindHttps,
			expectedString: "HTTPS",
		},
		{
			name:           "OCI_ECR string",
			kind:           helmRepository.KindECR,
			expectedString: "OCI_ECR",
		},
		{
			name:           "OCI_DOCR string",
			kind:           helmRepository.KindDocker,
			expectedString: "OCI_DOCR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.kind.String()
			assert.Equal(t, tt.expectedString, result)
		})
	}
}

func TestUpsertRequest_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     helmRepository.UpsertRequest
		expectError bool
	}{
		{
			name: "valid upsert request",
			request: helmRepository.UpsertRequest{
				Name: "test-repo",
				Kind: "HTTPS",
				URL:  "https://charts.example.com",
			},
			expectError: false,
		},
		{
			name: "valid upsert request with description",
			request: helmRepository.UpsertRequest{
				Name:        "test-repo",
				Kind:        "HTTPS",
				URL:         "https://charts.example.com",
				Description: stringPtr("Test description"),
			},
			expectError: false,
		},
		{
			name: "valid upsert request with skip tls",
			request: helmRepository.UpsertRequest{
				Name:               "test-repo",
				Kind:               "HTTPS",
				URL:                "https://charts.example.com",
				SkiTlsVerification: true,
			},
			expectError: false,
		},
		{
			name: "invalid upsert request with empty name",
			request: helmRepository.UpsertRequest{
				Name: "",
				Kind: "HTTPS",
				URL:  "https://charts.example.com",
			},
			expectError: true,
		},
		{
			name: "invalid upsert request with empty kind",
			request: helmRepository.UpsertRequest{
				Name: "test-repo",
				Kind: "",
				URL:  "https://charts.example.com",
			},
			expectError: true,
		},
		{
			name: "invalid upsert request with empty url",
			request: helmRepository.UpsertRequest{
				Name: "test-repo",
				Kind: "HTTPS",
				URL:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, tt.request.IsValid())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.request.IsValid())
			}
		})
	}
}

func TestAllowedKindValues(t *testing.T) {
	t.Parallel()

	// Verify all allowed values are present
	assert.Len(t, helmRepository.AllowedKindValues, 9)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindHttps)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindECR)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindDocker)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindScalewayCR)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindDockerHub)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindGithubCr)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindGitlabCr)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindPublicECR)
	assert.Contains(t, helmRepository.AllowedKindValues, helmRepository.KindGenericCR)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}
