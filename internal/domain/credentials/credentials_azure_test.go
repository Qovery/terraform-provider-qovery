//go:build unit

package credentials_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestUpsertAzureRequest_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       credentials.UpsertAzureRequest
		ExpectedError error
	}{
		{
			TestName: "fail_with_empty_name",
			Request: credentials.UpsertAzureRequest{
				AzureSubscriptionId: gofakeit.UUID(),
				AzureTenantId:       gofakeit.UUID(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAzureRequest,
		},
		{
			TestName: "fail_with_empty_azure_subscription_id",
			Request: credentials.UpsertAzureRequest{
				Name:          gofakeit.Name(),
				AzureTenantId: gofakeit.UUID(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAzureRequest,
		},
		{
			TestName: "fail_with_empty_azure_tenant_id",
			Request: credentials.UpsertAzureRequest{
				Name:                gofakeit.Name(),
				AzureSubscriptionId: gofakeit.UUID(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAzureRequest,
		},
		{
			TestName:      "fail_with_all_empty_fields",
			Request:       credentials.UpsertAzureRequest{},
			ExpectedError: credentials.ErrInvalidUpsertAzureRequest,
		},
		{
			TestName: "success_with_valid_request",
			Request: credentials.UpsertAzureRequest{
				Name:                gofakeit.Name(),
				AzureSubscriptionId: gofakeit.UUID(),
				AzureTenantId:       gofakeit.UUID(),
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
