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

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewHelmRepositoryService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  helmRepository.Repository
		ExpectError bool
		ExpectedErr error
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.HelmRepositoryRepository{},
			ExpectError: false,
		},
		{
			TestName:    "error_with_nil_repository",
			Repository:  nil,
			ExpectError: true,
			ExpectedErr: ErrInvalidRepository,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewHelmRepositoryService(tc.Repository)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tc.ExpectedErr != nil {
					assert.ErrorIs(t, err, tc.ExpectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestHelmRepositoryService_Create(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	validRequest := helmRepository.UpsertRequest{
		Name: gofakeit.Word(),
		Kind: "HTTPS",
		URL:  "https://charts.example.com",
	}

	invalidRequest := helmRepository.UpsertRequest{
		Name: "", // Empty name should fail validation
		Kind: "HTTPS",
		URL:  "https://charts.example.com",
	}

	expectedResult := &helmRepository.HelmRepository{
		ID:             uuid.New(),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           validRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        helmRepository.UpsertRequest
		SetupMock      func(*mocks_test.HelmRepositoryRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			Request:        invalidRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to create helm repository",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
				m.EXPECT().
					Create(mock.Anything, validOrgID, validRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm repository",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
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

			mockRepo := &mocks_test.HelmRepositoryRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewHelmRepositoryService(mockRepo)
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

func TestHelmRepositoryService_Get(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validRepositoryID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidRepositoryID := "invalid-uuid"
	emptyOrgID := ""
	emptyRepositoryID := ""

	expectedResult := &helmRepository.HelmRepository{
		ID:             uuid.MustParse(validRepositoryID),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           gofakeit.Word(),
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		RepositoryID   string
		SetupMock      func(*mocks_test.HelmRepositoryRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_empty_repository_id",
			OrganizationID: validOrgID,
			RepositoryID:   emptyRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid repository Id",
		},
		{
			TestName:       "error_invalid_repository_id",
			OrganizationID: validOrgID,
			RepositoryID:   invalidRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid repository Id",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validRepositoryID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get helm repository",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID, validRepositoryID).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.HelmRepositoryRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewHelmRepositoryService(mockRepo)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.OrganizationID, tc.RepositoryID)

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

func TestHelmRepositoryService_Update(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validRepositoryID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidRepositoryID := "invalid-uuid"
	emptyOrgID := ""
	emptyRepositoryID := ""

	validRequest := helmRepository.UpsertRequest{
		Name: gofakeit.Word(),
		Kind: "HTTPS",
		URL:  "https://charts.example.com",
	}

	invalidRequest := helmRepository.UpsertRequest{
		Name: "", // Empty name should fail validation
		Kind: "HTTPS",
		URL:  "https://charts.example.com",
	}

	expectedResult := &helmRepository.HelmRepository{
		ID:             uuid.MustParse(validRepositoryID),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           validRequest.Name,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		RepositoryID   string
		Request        helmRepository.UpsertRequest
		SetupMock      func(*mocks_test.HelmRepositoryRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			RepositoryID:   validRepositoryID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			RepositoryID:   validRepositoryID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_empty_repository_id",
			OrganizationID: validOrgID,
			RepositoryID:   emptyRepositoryID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid repository Id",
		},
		{
			TestName:       "error_invalid_repository_id",
			OrganizationID: validOrgID,
			RepositoryID:   invalidRepositoryID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid repository Id",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			RepositoryID:   validRepositoryID,
			Request:        invalidRequest,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "failed to update helm repository",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			RepositoryID:   validRepositoryID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validRepositoryID, validRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm repository",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			RepositoryID:   validRepositoryID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validRepositoryID, validRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.HelmRepositoryRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewHelmRepositoryService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.OrganizationID, tc.RepositoryID, tc.Request)

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

func TestHelmRepositoryService_Delete(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	validRepositoryID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	invalidRepositoryID := "invalid-uuid"
	emptyOrgID := ""
	emptyRepositoryID := ""

	testCases := []struct {
		TestName       string
		OrganizationID string
		RepositoryID   string
		SetupMock      func(*mocks_test.HelmRepositoryRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization Id",
		},
		{
			TestName:       "error_empty_repository_id",
			OrganizationID: validOrgID,
			RepositoryID:   emptyRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid repository Id",
		},
		{
			TestName:       "error_invalid_repository_id",
			OrganizationID: validOrgID,
			RepositoryID:   invalidRepositoryID,
			SetupMock:      func(m *mocks_test.HelmRepositoryRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid repository Id",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validRepositoryID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete helm repository",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			RepositoryID:   validRepositoryID,
			SetupMock: func(m *mocks_test.HelmRepositoryRepository) {
				m.EXPECT().
					Delete(mock.Anything, validOrgID, validRepositoryID).
					Return(nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.HelmRepositoryRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewHelmRepositoryService(mockRepo)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.OrganizationID, tc.RepositoryID)

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
