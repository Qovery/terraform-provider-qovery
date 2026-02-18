//go:build unit

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
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewCredentialsGcpService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Repository    credentials.GcpRepository
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_repository",
			Repository:    nil,
			ExpectedError: services.ErrInvalidRepository,
		},
		{
			TestName:   "success",
			Repository: &mocks_test.CredentialsGcpRepository{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			svc, err := services.NewCredentialsGcpService(tc.Repository)
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

func TestCredentialsGcpService_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        credentials.UpsertGcpRequest
		SetupMock      func(*mocks_test.CredentialsGcpRepository)
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_empty_organization_id",
			OrganizationID: "",
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: gofakeit.UUID(),
			},
			SetupMock:     func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError: credentials.ErrFailedToCreateGcpCredentials,
		},
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "invalid-uuid",
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: gofakeit.UUID(),
			},
			SetupMock:     func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError: credentials.ErrFailedToCreateGcpCredentials,
		},
		{
			TestName:       "fail_with_invalid_request",
			OrganizationID: uuid.NewString(),
			Request: credentials.UpsertGcpRequest{
				Name: "", // Invalid - empty name
			},
			SetupMock:     func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError: credentials.ErrFailedToCreateGcpCredentials,
		},
		{
			TestName:       "fail_with_repository_error",
			OrganizationID: uuid.NewString(),
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: gofakeit.UUID(),
			},
			SetupMock: func(m *mocks_test.CredentialsGcpRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("repository error"))
			},
			ExpectedError: credentials.ErrFailedToCreateGcpCredentials,
		},
		{
			TestName:       "success",
			OrganizationID: uuid.NewString(),
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: gofakeit.UUID(),
			},
			SetupMock: func(m *mocks_test.CredentialsGcpRepository) {
				creds := &credentials.Credentials{
					ID:             uuid.New(),
					OrganizationID: uuid.New(),
					Name:           gofakeit.Name(),
				}
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
					Return(creds, nil)
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsGcpRepository{}
			tc.SetupMock(mockRepo)

			svc, err := services.NewCredentialsGcpService(mockRepo)
			assert.NoError(t, err)

			result, err := svc.Create(context.Background(), tc.OrganizationID, tc.Request)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsGcpService_Get(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		SetupMock      func(*mocks_test.CredentialsGcpRepository)
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_empty_organization_id",
			OrganizationID: "",
			CredentialsID:  uuid.NewString(),
			SetupMock:      func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError:  credentials.ErrFailedToGetGcpCredentials,
		},
		{
			TestName:       "fail_with_empty_credentials_id",
			OrganizationID: uuid.NewString(),
			CredentialsID:  "",
			SetupMock:      func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError:  credentials.ErrFailedToGetGcpCredentials,
		},
		{
			TestName:       "fail_with_repository_error",
			OrganizationID: uuid.NewString(),
			CredentialsID:  uuid.NewString(),
			SetupMock: func(m *mocks_test.CredentialsGcpRepository) {
				m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("repository error"))
			},
			ExpectedError: credentials.ErrFailedToGetGcpCredentials,
		},
		{
			TestName:       "success",
			OrganizationID: uuid.NewString(),
			CredentialsID:  uuid.NewString(),
			SetupMock: func(m *mocks_test.CredentialsGcpRepository) {
				creds := &credentials.Credentials{
					ID:             uuid.New(),
					OrganizationID: uuid.New(),
					Name:           gofakeit.Name(),
				}
				m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Return(creds, nil)
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsGcpRepository{}
			tc.SetupMock(mockRepo)

			svc, err := services.NewCredentialsGcpService(mockRepo)
			assert.NoError(t, err)

			result, err := svc.Get(context.Background(), tc.OrganizationID, tc.CredentialsID)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsGcpService_Update(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		Request        credentials.UpsertGcpRequest
		SetupMock      func(*mocks_test.CredentialsGcpRepository)
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_empty_organization_id",
			OrganizationID: "",
			CredentialsID:  uuid.NewString(),
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: gofakeit.UUID(),
			},
			SetupMock:     func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError: credentials.ErrFailedToUpdateGcpCredentials,
		},
		{
			TestName:       "fail_with_empty_credentials_id",
			OrganizationID: uuid.NewString(),
			CredentialsID:  "",
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: gofakeit.UUID(),
			},
			SetupMock:     func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError: credentials.ErrFailedToUpdateGcpCredentials,
		},
		{
			TestName:       "fail_with_invalid_request",
			OrganizationID: uuid.NewString(),
			CredentialsID:  uuid.NewString(),
			Request: credentials.UpsertGcpRequest{
				Name: "", // Invalid
			},
			SetupMock:     func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError: credentials.ErrFailedToUpdateGcpCredentials,
		},
		{
			TestName:       "success",
			OrganizationID: uuid.NewString(),
			CredentialsID:  uuid.NewString(),
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: gofakeit.UUID(),
			},
			SetupMock: func(m *mocks_test.CredentialsGcpRepository) {
				creds := &credentials.Credentials{
					ID:             uuid.New(),
					OrganizationID: uuid.New(),
					Name:           gofakeit.Name(),
				}
				m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(creds, nil)
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsGcpRepository{}
			tc.SetupMock(mockRepo)

			svc, err := services.NewCredentialsGcpService(mockRepo)
			assert.NoError(t, err)

			result, err := svc.Update(context.Background(), tc.OrganizationID, tc.CredentialsID, tc.Request)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCredentialsGcpService_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		SetupMock      func(*mocks_test.CredentialsGcpRepository)
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_empty_organization_id",
			OrganizationID: "",
			CredentialsID:  uuid.NewString(),
			SetupMock:      func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError:  credentials.ErrFailedToDeleteGcpCredentials,
		},
		{
			TestName:       "fail_with_empty_credentials_id",
			OrganizationID: uuid.NewString(),
			CredentialsID:  "",
			SetupMock:      func(m *mocks_test.CredentialsGcpRepository) {},
			ExpectedError:  credentials.ErrFailedToDeleteGcpCredentials,
		},
		{
			TestName:       "fail_with_repository_error",
			OrganizationID: uuid.NewString(),
			CredentialsID:  uuid.NewString(),
			SetupMock: func(m *mocks_test.CredentialsGcpRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("repository error"))
			},
			ExpectedError: credentials.ErrFailedToDeleteGcpCredentials,
		},
		{
			TestName:       "success",
			OrganizationID: uuid.NewString(),
			CredentialsID:  uuid.NewString(),
			SetupMock: func(m *mocks_test.CredentialsGcpRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.CredentialsGcpRepository{}
			tc.SetupMock(mockRepo)

			svc, err := services.NewCredentialsGcpService(mockRepo)
			assert.NoError(t, err)

			err = svc.Delete(context.Background(), tc.OrganizationID, tc.CredentialsID)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				return
			}

			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
