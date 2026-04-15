//go:build unit

package credentials_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

func TestUpsertEksAnywhereVsphereRequest_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName      string
		request       credentials.UpsertEksAnywhereVsphereRequest
		expectedError error
	}{
		{
			testName: "success_with_static_credentials",
			request: credentials.UpsertEksAnywhereVsphereRequest{
				Name:            gofakeit.Name(),
				VsphereUser:     gofakeit.Username(),
				VspherePassword: gofakeit.Password(true, true, true, true, false, 12),
				StaticCredentials: &credentials.VsphereStaticCredentials{
					AccessKeyID:     gofakeit.Word(),
					SecretAccessKey: gofakeit.Word(),
				},
			},
		},
		{
			testName: "success_with_role_credentials",
			request: credentials.UpsertEksAnywhereVsphereRequest{
				Name:            gofakeit.Name(),
				VsphereUser:     gofakeit.Username(),
				VspherePassword: gofakeit.Password(true, true, true, true, false, 12),
				RoleCredentials: &credentials.VsphereRoleCredentials{
					RoleArn: gofakeit.Word(),
				},
			},
		},
		{
			testName: "fail_with_both_authentication_methods",
			request: credentials.UpsertEksAnywhereVsphereRequest{
				Name:            gofakeit.Name(),
				VsphereUser:     gofakeit.Username(),
				VspherePassword: gofakeit.Password(true, true, true, true, false, 12),
				StaticCredentials: &credentials.VsphereStaticCredentials{
					AccessKeyID:     gofakeit.Word(),
					SecretAccessKey: gofakeit.Word(),
				},
				RoleCredentials: &credentials.VsphereRoleCredentials{
					RoleArn: gofakeit.Word(),
				},
			},
			expectedError: credentials.ErrInvalidUpsertEksAnywhereVsphereRequest,
		},
		{
			testName: "fail_without_authentication_method",
			request: credentials.UpsertEksAnywhereVsphereRequest{
				Name:            gofakeit.Name(),
				VsphereUser:     gofakeit.Username(),
				VspherePassword: gofakeit.Password(true, true, true, true, false, 12),
			},
			expectedError: credentials.ErrInvalidUpsertEksAnywhereVsphereRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			err := tc.request.Validate()

			if tc.expectedError != nil {
				assert.ErrorContains(t, err, tc.expectedError.Error())
				assert.False(t, tc.request.IsValid())
				return
			}

			assert.NoError(t, err)
			assert.True(t, tc.request.IsValid())
		})
	}
}
