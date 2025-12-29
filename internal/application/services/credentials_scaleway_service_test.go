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

func TestNewCredentialsScalewayService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  credentials.ScalewayRepository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.CredentialsScalewayRepository{},
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
			service, err := NewCredentialsScalewayService(tc.Repository)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				assert.ErrorIs(t, err, ErrInvalidRepository)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestCredentialsScalewayService_Create(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	validRequest := credentials.UpsertScalewayRequest{
		Name:                   gofakeit.Word(),
		ScalewayProjectID:      gofakeit.UUID(),
		ScalewayAccessKey:      gofakeit.Password(true, true, true, false, false, 32),
		ScalewaySecretKey:      gofakeit.Password(true, true, true, false, false, 32),
		ScalewayOrganizationID: gofakeit.UUID(),
	}

	invalidRequest := credentials.UpsertScalewayRequest{
		Name: "", // Invalid: empty name
	}

	expectedResult := &credentials.Credentials{
		ID:             uuid.New(),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           validRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        credentials.UpsertScalewayRequest
		SetupMock      func(*mocks_test.CredentialsScalewayRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			Request:        invalidRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to create scaleway credentials",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create scaleway credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
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

			mockRepo := &mocks_test.CredentialsScalewayRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsScalewayService(mockRepo)
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
				assert.Equal(t, expectedResult.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsScalewayService_Get(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validCredentialsID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidCredentialsID := "invalid-uuid"
	emptyOrgID := ""
	emptyCredentialsID := ""

	expectedResult := &credentials.Credentials{
		ID:             uuid.MustParse(validCredentialsID),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           gofakeit.Word(),
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		SetupMock      func(*mocks_test.CredentialsScalewayRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_empty_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  emptyCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_invalid_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  invalidCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validCredentialsID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get scaleway credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validCredentialsID).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsScalewayRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsScalewayService(mockRepo)
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
				assert.Equal(t, expectedResult.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsScalewayService_Update(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validCredentialsID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidCredentialsID := "invalid-uuid"
	emptyOrgID := ""
	emptyCredentialsID := ""

	validRequest := credentials.UpsertScalewayRequest{
		Name:                   gofakeit.Word(),
		ScalewayProjectID:      gofakeit.UUID(),
		ScalewayAccessKey:      gofakeit.Password(true, true, true, false, false, 32),
		ScalewaySecretKey:      gofakeit.Password(true, true, true, false, false, 32),
		ScalewayOrganizationID: gofakeit.UUID(),
	}

	invalidRequest := credentials.UpsertScalewayRequest{
		Name: "", // Invalid: empty name
	}

	expectedResult := &credentials.Credentials{
		ID:             uuid.MustParse(validCredentialsID),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           validRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		Request        credentials.UpsertScalewayRequest
		SetupMock      func(*mocks_test.CredentialsScalewayRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			CredentialsID:  validCredentialsID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			CredentialsID:  validCredentialsID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_empty_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  emptyCredentialsID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_invalid_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  invalidCredentialsID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			CredentialsID:  validCredentialsID,
			Request:        invalidRequest,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to update scaleway credentials",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			CredentialsID:  validCredentialsID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validCredentialsID, validRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update scaleway credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			CredentialsID:  validCredentialsID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validCredentialsID, validRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsScalewayRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsScalewayService(mockRepo)
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
				assert.Equal(t, expectedResult.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsScalewayService_Delete(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validCredentialsID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidCredentialsID := "invalid-uuid"
	emptyOrgID := ""
	emptyCredentialsID := ""

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		SetupMock      func(*mocks_test.CredentialsScalewayRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_empty_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  emptyCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_invalid_credentials_id",
			OrganizationID: validOrgID,
			CredentialsID:  invalidCredentialsID,
			SetupMock:      func(m *mocks_test.CredentialsScalewayRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid credentials id param",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validCredentialsID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete scaleway credentials",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			CredentialsID:  validCredentialsID,
			SetupMock: func(m *mocks_test.CredentialsScalewayRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validCredentialsID).
					Return(nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsScalewayRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewCredentialsScalewayService(mockRepo)
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
