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
				StaticCredentials: &credentials.AwsStaticCredentials{
					AccessKeyID:     gofakeit.Word(),
					SecretAccessKey: gofakeit.Word(),
				},
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
		{
			TestName: "fail_with_invalid_access_key_id",
			Request: credentials.UpsertAwsRequest{
				Name: gofakeit.Name(),
				StaticCredentials: &credentials.AwsStaticCredentials{
					SecretAccessKey: gofakeit.Word(),
				},
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
		{
			TestName: "fail_with_secret_access_key",
			Request: credentials.UpsertAwsRequest{
				Name: gofakeit.Name(),
				StaticCredentials: &credentials.AwsStaticCredentials{
					AccessKeyID: gofakeit.Word(),
				},
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
		},
		{
			TestName: "success_with_static_credentials",
			Request: credentials.UpsertAwsRequest{
				Name: gofakeit.Name(),
				StaticCredentials: &credentials.AwsStaticCredentials{
					AccessKeyID:     gofakeit.Word(),
					SecretAccessKey: gofakeit.Word(),
				},
			},
		},
		{
			TestName: "success_with_role_credentials",
			Request: credentials.UpsertAwsRequest{
				Name: gofakeit.Name(),
				RoleCredentials: &credentials.AwsRoleCredentials{
					RoleArn: gofakeit.Word(),
				},
			},
		},
		{
			TestName: "fail_with_missing_role_arn",
			Request: credentials.UpsertAwsRequest{
				Name:            gofakeit.Name(),
				RoleCredentials: &credentials.AwsRoleCredentials{},
			},
			ExpectedError: credentials.ErrInvalidUpsertAwsRequest,
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
