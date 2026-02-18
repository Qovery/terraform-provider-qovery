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
				GcpCredentials: gofakeit.UUID(),
			},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName: "fail_with_empty_gcp_credentials",
			Request: credentials.UpsertGcpRequest{
				Name: gofakeit.Name(),
			},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName:      "fail_with_all_empty_fields",
			Request:       credentials.UpsertGcpRequest{},
			ExpectedError: credentials.ErrInvalidUpsertGcpRequest,
		},
		{
			TestName: "success_with_valid_request",
			Request: credentials.UpsertGcpRequest{
				Name:           gofakeit.Name(),
				GcpCredentials: `{"type": "service_account", "project_id": "test-project"}`,
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
