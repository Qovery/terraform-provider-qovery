//go:build unit || !integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEksAnywhereVsphereCredentials_toUpsertEksAnywhereVsphereRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		credentials         EksAnywhereVsphereCredentials
		expectedError       error
		expectedStaticCreds bool
		expectedRoleCreds   bool
	}{
		{
			name: "success_with_role_credentials",
			credentials: EksAnywhereVsphereCredentials{
				Name:            types.StringValue("test-creds"),
				VsphereUser:     types.StringValue("vsphere-user"),
				VspherePassword: types.StringValue("vsphere-password"),
				RoleArn:         types.StringValue("arn:aws:iam::123456789012:role/test-role"),
				AccessKeyId:     types.StringNull(),
				SecretAccessKey: types.StringNull(),
			},
			expectedRoleCreds: true,
		},
		{
			name: "success_with_static_credentials",
			credentials: EksAnywhereVsphereCredentials{
				Name:            types.StringValue("test-creds"),
				VsphereUser:     types.StringValue("vsphere-user"),
				VspherePassword: types.StringValue("vsphere-password"),
				RoleArn:         types.StringNull(),
				AccessKeyId:     types.StringValue("access-key-id"),
				SecretAccessKey: types.StringValue("secret-access-key"),
			},
			expectedStaticCreds: true,
		},
		{
			name: "fail_with_both_authentication_methods",
			credentials: EksAnywhereVsphereCredentials{
				Name:            types.StringValue("test-creds"),
				VsphereUser:     types.StringValue("vsphere-user"),
				VspherePassword: types.StringValue("vsphere-password"),
				RoleArn:         types.StringValue("arn:aws:iam::123456789012:role/test-role"),
				AccessKeyId:     types.StringValue("access-key-id"),
				SecretAccessKey: types.StringValue("secret-access-key"),
			},
			expectedError: errEksAnywhereVsphereMutuallyExclusiveAuthMethods,
		},
		{
			name: "fail_with_partial_static_credentials",
			credentials: EksAnywhereVsphereCredentials{
				Name:            types.StringValue("test-creds"),
				VsphereUser:     types.StringValue("vsphere-user"),
				VspherePassword: types.StringValue("vsphere-password"),
				RoleArn:         types.StringNull(),
				AccessKeyId:     types.StringValue("access-key-id"),
				SecretAccessKey: types.StringNull(),
			},
			expectedError: errEksAnywhereVsphereStaticCredentialsPair,
		},
		{
			name: "fail_without_authentication_method",
			credentials: EksAnywhereVsphereCredentials{
				Name:            types.StringValue("test-creds"),
				VsphereUser:     types.StringValue("vsphere-user"),
				VspherePassword: types.StringValue("vsphere-password"),
				RoleArn:         types.StringNull(),
				AccessKeyId:     types.StringNull(),
				SecretAccessKey: types.StringNull(),
			},
			expectedError: errEksAnywhereVsphereMissingAuthMethod,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			request, err := tc.credentials.toUpsertEksAnywhereVsphereRequest()

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, request)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, request)
			assert.Equal(t, tc.expectedRoleCreds, request.RoleCredentials != nil)
			assert.Equal(t, tc.expectedStaticCreds, request.StaticCredentials != nil)
		})
	}
}
