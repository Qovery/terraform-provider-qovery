//go:build unit
// +build unit

package credentials_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestUpsertAwsRequest_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       credentials.UpsertAwsRequest
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_name",
			Request: credentials.UpsertAwsRequest{
				AccessKeyID:     gofakeit.Word(),
				SecretAccessKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
		{
			TestName: "fail_with_invalid_access_key_id",
			Request: credentials.UpsertAwsRequest{
				Name:            gofakeit.Name(),
				SecretAccessKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
		{
			TestName: "fail_with_secret_access_key",
			Request: credentials.UpsertAwsRequest{
				Name:        gofakeit.Name(),
				AccessKeyID: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
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
