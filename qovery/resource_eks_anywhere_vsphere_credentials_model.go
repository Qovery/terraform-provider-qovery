package qovery

import (
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

type EksAnywhereVsphereCredentials struct {
	Id              types.String `tfsdk:"id"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	Name            types.String `tfsdk:"name"`
	VsphereUser     types.String `tfsdk:"vsphere_user"`
	VspherePassword types.String `tfsdk:"vsphere_password"`
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	RoleArn         types.String `tfsdk:"role_arn"`
}

type EksAnywhereVsphereCredentialsDataSource struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

var (
	errEksAnywhereVsphereMutuallyExclusiveAuthMethods = errors.New("`role_arn` cannot be set with `access_key_id` or `secret_access_key`; choose one authentication method")
	errEksAnywhereVsphereStaticCredentialsPair        = errors.New("`access_key_id` and `secret_access_key` must be set together")
	errEksAnywhereVsphereMissingAuthMethod            = errors.New("either `role_arn` or both `access_key_id` and `secret_access_key` must be set")
)

func (creds EksAnywhereVsphereCredentials) toUpsertEksAnywhereVsphereRequest() (*credentials.UpsertEksAnywhereVsphereRequest, error) {
	hasRoleArn := hasNonEmptyStringValue(creds.RoleArn)
	hasAccessKeyID := hasNonEmptyStringValue(creds.AccessKeyId)
	hasSecretAccessKey := hasNonEmptyStringValue(creds.SecretAccessKey)

	if hasRoleArn && (hasAccessKeyID || hasSecretAccessKey) {
		return nil, errEksAnywhereVsphereMutuallyExclusiveAuthMethods
	}

	if hasRoleArn {
		return &credentials.UpsertEksAnywhereVsphereRequest{
			Name:            ToString(creds.Name),
			VsphereUser:     ToString(creds.VsphereUser),
			VspherePassword: ToString(creds.VspherePassword),
			RoleCredentials: &credentials.VsphereRoleCredentials{
				RoleArn: ToString(creds.RoleArn),
			},
		}, nil
	}

	if hasAccessKeyID != hasSecretAccessKey {
		return nil, errEksAnywhereVsphereStaticCredentialsPair
	}

	if hasAccessKeyID {
		return &credentials.UpsertEksAnywhereVsphereRequest{
			Name:            ToString(creds.Name),
			VsphereUser:     ToString(creds.VsphereUser),
			VspherePassword: ToString(creds.VspherePassword),
			StaticCredentials: &credentials.VsphereStaticCredentials{
				AccessKeyID:     ToString(creds.AccessKeyId),
				SecretAccessKey: ToString(creds.SecretAccessKey),
			},
		}, nil
	}

	return nil, errEksAnywhereVsphereMissingAuthMethod
}

func hasNonEmptyStringValue(value types.String) bool {
	if value.IsNull() || value.IsUnknown() {
		return false
	}

	return strings.TrimSpace(value.ValueString()) != ""
}

func convertDomainCredentialsToEksAnywhereVsphereCredentials(creds *credentials.Credentials, plan EksAnywhereVsphereCredentials) EksAnywhereVsphereCredentials {
	return EksAnywhereVsphereCredentials{
		Id:              FromString(creds.ID.String()),
		OrganizationId:  FromString(creds.OrganizationID.String()),
		Name:            FromString(creds.Name),
		VsphereUser:     plan.VsphereUser,
		VspherePassword: plan.VspherePassword,
		AccessKeyId:     plan.AccessKeyId,
		SecretAccessKey: plan.SecretAccessKey,
		RoleArn:         plan.RoleArn,
	}
}

func convertDomainCredentialsToEksAnywhereVsphereCredentialsDataSource(creds *credentials.Credentials) EksAnywhereVsphereCredentialsDataSource {
	return EksAnywhereVsphereCredentialsDataSource{
		Id:             FromString(creds.ID.String()),
		OrganizationId: FromString(creds.OrganizationID.String()),
		Name:           FromString(creds.Name),
	}
}
