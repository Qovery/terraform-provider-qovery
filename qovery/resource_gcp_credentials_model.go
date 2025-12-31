package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// GCPCredentials represents the Terraform model for GCP credentials resource.
type GCPCredentials struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	GcpCredentials types.String `tfsdk:"gcp_credentials"`
}

// GCPCredentialsDataSource represents the Terraform model for GCP credentials data source.
type GCPCredentialsDataSource struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

// toUpsertGcpRequest converts the Terraform model to a domain request.
func (creds GCPCredentials) toUpsertGcpRequest() credentials.UpsertGcpRequest {
	return credentials.UpsertGcpRequest{
		Name:           ToString(creds.Name),
		GcpCredentials: ToString(creds.GcpCredentials),
	}
}

// convertDomainCredentialsToGCPCredentials converts domain credentials to Terraform model.
// Note: GcpCredentials is write-only, so it must be preserved from the plan.
func convertDomainCredentialsToGCPCredentials(creds *credentials.Credentials, plan GCPCredentials) GCPCredentials {
	return GCPCredentials{
		Id:             FromString(creds.ID.String()),
		OrganizationId: FromString(creds.OrganizationID.String()),
		Name:           FromString(creds.Name),
		GcpCredentials: plan.GcpCredentials,
	}
}

// convertDomainCredentialsToGCPCredentialsDataSource converts domain credentials to data source model.
func convertDomainCredentialsToGCPCredentialsDataSource(creds *credentials.Credentials) GCPCredentialsDataSource {
	return GCPCredentialsDataSource{
		Id:             FromString(creds.ID.String()),
		OrganizationId: FromString(creds.OrganizationID.String()),
		Name:           FromString(creds.Name),
	}
}
