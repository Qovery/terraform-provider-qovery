//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

type CredentialsScalewayServiceTestSuite struct {
	suite.Suite

	repository *mocks_test.CredentialsScalewayRepository
	service    credentials.ScalewayService
}

func (ts *CredentialsScalewayServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize repository
	credsRepository := mocks_test.NewCredentialsScalewayRepository(t)

	// Initialize service
	credsService, err := services.NewCredentialsScalewayService(credsRepository)
	require.NoError(t, err)
	require.NotNil(t, credsService)

	ts.repository = credsRepository
	ts.service = credsService
}

func (ts *CredentialsScalewayServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	credsService, err := services.NewCredentialsScalewayService(nil)
	assert.Nil(t, credsService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *CredentialsScalewayServiceTestSuite) TestNew_Success() {
	t := ts.T()

	credsService, err := services.NewCredentialsScalewayService(mocks_test.NewCredentialsScalewayRepository(t))
	assert.Nil(t, err)
	assert.NotNil(t, credsService)
}

func (ts *CredentialsScalewayServiceTestSuite) TestCreate_FailWithInvalidOrganizationID() {
	t := ts.T()

	testCases := []struct {
		TestName       string
		OrganizationID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:       "invalid_uuid",
			OrganizationID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Create(context.Background(), tc.OrganizationID, assertNewCredentialsUpsertScalewayRequest(t))
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToCreateScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestCreate_FailWithInvalidCreateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CreateRequest credentials.UpsertScalewayRequest
	}{
		{
			TestName: "invalid_name",
			CreateRequest: credentials.UpsertScalewayRequest{
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_scaleway_project_id",
			CreateRequest: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_scaleway_access_key",
			CreateRequest: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_scaleway_secret_key",
			CreateRequest: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Create(context.Background(), gofakeit.UUID(), tc.CreateRequest)
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToCreateScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidUpsertScalewayRequest.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestCreate_FailUnexpectedRepositoryError() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)
	createRequest := assertNewCredentialsUpsertScalewayRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedCreds.OrganizationID.String(), createRequest).
		Return(nil, errors.New(""))

	creds, err := ts.service.Create(context.Background(), expectedCreds.OrganizationID.String(), createRequest)
	assert.Nil(t, creds)
	assert.ErrorContains(t, err, credentials.ErrFailedToCreateScalewayCredentials.Error())
}

func (ts *CredentialsScalewayServiceTestSuite) TestCreate_Success() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)
	createRequest := assertNewCredentialsUpsertScalewayRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedCreds.OrganizationID.String(), createRequest).
		Return(expectedCreds, nil)

	creds, err := ts.service.Create(context.Background(), expectedCreds.OrganizationID.String(), createRequest)
	assert.Nil(t, err)
	assertEqualCredentials(t, expectedCreds, creds)
}

func (ts *CredentialsScalewayServiceTestSuite) TestGet_FailWithInvalidOrganizationID() {
	t := ts.T()

	testCases := []struct {
		TestName       string
		OrganizationID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:       "invalid_uuid",
			OrganizationID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Get(context.Background(), tc.OrganizationID, gofakeit.UUID())
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToGetScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestGet_FailWithInvalidCredentialsID() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CredentialsID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:      "invalid_uuid",
			CredentialsID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Get(context.Background(), gofakeit.UUID(), tc.CredentialsID)
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToGetScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidCredentialsIDParam.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestGet_FailCredentialsNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Get(mock.Anything, fakeID, fakeID).
		Return(nil, errors.New(""))

	creds, err := ts.service.Get(context.Background(), fakeID, fakeID)
	assert.Nil(t, creds)
	assert.ErrorContains(t, err, credentials.ErrFailedToGetScalewayCredentials.Error())
}

func (ts *CredentialsScalewayServiceTestSuite) TestGet_Success() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedCreds.ID.String(), expectedCreds.ID.String()).
		Return(expectedCreds, nil)

	creds, err := ts.service.Get(context.Background(), expectedCreds.ID.String(), expectedCreds.ID.String())
	assert.Nil(t, err)
	assertEqualCredentials(t, expectedCreds, creds)
}

func (ts *CredentialsScalewayServiceTestSuite) TestUpdate_FailWithInvalidOrganizationID() {
	t := ts.T()

	testCases := []struct {
		TestName       string
		OrganizationID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:       "invalid_uuid",
			OrganizationID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Update(context.Background(), tc.OrganizationID, gofakeit.UUID(), assertNewCredentialsUpsertScalewayRequest(t))
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestUpdate_FailWithInvalidCredentialsID() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CredentialsID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:      "invalid_uuid",
			CredentialsID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Update(context.Background(), gofakeit.UUID(), tc.CredentialsID, assertNewCredentialsUpsertScalewayRequest(t))
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidCredentialsIDParam.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestUpdate_FailCredentialsNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()
	updateRequest := assertNewCredentialsUpsertScalewayRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, fakeID, fakeID, updateRequest).
		Return(nil, errors.New(""))

	creds, err := ts.service.Update(context.Background(), fakeID, fakeID, updateRequest)
	assert.Nil(t, creds)
	assert.ErrorContains(t, err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
}

func (ts *CredentialsScalewayServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		UpdateRequest credentials.UpsertScalewayRequest
	}{
		{
			TestName: "invalid_name",
			UpdateRequest: credentials.UpsertScalewayRequest{
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_scaleway_project_id",
			UpdateRequest: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_scaleway_access_key",
			UpdateRequest: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_scaleway_secret_key",
			UpdateRequest: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Update(context.Background(), gofakeit.UUID(), gofakeit.UUID(), tc.UpdateRequest)
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidUpsertScalewayRequest.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	newCreds := assertCreateCredentials(t)
	updateRequest := assertNewCredentialsUpsertScalewayRequest(t)
	updatedCreds := assertApplyCredentialsUpdateScalewayRequest(t, newCreds, updateRequest)

	ts.repository.EXPECT().
		Update(mock.Anything, updatedCreds.OrganizationID.String(), updatedCreds.ID.String(), updateRequest).
		Return(updatedCreds, nil)

	creds, err := ts.service.Update(context.Background(), updatedCreds.OrganizationID.String(), updatedCreds.ID.String(), updateRequest)
	assert.Nil(t, err)
	assertEqualCredentials(t, updatedCreds, creds)
}

func (ts *CredentialsScalewayServiceTestSuite) TestDelete_FailWithInvalidOrganizationID() {
	t := ts.T()

	testCases := []struct {
		TestName       string
		OrganizationID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:       "invalid_uuid",
			OrganizationID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			err := ts.service.Delete(context.Background(), tc.OrganizationID, gofakeit.UUID())
			assert.ErrorContains(t, err, credentials.ErrFailedToDeleteScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestDelete_FailWithInvalidCredentialsID() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CredentialsID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:      "invalid_uuid",
			CredentialsID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			err := ts.service.Delete(context.Background(), gofakeit.UUID(), tc.CredentialsID)
			assert.ErrorContains(t, err, credentials.ErrFailedToDeleteScalewayCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidCredentialsIDParam.Error())
		})
	}
}

func (ts *CredentialsScalewayServiceTestSuite) TestDelete_FailCredentialsNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID, fakeID).
		Return(errors.New(""))

	err := ts.service.Delete(context.Background(), fakeID, fakeID)
	assert.ErrorContains(t, err, credentials.ErrFailedToDeleteScalewayCredentials.Error())
}

func (ts *CredentialsScalewayServiceTestSuite) TestDelete_Success() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, expectedCreds.ID.String(), expectedCreds.ID.String()).
		Return(nil)

	err := ts.service.Delete(context.Background(), expectedCreds.ID.String(), expectedCreds.ID.String())
	assert.Nil(t, err)
}

func TestCredentialsScalewayServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialsScalewayServiceTestSuite))
}

func assertNewCredentialsUpsertScalewayRequest(t *testing.T) credentials.UpsertScalewayRequest {
	req := credentials.UpsertScalewayRequest{
		Name:              gofakeit.Name(),
		ScalewayProjectID: gofakeit.UUID(),
		ScalewayAccessKey: gofakeit.Word(),
		ScalewaySecretKey: gofakeit.Word(),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertApplyCredentialsUpdateScalewayRequest(t *testing.T, creds *credentials.Credentials, update credentials.UpsertScalewayRequest) *credentials.Credentials {
	updatedCreds, err := credentials.NewCredentials(credentials.NewCredentialsParams{
		CredentialsID:  creds.ID.String(),
		OrganizationID: creds.OrganizationID.String(),
		Name:           update.Name,
	})
	require.NoError(t, err)
	require.NotNil(t, updatedCreds)
	require.NoError(t, updatedCreds.Validate())

	return updatedCreds
}
