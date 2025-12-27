//go:build unit && !integration
// +build unit,!integration

package services

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewCredentialsAwsService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  credentials.AwsRepository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.CredentialsAwsRepository{},
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
			service, err := NewCredentialsAwsService(tc.Repository)
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

func TestCredentialsAwsService_Create(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	validRequest := credentials.UpsertAwsRequest{
		Name: gofakeit.Word(),
		StaticCredentials: &credentials.AwsStaticCredentials{
			AccessKeyID:     gofakeit.UUID(),
			SecretAccessKey: gofakeit.UUID(),
		},
	}

	invalidRequest := credentials.UpsertAwsRequest{
		Name: "",
	}

	expectedResult := &credentials.Credentials{
		ID:             uuid.New(),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           validRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        credentials.UpsertAwsRequest
		SetupMock      func(*mocks_test.CredentialsAwsRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			Request:        invalidRequest,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to create aws credentials",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create aws credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsAwsRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsAwsService(mockRepo)
			require.NoError(t, err)

			result, err := service.Create(context.Background(), tc.OrganizationID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.ID, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsAwsService_Get(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validCredID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidCredID := "invalid-uuid"
	emptyOrgID := ""
	emptyCredID := ""

	expectedResult := &credentials.Credentials{
		ID:             uuid.MustParse(validCredID),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           gofakeit.Word(),
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		SetupMock      func(*mocks_test.CredentialsAwsRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			CredentialsID:  validCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			CredentialsID:  validCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_empty_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  emptyCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_invalid_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  invalidCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			CredentialsID:  validCredID,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validCredID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get aws credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			CredentialsID:  validCredID,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validCredID).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsAwsRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsAwsService(mockRepo)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.OrganizationID, tc.CredentialsID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.ID, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsAwsService_Update(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validCredID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidCredID := "invalid-uuid"

	validRequest := credentials.UpsertAwsRequest{
		Name: gofakeit.Word(),
		StaticCredentials: &credentials.AwsStaticCredentials{
			AccessKeyID:     gofakeit.UUID(),
			SecretAccessKey: gofakeit.UUID(),
		},
	}

	invalidRequest := credentials.UpsertAwsRequest{
		Name: "",
	}

	expectedResult := &credentials.Credentials{
		ID:             uuid.MustParse(validCredID),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           validRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		Request        credentials.UpsertAwsRequest
		SetupMock      func(*mocks_test.CredentialsAwsRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			CredentialsID:  validCredID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  invalidCredID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			CredentialsID:  validCredID,
			Request:        invalidRequest,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to update aws credentials",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			CredentialsID:  validCredID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validCredID, validRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update aws credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			CredentialsID:  validCredID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validCredID, validRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsAwsRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsAwsService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.OrganizationID, tc.CredentialsID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.ID, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsAwsService_Delete(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validCredID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidCredID := "invalid-uuid"
	emptyOrgID := ""
	emptyCredID := ""

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		SetupMock      func(*mocks_test.CredentialsAwsRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			CredentialsID:  validCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			CredentialsID:  validCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_empty_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  emptyCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_invalid_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  invalidCredID,
			SetupMock:      func(m *mocks_test.CredentialsAwsRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			CredentialsID:  validCredID,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validCredID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete aws credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			CredentialsID:  validCredID,
			SetupMock: func(m *mocks_test.CredentialsAwsRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validCredID).
					Return(nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsAwsRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsAwsService(mockRepo)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.OrganizationID, tc.CredentialsID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
