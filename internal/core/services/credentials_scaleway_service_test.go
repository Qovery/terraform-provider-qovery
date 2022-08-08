package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/core/repositories/inmem"
	"github.com/qovery/terraform-provider-qovery/internal/core/repositories/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/core/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestNewCredentialsScalewayService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Repository    credentials.ScalewayRepository
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_repository",
			Repository:    nil,
			ExpectedError: services.ErrInvalidRepository,
		},
		{
			TestName:   "success_with_repository",
			Repository: mocks_test.NewCredentialsScalewayRepository(t),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			svc, err := services.NewCredentialsScalewayService(tc.Repository)
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

func TestCredentialsScalewayService_Create(t *testing.T) {
	t.Parallel()

	// Setup request
	organizationID := uuid.New()
	request := credentials.UpsertScalewayRequest{
		Name:              gofakeit.Name(),
		ScalewayProjectID: gofakeit.Word(),
		ScalewayAccessKey: gofakeit.Word(),
		ScalewaySecretKey: gofakeit.Word(),
	}

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsScalewayService(inmem.NewCredentialsScalewayInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        credentials.UpsertScalewayRequest
		Expected       *credentials.Credentials
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_invalid_request_name",
			OrganizationID: organizationID.String(),
			Request: credentials.UpsertScalewayRequest{
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName:       "fail_with_invalid_scaleway_project_id",
			OrganizationID: organizationID.String(),
			Request: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName:       "fail_with_invalid_scaleway_access_key",
			OrganizationID: organizationID.String(),
			Request: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName:       "fail_with_invalid_scaleway_secret_key",
			OrganizationID: organizationID.String(),
			Request: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName:       "success",
			OrganizationID: organizationID.String(),
			Request:        request,
			Expected: &credentials.Credentials{
				OrganizationID: organizationID,
				Name:           request.Name,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := awsCredentialsService.Create(context.Background(), tc.OrganizationID, tc.Request)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.ErrorContains(t, err, credentials.ErrFailedToCreateScalewayCredentials.Error())
				assert.Nil(t, creds)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, creds)
			assert.NotEmpty(t, creds.ID)
			assert.Equal(t, tc.Expected.OrganizationID, creds.OrganizationID)
			assert.Equal(t, tc.Expected.Name, creds.Name)
		})
	}
}

func TestCredentialsScalewayService_Get(t *testing.T) {
	t.Parallel()

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsScalewayService(inmem.NewCredentialsScalewayInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	newCreds, err := awsCredentialsService.Create(context.Background(), uuid.NewString(), credentials.UpsertScalewayRequest{
		Name:              gofakeit.Name(),
		ScalewayProjectID: gofakeit.Word(),
		ScalewayAccessKey: gofakeit.Word(),
		ScalewaySecretKey: gofakeit.Word(),
	})
	require.NoError(t, err)
	require.NotNil(t, newCreds)

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		Expected       *credentials.Credentials
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "",
			CredentialsID:  newCreds.ID.String(),
			ExpectedError:  credentials.ErrInvalidOrganizationIDParam,
		},
		{
			TestName:       "fail_with_invalid_credentials_id",
			OrganizationID: newCreds.OrganizationID.String(),
			CredentialsID:  "",
			ExpectedError:  credentials.ErrInvalidCredentialsIDParam,
		},
		{
			TestName:       "success",
			OrganizationID: newCreds.OrganizationID.String(),
			CredentialsID:  newCreds.ID.String(),
			Expected:       newCreds,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := awsCredentialsService.Get(context.Background(), tc.OrganizationID, tc.CredentialsID)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, creds)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, creds)
			assert.Equal(t, tc.Expected.ID, creds.ID)
			assert.Equal(t, tc.Expected.OrganizationID, creds.OrganizationID)
			assert.Equal(t, tc.Expected.Name, creds.Name)
		})
	}
}

func TestCredentialsScalewayService_Update(t *testing.T) {
	t.Parallel()

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsScalewayService(inmem.NewCredentialsScalewayInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	newCreds, err := awsCredentialsService.Create(context.Background(), uuid.NewString(), credentials.UpsertScalewayRequest{
		Name:              gofakeit.Name(),
		ScalewayProjectID: gofakeit.Word(),
		ScalewayAccessKey: gofakeit.Word(),
		ScalewaySecretKey: gofakeit.Word(),
	})
	require.NoError(t, err)
	require.NotNil(t, newCreds)

	updateRequest := credentials.UpsertScalewayRequest{
		Name:              gofakeit.Name(),
		ScalewayProjectID: gofakeit.Word(),
		ScalewayAccessKey: gofakeit.Word(),
		ScalewaySecretKey: gofakeit.Word(),
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		Request        credentials.UpsertScalewayRequest
		Expected       *credentials.Credentials
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "",
			CredentialsID:  newCreds.ID.String(),
			ExpectedError:  credentials.ErrInvalidOrganizationIDParam,
		},
		{
			TestName:       "fail_with_invalid_credentials_id",
			OrganizationID: newCreds.OrganizationID.String(),
			CredentialsID:  "",
			ExpectedError:  credentials.ErrInvalidCredentialsIDParam,
		},
		{
			TestName:       "fail_with_invalid_upsert_aws_request",
			OrganizationID: newCreds.OrganizationID.String(),
			CredentialsID:  newCreds.ID.String(),
			ExpectedError:  credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName:       "success",
			OrganizationID: newCreds.OrganizationID.String(),
			CredentialsID:  newCreds.ID.String(),
			Request:        updateRequest,
			Expected: &credentials.Credentials{
				ID:             newCreds.ID,
				OrganizationID: newCreds.OrganizationID,
				Name:           updateRequest.Name,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := awsCredentialsService.Update(context.Background(), tc.OrganizationID, tc.CredentialsID, tc.Request)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.ErrorContains(t, err, credentials.ErrFailedToUpdateScalewayCredentials.Error())
				assert.Nil(t, creds)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, creds)
			assert.Equal(t, tc.Expected.ID, creds.ID)
			assert.Equal(t, tc.Expected.OrganizationID, creds.OrganizationID)
			assert.Equal(t, tc.Expected.Name, creds.Name)
		})
	}
}

func TestCredentialsScalewayService_Delete(t *testing.T) {
	t.Parallel()

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsScalewayService(inmem.NewCredentialsScalewayInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	newCreds, err := awsCredentialsService.Create(context.Background(), uuid.NewString(), credentials.UpsertScalewayRequest{
		Name:              gofakeit.Name(),
		ScalewayProjectID: gofakeit.Word(),
		ScalewayAccessKey: gofakeit.Word(),
		ScalewaySecretKey: gofakeit.Word(),
	})
	require.NoError(t, err)
	require.NotNil(t, newCreds)

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrganizationID: "",
			CredentialsID:  newCreds.ID.String(),
			ExpectedError:  credentials.ErrInvalidOrganizationIDParam,
		},
		{
			TestName:       "fail_with_invalid_credentials_id",
			OrganizationID: newCreds.OrganizationID.String(),
			CredentialsID:  "",
			ExpectedError:  credentials.ErrInvalidCredentialsIDParam,
		},
		{
			TestName:       "success",
			OrganizationID: newCreds.OrganizationID.String(),
			CredentialsID:  newCreds.ID.String(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			err := awsCredentialsService.Delete(context.Background(), tc.OrganizationID, tc.CredentialsID)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.ErrorContains(t, err, credentials.ErrFailedToDeleteScalewayCredentials.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
