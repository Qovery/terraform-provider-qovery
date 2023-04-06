package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

type AWSCredentials struct {
	Id              types.String `tfsdk:"id"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	Name            types.String `tfsdk:"name"`
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

type AWSCredentialsDataSource struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

func (creds AWSCredentials) toUpsertAwsRequest() credentials.UpsertAwsRequest {
	return credentials.UpsertAwsRequest{
		Name:            ToString(creds.Name),
		AccessKeyID:     ToString(creds.AccessKeyId),
		SecretAccessKey: ToString(creds.SecretAccessKey),
	}
}

func convertDomainCredentialsToAWSCredentials(creds *credentials.Credentials, plan AWSCredentials) AWSCredentials {
	return AWSCredentials{
		Id:              FromString(creds.ID.String()),
		OrganizationId:  FromString(creds.OrganizationID.String()),
		Name:            FromString(creds.Name),
		AccessKeyId:     plan.AccessKeyId,
		SecretAccessKey: plan.SecretAccessKey,
	}
}

func convertDomainCredentialsToAWSCredentialsDataSource(creds *credentials.Credentials) AWSCredentialsDataSource {
	return AWSCredentialsDataSource{
		Id:             FromString(creds.ID.String()),
		OrganizationId: FromString(creds.OrganizationID.String()),
		Name:           FromString(creds.Name),
	}
}
