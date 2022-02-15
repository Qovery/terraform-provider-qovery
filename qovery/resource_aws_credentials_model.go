package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
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

func (creds AWSCredentials) toUpsertAWSCredentialsRequest() qovery.AwsCredentialsRequest {
	return qovery.AwsCredentialsRequest{
		Name:            toString(creds.Name),
		AccessKeyId:     toStringPointer(creds.AccessKeyId),
		SecretAccessKey: toStringPointer(creds.SecretAccessKey),
	}
}

func convertResponseToAWSCredentials(creds *qovery.ClusterCredentialsResponse, plan AWSCredentials) AWSCredentials {
	return AWSCredentials{
		Id:              fromStringPointer(creds.Id),
		Name:            fromStringPointer(creds.Name),
		OrganizationId:  plan.OrganizationId,
		AccessKeyId:     plan.AccessKeyId,
		SecretAccessKey: plan.SecretAccessKey,
	}
}

func convertResponseToAWSCredentialsDataSource(creds *qovery.ClusterCredentialsResponse, plan AWSCredentialsDataSource) AWSCredentialsDataSource {
	return AWSCredentialsDataSource{
		Id:             fromStringPointer(creds.Id),
		Name:           fromStringPointer(creds.Name),
		OrganizationId: plan.OrganizationId,
	}
}
