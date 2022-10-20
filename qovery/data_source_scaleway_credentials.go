package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSource = scalewayCredentialsDataSource{}

type scalewayCredentialsDataSource struct {
	scalewayCredentialsService credentials.ScalewayService
}

func NewScalewayCredentialsDataSource(service credentials.ScalewayService) func() datasource.DataSource {
	return func() datasource.DataSource {
		return scalewayCredentialsDataSource{
			scalewayCredentialsService: service,
		}
	}
}

func (d scalewayCredentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scaleway_credentials"
}

func (d scalewayCredentialsDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing Scaleway credentials.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the credentials.",
				Type:        types.StringType,
				Required:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the Scaleway credentials.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

// Read qovery scalewayCredentials data source
func (d scalewayCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data ScalewayCredentialsDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API
	creds, err := d.scalewayCredentialsService.Get(ctx, data.OrganizationId.Value, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on scaleway credentials read", err.Error())
		return
	}

	state := convertDomainCredentialsToScalewayCredentialsDataSource(creds)
	tflog.Trace(ctx, "read scaleway credentials", map[string]interface{}{"credentials_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
