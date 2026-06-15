//go:build unit || !integration

package qovery

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ----------------------------------------------------------------

func makeEndpointObj(typ, region, projectID string) types.Object {
	vals := map[string]attr.Value{
		"type":       strOrNull(typ),
		"region":     strOrNull(region),
		"project_id": strOrNull(projectID),
	}
	return types.ObjectValueMust(secretManagerEndpointAttrTypes, vals)
}

func makeAuthObj(typ, roleArn, region, accessKey, secretKey, jsonCreds string) types.Object {
	vals := map[string]attr.Value{
		"type":             strOrNull(typ),
		"role_arn":         strOrNull(roleArn),
		"region":           strOrNull(region),
		"access_key":       strOrNull(accessKey),
		"secret_key":       strOrNull(secretKey),
		"json_credentials": strOrNull(jsonCreds),
	}
	return types.ObjectValueMust(secretManagerAuthAttrTypes, vals)
}

func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func makeAccessObj(id, name string, endpoint, auth types.Object) types.Object {
	return types.ObjectValueMust(secretManagerAccessAttrTypes, map[string]attr.Value{
		"id":             strOrNull(id),
		"name":           types.StringValue(name),
		"endpoint":       endpoint,
		"authentication": auth,
	})
}

func createSecretManagerAccesses(objs ...types.Object) types.Set {
	elems := make([]attr.Value, len(objs))
	for i, o := range objs {
		elems[i] = o
	}
	return types.SetValueMust(types.ObjectType{AttrTypes: secretManagerAccessAttrTypes}, elems)
}

func awsParamEndpointObj(region string) types.Object {
	return makeEndpointObj("AWS_PARAMETER_STORE", region, "")
}

func awsSecretEndpointObj(region string) types.Object {
	return makeEndpointObj("AWS_SECRET_MANAGER", region, "")
}

func gcpSecretEndpointObj(region, projectID string) types.Object {
	return makeEndpointObj("GCP_SECRET_MANAGER", region, projectID)
}

func autoAuthObj() types.Object {
	return makeAuthObj("AUTOMATICALLY_CONFIGURED", "", "", "", "", "")
}

func roleArnAuthObj(roleArn string) types.Object {
	return makeAuthObj("AWS_ROLE_ARN", roleArn, "", "", "", "")
}

func staticCredsAuthObj(region, accessKey, secretKey string) types.Object {
	return makeAuthObj("AWS_STATIC_CREDENTIALS", "", region, accessKey, secretKey, "")
}

func gcpJsonCredsAuthObj(jsonCreds string) types.Object {
	return makeAuthObj("GCP_JSON_CREDENTIALS", "", "", "", "", jsonCreds)
}

func makeAPIAccess(id, name string, endpoint qovery.SecretManagerEndpointConfigurationDto, auth qovery.SecretManagerAuthenticationDto) qovery.SecretManagerAccess {
	return qovery.SecretManagerAccess{
		Id:             id,
		Name:           name,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Endpoint:       endpoint,
		Authentication: auth,
	}
}

func awsParamAPIEndpoint(region string) qovery.SecretManagerEndpointConfigurationDto {
	return qovery.AwsParameterStoreEndpointDtoAsSecretManagerEndpointConfigurationDto(&qovery.AwsParameterStoreEndpointDto{
		Mode:   "AWS_PARAMETER_STORE",
		Region: region,
	})
}

func awsSecretAPIEndpoint(region string) qovery.SecretManagerEndpointConfigurationDto {
	return qovery.AwsSecretManagerEndpointDtoAsSecretManagerEndpointConfigurationDto(&qovery.AwsSecretManagerEndpointDto{
		Mode:   "AWS_SECRET_MANAGER",
		Region: region,
	})
}

func gcpSecretAPIEndpoint(region, projectID string) qovery.SecretManagerEndpointConfigurationDto {
	return qovery.GcpSecretManagerEndpointDtoAsSecretManagerEndpointConfigurationDto(&qovery.GcpSecretManagerEndpointDto{
		Mode:      "GCP_SECRET_MANAGER",
		Region:    region,
		ProjectId: projectID,
	})
}

func autoAPIAuth() qovery.SecretManagerAuthenticationDto {
	return qovery.AutomaticallyConfiguredAuthDtoAsSecretManagerAuthenticationDto(&qovery.AutomaticallyConfiguredAuthDto{
		Mode: "AUTOMATICALLY_CONFIGURED",
	})
}

func roleArnAPIAuth(roleArn string) qovery.SecretManagerAuthenticationDto {
	return qovery.AwsRoleArnAuthDtoAsSecretManagerAuthenticationDto(&qovery.AwsRoleArnAuthDto{
		Mode:    "AWS_ROLE_ARN",
		RoleArn: roleArn,
	})
}

func staticCredsAPIAuth(region, accessKey, secretKey string) qovery.SecretManagerAuthenticationDto {
	dto := &qovery.AwsStaticCredentialsAuthDto{
		Mode:      "AWS_STATIC_CREDENTIALS",
		Region:    region,
		AccessKey: accessKey,
	}
	dto.SetSecretKey(secretKey)
	return qovery.AwsStaticCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto)
}

func gcpJsonCredsAPIAuth(jsonCreds string) qovery.SecretManagerAuthenticationDto {
	dto := qovery.NewGcpJsonCredentialsAuthDto("GCP_JSON_CREDENTIALS")
	dto.SetJsonCredentials(jsonCreds)
	return qovery.GcpJsonCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto)
}

// --- toQoverySecretManagerAccessRequests ------------------------------------

func TestAcc_ToQoverySecretManagerAccessRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		secretManagerAccesses types.Set
		wantNil               bool
		wantLen               int
		wantError             bool
		errorContains         string
		check                 func(t *testing.T, reqs []qovery.SecretManagerAccessRequest)
	}{
		{
			name:                  "null_set_returns_nil",
			secretManagerAccesses: types.SetNull(types.ObjectType{AttrTypes: secretManagerAccessAttrTypes}),
			wantNil:               true,
		},
		{
			name:                  "unknown_set_returns_nil",
			secretManagerAccesses: types.SetUnknown(types.ObjectType{AttrTypes: secretManagerAccessAttrTypes}),
			wantNil:               true,
		},
		{
			name:                  "empty_set_returns_empty_slice",
			secretManagerAccesses: createSecretManagerAccesses(),
			wantLen:               0,
		},
		{
			name: "aws_parameter_store_endpoint_automatically_configured_auth",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "my-access",
				awsParamEndpointObj("us-east-1"),
				autoAuthObj(),
			)),
			wantLen: 1,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				r := reqs[0]
				assert.Equal(t, "my-access", r.Name)
				require.NotNil(t, r.Endpoint.AwsParameterStoreEndpointDto)
				assert.Equal(t, "AWS_PARAMETER_STORE", r.Endpoint.AwsParameterStoreEndpointDto.Mode)
				assert.Equal(t, "us-east-1", r.Endpoint.AwsParameterStoreEndpointDto.Region)
				require.NotNil(t, r.Authentication.AutomaticallyConfiguredAuthDto)
				assert.Equal(t, "AUTOMATICALLY_CONFIGURED", r.Authentication.AutomaticallyConfiguredAuthDto.Mode)
			},
		},
		{
			name: "aws_secret_manager_endpoint_aws_role_arn_auth",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "s3-access",
				awsSecretEndpointObj("eu-west-1"),
				roleArnAuthObj("arn:aws:iam::123456789012:role/my-role"),
			)),
			wantLen: 1,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				r := reqs[0]
				require.NotNil(t, r.Endpoint.AwsSecretManagerEndpointDto)
				assert.Equal(t, "AWS_SECRET_MANAGER", r.Endpoint.AwsSecretManagerEndpointDto.Mode)
				assert.Equal(t, "eu-west-1", r.Endpoint.AwsSecretManagerEndpointDto.Region)
				require.NotNil(t, r.Authentication.AwsRoleArnAuthDto)
				assert.Equal(t, "arn:aws:iam::123456789012:role/my-role", r.Authentication.AwsRoleArnAuthDto.RoleArn)
			},
		},
		{
			name: "gcp_secret_manager_endpoint_gcp_json_credentials_auth",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "gcp-access",
				gcpSecretEndpointObj("us-central1", "my-gcp-project"),
				gcpJsonCredsAuthObj(`{"type":"service_account"}`),
			)),
			wantLen: 1,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				r := reqs[0]
				require.NotNil(t, r.Endpoint.GcpSecretManagerEndpointDto)
				assert.Equal(t, "GCP_SECRET_MANAGER", r.Endpoint.GcpSecretManagerEndpointDto.Mode)
				assert.Equal(t, "us-central1", r.Endpoint.GcpSecretManagerEndpointDto.Region)
				assert.Equal(t, "my-gcp-project", r.Endpoint.GcpSecretManagerEndpointDto.ProjectId)
				require.NotNil(t, r.Authentication.GcpJsonCredentialsAuthDto)
				jc, ok := r.Authentication.GcpJsonCredentialsAuthDto.GetJsonCredentialsOk()
				require.True(t, ok)
				assert.Equal(t, `{"type":"service_account"}`, *jc)
			},
		},
		{
			name: "aws_static_credentials_auth",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "static-access",
				awsParamEndpointObj("us-east-2"),
				staticCredsAuthObj("us-east-2", "AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI"),
			)),
			wantLen: 1,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				dto := reqs[0].Authentication.AwsStaticCredentialsAuthDto
				require.NotNil(t, dto)
				assert.Equal(t, "us-east-2", dto.Region)
				assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", dto.AccessKey)
				sk, ok := dto.GetSecretKeyOk()
				require.True(t, ok)
				assert.Equal(t, "wJalrXUtnFEMI", *sk)
			},
		},
		{
			name: "element_with_id_sets_id_on_request",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("uuid-1234", "named",
				awsParamEndpointObj("us-east-1"),
				autoAuthObj(),
			)),
			wantLen: 1,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				assert.True(t, reqs[0].HasId())
				assert.Equal(t, "uuid-1234", reqs[0].GetId())
			},
		},
		{
			name: "element_with_null_id_omits_id",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "named",
				awsParamEndpointObj("us-east-1"),
				autoAuthObj(),
			)),
			wantLen: 1,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				assert.False(t, reqs[0].HasId())
			},
		},
		{
			name: "element_with_explicit_empty_string_id_omits_id",
			secretManagerAccesses: createSecretManagerAccesses(func() types.Object {
				// Use a non-null but empty string for id
				obj := makeAccessObj("", "named", awsParamEndpointObj("us-east-1"), autoAuthObj())
				// Override id with an actual empty types.StringValue (not null)
				attrs := map[string]attr.Value{
					"id":             types.StringValue(""),
					"name":           types.StringValue("named"),
					"endpoint":       awsParamEndpointObj("us-east-1"),
					"authentication": autoAuthObj(),
				}
				_ = obj
				return types.ObjectValueMust(secretManagerAccessAttrTypes, attrs)
			}()),
			wantLen: 1,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				assert.False(t, reqs[0].HasId())
			},
		},
		{
			name: "multiple_elements_all_converted",
			secretManagerAccesses: createSecretManagerAccesses(
				makeAccessObj("id-1", "access-1", awsParamEndpointObj("us-east-1"), autoAuthObj()),
				makeAccessObj("id-2", "access-2", awsSecretEndpointObj("eu-west-1"), roleArnAuthObj("arn:aws:iam::999:role/r")),
			),
			wantLen: 2,
			check: func(t *testing.T, reqs []qovery.SecretManagerAccessRequest) {
				assert.Equal(t, "access-1", reqs[0].Name)
				assert.Equal(t, "access-2", reqs[1].Name)
			},
		},
		{
			name: "gcp_secret_manager_missing_project_id_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "gcp",
				makeEndpointObj("GCP_SECRET_MANAGER", "us-central1", ""),
				autoAuthObj(),
			)),
			wantError:     true,
			errorContains: "project_id is required",
		},
		{
			name: "aws_role_arn_missing_role_arn_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "x",
				awsParamEndpointObj("us-east-1"),
				makeAuthObj("AWS_ROLE_ARN", "", "", "", "", ""),
			)),
			wantError:     true,
			errorContains: "role_arn is required",
		},
		{
			name: "aws_static_credentials_missing_region_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "x",
				awsParamEndpointObj("us-east-1"),
				makeAuthObj("AWS_STATIC_CREDENTIALS", "", "", "key", "secret", ""),
			)),
			wantError:     true,
			errorContains: "region is required",
		},
		{
			name: "aws_static_credentials_missing_access_key_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "x",
				awsParamEndpointObj("us-east-1"),
				makeAuthObj("AWS_STATIC_CREDENTIALS", "", "us-east-1", "", "secret", ""),
			)),
			wantError:     true,
			errorContains: "access_key is required",
		},
		{
			name: "aws_static_credentials_missing_secret_key_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "x",
				awsParamEndpointObj("us-east-1"),
				makeAuthObj("AWS_STATIC_CREDENTIALS", "", "us-east-1", "key", "", ""),
			)),
			wantError:     true,
			errorContains: "secret_key is required",
		},
		{
			name: "gcp_json_credentials_missing_json_credentials_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "x",
				awsParamEndpointObj("us-east-1"),
				makeAuthObj("GCP_JSON_CREDENTIALS", "", "", "", "", ""),
			)),
			wantError:     true,
			errorContains: "json_credentials is required",
		},
		{
			name: "unsupported_endpoint_type_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "x",
				makeEndpointObj("UNKNOWN_ENDPOINT", "us-east-1", ""),
				autoAuthObj(),
			)),
			wantError:     true,
			errorContains: "unsupported endpoint type",
		},
		{
			name: "unsupported_auth_type_returns_error",
			secretManagerAccesses: createSecretManagerAccesses(makeAccessObj("", "x",
				awsParamEndpointObj("us-east-1"),
				makeAuthObj("UNKNOWN_AUTH", "", "", "", "", ""),
			)),
			wantError:     true,
			errorContains: "unsupported authentication type",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reqs, err := toQoverySecretManagerAccessRequests(tc.secretManagerAccesses)

			if tc.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
				return
			}

			require.NoError(t, err)

			if tc.wantNil {
				assert.Nil(t, reqs)
				return
			}

			require.NotNil(t, reqs)
			assert.Len(t, reqs, tc.wantLen)

			if tc.check != nil {
				tc.check(t, reqs)
			}
		})
	}
}

// --- fromQoverySecretManagerAccesses ----------------------------------------

func TestAcc_FromQoverySecretManagerAccesses(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	setType := types.ObjectType{AttrTypes: secretManagerAccessAttrTypes}

	tests := []struct {
		name      string
		accesses  []qovery.SecretManagerAccess
		planSet   types.Set
		checkList func(t *testing.T, result types.Set)
	}{
		{
			name:     "plan_null_accesses_nil_returns_null_set",
			accesses: nil,
			planSet:  types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				assert.True(t, result.IsNull())
			},
		},
		{
			name:     "plan_unknown_accesses_nil_returns_unknown_set",
			accesses: nil,
			planSet:  types.SetUnknown(setType),
			checkList: func(t *testing.T, result types.Set) {
				assert.True(t, result.IsUnknown())
			},
		},
		{
			name:     "plan_set_accesses_nil_returns_null_set",
			accesses: nil,
			planSet:  createSecretManagerAccesses(makeAccessObj("id-1", "a1", awsParamEndpointObj("us-east-1"), autoAuthObj())),
			checkList: func(t *testing.T, result types.Set) {
				assert.True(t, result.IsNull())
			},
		},
		{
			name: "plan_null_accesses_non_nil_builds_set",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("srv-id", "api-access", awsParamAPIEndpoint("us-east-1"), autoAPIAuth()),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				require.False(t, result.IsNull())
				require.False(t, result.IsUnknown())
				assert.Len(t, result.Elements(), 1)
				obj := result.Elements()[0].(types.Object)
				assert.Equal(t, "api-access", obj.Attributes()["name"].(types.String).ValueString())
				assert.Equal(t, "srv-id", obj.Attributes()["id"].(types.String).ValueString())
			},
		},
		{
			name: "plan_unknown_accesses_non_nil_builds_set",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("srv-id", "api-access", awsParamAPIEndpoint("us-east-1"), autoAPIAuth()),
			},
			planSet: types.SetUnknown(setType),
			checkList: func(t *testing.T, result types.Set) {
				require.False(t, result.IsNull())
				require.False(t, result.IsUnknown())
				assert.Len(t, result.Elements(), 1)
			},
		},
		{
			name: "plan_set_accesses_non_nil_builds_set",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("srv-id-1", "named-access", awsSecretAPIEndpoint("eu-west-1"), roleArnAPIAuth("arn:aws:iam::1:role/r")),
			},
			planSet: createSecretManagerAccesses(makeAccessObj("srv-id-1", "named-access", awsSecretEndpointObj("eu-west-1"), roleArnAuthObj("arn:aws:iam::1:role/r"))),
			checkList: func(t *testing.T, result types.Set) {
				require.False(t, result.IsNull())
				require.Len(t, result.Elements(), 1)
				obj := result.Elements()[0].(types.Object)
				assert.Equal(t, "srv-id-1", obj.Attributes()["id"].(types.String).ValueString())
				assert.Equal(t, "named-access", obj.Attributes()["name"].(types.String).ValueString())
			},
		},
		{
			name: "plan_order_preserved",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-b", "access-b", awsParamAPIEndpoint("us-east-1"), autoAPIAuth()),
				makeAPIAccess("id-a", "access-a", awsParamAPIEndpoint("us-east-1"), autoAPIAuth()),
			},
			planSet: createSecretManagerAccesses(
				makeAccessObj("id-a", "access-a", awsParamEndpointObj("us-east-1"), autoAuthObj()),
				makeAccessObj("id-b", "access-b", awsParamEndpointObj("us-east-1"), autoAuthObj()),
			),
			checkList: func(t *testing.T, result types.Set) {
				require.Len(t, result.Elements(), 2)
				names := make([]string, 0, 2)
				for _, elem := range result.Elements() {
					names = append(names, elem.(types.Object).Attributes()["name"].(types.String).ValueString())
				}
				assert.ElementsMatch(t, []string{"access-a", "access-b"}, names)
			},
		},
		{
			name: "server_side_extra_appended_after_plan_items",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-plan", "plan-access", awsParamAPIEndpoint("us-east-1"), autoAPIAuth()),
				makeAPIAccess("id-extra", "extra-access", awsParamAPIEndpoint("us-east-1"), autoAPIAuth()),
			},
			planSet: createSecretManagerAccesses(
				makeAccessObj("id-plan", "plan-access", awsParamEndpointObj("us-east-1"), autoAuthObj()),
			),
			checkList: func(t *testing.T, result types.Set) {
				require.Len(t, result.Elements(), 2)
				names := make([]string, 0, 2)
				for _, elem := range result.Elements() {
					names = append(names, elem.(types.Object).Attributes()["name"].(types.String).ValueString())
				}
				assert.ElementsMatch(t, []string{"plan-access", "extra-access"}, names)
			},
		},
		{
			name: "aws_parameter_store_endpoint_fields_mapped",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", awsParamAPIEndpoint("ap-southeast-1"), autoAPIAuth()),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				endpointObj := result.Elements()[0].(types.Object).Attributes()["endpoint"].(types.Object)
				assert.Equal(t, "AWS_PARAMETER_STORE", endpointObj.Attributes()["type"].(types.String).ValueString())
				assert.Equal(t, "ap-southeast-1", endpointObj.Attributes()["region"].(types.String).ValueString())
				assert.True(t, endpointObj.Attributes()["project_id"].(types.String).IsNull())
			},
		},
		{
			name: "aws_secret_manager_endpoint_fields_mapped",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", awsSecretAPIEndpoint("us-west-2"), autoAPIAuth()),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				endpointObj := result.Elements()[0].(types.Object).Attributes()["endpoint"].(types.Object)
				assert.Equal(t, "AWS_SECRET_MANAGER", endpointObj.Attributes()["type"].(types.String).ValueString())
				assert.Equal(t, "us-west-2", endpointObj.Attributes()["region"].(types.String).ValueString())
				assert.True(t, endpointObj.Attributes()["project_id"].(types.String).IsNull())
			},
		},
		{
			name: "gcp_secret_manager_endpoint_fields_mapped",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", gcpSecretAPIEndpoint("us-central1", "gcp-proj"), autoAPIAuth()),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				endpointObj := result.Elements()[0].(types.Object).Attributes()["endpoint"].(types.Object)
				assert.Equal(t, "GCP_SECRET_MANAGER", endpointObj.Attributes()["type"].(types.String).ValueString())
				assert.Equal(t, "us-central1", endpointObj.Attributes()["region"].(types.String).ValueString())
				assert.Equal(t, "gcp-proj", endpointObj.Attributes()["project_id"].(types.String).ValueString())
			},
		},
		{
			name: "auth_automatically_configured_mapped",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", awsParamAPIEndpoint("us-east-1"), autoAPIAuth()),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				authObj := result.Elements()[0].(types.Object).Attributes()["authentication"].(types.Object)
				assert.Equal(t, "AUTOMATICALLY_CONFIGURED", authObj.Attributes()["type"].(types.String).ValueString())
				assert.True(t, authObj.Attributes()["role_arn"].(types.String).IsNull())
				assert.True(t, authObj.Attributes()["region"].(types.String).IsNull())
			},
		},
		{
			name: "auth_aws_role_arn_mapped",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", awsParamAPIEndpoint("us-east-1"), roleArnAPIAuth("arn:aws:iam::9999:role/r")),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				authObj := result.Elements()[0].(types.Object).Attributes()["authentication"].(types.Object)
				assert.Equal(t, "AWS_ROLE_ARN", authObj.Attributes()["type"].(types.String).ValueString())
				assert.Equal(t, "arn:aws:iam::9999:role/r", authObj.Attributes()["role_arn"].(types.String).ValueString())
			},
		},
		{
			name: "auth_aws_static_credentials_secret_key_from_api",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", awsParamAPIEndpoint("us-east-1"), staticCredsAPIAuth("us-east-1", "AKID", "s3cr3t")),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				authObj := result.Elements()[0].(types.Object).Attributes()["authentication"].(types.Object)
				assert.Equal(t, "AWS_STATIC_CREDENTIALS", authObj.Attributes()["type"].(types.String).ValueString())
				assert.Equal(t, "us-east-1", authObj.Attributes()["region"].(types.String).ValueString())
				assert.Equal(t, "AKID", authObj.Attributes()["access_key"].(types.String).ValueString())
				assert.Equal(t, "s3cr3t", authObj.Attributes()["secret_key"].(types.String).ValueString())
			},
		},
		{
			name: "auth_aws_static_credentials_secret_key_from_plan_when_redacted",
			accesses: []qovery.SecretManagerAccess{
				// API returns empty secret_key (redacted)
				makeAPIAccess("id-1", "a", awsParamAPIEndpoint("us-east-1"), func() qovery.SecretManagerAuthenticationDto {
					dto := &qovery.AwsStaticCredentialsAuthDto{
						Mode:      "AWS_STATIC_CREDENTIALS",
						Region:    "us-east-1",
						AccessKey: "AKID",
					}
					// secret_key deliberately not set → nil
					return qovery.AwsStaticCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto)
				}()),
			},
			planSet: createSecretManagerAccesses(makeAccessObj("id-1", "a",
				awsParamEndpointObj("us-east-1"),
				staticCredsAuthObj("us-east-1", "AKID", "plan-secret"),
			)),
			checkList: func(t *testing.T, result types.Set) {
				authObj := result.Elements()[0].(types.Object).Attributes()["authentication"].(types.Object)
				assert.Equal(t, "plan-secret", authObj.Attributes()["secret_key"].(types.String).ValueString())
			},
		},
		{
			name: "auth_gcp_json_credentials_from_api",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", gcpSecretAPIEndpoint("us-central1", "p"), gcpJsonCredsAPIAuth(`{"k":"v"}`)),
			},
			planSet: types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				authObj := result.Elements()[0].(types.Object).Attributes()["authentication"].(types.Object)
				assert.Equal(t, "GCP_JSON_CREDENTIALS", authObj.Attributes()["type"].(types.String).ValueString())
				assert.Equal(t, `{"k":"v"}`, authObj.Attributes()["json_credentials"].(types.String).ValueString())
			},
		},
		{
			name: "auth_gcp_json_credentials_from_plan_when_redacted",
			accesses: []qovery.SecretManagerAccess{
				makeAPIAccess("id-1", "a", gcpSecretAPIEndpoint("us-central1", "p"), func() qovery.SecretManagerAuthenticationDto {
					dto := qovery.NewGcpJsonCredentialsAuthDto("GCP_JSON_CREDENTIALS")
					// json_credentials deliberately not set → nil
					return qovery.GcpJsonCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto)
				}()),
			},
			planSet: createSecretManagerAccesses(makeAccessObj("id-1", "a",
				gcpSecretEndpointObj("us-central1", "p"),
				gcpJsonCredsAuthObj(`{"from":"plan"}`),
			)),
			checkList: func(t *testing.T, result types.Set) {
				authObj := result.Elements()[0].(types.Object).Attributes()["authentication"].(types.Object)
				assert.Equal(t, `{"from":"plan"}`, authObj.Attributes()["json_credentials"].(types.String).ValueString())
			},
		},
		{
			name:     "empty_accesses_slice_returns_empty_set",
			accesses: []qovery.SecretManagerAccess{},
			planSet:  createSecretManagerAccesses(makeAccessObj("id-1", "a", awsParamEndpointObj("us-east-1"), autoAuthObj())),
			checkList: func(t *testing.T, result types.Set) {
				require.False(t, result.IsNull())
				assert.Len(t, result.Elements(), 0)
			},
		},
		{
			// Removing secret_manager_accesses from config (planSet=null) after a previous
			// apply must produce null in state, not an empty set — otherwise Terraform reports
			// "provider produced inconsistent result after apply".
			name:     "plan_null_empty_accesses_slice_returns_null_set",
			accesses: []qovery.SecretManagerAccess{},
			planSet:  types.SetNull(setType),
			checkList: func(t *testing.T, result types.Set) {
				assert.True(t, result.IsNull())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := fromQoverySecretManagerAccesses(ctx, tc.accesses, tc.planSet)
			tc.checkList(t, result)
		})
	}
}
