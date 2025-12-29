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

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewOrganizationService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  organization.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.OrganizationRepository{},
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
			service, err := NewOrganizationService(tc.Repository)
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

func TestOrganizationService_Get(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	expectedResult := &organization.Organization{
		ID:   uuid.MustParse(validOrgID),
		Name: gofakeit.Word(),
		Plan: organization.PlanFree,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		SetupMock      func(*mocks_test.OrganizationRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			SetupMock:      func(m *mocks_test.OrganizationRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			SetupMock:      func(m *mocks_test.OrganizationRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			SetupMock: func(m *mocks_test.OrganizationRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get organization",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			SetupMock: func(m *mocks_test.OrganizationRepository) {
				m.EXPECT().
					Get(mock.Anything, validOrgID).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.OrganizationRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewOrganizationService(mockRepo)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.OrganizationID)

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

func TestOrganizationService_Update(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	validRequest := organization.UpdateRequest{
		Name:        gofakeit.Word(),
		Description: nil,
	}

	invalidRequest := organization.UpdateRequest{
		Name:        "",
		Description: nil,
	}

	expectedResult := &organization.Organization{
		ID:          uuid.MustParse(validOrgID),
		Name:        validRequest.Name,
		Plan:        organization.PlanFree,
		Description: validRequest.Description,
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        organization.UpdateRequest
		SetupMock      func(*mocks_test.OrganizationRepository)
		ExpectError    bool
		ErrorContains  string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.OrganizationRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			Request:        validRequest,
			SetupMock:      func(m *mocks_test.OrganizationRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			Request:        invalidRequest,
			SetupMock:      func(m *mocks_test.OrganizationRepository) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization update request",
		},
		{
			TestName:       "error_repository_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.OrganizationRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update organization",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMock: func(m *mocks_test.OrganizationRepository) {
				m.EXPECT().
					Update(mock.Anything, validOrgID, validRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.OrganizationRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewOrganizationService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.OrganizationID, tc.Request)

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
