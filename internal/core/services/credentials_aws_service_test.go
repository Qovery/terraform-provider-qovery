//go:build unit
// +build unit

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

func TestNewCredentialsAwsService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Repository    credentials.AwsRepository
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_repository",
			Repository:    nil,
			ExpectedError: services.ErrInvalidRepository,
		},
		{
			TestName:   "success_with_repository",
			Repository: mocks_test.NewCredentialsAwsRepository(t),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			svc, err := services.NewCredentialsAwsService(tc.Repository)
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

func TestCredentialsAwsService_Create(t *testing.T) {
	t.Parallel()

	// Setup request
	organizationID := uuid.New()
	request := credentials.UpsertAwsRequest{
		Name:            gofakeit.Name(),
		AccessKeyID:     gofakeit.Word(),
		SecretAccessKey: gofakeit.Word(),
	}

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsAwsService(inmem.NewCredentialsAwsInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	testCases := []struct {
		TestName       string
		OrganizationID string
		Request        credentials.UpsertAwsRequest
		Expected       *credentials.Credentials
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_invalid_request_name",
			OrganizationID: organizationID.String(),
			Request: credentials.UpsertAwsRequest{
				AccessKeyID:     gofakeit.Word(),
				SecretAccessKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
		{
			TestName:       "fail_with_invalid_access_key_id",
			OrganizationID: organizationID.String(),
			Request: credentials.UpsertAwsRequest{
				Name:            gofakeit.Name(),
				SecretAccessKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
		{
			TestName:       "fail_with_invalid_secret_access_key",
			OrganizationID: organizationID.String(),
			Request: credentials.UpsertAwsRequest{
				Name:        gofakeit.Name(),
				AccessKeyID: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
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
				assert.ErrorContains(t, err, credentials.ErrFailedToCreateAwsCredentials.Error())
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

func TestCredentialsAwsService_Get(t *testing.T) {
	t.Parallel()

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsAwsService(inmem.NewCredentialsAwsInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	newCreds, err := awsCredentialsService.Create(context.Background(), uuid.NewString(), credentials.UpsertAwsRequest{
		Name:            gofakeit.Name(),
		SecretAccessKey: gofakeit.Word(),
		AccessKeyID:     gofakeit.Word(),
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

func TestCredentialsAwsService_Update(t *testing.T) {
	t.Parallel()

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsAwsService(inmem.NewCredentialsAwsInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	newCreds, err := awsCredentialsService.Create(context.Background(), uuid.NewString(), credentials.UpsertAwsRequest{
		Name:            gofakeit.Name(),
		SecretAccessKey: gofakeit.Word(),
		AccessKeyID:     gofakeit.Word(),
	})
	require.NoError(t, err)
	require.NotNil(t, newCreds)

	updateRequest := credentials.UpsertAwsRequest{
		Name:            gofakeit.Name(),
		SecretAccessKey: gofakeit.Word(),
		AccessKeyID:     gofakeit.Word(),
	}

	testCases := []struct {
		TestName       string
		OrganizationID string
		CredentialsID  string
		Request        credentials.UpsertAwsRequest
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
			ExpectedError:  credentials.ErrInvalidUpsertAwsRequest,
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
				assert.ErrorContains(t, err, credentials.ErrFailedToUpdateAwsCredentials.Error())
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

func TestCredentialsAwsService_Delete(t *testing.T) {
	t.Parallel()

	// Initialize service
	awsCredentialsService, err := services.NewCredentialsAwsService(inmem.NewCredentialsAwsInmem())
	require.NoError(t, err)
	require.NotNil(t, awsCredentialsService)

	newCreds, err := awsCredentialsService.Create(context.Background(), uuid.NewString(), credentials.UpsertAwsRequest{
		Name:            gofakeit.Name(),
		SecretAccessKey: gofakeit.Word(),
		AccessKeyID:     gofakeit.Word(),
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
				assert.ErrorContains(t, err, credentials.ErrFailedToDeleteAwsCredentials.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
