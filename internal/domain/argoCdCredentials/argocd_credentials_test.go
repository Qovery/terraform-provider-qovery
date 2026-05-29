//go:build unit && !integration

package argoCdCredentials_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdCredentials"
)

func TestUpsertRequest_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Request     argoCdCredentials.UpsertRequest
		ExpectError bool
	}{
		{
			TestName:    "fail_with_missing_argocd_url",
			Request:     argoCdCredentials.UpsertRequest{ArgocdToken: gofakeit.UUID()},
			ExpectError: true,
		},
		{
			TestName:    "fail_with_missing_argocd_token",
			Request:     argoCdCredentials.UpsertRequest{ArgocdUrl: gofakeit.URL()},
			ExpectError: true,
		},
		{
			TestName: "success",
			Request:  argoCdCredentials.UpsertRequest{ArgocdUrl: gofakeit.URL(), ArgocdToken: gofakeit.UUID()},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			err := tc.Request.Validate()
			if tc.ExpectError {
				assert.ErrorContains(t, err, argoCdCredentials.ErrInvalidUpsertRequest.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
