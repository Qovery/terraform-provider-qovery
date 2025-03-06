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
				AwsStaticClusterCredentials: nil,
				AwsRoleClusterCredentials:   nil,
				ScalewayClusterCredentials:  nil,
				GenericClusterCredentials: &qovery.GenericClusterCredentials{
					Id:                   gofakeit.UUID(),
					Name:                 gofakeit.Name(),
					ObjectType:           "OTHER",
					AdditionalProperties: map[string]interface{}{},
				},
			},
			ExpectedError: credentials.ErrInvalidCredentialsOrganizationID,
		},
		{
			TestName:       "success",
			OrganizationID: gofakeit.UUID(),
			Credentials: &qovery.ClusterCredentials{
				AwsStaticClusterCredentials: nil,
				AwsRoleClusterCredentials:   nil,
				ScalewayClusterCredentials:  nil,
				GenericClusterCredentials: &qovery.GenericClusterCredentials{
					Id:                   gofakeit.UUID(),
					Name:                 gofakeit.Name(),
					ObjectType:           "OTHER",
					AdditionalProperties: map[string]interface{}{},
				},
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
			assert.Equal(t, tc.Credentials.GenericClusterCredentials.Id, creds.ID.String())
			assert.Equal(t, tc.Credentials.GenericClusterCredentials.Name, creds.Name)
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
				Name: gofakeit.Name(),
				StaticCredentials: &credentials.AwsStaticCredentials{
					AccessKeyID:     gofakeit.Word(),
					SecretAccessKey: gofakeit.Word(),
				},
			},
		},
		{
			TestName: "success",
			Request: credentials.UpsertAwsRequest{
				Name: gofakeit.Name(),
				RoleCredentials: &credentials.AwsRoleCredentials{
					RoleArn: gofakeit.Word(),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			req := newQoveryAwsCredentialsRequestFromDomain(tc.Request)

			if tc.Request.StaticCredentials != nil {
				assert.Equal(t, tc.Request.Name, req.AwsStaticCredentialsRequest.Name)
				assert.Equal(t, tc.Request.StaticCredentials.AccessKeyID, req.AwsStaticCredentialsRequest.AccessKeyId)
				assert.Equal(t, tc.Request.StaticCredentials.SecretAccessKey, req.AwsStaticCredentialsRequest.SecretAccessKey)
			}
			if tc.Request.RoleCredentials != nil {
				assert.Equal(t, tc.Request.Name, req.AwsRoleCredentialsRequest.Name)
				assert.Equal(t, tc.Request.RoleCredentials.RoleArn, req.AwsRoleCredentialsRequest.RoleArn)
			}
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
