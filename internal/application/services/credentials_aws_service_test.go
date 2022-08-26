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

type CredentialsAwsServiceTestSuite struct {
	suite.Suite

	repository *mocks_test.CredentialsAwsRepository
	service    credentials.AwsService
}

func (ts *CredentialsAwsServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize repository
	credsRepository := mocks_test.NewCredentialsAwsRepository(t)

	// Initialize service
	credsService, err := services.NewCredentialsAwsService(credsRepository)
	require.NoError(t, err)
	require.NotNil(t, credsService)

	ts.repository = credsRepository
	ts.service = credsService
}

func (ts *CredentialsAwsServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	credsService, err := services.NewCredentialsAwsService(nil)
	assert.Nil(t, credsService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *CredentialsAwsServiceTestSuite) TestNew_Success() {
	t := ts.T()

	credsService, err := services.NewCredentialsAwsService(mocks_test.NewCredentialsAwsRepository(t))
	assert.Nil(t, err)
	assert.NotNil(t, credsService)
}

func (ts *CredentialsAwsServiceTestSuite) TestCreate_FailWithInvalidOrganizationID() {
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
			creds, err := ts.service.Create(context.Background(), tc.OrganizationID, assertNewCredentialsUpsertAwsRequest(t))
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToCreateAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestCreate_FailWithInvalidCreateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CreateRequest credentials.UpsertAwsRequest
	}{
		{
			TestName: "invalid_name",
			CreateRequest: credentials.UpsertAwsRequest{
				AccessKeyID:     gofakeit.Word(),
				SecretAccessKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_access_key_id",
			CreateRequest: credentials.UpsertAwsRequest{
				Name:            gofakeit.Name(),
				SecretAccessKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_secret_access_key",
			CreateRequest: credentials.UpsertAwsRequest{
				Name:        gofakeit.Name(),
				AccessKeyID: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Create(context.Background(), gofakeit.UUID(), tc.CreateRequest)
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToCreateAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidUpsertAwsRequest.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestCreate_FailUnexpectedRepositoryError() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)
	createRequest := assertNewCredentialsUpsertAwsRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedCreds.OrganizationID.String(), createRequest).
		Return(nil, errors.New(""))

	creds, err := ts.service.Create(context.Background(), expectedCreds.OrganizationID.String(), createRequest)
	assert.Nil(t, creds)
	assert.ErrorContains(t, err, credentials.ErrFailedToCreateAwsCredentials.Error())
}

func (ts *CredentialsAwsServiceTestSuite) TestCreate_Success() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)
	createRequest := assertNewCredentialsUpsertAwsRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedCreds.OrganizationID.String(), createRequest).
		Return(expectedCreds, nil)

	creds, err := ts.service.Create(context.Background(), expectedCreds.OrganizationID.String(), createRequest)
	assert.Nil(t, err)
	assertEqualCredentials(t, expectedCreds, creds)
}

func (ts *CredentialsAwsServiceTestSuite) TestGet_FailWithInvalidOrganizationID() {
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
			assert.ErrorContains(t, err, credentials.ErrFailedToGetAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestGet_FailWithInvalidCredentialsID() {
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
			assert.ErrorContains(t, err, credentials.ErrFailedToGetAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidCredentialsIDParam.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestGet_FailCredentialsNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Get(mock.Anything, fakeID, fakeID).
		Return(nil, errors.New(""))

	creds, err := ts.service.Get(context.Background(), fakeID, fakeID)
	assert.Nil(t, creds)
	assert.ErrorContains(t, err, credentials.ErrFailedToGetAwsCredentials.Error())
}

func (ts *CredentialsAwsServiceTestSuite) TestGet_Success() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedCreds.ID.String(), expectedCreds.ID.String()).
		Return(expectedCreds, nil)

	creds, err := ts.service.Get(context.Background(), expectedCreds.ID.String(), expectedCreds.ID.String())
	assert.Nil(t, err)
	assertEqualCredentials(t, expectedCreds, creds)
}

func (ts *CredentialsAwsServiceTestSuite) TestUpdate_FailWithInvalidOrganizationID() {
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
			creds, err := ts.service.Update(context.Background(), tc.OrganizationID, gofakeit.UUID(), assertNewCredentialsUpsertAwsRequest(t))
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToUpdateAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestUpdate_FailWithInvalidCredentialsID() {
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
			creds, err := ts.service.Update(context.Background(), gofakeit.UUID(), tc.CredentialsID, assertNewCredentialsUpsertAwsRequest(t))
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToUpdateAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidCredentialsIDParam.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestUpdate_FailCredentialsNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()
	updateRequest := assertNewCredentialsUpsertAwsRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, fakeID, fakeID, updateRequest).
		Return(nil, errors.New(""))

	creds, err := ts.service.Update(context.Background(), fakeID, fakeID, updateRequest)
	assert.Nil(t, creds)
	assert.ErrorContains(t, err, credentials.ErrFailedToUpdateAwsCredentials.Error())
}

func (ts *CredentialsAwsServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		UpdateRequest credentials.UpsertAwsRequest
	}{
		{
			TestName: "invalid_name",
			UpdateRequest: credentials.UpsertAwsRequest{
				AccessKeyID:     gofakeit.Word(),
				SecretAccessKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_access_key_id",
			UpdateRequest: credentials.UpsertAwsRequest{
				Name:            gofakeit.Name(),
				SecretAccessKey: gofakeit.Word(),
			},
		},
		{
			TestName: "invalid_secret_access_key",
			UpdateRequest: credentials.UpsertAwsRequest{
				Name:        gofakeit.Name(),
				AccessKeyID: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := ts.service.Update(context.Background(), gofakeit.UUID(), gofakeit.UUID(), tc.UpdateRequest)
			assert.Nil(t, creds)
			assert.ErrorContains(t, err, credentials.ErrFailedToUpdateAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidUpsertAwsRequest.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	newCreds := assertCreateCredentials(t)
	updateRequest := assertNewCredentialsUpsertAwsRequest(t)
	updatedCreds := assertApplyCredentialsUpdateAwsRequest(t, newCreds, updateRequest)

	ts.repository.EXPECT().
		Update(mock.Anything, updatedCreds.OrganizationID.String(), updatedCreds.ID.String(), updateRequest).
		Return(updatedCreds, nil)

	creds, err := ts.service.Update(context.Background(), updatedCreds.OrganizationID.String(), updatedCreds.ID.String(), updateRequest)
	assert.Nil(t, err)
	assertEqualCredentials(t, updatedCreds, creds)
}

func (ts *CredentialsAwsServiceTestSuite) TestDelete_FailWithInvalidOrganizationID() {
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
			assert.ErrorContains(t, err, credentials.ErrFailedToDeleteAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestDelete_FailWithInvalidCredentialsID() {
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
			assert.ErrorContains(t, err, credentials.ErrFailedToDeleteAwsCredentials.Error())
			assert.ErrorContains(t, err, credentials.ErrInvalidCredentialsIDParam.Error())
		})
	}
}

func (ts *CredentialsAwsServiceTestSuite) TestDelete_FailCredentialsNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID, fakeID).
		Return(errors.New(""))

	err := ts.service.Delete(context.Background(), fakeID, fakeID)
	assert.ErrorContains(t, err, credentials.ErrFailedToDeleteAwsCredentials.Error())
}

func (ts *CredentialsAwsServiceTestSuite) TestDelete_Success() {
	t := ts.T()

	expectedCreds := assertCreateCredentials(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, expectedCreds.ID.String(), expectedCreds.ID.String()).
		Return(nil)

	err := ts.service.Delete(context.Background(), expectedCreds.ID.String(), expectedCreds.ID.String())
	assert.Nil(t, err)
}

func TestCredentialsAwsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialsAwsServiceTestSuite))
}

func assertNewCredentialsUpsertAwsRequest(t *testing.T) credentials.UpsertAwsRequest {
	req := credentials.UpsertAwsRequest{
		Name:            gofakeit.Name(),
		AccessKeyID:     gofakeit.UUID(),
		SecretAccessKey: gofakeit.Word(),
	}
	require.NoError(t, req.Validate())

	return req
}

//
//func assertCreateCredentials(t *testing.T) *credentials.Credentials {
//	creds, err := credentials.NewCredentials(credentials.NewCredentialsParams{
//		CredentialsID:  gofakeit.UUID(),
//		OrganizationID: gofakeit.UUID(),
//		Name:           gofakeit.Name(),
//	})
//	require.NoError(t, err)
//	require.NotNil(t, creds)
//	require.NoError(t, creds.Validate())
//
//	return creds
//}
//
//func assertEqualCredentials(t *testing.T, expected *credentials.Credentials, actual *credentials.Credentials) {
//	assert.Equal(t, expected.ID, actual.ID)
//	assert.Equal(t, expected.OrganizationID, actual.OrganizationID)
//	assert.Equal(t, expected.Name, actual.Name)
//}

func assertApplyCredentialsUpdateAwsRequest(t *testing.T, creds *credentials.Credentials, update credentials.UpsertAwsRequest) *credentials.Credentials {
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
