package credentials_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestUpsertScalewayRequest_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Request       credentials.UpsertScalewayRequest
		ExpectedError error
	}{
		{
			TestName: "fail_with_invalid_name",
			Request: credentials.UpsertScalewayRequest{
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName: "fail_with_invalid_scaleway_project_id",
			Request: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName: "fail_with_scaleway_access_key",
			Request: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName: "fail_with_scaleway_secret_key",
			Request: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
			},
			ExpectedError: credentials.ErrInvalidUpsertScalewayRequest,
		},
		{
			TestName: "success",
			Request: credentials.UpsertScalewayRequest{
				Name:              gofakeit.Name(),
				ScalewayProjectID: gofakeit.Word(),
				ScalewayAccessKey: gofakeit.Word(),
				ScalewaySecretKey: gofakeit.Word(),
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
