//go:build unit && !integration
// +build unit,!integration

package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

// Helper function to create an HTTP response with a body
func createHTTPResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func TestNewSecretService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  secret.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.SecretRepository{},
			ExpectError: false,
		},
		{
			TestName:    "error_with_nil_repository",
			Repository:  nil,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewSecretService(tc.Repository)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				assert.Equal(t, ErrInvalidRepository, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestSecretService_List(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""

	expectedSecrets := secret.Secrets{
		{
			ID:          uuid.New(),
			Scope:       variable.ScopeEnvironment,
			Key:         "SECRET_KEY",
			Type:        "VALUE",
			Description: "Test secret",
		},
		{
			ID:          uuid.New(),
			Scope:       variable.ScopeEnvironment,
			Key:         "ANOTHER_SECRET",
			Type:        "VALUE",
			Description: "Another test secret",
		},
	}

	testCases := []struct {
		TestName      string
		ResourceID    string
		SetupMock     func(*mocks_test.SecretRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_resource_id",
			ResourceID:    emptyResourceID,
			SetupMock:     func(m *mocks_test.SecretRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid resource id param",
		},
		{
			TestName:      "error_invalid_resource_id",
			ResourceID:    invalidResourceID,
			SetupMock:     func(m *mocks_test.SecretRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid resource id param",
		},
		{
			TestName:   "error_repository_failure",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to list secrets",
		},
		{
			TestName:   "success",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(expectedSecrets, nil)
			},
			ExpectError: false,
		},
		{
			TestName:   "success_empty_list",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.SecretRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewSecretService(mockRepo)
			require.NoError(t, err)

			result, err := service.List(context.Background(), tc.ResourceID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSecretService_Update(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""

	secretID1 := uuid.New().String()
	secretID2 := uuid.New().String()
	secretID3 := uuid.New().String()

	validSecretsRequest := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key:         "NEW_SECRET",
					Value:       "new-value",
					Description: "New secret",
				},
			},
		},
		Update: []secret.DiffUpdateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key:         "EXISTING_SECRET",
					Value:       "updated-value",
					Description: "Updated secret",
				},
				SecretID: secretID1,
			},
		},
		Delete: []secret.DiffDeleteRequest{
			{
				SecretID: secretID2,
			},
		},
	}

	invalidSecretsRequest := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key: "", // Invalid: empty key
				},
			},
		},
	}

	emptySecretsRequest := secret.DiffRequest{}
	emptyAliasesRequest := secret.DiffRequest{}
	emptyOverridesRequest := secret.DiffRequest{}

	existingSecrets := secret.Secrets{
		{
			ID:    uuid.MustParse(secretID3),
			Scope: variable.ScopeEnvironment,
			Key:   "EXISTING_SECRET",
			Type:  "VALUE",
		},
	}

	createdSecret := &secret.Secret{
		ID:    uuid.New(),
		Scope: variable.ScopeEnvironment,
		Key:   "NEW_SECRET",
		Type:  "VALUE",
	}

	updatedSecret := &secret.Secret{
		ID:    uuid.MustParse(secretID1),
		Scope: variable.ScopeEnvironment,
		Key:   "EXISTING_SECRET",
		Type:  "VALUE",
	}

	testCases := []struct {
		TestName                 string
		ResourceID               string
		SecretsRequest           secret.DiffRequest
		AliasesRequest           secret.DiffRequest
		OverridesRequest         secret.DiffRequest
		OverrideAuthorizedScopes map[variable.Scope]struct{}
		SetupMock                func(*mocks_test.SecretRepository)
		ExpectError              bool
		ErrorContains            string
	}{
		{
			TestName:                 "error_empty_resource_id",
			ResourceID:               emptyResourceID,
			SecretsRequest:           validSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock:                func(m *mocks_test.SecretRepository) {},
			ExpectError:              true,
			ErrorContains:            "invalid resource id param",
		},
		{
			TestName:                 "error_invalid_resource_id",
			ResourceID:               invalidResourceID,
			SecretsRequest:           validSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock:                func(m *mocks_test.SecretRepository) {},
			ExpectError:              true,
			ErrorContains:            "invalid resource id param",
		},
		{
			TestName:                 "error_invalid_secrets_request",
			ResourceID:               validResourceID,
			SecretsRequest:           invalidSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock:                func(m *mocks_test.SecretRepository) {},
			ExpectError:              true,
			ErrorContains:            "failed to update secrets",
		},
		{
			TestName:                 "error_delete_failure",
			ResourceID:               validResourceID,
			SecretsRequest:           validSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					Delete(mock.Anything, validResourceID, secretID2).
					Return(apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, secretID2, createHTTPResponse(http.StatusInternalServerError), errors.New("delete error")))
			},
			ExpectError:   true,
			ErrorContains: "failed to update secrets",
		},
		{
			TestName:                 "error_update_failure",
			ResourceID:               validResourceID,
			SecretsRequest:           validSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					Delete(mock.Anything, validResourceID, secretID2).
					Return(nil)
				m.EXPECT().
					Update(mock.Anything, validResourceID, secretID1, mock.Anything).
					Return(nil, errors.New("update error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update secrets",
		},
		{
			TestName:                 "error_create_failure",
			ResourceID:               validResourceID,
			SecretsRequest:           validSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					Delete(mock.Anything, validResourceID, secretID2).
					Return(nil)
				m.EXPECT().
					Update(mock.Anything, validResourceID, secretID1, mock.Anything).
					Return(updatedSecret, nil)
				m.EXPECT().
					Create(mock.Anything, validResourceID, mock.Anything).
					Return(nil, errors.New("create error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update secrets",
		},
		{
			TestName:                 "error_list_failure_after_update",
			ResourceID:               validResourceID,
			SecretsRequest:           validSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					Delete(mock.Anything, validResourceID, secretID2).
					Return(nil)
				m.EXPECT().
					Update(mock.Anything, validResourceID, secretID1, mock.Anything).
					Return(updatedSecret, nil)
				m.EXPECT().
					Create(mock.Anything, validResourceID, mock.Anything).
					Return(createdSecret, nil)
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(nil, errors.New("list error"))
			},
			ExpectError: true,
		},
		{
			TestName:                 "success_secrets_only",
			ResourceID:               validResourceID,
			SecretsRequest:           validSecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					Delete(mock.Anything, validResourceID, secretID2).
					Return(nil)
				m.EXPECT().
					Update(mock.Anything, validResourceID, secretID1, mock.Anything).
					Return(updatedSecret, nil)
				m.EXPECT().
					Create(mock.Anything, validResourceID, mock.Anything).
					Return(createdSecret, nil)
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
			},
			ExpectError: false,
		},
		{
			TestName:                 "success_empty_requests",
			ResourceID:               validResourceID,
			SecretsRequest:           emptySecretsRequest,
			AliasesRequest:           emptyAliasesRequest,
			OverridesRequest:         emptyOverridesRequest,
			OverrideAuthorizedScopes: make(map[variable.Scope]struct{}),
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.SecretRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewSecretService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(
				context.Background(),
				tc.ResourceID,
				tc.SecretsRequest,
				tc.AliasesRequest,
				tc.OverridesRequest,
				tc.OverrideAuthorizedScopes,
			)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSecretService_Update_WithAliases(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	aliasSecretID := uuid.New().String()
	aliasedSecretID := uuid.New().String()

	emptySecretsRequest := secret.DiffRequest{}
	emptyOverridesRequest := secret.DiffRequest{}

	aliasesRequest := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key:         "ALIAS_KEY",
					Value:       "ORIGINAL_KEY",
					Description: "Alias secret",
				},
			},
		},
		Update: []secret.DiffUpdateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key:         "EXISTING_ALIAS",
					Value:       "NEW_ORIGINAL_KEY",
					Description: "Updated alias",
				},
				SecretID: aliasSecretID,
			},
		},
		Delete: []secret.DiffDeleteRequest{
			{
				SecretID: uuid.New().String(),
			},
		},
	}

	invalidAliasesRequest := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key: "", // Invalid: empty key
				},
			},
		},
	}

	existingSecrets := secret.Secrets{
		{
			ID:    uuid.MustParse(aliasedSecretID),
			Scope: variable.ScopeEnvironment,
			Key:   "ORIGINAL_KEY",
			Type:  "VALUE",
		},
		{
			ID:    uuid.New(),
			Scope: variable.ScopeEnvironment,
			Key:   "NEW_ORIGINAL_KEY",
			Type:  "VALUE",
		},
		{
			ID:    uuid.New(),
			Scope: variable.ScopeBuiltIn,
			Key:   "BUILT_IN_KEY",
			Type:  "BUILT_IN",
		},
	}

	createdAlias := &secret.Secret{
		ID:    uuid.New(),
		Scope: variable.ScopeEnvironment,
		Key:   "ALIAS_KEY",
		Type:  "ALIAS",
	}

	updatedAlias := &secret.Secret{
		ID:    uuid.MustParse(aliasSecretID),
		Scope: variable.ScopeEnvironment,
		Key:   "EXISTING_ALIAS",
		Type:  "ALIAS",
	}

	testCases := []struct {
		TestName       string
		AliasesRequest secret.DiffRequest
		SetupMock      func(*mocks_test.SecretRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_invalid_aliases_request",
			AliasesRequest: invalidAliasesRequest,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:       "error_alias_delete_failure_non_404",
			AliasesRequest: aliasesRequest,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, "test", createHTTPResponse(http.StatusInternalServerError), errors.New("delete error")))
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:       "success_alias_delete_404_ignored",
			AliasesRequest: aliasesRequest,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				// Delete for the delete request (ignored 404)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, "test", createHTTPResponse(http.StatusNotFound), errors.New("not found"))).
					Once()
				// Delete for the update request (ignored 404)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, aliasSecretID).
					Return(apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, aliasSecretID, createHTTPResponse(http.StatusNotFound), errors.New("not found"))).
					Once()
				m.EXPECT().
					CreateAlias(mock.Anything, validResourceID, mock.Anything, mock.Anything).
					Return(updatedAlias, nil)
				m.EXPECT().
					CreateAlias(mock.Anything, validResourceID, mock.Anything, aliasedSecretID).
					Return(createdAlias, nil)
			},
			ExpectError: false,
		},
		{
			TestName:       "error_alias_create_failure",
			AliasesRequest: aliasesRequest,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, aliasSecretID).
					Return(nil)
				m.EXPECT().
					CreateAlias(mock.Anything, validResourceID, mock.Anything, mock.Anything).
					Return(nil, errors.New("create alias error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:       "success_with_aliases",
			AliasesRequest: aliasesRequest,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, aliasSecretID).
					Return(nil)
				m.EXPECT().
					CreateAlias(mock.Anything, validResourceID, mock.Anything, mock.Anything).
					Return(updatedAlias, nil)
				m.EXPECT().
					CreateAlias(mock.Anything, validResourceID, mock.Anything, aliasedSecretID).
					Return(createdAlias, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.SecretRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewSecretService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(
				context.Background(),
				validResourceID,
				emptySecretsRequest,
				tc.AliasesRequest,
				emptyOverridesRequest,
				make(map[variable.Scope]struct{}),
			)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSecretService_Update_WithOverrides(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	overrideSecretID := uuid.New().String()
	overriddenSecretID := uuid.New().String()

	emptySecretsRequest := secret.DiffRequest{}
	emptyAliasesRequest := secret.DiffRequest{}

	overridesRequest := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key:         "OVERRIDDEN_KEY",
					Value:       "override-value",
					Description: "Override secret",
				},
			},
		},
		Update: []secret.DiffUpdateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key:         "EXISTING_OVERRIDE",
					Value:       "new-override-value",
					Description: "Updated override",
				},
				SecretID: overrideSecretID,
			},
		},
		Delete: []secret.DiffDeleteRequest{
			{
				SecretID: uuid.New().String(),
			},
		},
	}

	invalidOverridesRequest := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{
			{
				UpsertRequest: secret.UpsertRequest{
					Key: "", // Invalid: empty key
				},
			},
		},
	}

	overrideAuthorizedScopes := map[variable.Scope]struct{}{
		variable.ScopeProject: {},
	}

	existingSecrets := secret.Secrets{
		{
			ID:    uuid.MustParse(overriddenSecretID),
			Scope: variable.ScopeProject,
			Key:   "OVERRIDDEN_KEY",
			Type:  "VALUE",
		},
		{
			ID:    uuid.New(),
			Scope: variable.ScopeProject,
			Key:   "EXISTING_OVERRIDE",
			Type:  "VALUE",
		},
	}

	createdOverride := &secret.Secret{
		ID:    uuid.New(),
		Scope: variable.ScopeEnvironment,
		Key:   "OVERRIDDEN_KEY",
		Type:  "OVERRIDE",
	}

	updatedOverride := &secret.Secret{
		ID:    uuid.MustParse(overrideSecretID),
		Scope: variable.ScopeEnvironment,
		Key:   "EXISTING_OVERRIDE",
		Type:  "OVERRIDE",
	}

	testCases := []struct {
		TestName                 string
		OverridesRequest         secret.DiffRequest
		OverrideAuthorizedScopes map[variable.Scope]struct{}
		SetupMock                func(*mocks_test.SecretRepository)
		ExpectError              bool
		ErrorContains            string
	}{
		{
			TestName:                 "error_invalid_overrides_request",
			OverridesRequest:         invalidOverridesRequest,
			OverrideAuthorizedScopes: overrideAuthorizedScopes,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:                 "error_override_delete_failure_non_404",
			OverridesRequest:         overridesRequest,
			OverrideAuthorizedScopes: overrideAuthorizedScopes,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, "test", createHTTPResponse(http.StatusInternalServerError), errors.New("delete error")))
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:                 "success_override_delete_404_ignored",
			OverridesRequest:         overridesRequest,
			OverrideAuthorizedScopes: overrideAuthorizedScopes,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				// Delete for the delete request (ignored 404)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, "test", createHTTPResponse(http.StatusNotFound), errors.New("not found"))).
					Once()
				// Delete for the update request (ignored 404)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, overrideSecretID).
					Return(apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, overrideSecretID, createHTTPResponse(http.StatusNotFound), errors.New("not found"))).
					Once()
				m.EXPECT().
					CreateOverride(mock.Anything, validResourceID, mock.Anything, mock.Anything).
					Return(updatedOverride, nil)
				m.EXPECT().
					CreateOverride(mock.Anything, validResourceID, mock.Anything, overriddenSecretID).
					Return(createdOverride, nil)
			},
			ExpectError: false,
		},
		{
			TestName:                 "error_override_create_failure",
			OverridesRequest:         overridesRequest,
			OverrideAuthorizedScopes: overrideAuthorizedScopes,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, overrideSecretID).
					Return(nil)
				m.EXPECT().
					CreateOverride(mock.Anything, validResourceID, mock.Anything, mock.Anything).
					Return(nil, errors.New("create override error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:                 "success_with_overrides",
			OverridesRequest:         overridesRequest,
			OverrideAuthorizedScopes: overrideAuthorizedScopes,
			SetupMock: func(m *mocks_test.SecretRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingSecrets, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, overrideSecretID).
					Return(nil)
				m.EXPECT().
					CreateOverride(mock.Anything, validResourceID, mock.Anything, mock.Anything).
					Return(updatedOverride, nil)
				m.EXPECT().
					CreateOverride(mock.Anything, validResourceID, mock.Anything, overriddenSecretID).
					Return(createdOverride, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.SecretRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewSecretService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(
				context.Background(),
				validResourceID,
				emptySecretsRequest,
				emptyAliasesRequest,
				tc.OverridesRequest,
				tc.OverrideAuthorizedScopes,
			)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
