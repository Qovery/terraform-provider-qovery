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

func toQoverySecretManagerAccessRequests(list types.List) ([]qovery.SecretManagerAccessRequest, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	requests := make([]qovery.SecretManagerAccessRequest, 0, len(list.Elements()))
	for _, elem := range list.Elements() {
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
	endpointType := attrs["type"].(types.String).ValueString()
	region := attrs["region"].(types.String).ValueString()
	projectId := attrs["project_id"].(types.String).ValueString()

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
		return qovery.GcpSecretManagerEndpointDtoAsSecretManagerEndpointConfigurationDto(&qovery.GcpSecretManagerEndpointDto{
			Mode:      endpointType,
			Region:    region,
			ProjectId: projectId,
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
	authType := attrs["type"].(types.String).ValueString()

	switch authType {
	case "AUTOMATICALLY_CONFIGURED":
		return qovery.AutomaticallyConfiguredAuthDtoAsSecretManagerAuthenticationDto(&qovery.AutomaticallyConfiguredAuthDto{
			Mode: authType,
		}), nil
	case "AWS_ROLE_ARN":
		return qovery.AwsRoleArnAuthDtoAsSecretManagerAuthenticationDto(&qovery.AwsRoleArnAuthDto{
			Mode:    authType,
			RoleArn: attrs["role_arn"].(types.String).ValueString(),
		}), nil
	case "AWS_STATIC_CREDENTIALS":
		dto := &qovery.AwsStaticCredentialsAuthDto{
			Mode:      authType,
			Region:    attrs["region"].(types.String).ValueString(),
			AccessKey: attrs["access_key"].(types.String).ValueString(),
		}
		if sk := attrs["secret_key"].(types.String); !sk.IsNull() && !sk.IsUnknown() && sk.ValueString() != "" {
			dto.SetSecretKey(sk.ValueString())
		}
		return qovery.AwsStaticCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto), nil
	case "GCP_JSON_CREDENTIALS":
		dto := qovery.NewGcpJsonCredentialsAuthDto(authType)
		if jc := attrs["json_credentials"].(types.String); !jc.IsNull() && !jc.IsUnknown() && jc.ValueString() != "" {
			dto.SetJsonCredentials(jc.ValueString())
		}
		return qovery.GcpJsonCredentialsAuthDtoAsSecretManagerAuthenticationDto(dto), nil
	default:
		return qovery.SecretManagerAuthenticationDto{}, errors.Errorf("unsupported authentication type: %q", authType)
	}
}

// fromQoverySecretManagerAccesses converts API response objects to a Terraform list.
// planList is used as fallback for sensitive fields the API may redact in responses.
func fromQoverySecretManagerAccesses(ctx context.Context, accesses []qovery.SecretManagerAccess, planList types.List) types.List {
	listType := types.ObjectType{AttrTypes: secretManagerAccessAttrTypes}

	if accesses == nil {
		return types.ListNull(listType)
	}

	planByName := make(map[string]types.Object)
	if !planList.IsNull() && !planList.IsUnknown() {
		for _, elem := range planList.Elements() {
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

	list, diags := types.ListValueFrom(ctx, listType, elements)
	if diags.HasError() {
		return types.ListNull(listType)
	}
	return list
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
