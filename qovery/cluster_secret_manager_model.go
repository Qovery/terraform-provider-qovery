package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
)

var secretManagerEndpointAttrTypes = map[string]attr.Type{
	"type":       types.StringType,
	"region":     types.StringType,
	"project_id": types.StringType,
}

var secretManagerAuthAttrTypes = map[string]attr.Type{
	"type":             types.StringType,
	"role_arn":         types.StringType,
	"region":           types.StringType,
	"access_key":       types.StringType,
	"secret_key":       types.StringType,
	"json_credentials": types.StringType,
}

var secretManagerAccessAttrTypes = map[string]attr.Type{
	"id":             types.StringType,
	"name":           types.StringType,
	"endpoint":       types.ObjectType{AttrTypes: secretManagerEndpointAttrTypes},
	"authentication": types.ObjectType{AttrTypes: secretManagerAuthAttrTypes},
}

func toQoverySecretManagerAccessRequests(set types.Set) ([]qovery.SecretManagerAccessRequest, error) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}

	requests := make([]qovery.SecretManagerAccessRequest, 0, len(set.Elements()))
	for _, elem := range set.Elements() {
		obj := elem.(types.Object)
		attrs := obj.Attributes()

		name := attrs["name"].(types.String).ValueString()
		id := attrs["id"].(types.String)

		endpoint, err := toQoverySecretManagerEndpoint(attrs["endpoint"].(types.Object))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse endpoint for secret manager access %q", name)
		}

		auth, err := toQoverySecretManagerAuth(attrs["authentication"].(types.Object))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse authentication for secret manager access %q", name)
		}

		req := qovery.SecretManagerAccessRequest{
			Name:           name,
			Endpoint:       endpoint,
			Authentication: auth,
		}
		if !id.IsNull() && !id.IsUnknown() && id.ValueString() != "" {
			req.SetId(id.ValueString())
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func toQoverySecretManagerEndpoint(obj types.Object) (qovery.SecretManagerEndpointConfigurationDto, error) {
	if obj.IsNull() || obj.IsUnknown() {
		return qovery.SecretManagerEndpointConfigurationDto{}, errors.New("endpoint is required")
	}

	attrs := obj.Attributes()

	typeAttr := attrs["type"].(types.String)
	if typeAttr.IsNull() || typeAttr.IsUnknown() || typeAttr.ValueString() == "" {
		return qovery.SecretManagerEndpointConfigurationDto{}, errors.New("endpoint.type is required")
	}
	regionAttr := attrs["region"].(types.String)
	if regionAttr.IsNull() || regionAttr.IsUnknown() || regionAttr.ValueString() == "" {
		return qovery.SecretManagerEndpointConfigurationDto{}, errors.New("endpoint.region is required")
	}
	endpointType := typeAttr.ValueString()
	region := regionAttr.ValueString()

	switch endpointType {
	case "AWS_PARAMETER_STORE":
		return qovery.AwsParameterStoreEndpointDtoAsSecretManagerEndpointConfigurationDto(&qovery.AwsParameterStoreEndpointDto{
			Mode:   endpointType,
			Region: region,
		}), nil
	case "AWS_SECRET_MANAGER":
		return qovery.AwsSecretManagerEndpointDtoAsSecretManagerEndpointConfigurationDto(&qovery.AwsSecretManagerEndpointDto{
			Mode:   endpointType,
			Region: region,
		}), nil
	case "GCP_SECRET_MANAGER":
		projectAttr := attrs["project_id"].(types.String)
		if projectAttr.IsNull() || projectAttr.IsUnknown() || projectAttr.ValueString() == "" {
			return qovery.SecretManagerEndpointConfigurationDto{}, errors.New("endpoint.project_id is required when endpoint.type is GCP_SECRET_MANAGER")
		}
		return qovery.GcpSecretManagerEndpointDtoAsSecretManagerEndpointConfigurationDto(&qovery.GcpSecretManagerEndpointDto{
			Mode:      endpointType,
			Region:    region,
			ProjectId: projectAttr.ValueString(),
		}), nil
	default:
		return qovery.SecretManagerEndpointConfigurationDto{}, errors.Errorf("unsupported endpoint type: %q", endpointType)
	}
}

func toQoverySecretManagerAuth(obj types.Object) (qovery.SecretManagerAuthenticationDto, error) {
	if obj.IsNull() || obj.IsUnknown() {
		return qovery.SecretManagerAuthenticationDto{}, errors.New("authentication is required")
	}

	attrs := obj.Attributes()
	typeAttr := attrs["type"].(types.String)
	if typeAttr.IsNull() || typeAttr.IsUnknown() || typeAttr.ValueString() == "" {
		return qovery.SecretManagerAuthenticationDto{}, errors.New("authentication.type is required")
	}
	authType := typeAttr.ValueString()

	switch authType {
	case "AUTOMATICALLY_CONFIGURED":
		return qovery.AutomaticallyConfiguredAuthDtoAsSecretManagerAuthenticationDto(&qovery.AutomaticallyConfiguredAuthDto{
			Mode: authType,
		}), nil
	case "AWS_ROLE_ARN":
		roleArnAttr := attrs["role_arn"].(types.String)
		if roleArnAttr.IsNull() || roleArnAttr.IsUnknown() || roleArnAttr.ValueString() == "" {
			return qovery.SecretManagerAuthenticationDto{}, errors.New("authentication.role_arn is required when authentication.type is AWS_ROLE_ARN")
		}
		return qovery.AwsRoleArnAuthDtoAsSecretManagerAuthenticationDto(&qovery.AwsRoleArnAuthDto{
			Mode:    authType,
			RoleArn: roleArnAttr.ValueString(),
		}), nil
	case "AWS_STATIC_CREDENTIALS":
		regionAttr := attrs["region"].(types.String)
		if regionAttr.IsNull() || regionAttr.IsUnknown() || regionAttr.ValueString() == "" {
			return qovery.SecretManagerAuthenticationDto{}, errors.New("authentication.region is required when authentication.type is AWS_STATIC_CREDENTIALS")
		}
		accessKeyAttr := attrs["access_key"].(types.String)
		if accessKeyAttr.IsNull() || accessKeyAttr.IsUnknown() || accessKeyAttr.ValueString() == "" {
			return qovery.SecretManagerAuthenticationDto{}, errors.New("authentication.access_key is required when authentication.type is AWS_STATIC_CREDENTIALS")
		}
		secretKeyAttr := attrs["secret_key"].(types.String)
		if secretKeyAttr.IsNull() || secretKeyAttr.IsUnknown() || secretKeyAttr.ValueString() == "" {
			return qovery.SecretManagerAuthenticationDto{}, errors.New("authentication.secret_key is required when authentication.type is AWS_STATIC_CREDENTIALS")
		}
		dto := &qovery.AwsStaticCredentialsAuthDto{
			Mode:      authType,
			Region:    regionAttr.ValueString(),
			AccessKey: accessKeyAttr.ValueString(),
		}
		dto.SetSecretKey(secretKeyAttr.ValueString())
		return qovery.AwsStaticCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto), nil
	case "GCP_JSON_CREDENTIALS":
		jsonCredsAttr := attrs["json_credentials"].(types.String)
		if jsonCredsAttr.IsNull() || jsonCredsAttr.IsUnknown() || jsonCredsAttr.ValueString() == "" {
			return qovery.SecretManagerAuthenticationDto{}, errors.New("authentication.json_credentials is required when authentication.type is GCP_JSON_CREDENTIALS")
		}
		dto := qovery.NewGcpJsonCredentialsAuthDto(authType)
		dto.SetJsonCredentials(jsonCredsAttr.ValueString())
		return qovery.GcpJsonCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto), nil
	default:
		return qovery.SecretManagerAuthenticationDto{}, errors.Errorf("unsupported authentication type: %q", authType)
	}
}

// fromQoverySecretManagerAccesses converts API response objects to a Terraform set.
// planSet is used as fallback for sensitive fields the API may redact in responses.
func fromQoverySecretManagerAccesses(ctx context.Context, accesses []qovery.SecretManagerAccess, planSet types.Set) types.Set {
	setType := types.ObjectType{AttrTypes: secretManagerAccessAttrTypes}

	// If the API returned nothing (nil = field absent), preserve the plan state to avoid noise.
	if accesses == nil {
		if planSet.IsUnknown() {
			return types.SetUnknown(setType)
		}
		return types.SetNull(setType)
	}

	// Empty slice + null plan: user removed the attribute and the API confirmed deletion.
	if len(accesses) == 0 && planSet.IsNull() {
		return types.SetNull(setType)
	}

	planByName := make(map[string]types.Object)
	if !planSet.IsNull() && !planSet.IsUnknown() {
		for _, elem := range planSet.Elements() {
			obj := elem.(types.Object)
			name := obj.Attributes()["name"].(types.String).ValueString()
			planByName[name] = obj
		}
	}

	elements := make([]attr.Value, 0, len(accesses))
	for _, access := range accesses {
		obj := types.ObjectValueMust(secretManagerAccessAttrTypes, map[string]attr.Value{
			"id":             types.StringValue(access.Id),
			"name":           types.StringValue(access.Name),
			"endpoint":       fromQoverySecretManagerEndpoint(access.Endpoint),
			"authentication": fromQoverySecretManagerAuth(access.Authentication, planByName[access.Name]),
		})
		elements = append(elements, obj)
	}

	set, diags := types.SetValueFrom(ctx, setType, elements)
	if diags.HasError() {
		return types.SetNull(setType)
	}
	return set
}

func fromQoverySecretManagerEndpoint(endpoint qovery.SecretManagerEndpointConfigurationDto) types.Object {
	vals := map[string]attr.Value{
		"type":       types.StringNull(),
		"region":     types.StringNull(),
		"project_id": types.StringNull(),
	}

	switch {
	case endpoint.AwsParameterStoreEndpointDto != nil:
		dto := endpoint.AwsParameterStoreEndpointDto
		vals["type"] = types.StringValue(dto.Mode)
		vals["region"] = types.StringValue(dto.Region)
	case endpoint.AwsSecretManagerEndpointDto != nil:
		dto := endpoint.AwsSecretManagerEndpointDto
		vals["type"] = types.StringValue(dto.Mode)
		vals["region"] = types.StringValue(dto.Region)
	case endpoint.GcpSecretManagerEndpointDto != nil:
		dto := endpoint.GcpSecretManagerEndpointDto
		vals["type"] = types.StringValue(dto.Mode)
		vals["region"] = types.StringValue(dto.Region)
		vals["project_id"] = types.StringValue(dto.ProjectId)
	}

	return types.ObjectValueMust(secretManagerEndpointAttrTypes, vals)
}

func fromQoverySecretManagerAuth(auth qovery.SecretManagerAuthenticationDto, planObj types.Object) types.Object {
	vals := map[string]attr.Value{
		"type":             types.StringNull(),
		"role_arn":         types.StringNull(),
		"region":           types.StringNull(),
		"access_key":       types.StringNull(),
		"secret_key":       types.StringNull(),
		"json_credentials": types.StringNull(),
	}

	// Sensitive fields may be redacted by the API; fall back to plan values.
	planSecretKey := types.StringNull()
	planJsonCreds := types.StringNull()
	if !planObj.IsNull() && !planObj.IsUnknown() {
		if planAuth, ok := planObj.Attributes()["authentication"]; ok {
			planAuthObj := planAuth.(types.Object)
			if !planAuthObj.IsNull() && !planAuthObj.IsUnknown() {
				if v, ok := planAuthObj.Attributes()["secret_key"]; ok {
					planSecretKey = v.(types.String)
				}
				if v, ok := planAuthObj.Attributes()["json_credentials"]; ok {
					planJsonCreds = v.(types.String)
				}
			}
		}
	}

	switch {
	case auth.AutomaticallyConfiguredAuthDto != nil:
		vals["type"] = types.StringValue(auth.AutomaticallyConfiguredAuthDto.Mode)
	case auth.AwsRoleArnAuthDto != nil:
		dto := auth.AwsRoleArnAuthDto
		vals["type"] = types.StringValue(dto.Mode)
		vals["role_arn"] = types.StringValue(dto.RoleArn)
	case auth.AwsStaticCredentialsAuthDto != nil:
		dto := auth.AwsStaticCredentialsAuthDto
		vals["type"] = types.StringValue(dto.Mode)
		vals["region"] = types.StringValue(dto.Region)
		vals["access_key"] = types.StringValue(dto.AccessKey)
		if sk, ok := dto.GetSecretKeyOk(); ok && sk != nil && *sk != "" {
			vals["secret_key"] = types.StringValue(*sk)
		} else {
			vals["secret_key"] = planSecretKey
		}
	case auth.GcpJsonCredentialsAuthDto != nil:
		dto := auth.GcpJsonCredentialsAuthDto
		vals["type"] = types.StringValue(dto.Mode)
		if jc, ok := dto.GetJsonCredentialsOk(); ok && jc != nil && *jc != "" {
			vals["json_credentials"] = types.StringValue(*jc)
		} else {
			vals["json_credentials"] = planJsonCreds
		}
	}

	return types.ObjectValueMust(secretManagerAuthAttrTypes, vals)
}
