package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// AzureCredentialsDataSource represents the Terraform model for Azure credentials data source.
// Note: Azure credentials must be created via the Qovery console (provisioning requires server-side scripts).
// This data source provides read-only access to existing credentials.
type AzureCredentialsDataSource struct {
	Id                       types.String `tfsdk:"id"`
	OrganizationId           types.String `tfsdk:"organization_id"`
	Name                     types.String `tfsdk:"name"`
	AzureSubscriptionId      types.String `tfsdk:"azure_subscription_id"`
	AzureTenantId            types.String `tfsdk:"azure_tenant_id"`
	AzureApplicationId       types.String `tfsdk:"azure_application_id"`
	AzureApplicationObjectId types.String `tfsdk:"azure_application_object_id"`
}

// convertDomainAzureCredentialsToDataSource converts domain credentials to data source model.
func convertDomainAzureCredentialsToDataSource(creds *credentials.AzureCredentials) AzureCredentialsDataSource {
	return AzureCredentialsDataSource{
		Id:                       FromString(creds.ID.String()),
		OrganizationId:           FromString(creds.OrganizationID.String()),
		Name:                     FromString(creds.Name),
		AzureSubscriptionId:      FromString(creds.AzureSubscriptionId),
		AzureTenantId:            FromString(creds.AzureTenantId),
		AzureApplicationId:       FromString(creds.AzureApplicationId),
		AzureApplicationObjectId: FromString(creds.AzureApplicationObjectId),
	}
}
