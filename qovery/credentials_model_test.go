//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestCredentialIdentifierFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Prior    types.String
		API      *string
		Expect   types.String
	}{
		{TestName: "api_value_wins", Prior: types.StringValue("old"), API: new("new"), Expect: types.StringValue("new")},
		{TestName: "api_value_fills_a_null_prior", Prior: types.StringNull(), API: new("new"), Expect: types.StringValue("new")},
		{TestName: "nil_api_value_clears_the_prior", Prior: types.StringValue("old"), API: nil, Expect: types.StringNull()},
		{TestName: "nil_api_value_keeps_a_null_prior", Prior: types.StringNull(), API: nil, Expect: types.StringNull()},
		{TestName: "prior_that_q_core_trims_to_the_api_value_is_kept", Prior: types.StringValue(" AKIA\n"), API: new("AKIA"), Expect: types.StringValue(" AKIA\n")},
		{TestName: "unknown_prior_takes_the_api_value", Prior: types.StringUnknown(), API: new("new"), Expect: types.StringValue("new")},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expect, credentialIdentifierFromAPI(tc.Prior, tc.API))
		})
	}
}

func newTestCredentials(identifiers credentials.Identifiers) *credentials.Credentials {
	return &credentials.Credentials{ID: uuid.New(), OrganizationID: uuid.New(), Name: "creds", Identifiers: identifiers}
}

func TestConvertDomainCredentialsToAWSCredentials(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName          string
		Identifiers       credentials.Identifiers
		Prior             AWSCredentials
		ExpectAccessKeyID types.String
		ExpectRoleArn     types.String
	}{
		{
			TestName:          "static_credentials_read_the_access_key_id",
			Identifiers:       credentials.Identifiers{AccessKeyID: new("AKIA_API")},
			Prior:             AWSCredentials{AccessKeyId: types.StringValue("AKIA_STATE"), SecretAccessKey: types.StringValue("secret"), RoleArn: types.StringNull()},
			ExpectAccessKeyID: types.StringValue("AKIA_API"),
			ExpectRoleArn:     types.StringNull(),
		},
		{
			TestName:          "role_credentials_read_the_role_arn",
			Identifiers:       credentials.Identifiers{RoleArn: new("arn:aws:iam::123456789012:role/api")},
			Prior:             AWSCredentials{AccessKeyId: types.StringNull(), SecretAccessKey: types.StringNull(), RoleArn: types.StringValue("arn:aws:iam::123456789012:role/state")},
			ExpectAccessKeyID: types.StringNull(),
			ExpectRoleArn:     types.StringValue("arn:aws:iam::123456789012:role/api"),
		},
		{
			TestName:          "a_switch_to_a_role_from_the_console_shows_up",
			Identifiers:       credentials.Identifiers{RoleArn: new("arn:aws:iam::123456789012:role/console")},
			Prior:             AWSCredentials{AccessKeyId: types.StringValue("AKIA_STATE"), SecretAccessKey: types.StringValue("secret"), RoleArn: types.StringNull()},
			ExpectAccessKeyID: types.StringNull(),
			ExpectRoleArn:     types.StringValue("arn:aws:iam::123456789012:role/console"),
		},
		{
			TestName:          "import_records_the_identifier",
			Identifiers:       credentials.Identifiers{AccessKeyID: new("AKIA_API")},
			Prior:             AWSCredentials{},
			ExpectAccessKeyID: types.StringValue("AKIA_API"),
			ExpectRoleArn:     types.StringNull(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			creds := newTestCredentials(tc.Identifiers)
			got := convertDomainCredentialsToAWSCredentials(creds, tc.Prior)

			assert.Equal(t, creds.ID.String(), got.Id.ValueString())
			assert.Equal(t, creds.OrganizationID.String(), got.OrganizationId.ValueString())
			assert.Equal(t, "creds", got.Name.ValueString())
			assert.Equal(t, tc.ExpectAccessKeyID, got.AccessKeyId)
			assert.Equal(t, tc.ExpectRoleArn, got.RoleArn)
			assert.Equal(t, tc.Prior.SecretAccessKey, got.SecretAccessKey, "the API never returns the secret: it comes from prior")
		})
	}
}

func TestConvertDomainCredentialsToScalewayCredentials(t *testing.T) {
	t.Parallel()

	creds := newTestCredentials(credentials.Identifiers{
		ScalewayAccessKey:      new("SCW_API"),
		ScalewayProjectID:      new("project-api"),
		ScalewayOrganizationID: new("organization-api"),
	})
	prior := ScalewayCredentials{
		ScalewayAccessKey:      types.StringValue("SCW_STATE"),
		ScalewaySecretKey:      types.StringValue("secret"),
		ScalewayProjectId:      types.StringValue("project-state"),
		ScalewayOrganizationId: types.StringValue("organization-state"),
	}

	got := convertDomainCredentialsToScalewayCredentials(creds, prior)

	assert.Equal(t, types.StringValue("SCW_API"), got.ScalewayAccessKey)
	assert.Equal(t, types.StringValue("project-api"), got.ScalewayProjectId)
	assert.Equal(t, types.StringValue("organization-api"), got.ScalewayOrganizationId)
	assert.Equal(t, prior.ScalewaySecretKey, got.ScalewaySecretKey, "the API never returns the secret: it comes from prior")
}

func TestConvertDomainCredentialsToGCPCredentials(t *testing.T) {
	t.Parallel()

	workloadIdentity := credentials.Identifiers{
		ServiceAccountEmail:              new("api@project.iam.gserviceaccount.com"),
		WorkloadIdentityProviderResource: new("projects/1/locations/global/workloadIdentityPools/api/providers/api"),
	}

	testCases := []struct {
		TestName                string
		Identifiers             credentials.Identifiers
		Prior                   GCPCredentials
		ExpectEmail             types.String
		ExpectProviderResource  types.String
		ExpectGcpCredentialsKey types.String
	}{
		{
			TestName:    "workload_identity_federation_reads_its_fields",
			Identifiers: workloadIdentity,
			Prior: GCPCredentials{
				GcpCredentials:                   types.StringNull(),
				ServiceAccountEmail:              types.StringValue("state@project.iam.gserviceaccount.com"),
				WorkloadIdentityProviderResource: types.StringValue("projects/1/locations/global/workloadIdentityPools/state/providers/state"),
			},
			ExpectEmail:             types.StringValue("api@project.iam.gserviceaccount.com"),
			ExpectProviderResource:  types.StringValue("projects/1/locations/global/workloadIdentityPools/api/providers/api"),
			ExpectGcpCredentialsKey: types.StringNull(),
		},
		{
			TestName:    "service_account_key_keeps_the_key_and_reports_no_federation_field",
			Identifiers: credentials.Identifiers{},
			Prior: GCPCredentials{
				GcpCredentials:                   types.StringValue(`{"type":"service_account"}`),
				ServiceAccountEmail:              types.StringNull(),
				WorkloadIdentityProviderResource: types.StringNull(),
			},
			ExpectEmail:             types.StringNull(),
			ExpectProviderResource:  types.StringNull(),
			ExpectGcpCredentialsKey: types.StringValue(`{"type":"service_account"}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			got := convertDomainCredentialsToGCPCredentials(newTestCredentials(tc.Identifiers), tc.Prior)

			assert.Equal(t, tc.ExpectEmail, got.ServiceAccountEmail)
			assert.Equal(t, tc.ExpectProviderResource, got.WorkloadIdentityProviderResource)
			assert.Equal(t, tc.ExpectGcpCredentialsKey, got.GcpCredentials)
		})
	}
}

func TestConvertDomainCredentialsToEksAnywhereVsphereCredentials(t *testing.T) {
	t.Parallel()

	creds := newTestCredentials(credentials.Identifiers{
		VsphereUser: new("user-api"),
		RoleArn:     new("arn:aws:iam::123456789012:role/api"),
	})
	prior := EksAnywhereVsphereCredentials{
		VsphereUser:     types.StringValue("user-state"),
		VspherePassword: types.StringValue("vsphere-secret"),
		AccessKeyId:     types.StringValue("AKIA_STATE"),
		SecretAccessKey: types.StringValue("aws-secret"),
		RoleArn:         types.StringNull(),
	}

	got := convertDomainCredentialsToEksAnywhereVsphereCredentials(creds, prior)

	assert.Equal(t, types.StringValue("user-api"), got.VsphereUser)
	assert.Equal(t, types.StringNull(), got.AccessKeyId, "the API reports role credentials")
	assert.Equal(t, types.StringValue("arn:aws:iam::123456789012:role/api"), got.RoleArn)
	assert.Equal(t, prior.VspherePassword, got.VspherePassword, "the API never returns the secret: it comes from prior")
	assert.Equal(t, prior.SecretAccessKey, got.SecretAccessKey, "the API never returns the secret: it comes from prior")
}
