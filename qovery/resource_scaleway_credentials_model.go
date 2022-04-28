package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type ScalewayCredentials struct {
	Id                types.String `tfsdk:"id"`
	OrganizationId    types.String `tfsdk:"organization_id"`
	Name              types.String `tfsdk:"name"`
	ScalewayAccessKey types.String `tfsdk:"scaleway_access_key"`
	ScalewaySecretKey types.String `tfsdk:"scaleway_secret_key"`
	ScalewayProjectId types.String `tfsdk:"scaleway_project_id"`
}

type ScalewayCredentialsDataSource struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

func (creds ScalewayCredentials) toUpsertScalewayCredentialsRequest() qovery.ScalewayCredentialsRequest {
	return qovery.ScalewayCredentialsRequest{
		Name:              toString(creds.Name),
		ScalewayAccessKey: toStringPointer(creds.ScalewayAccessKey),
		ScalewaySecretKey: toStringPointer(creds.ScalewaySecretKey),
		ScalewayProjectId: toStringPointer(creds.ScalewayProjectId),
	}
}

func convertResponseToScalewayCredentials(creds *qovery.ClusterCredentials, plan ScalewayCredentials) ScalewayCredentials {
	return ScalewayCredentials{
		Id:                fromStringPointer(creds.Id),
		Name:              fromStringPointer(creds.Name),
		OrganizationId:    plan.OrganizationId,
		ScalewayAccessKey: plan.ScalewayAccessKey,
		ScalewaySecretKey: plan.ScalewaySecretKey,
		ScalewayProjectId: plan.ScalewayProjectId,
	}
}

func convertResponseToScalewayCredentialsDataSource(creds *qovery.ClusterCredentials, plan ScalewayCredentialsDataSource) ScalewayCredentialsDataSource {
	return ScalewayCredentialsDataSource{
		Id:             fromStringPointer(creds.Id),
		Name:           fromStringPointer(creds.Name),
		OrganizationId: plan.OrganizationId,
	}
}
