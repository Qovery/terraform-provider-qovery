//go:build unit && !integration

package qoveryapi

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdCredentials"
)

func TestNewDomainArgoCdCredentialsFromQovery(t *testing.T) {
	t.Parallel()

	validID := gofakeit.UUID()
	validClusterID := gofakeit.UUID()

	testCases := []struct {
		TestName      string
		Response      *qovery.ArgoCdCredentialsResponse
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "fail_with_invalid_id",
			Response:      &qovery.ArgoCdCredentialsResponse{Id: "not-a-uuid", ClusterId: validClusterID, ArgocdUrl: gofakeit.URL(), ArgocdToken: "REDACTED"},
			ExpectError:   true,
			ErrorContains: argoCdCredentials.ErrInvalidClusterIDParam.Error(),
		},
		{
			TestName:      "fail_with_invalid_cluster_id",
			Response:      &qovery.ArgoCdCredentialsResponse{Id: validID, ClusterId: "not-a-uuid", ArgocdUrl: gofakeit.URL(), ArgocdToken: "REDACTED"},
			ExpectError:   true,
			ErrorContains: argoCdCredentials.ErrInvalidClusterIDParam.Error(),
		},
		{
			TestName: "success",
			Response: &qovery.ArgoCdCredentialsResponse{Id: validID, ClusterId: validClusterID, ArgocdUrl: "https://argocd.example.com", ArgocdToken: "REDACTED"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			creds, err := newDomainArgoCdCredentialsFromQovery(tc.Response)
			if tc.ExpectError {
				assert.ErrorContains(t, err, tc.ErrorContains)
				assert.Nil(t, creds)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, creds)
			assert.Equal(t, tc.Response.Id, creds.ID.String())
			assert.Equal(t, tc.Response.ClusterId, creds.ClusterID.String())
			assert.Equal(t, tc.Response.ArgocdUrl, creds.ArgocdUrl)
			assert.Equal(t, tc.Response.ArgocdToken, creds.ArgocdToken)
		})
	}
}
