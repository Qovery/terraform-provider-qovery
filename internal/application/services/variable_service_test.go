//go:build unit && !integration
// +build unit,!integration

package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewVariableService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  variable.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.VariableRepository{},
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
			service, err := NewVariableService(tc.Repository)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestVariableService_List(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""

	expectedVariables := variable.Variables{
		{
			ID:          uuid.MustParse(gofakeit.UUID()),
			Scope:       variable.ScopeEnvironment,
			Key:         "TEST_VAR_1",
			Value:       "value1",
			Type:        "VALUE",
			Description: "Test variable 1",
		},
		{
			ID:          uuid.MustParse(gofakeit.UUID()),
			Scope:       variable.ScopeEnvironment,
			Key:         "TEST_VAR_2",
			Value:       "value2",
			Type:        "VALUE",
			Description: "Test variable 2",
		},
	}

	testCases := []struct {
		TestName      string
		ResourceID    string
		SetupMock     func(*mocks_test.VariableRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_resource_id",
			ResourceID:    emptyResourceID,
			SetupMock:     func(m *mocks_test.VariableRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid resource id param",
		},
		{
			TestName:      "error_invalid_resource_id",
			ResourceID:    invalidResourceID,
			SetupMock:     func(m *mocks_test.VariableRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid resource id param",
		},
		{
			TestName:   "error_repository_failure",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to list variables",
		},
		{
			TestName:   "success",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(expectedVariables, nil)
			},
			ExpectError: false,
		},
		{
			TestName:   "success_empty_list",
			ResourceID: validResourceID,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(variable.Variables{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.VariableRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewVariableService(mockRepo)
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

func TestVariableService_Update(t *testing.T) {
	t.Parallel()

	validResourceID := gofakeit.UUID()
	invalidResourceID := "invalid-uuid"
	emptyResourceID := ""

	validVariableID := gofakeit.UUID()
	aliasedVariableID := gofakeit.UUID()
	overriddenVariableID := gofakeit.UUID()

	validEnvironmentVariablesRequest := variable.DiffRequest{
		Create: []variable.DiffCreateRequest{
			{
				UpsertRequest: variable.UpsertRequest{
					Key:         "NEW_VAR",
					Value:       "new_value",
					Description: "New variable",
				},
			},
		},
		Update: []variable.DiffUpdateRequest{
			{
				UpsertRequest: variable.UpsertRequest{
					Key:         "EXISTING_VAR",
					Value:       "updated_value",
					Description: "Updated variable",
				},
				VariableID: validVariableID,
			},
		},
		Delete: []variable.DiffDeleteRequest{
			{
				VariableID: gofakeit.UUID(),
			},
		},
	}

	invalidEnvironmentVariablesRequest := variable.DiffRequest{
		Create: []variable.DiffCreateRequest{
			{
				UpsertRequest: variable.UpsertRequest{
					Key:   "", // Invalid: empty key
					Value: "value",
				},
			},
		},
	}

	emptyAliasRequest := variable.DiffRequest{}
	emptyOverrideRequest := variable.DiffRequest{}
	emptyAuthorizedScopes := make(map[variable.Scope]struct{})

	existingVariables := variable.Variables{
		{
			ID:    uuid.MustParse(aliasedVariableID),
			Scope: variable.ScopeProject,
			Key:   "PROJECT_VAR",
			Value: "project_value",
			Type:  "VALUE",
		},
		{
			ID:    uuid.MustParse(overriddenVariableID),
			Scope: variable.ScopeProject,
			Key:   "OVERRIDE_VAR",
			Value: "original_value",
			Type:  "VALUE",
		},
	}

	testCases := []struct {
		TestName                            string
		ResourceID                          string
		EnvironmentVariablesRequest         variable.DiffRequest
		EnvironmentVariableAliasesRequest   variable.DiffRequest
		EnvironmentVariableOverridesRequest variable.DiffRequest
		OverrideAuthorizedScopes            map[variable.Scope]struct{}
		SetupMock                           func(*mocks_test.VariableRepository)
		ExpectError                         bool
		ErrorContains                       string
	}{
		{
			TestName:                            "error_empty_resource_id",
			ResourceID:                          emptyResourceID,
			EnvironmentVariablesRequest:         validEnvironmentVariablesRequest,
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock:                           func(m *mocks_test.VariableRepository) {},
			ExpectError:                         true,
			ErrorContains:                       "invalid resource id param",
		},
		{
			TestName:                            "error_invalid_resource_id",
			ResourceID:                          invalidResourceID,
			EnvironmentVariablesRequest:         validEnvironmentVariablesRequest,
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock:                           func(m *mocks_test.VariableRepository) {},
			ExpectError:                         true,
			ErrorContains:                       "invalid resource id param",
		},
		{
			TestName:                            "error_invalid_environment_variables_request",
			ResourceID:                          validResourceID,
			EnvironmentVariablesRequest:         invalidEnvironmentVariablesRequest,
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock:                           func(m *mocks_test.VariableRepository) {},
			ExpectError:                         true,
			ErrorContains:                       "failed to update variables",
		},
		{
			TestName:   "error_repository_delete_failure",
			ResourceID: validResourceID,
			EnvironmentVariablesRequest: variable.DiffRequest{
				Delete: []variable.DiffDeleteRequest{
					{
						VariableID: validVariableID,
					},
				},
			},
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					Delete(mock.Anything, validResourceID, validVariableID).
					Return(apierrors.NewDeleteAPIError("variable", validVariableID, &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("")),
					}, errors.New("delete error")))
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:   "error_repository_update_failure",
			ResourceID: validResourceID,
			EnvironmentVariablesRequest: variable.DiffRequest{
				Update: []variable.DiffUpdateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "VAR",
							Value: "value",
						},
						VariableID: validVariableID,
					},
				},
			},
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					Update(mock.Anything, validResourceID, validVariableID, mock.Anything).
					Return(nil, errors.New("update error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:   "error_repository_create_failure",
			ResourceID: validResourceID,
			EnvironmentVariablesRequest: variable.DiffRequest{
				Create: []variable.DiffCreateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "VAR",
							Value: "value",
						},
					},
				},
			},
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					Create(mock.Anything, validResourceID, mock.Anything).
					Return(nil, errors.New("create error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update variables",
		},
		{
			TestName:   "error_repository_list_failure",
			ResourceID: validResourceID,
			EnvironmentVariablesRequest: variable.DiffRequest{
				Create: []variable.DiffCreateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "VAR",
							Value: "value",
						},
					},
				},
			},
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				createdVar := &variable.Variable{
					ID:    uuid.MustParse(gofakeit.UUID()),
					Scope: variable.ScopeEnvironment,
					Key:   "VAR",
					Value: "value",
					Type:  "VALUE",
				}
				m.EXPECT().
					Create(mock.Anything, validResourceID, mock.Anything).
					Return(createdVar, nil)
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(nil, errors.New("list error"))
			},
			ExpectError: true,
		},
		{
			TestName:   "success_with_environment_variables_only",
			ResourceID: validResourceID,
			EnvironmentVariablesRequest: variable.DiffRequest{
				Create: []variable.DiffCreateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "NEW_VAR",
							Value: "new_value",
						},
					},
				},
				Update: []variable.DiffUpdateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "UPDATED_VAR",
							Value: "updated_value",
						},
						VariableID: validVariableID,
					},
				},
				Delete: []variable.DiffDeleteRequest{
					{
						VariableID: gofakeit.UUID(),
					},
				},
			},
			EnvironmentVariableAliasesRequest:   emptyAliasRequest,
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					Delete(mock.Anything, validResourceID, mock.Anything).
					Return(nil)
				m.EXPECT().
					Update(mock.Anything, validResourceID, validVariableID, mock.Anything).
					Return(&variable.Variable{
						ID:    uuid.MustParse(validVariableID),
						Scope: variable.ScopeEnvironment,
						Key:   "UPDATED_VAR",
						Value: "updated_value",
						Type:  "VALUE",
					}, nil)
				m.EXPECT().
					Create(mock.Anything, validResourceID, mock.Anything).
					Return(&variable.Variable{
						ID:    uuid.MustParse(gofakeit.UUID()),
						Scope: variable.ScopeEnvironment,
						Key:   "NEW_VAR",
						Value: "new_value",
						Type:  "VALUE",
					}, nil)
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingVariables, nil)
			},
			ExpectError: false,
		},
		{
			TestName:                    "success_with_aliases",
			ResourceID:                  validResourceID,
			EnvironmentVariablesRequest: emptyAliasRequest,
			EnvironmentVariableAliasesRequest: variable.DiffRequest{
				Create: []variable.DiffCreateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "ALIAS_VAR",
							Value: "PROJECT_VAR",
						},
					},
				},
			},
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingVariables, nil)
				m.EXPECT().
					CreateAlias(mock.Anything, validResourceID, mock.Anything, aliasedVariableID).
					Return(&variable.Variable{
						ID:    uuid.MustParse(gofakeit.UUID()),
						Scope: variable.ScopeEnvironment,
						Key:   "ALIAS_VAR",
						Value: "PROJECT_VAR",
						Type:  "ALIAS",
					}, nil)
			},
			ExpectError: false,
		},
		{
			TestName:                          "success_with_overrides",
			ResourceID:                        validResourceID,
			EnvironmentVariablesRequest:       emptyAliasRequest,
			EnvironmentVariableAliasesRequest: emptyAliasRequest,
			EnvironmentVariableOverridesRequest: variable.DiffRequest{
				Create: []variable.DiffCreateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "OVERRIDE_VAR",
							Value: "overridden_value",
						},
					},
				},
			},
			OverrideAuthorizedScopes: map[variable.Scope]struct{}{
				variable.ScopeProject: {},
			},
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingVariables, nil)
				m.EXPECT().
					CreateOverride(mock.Anything, validResourceID, mock.Anything, overriddenVariableID).
					Return(&variable.Variable{
						ID:    uuid.MustParse(gofakeit.UUID()),
						Scope: variable.ScopeEnvironment,
						Key:   "OVERRIDE_VAR",
						Value: "overridden_value",
						Type:  "OVERRIDE",
					}, nil)
			},
			ExpectError: false,
		},
		{
			TestName:                    "success_alias_update_with_delete_and_recreate",
			ResourceID:                  validResourceID,
			EnvironmentVariablesRequest: emptyAliasRequest,
			EnvironmentVariableAliasesRequest: variable.DiffRequest{
				Update: []variable.DiffUpdateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "ALIAS_VAR",
							Value: "PROJECT_VAR",
						},
						VariableID: validVariableID,
					},
				},
			},
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingVariables, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, validVariableID).
					Return(nil)
				m.EXPECT().
					CreateAlias(mock.Anything, validResourceID, mock.Anything, aliasedVariableID).
					Return(&variable.Variable{
						ID:    uuid.MustParse(gofakeit.UUID()),
						Scope: variable.ScopeEnvironment,
						Key:   "ALIAS_VAR",
						Value: "PROJECT_VAR",
						Type:  "ALIAS",
					}, nil)
			},
			ExpectError: false,
		},
		{
			TestName:                    "success_alias_delete_with_404_ignored",
			ResourceID:                  validResourceID,
			EnvironmentVariablesRequest: emptyAliasRequest,
			EnvironmentVariableAliasesRequest: variable.DiffRequest{
				Delete: []variable.DiffDeleteRequest{
					{
						VariableID: validVariableID,
					},
				},
			},
			EnvironmentVariableOverridesRequest: emptyOverrideRequest,
			OverrideAuthorizedScopes:            emptyAuthorizedScopes,
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingVariables, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, validVariableID).
					Return(apierrors.NewDeleteAPIError("variable", validVariableID, &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(bytes.NewBufferString("")),
					}, errors.New("not found")))
			},
			ExpectError: false,
		},
		{
			TestName:                          "success_override_update_with_delete_and_recreate",
			ResourceID:                        validResourceID,
			EnvironmentVariablesRequest:       emptyAliasRequest,
			EnvironmentVariableAliasesRequest: emptyAliasRequest,
			EnvironmentVariableOverridesRequest: variable.DiffRequest{
				Update: []variable.DiffUpdateRequest{
					{
						UpsertRequest: variable.UpsertRequest{
							Key:   "OVERRIDE_VAR",
							Value: "new_overridden_value",
						},
						VariableID: validVariableID,
					},
				},
			},
			OverrideAuthorizedScopes: map[variable.Scope]struct{}{
				variable.ScopeProject: {},
			},
			SetupMock: func(m *mocks_test.VariableRepository) {
				m.EXPECT().
					List(mock.Anything, validResourceID).
					Return(existingVariables, nil)
				m.EXPECT().
					Delete(mock.Anything, validResourceID, validVariableID).
					Return(nil)
				m.EXPECT().
					CreateOverride(mock.Anything, validResourceID, mock.Anything, overriddenVariableID).
					Return(&variable.Variable{
						ID:    uuid.MustParse(gofakeit.UUID()),
						Scope: variable.ScopeEnvironment,
						Key:   "OVERRIDE_VAR",
						Value: "new_overridden_value",
						Type:  "OVERRIDE",
					}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.VariableRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewVariableService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(
				context.Background(),
				tc.ResourceID,
				tc.EnvironmentVariablesRequest,
				tc.EnvironmentVariableAliasesRequest,
				tc.EnvironmentVariableOverridesRequest,
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
