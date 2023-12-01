package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

type ScalewayCredentials struct {
	Id                     types.String `tfsdk:"id"`
	OrganizationId         types.String `tfsdk:"organization_id"`
	Name                   types.String `tfsdk:"name"`
	ScalewayAccessKey      types.String `tfsdk:"scaleway_access_key"`
	ScalewaySecretKey      types.String `tfsdk:"scaleway_secret_key"`
	ScalewayProjectId      types.String `tfsdk:"scaleway_project_id"`
	ScalewayOrganizationId types.String `tfsdk:"scaleway_organization_id"`
}

type ScalewayCredentialsDataSource struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

func (creds ScalewayCredentials) toUpsertScalewayRequest() credentials.UpsertScalewayRequest {
	return credentials.UpsertScalewayRequest{
		Name:                   ToString(creds.Name),
		ScalewayProjectID:      ToString(creds.ScalewayProjectId),
		ScalewayAccessKey:      ToString(creds.ScalewayAccessKey),
		ScalewaySecretKey:      ToString(creds.ScalewaySecretKey),
		ScalewayOrganizationID: ToString(creds.ScalewayOrganizationId),
	}
}

func convertDomainCredentialsToScalewayCredentials(creds *credentials.Credentials, plan ScalewayCredentials) ScalewayCredentials {
	return ScalewayCredentials{
		Id:                     FromString(creds.ID.String()),
		OrganizationId:         FromString(creds.OrganizationID.String()),
		Name:                   FromString(creds.Name),
		ScalewayProjectId:      plan.ScalewayProjectId,
		ScalewayAccessKey:      plan.ScalewayAccessKey,
		ScalewaySecretKey:      plan.ScalewaySecretKey,
		ScalewayOrganizationId: plan.ScalewayOrganizationId,
	}
}

func convertDomainCredentialsToScalewayCredentialsDataSource(creds *credentials.Credentials) ScalewayCredentialsDataSource {
	return ScalewayCredentialsDataSource{
		Id:             FromString(creds.ID.String()),
		OrganizationId: FromString(creds.OrganizationID.String()),
		Name:           FromString(creds.Name),
	}
}
