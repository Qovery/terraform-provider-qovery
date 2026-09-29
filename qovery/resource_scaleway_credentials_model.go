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

// convertDomainCredentialsToScalewayCredentials reads the access key, project and organization
// from the API, so a value changed outside Terraform shows up in the plan. The API never returns
// the secret key, so it is kept from prior, the plan on apply and the state on refresh.
func convertDomainCredentialsToScalewayCredentials(creds *credentials.Credentials, prior ScalewayCredentials) ScalewayCredentials {
	return ScalewayCredentials{
		Id:                     FromString(creds.ID.String()),
		OrganizationId:         FromString(creds.OrganizationID.String()),
		Name:                   FromString(creds.Name),
		ScalewayProjectId:      credentialIdentifierFromAPI(prior.ScalewayProjectId, creds.Identifiers.ScalewayProjectID),
		ScalewayAccessKey:      credentialIdentifierFromAPI(prior.ScalewayAccessKey, creds.Identifiers.ScalewayAccessKey),
		ScalewaySecretKey:      prior.ScalewaySecretKey,
		ScalewayOrganizationId: credentialIdentifierFromAPI(prior.ScalewayOrganizationId, creds.Identifiers.ScalewayOrganizationID),
	}
}

func convertDomainCredentialsToScalewayCredentialsDataSource(creds *credentials.Credentials) ScalewayCredentialsDataSource {
	return ScalewayCredentialsDataSource{
		Id:             FromString(creds.ID.String()),
		OrganizationId: FromString(creds.OrganizationID.String()),
		Name:           FromString(creds.Name),
	}
}
