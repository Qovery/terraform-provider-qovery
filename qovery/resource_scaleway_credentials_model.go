package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
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

func (creds ScalewayCredentials) toUpsertScalewayRequest() credentials.UpsertScalewayRequest {
	return credentials.UpsertScalewayRequest{
		Name:              toString(creds.Name),
		ScalewayProjectID: toString(creds.ScalewayProjectId),
		ScalewayAccessKey: toString(creds.ScalewayAccessKey),
		ScalewaySecretKey: toString(creds.ScalewaySecretKey),
	}
}

func convertDomainCredentialsToScalewayCredentials(creds *credentials.Credentials, plan ScalewayCredentials) ScalewayCredentials {
	return ScalewayCredentials{
		Id:                fromString(creds.ID.String()),
		OrganizationId:    fromString(creds.OrganizationID.String()),
		Name:              fromString(creds.Name),
		ScalewayProjectId: plan.ScalewayProjectId,
		ScalewayAccessKey: plan.ScalewayAccessKey,
		ScalewaySecretKey: plan.ScalewaySecretKey,
	}
}

func convertDomainCredentialsToScalewayCredentialsDataSource(creds *credentials.Credentials) ScalewayCredentialsDataSource {
	return ScalewayCredentialsDataSource{
		Id:             fromString(creds.ID.String()),
		OrganizationId: fromString(creds.OrganizationID.String()),
		Name:           fromString(creds.Name),
	}
}
