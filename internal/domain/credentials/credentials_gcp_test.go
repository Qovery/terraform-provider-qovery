//go:build unit

package credentials_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestUpsertGcpRequest_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       credentials.UpsertGcpRequest
		ExpectedError error
	}{
		{
			TestName: "fail_with_empty_name",
			Request: credentials.UpsertGcpRequest{
				ServiceAccountKey: &credentials.GcpServiceAccountKeyCredentials{GcpCredentials: gofakeit.UUID()},
			},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName:      "fail_with_no_auth_mode",
			Request:       credentials.UpsertGcpRequest{Name: gofakeit.Name()},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName: "fail_with_both_auth_modes",
			Request: credentials.UpsertGcpRequest{
				Name:              gofakeit.Name(),
				ServiceAccountKey: &credentials.GcpServiceAccountKeyCredentials{GcpCredentials: gofakeit.UUID()},
				WorkloadIdentity: &credentials.GcpWorkloadIdentityCredentials{
					ServiceAccountEmail:              "qovery@proj.iam.gserviceaccount.com",
					WorkloadIdentityProviderResource: "projects/123/locations/global/workloadIdentityPools/p/providers/pr",
				},
			},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName: "fail_with_empty_gcp_credentials",
			Request: credentials.UpsertGcpRequest{
				Name:              gofakeit.Name(),
				ServiceAccountKey: &credentials.GcpServiceAccountKeyCredentials{},
			},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName: "fail_with_wif_missing_provider_resource",
			Request: credentials.UpsertGcpRequest{
				Name: gofakeit.Name(),
				WorkloadIdentity: &credentials.GcpWorkloadIdentityCredentials{
					ServiceAccountEmail: "qovery@proj.iam.gserviceaccount.com",
				},
			},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName: "success_with_service_account_key",
			Request: credentials.UpsertGcpRequest{
				Name:              gofakeit.Name(),
				ServiceAccountKey: &credentials.GcpServiceAccountKeyCredentials{GcpCredentials: `{"type":"service_account","project_id":"test-project"}`},
			},
		},
		{
			TestName: "success_with_workload_identity",
			Request: credentials.UpsertGcpRequest{
				Name: gofakeit.Name(),
				WorkloadIdentity: &credentials.GcpWorkloadIdentityCredentials{
					ServiceAccountEmail:              "qovery@proj.iam.gserviceaccount.com",
					WorkloadIdentityProviderResource: "projects/123/locations/global/workloadIdentityPools/p/providers/pr",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			err := tc.Request.Validate()
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.False(t, tc.Request.IsValid())
				return
			}

			assert.NoError(t, err)
			assert.True(t, tc.Request.IsValid())
		})
	}
}
