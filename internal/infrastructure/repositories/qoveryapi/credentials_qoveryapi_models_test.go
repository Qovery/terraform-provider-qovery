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
					AdditionalProperties: map[string]any{},
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
					AdditionalProperties: map[string]any{},
				},
			},
		},
	}

	for _, tc := range testCases {
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

func TestNewQoveryGcpCredentialsRequestFromDomain(t *testing.T) {
	t.Parallel()

	t.Run("service_account_key", func(t *testing.T) {
		t.Parallel()
		json := `{"type":"service_account"}`
		req := newQoveryGcpCredentialsRequestFromDomain(credentials.UpsertGcpRequest{
			Name:              "key-creds",
			ServiceAccountKey: &credentials.GcpServiceAccountKeyCredentials{GcpCredentials: json},
		})
		assert.NotNil(t, req.GcpServiceAccountKeyCredentialsRequest)
		assert.Nil(t, req.GcpWorkloadIdentityFederationCredentialsRequest)
		assert.Equal(t, "key-creds", req.GcpServiceAccountKeyCredentialsRequest.Name)
		assert.Equal(t, json, req.GcpServiceAccountKeyCredentialsRequest.GcpCredentials)
	})

	t.Run("workload_identity_federation", func(t *testing.T) {
		t.Parallel()
		req := newQoveryGcpCredentialsRequestFromDomain(credentials.UpsertGcpRequest{
			Name: "wif-creds",
			WorkloadIdentity: &credentials.GcpWorkloadIdentityCredentials{
				ServiceAccountEmail:              "qovery@proj.iam.gserviceaccount.com",
				WorkloadIdentityProviderResource: "projects/123/locations/global/workloadIdentityPools/p/providers/pr",
			},
		})
		assert.NotNil(t, req.GcpWorkloadIdentityFederationCredentialsRequest)
		assert.Nil(t, req.GcpServiceAccountKeyCredentialsRequest)
		assert.Equal(t, "wif-creds", req.GcpWorkloadIdentityFederationCredentialsRequest.Name)
		assert.Equal(t, "qovery@proj.iam.gserviceaccount.com", req.GcpWorkloadIdentityFederationCredentialsRequest.ServiceAccountEmail)
		assert.Equal(t, "projects/123/locations/global/workloadIdentityPools/p/providers/pr", req.GcpWorkloadIdentityFederationCredentialsRequest.WorkloadIdentityProviderResource)
	})
}

func TestNewDomainCredentialsFromQovery_GcpWif(t *testing.T) {
	t.Parallel()

	orgID := gofakeit.UUID()
	credID := gofakeit.UUID()
	name := gofakeit.Name()

	creds, err := newDomainCredentialsFromQovery(orgID, &qovery.ClusterCredentials{
		GcpWorkloadIdentityFederationClusterCredentials: &qovery.GcpWorkloadIdentityFederationClusterCredentials{
			Id:         credID,
			Name:       name,
			ObjectType: "GcpWorkloadIdentityFederationClusterCredentials",
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, creds)
	assert.Equal(t, credID, creds.ID.String())
	assert.Equal(t, name, creds.Name)
	assert.Equal(t, orgID, creds.OrganizationID.String())
}

func TestNewDomainCredentialsFromQovery_Identifiers(t *testing.T) {
	t.Parallel()

	credID := gofakeit.UUID()
	testCases := []struct {
		TestName    string
		Credentials *qovery.ClusterCredentials
		Expected    credentials.Identifiers
	}{
		{
			TestName:    "aws_static",
			Credentials: &qovery.ClusterCredentials{AwsStaticClusterCredentials: &qovery.AwsStaticClusterCredentials{Id: credID, Name: "creds", AccessKeyId: "AKIA"}},
			Expected:    credentials.Identifiers{AccessKeyID: new("AKIA")},
		},
		{
			TestName:    "aws_role",
			Credentials: &qovery.ClusterCredentials{AwsRoleClusterCredentials: &qovery.AwsRoleClusterCredentials{Id: credID, Name: "creds", RoleArn: "arn:aws:iam::123456789012:role/qovery"}},
			Expected:    credentials.Identifiers{RoleArn: new("arn:aws:iam::123456789012:role/qovery")},
		},
		{
			TestName: "scaleway",
			Credentials: &qovery.ClusterCredentials{ScalewayClusterCredentials: &qovery.ScalewayClusterCredentials{
				Id: credID, Name: "creds", ScalewayAccessKey: "SCW", ScalewayProjectId: "project", ScalewayOrganizationId: "organization",
			}},
			Expected: credentials.Identifiers{ScalewayAccessKey: new("SCW"), ScalewayProjectID: new("project"), ScalewayOrganizationID: new("organization")},
		},
		{
			TestName:    "gcp_service_account_key",
			Credentials: &qovery.ClusterCredentials{GcpStaticClusterCredentials: &qovery.GcpStaticClusterCredentials{Id: credID, Name: "creds"}},
			Expected:    credentials.Identifiers{},
		},
		{
			TestName: "gcp_workload_identity_federation",
			Credentials: &qovery.ClusterCredentials{GcpWorkloadIdentityFederationClusterCredentials: &qovery.GcpWorkloadIdentityFederationClusterCredentials{
				Id: credID, Name: "creds", ProjectId: "project", ServiceAccountEmail: "qovery@project.iam.gserviceaccount.com", WorkloadIdentityProviderResource: "projects/1/locations/global/workloadIdentityPools/p/providers/p",
			}},
			Expected: credentials.Identifiers{ServiceAccountEmail: new("qovery@project.iam.gserviceaccount.com"), WorkloadIdentityProviderResource: new("projects/1/locations/global/workloadIdentityPools/p/providers/p")},
		},
		{
			TestName: "eks_anywhere_vsphere_static",
			Credentials: &qovery.ClusterCredentials{EksAnywhereVsphereClusterCredentials: &qovery.EksAnywhereVsphereClusterCredentials{
				Id: credID, Name: "creds", VsphereUser: "user", AccessKeyId: *qovery.NewNullableString(new("AKIA")),
			}},
			Expected: credentials.Identifiers{VsphereUser: new("user"), AccessKeyID: new("AKIA")},
		},
		{
			TestName: "eks_anywhere_vsphere_role",
			Credentials: &qovery.ClusterCredentials{EksAnywhereVsphereClusterCredentials: &qovery.EksAnywhereVsphereClusterCredentials{
				Id: credID, Name: "creds", VsphereUser: "user", RoleArn: *qovery.NewNullableString(new("arn:aws:iam::123456789012:role/qovery")),
			}},
			Expected: credentials.Identifiers{VsphereUser: new("user"), RoleArn: new("arn:aws:iam::123456789012:role/qovery")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			creds, err := newDomainCredentialsFromQovery(gofakeit.UUID(), tc.Credentials)

			assert.NoError(t, err)
			assert.Equal(t, credID, creds.ID.String())
			assert.Equal(t, "creds", creds.Name)
			assert.Equal(t, tc.Expected, creds.Identifiers)
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
