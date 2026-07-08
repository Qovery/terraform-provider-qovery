//go:build unit && !integration

package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newApiTokenCreateRequest() apitoken.CreateRequest {
	return apitoken.CreateRequest{
		Name:   gofakeit.Name(),
		RoleID: gofakeit.UUID(),
	}
}

func newApiToken() *apitoken.ApiToken {
	token := gofakeit.UUID()
	return &apitoken.ApiToken{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           gofakeit.Name(),
		RoleID:         gofakeit.UUID(),
		Token:          &token,
	}
}

func TestNewApiTokenService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Repository    apitoken.Repository
		ExpectedError error
	}{
		{TestName: "fail_with_nil_repository", Repository: nil, ExpectedError: services.ErrInvalidRepository},
		{TestName: "success", Repository: &mocks_test.ApiTokenRepository{}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			svc, err := services.NewApiTokenService(tc.Repository)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, svc)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, svc)
		})
	}
}

func TestApiTokenService_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        apitoken.CreateRequest
		SetupMock      func(*mocks_test.ApiTokenRepository)
		ErrorContains  string
	}{
		{
			TestName:       "fail_with_empty_organization_id",
			OrganizationID: "",
			Request:        newApiTokenCreateRequest(),
			SetupMock:      func(m *mocks_test.ApiTokenRepository) {},
			ErrorContains:  apitoken.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "not-a-uuid",
			Request:        newApiTokenCreateRequest(),
			SetupMock:      func(m *mocks_test.ApiTokenRepository) {},
			ErrorContains:  apitoken.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:       "fail_with_invalid_request",
			OrganizationID: gofakeit.UUID(),
			Request:        apitoken.CreateRequest{},
			SetupMock:      func(m *mocks_test.ApiTokenRepository) {},
			ErrorContains:  apitoken.ErrInvalidCreateRequest.Error(),
		},
		{
			TestName:       "fail_with_repository_error",
			OrganizationID: gofakeit.UUID(),
			Request:        newApiTokenCreateRequest(),
			SetupMock: func(m *mocks_test.ApiTokenRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("api error"))
			},
			ErrorContains: apitoken.ErrFailedToCreateApiToken.Error(),
		},
		{
			TestName:       "success",
			OrganizationID: gofakeit.UUID(),
			Request:        newApiTokenCreateRequest(),
			SetupMock: func(m *mocks_test.ApiTokenRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(newApiToken(), nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ApiTokenRepository{}
			tc.SetupMock(repo)

			svc, err := services.NewApiTokenService(repo)
			assert.NoError(t, err)

			res, err := svc.Create(context.Background(), tc.OrganizationID, tc.Request)
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
			repo.AssertExpectations(t)
		})
	}
}

func TestApiTokenService_Get(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		ApiTokenID     string
		SetupMock      func(*mocks_test.ApiTokenRepository)
		ErrorContains  string
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "not-a-uuid",
			ApiTokenID:     gofakeit.UUID(),
			SetupMock:      func(m *mocks_test.ApiTokenRepository) {},
			ErrorContains:  apitoken.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:       "fail_with_invalid_api_token_id",
			OrganizationID: gofakeit.UUID(),
			ApiTokenID:     "not-a-uuid",
			SetupMock:      func(m *mocks_test.ApiTokenRepository) {},
			ErrorContains:  apitoken.ErrInvalidApiTokenIdParam.Error(),
		},
		{
			TestName:       "fail_with_repository_error",
			OrganizationID: gofakeit.UUID(),
			ApiTokenID:     gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ApiTokenRepository) {
				m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("api error"))
			},
			ErrorContains: apitoken.ErrFailedToGetApiToken.Error(),
		},
		{
			TestName:       "success",
			OrganizationID: gofakeit.UUID(),
			ApiTokenID:     gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ApiTokenRepository) {
				m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(newApiToken(), nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ApiTokenRepository{}
			tc.SetupMock(repo)

			svc, err := services.NewApiTokenService(repo)
			assert.NoError(t, err)

			res, err := svc.Get(context.Background(), tc.OrganizationID, tc.ApiTokenID)
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
			repo.AssertExpectations(t)
		})
	}
}

func TestApiTokenService_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		ApiTokenID     string
		SetupMock      func(*mocks_test.ApiTokenRepository)
		ErrorContains  string
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "not-a-uuid",
			ApiTokenID:     gofakeit.UUID(),
			SetupMock:      func(m *mocks_test.ApiTokenRepository) {},
			ErrorContains:  apitoken.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:       "fail_with_invalid_api_token_id",
			OrganizationID: gofakeit.UUID(),
			ApiTokenID:     "not-a-uuid",
			SetupMock:      func(m *mocks_test.ApiTokenRepository) {},
			ErrorContains:  apitoken.ErrInvalidApiTokenIdParam.Error(),
		},
		{
			TestName:       "fail_with_repository_error",
			OrganizationID: gofakeit.UUID(),
			ApiTokenID:     gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ApiTokenRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("api error"))
			},
			ErrorContains: apitoken.ErrFailedToDeleteApiToken.Error(),
		},
		{
			TestName:       "success",
			OrganizationID: gofakeit.UUID(),
			ApiTokenID:     gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ApiTokenRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ApiTokenRepository{}
			tc.SetupMock(repo)

			svc, err := services.NewApiTokenService(repo)
			assert.NoError(t, err)

			err = svc.Delete(context.Background(), tc.OrganizationID, tc.ApiTokenID)
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				return
			}
			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}
