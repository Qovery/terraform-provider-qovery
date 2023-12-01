package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestNewDomainCredentialsFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrganizationID string
		Credentials    *qovery.ClusterCredentials
		ExpectedError  error
	}{
		{
			TestName:       "fail_with_nil_credentials",
			OrganizationID: gofakeit.UUID(),
			Credentials:    nil,
			ExpectedError:  credentials.ErrNilCredentials,
		},
		{
			TestName:       "fail_with_empty_organization_id",
			OrganizationID: "",
			Credentials: &qovery.ClusterCredentials{
				Id:   gofakeit.UUID(),
				Name: gofakeit.Name(),
			},
			ExpectedError: credentials.ErrInvalidCredentialsOrganizationID,
		},
		{
			TestName:       "success",
			OrganizationID: gofakeit.UUID(),
			Credentials: &qovery.ClusterCredentials{
				Id:   gofakeit.UUID(),
				Name: gofakeit.Name(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			creds, err := newDomainCredentialsFromQovery(tc.OrganizationID, tc.Credentials)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, creds)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, creds)
			assert.True(t, creds.IsValid())
			assert.Equal(t, tc.OrganizationID, creds.OrganizationID.String())
			assert.Equal(t, tc.Credentials.Id, creds.ID.String())
			assert.Equal(t, tc.Credentials.Name, creds.Name)
		})
	}
}

func TestNewQoveryAwsCredentialsRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  credentials.UpsertAwsRequest
	}{
		{
			TestName: "success",
			Request: credentials.UpsertAwsRequest{
				Name:            gofakeit.Name(),
				AccessKeyID:     gofakeit.Word(),
				SecretAccessKey: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryAwsCredentialsRequestFromDomain(tc.Request)

			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.AccessKeyID, req.AccessKeyId)
			assert.Equal(t, tc.Request.SecretAccessKey, req.SecretAccessKey)
		})
	}
}

func TestNewQoveryScalewayCredentialsRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Request  credentials.UpsertScalewayRequest
	}{
		{
			TestName: "success",
			Request: credentials.UpsertScalewayRequest{
				Name:                   gofakeit.Name(),
				ScalewayProjectID:      gofakeit.Word(),
				ScalewayAccessKey:      gofakeit.Word(),
				ScalewaySecretKey:      gofakeit.Word(),
				ScalewayOrganizationID: gofakeit.Word(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryScalewayCredentialsRequestFromDomain(tc.Request)

			assert.Equal(t, tc.Request.Name, req.Name)
			assert.Equal(t, tc.Request.ScalewayProjectID, req.ScalewayProjectId)
			assert.Equal(t, tc.Request.ScalewayAccessKey, req.ScalewayAccessKey)
			assert.Equal(t, tc.Request.ScalewaySecretKey, req.ScalewaySecretKey)
			assert.Equal(t, tc.Request.ScalewayOrganizationID, req.ScalewayOrganizationId)
		})
	}
}
