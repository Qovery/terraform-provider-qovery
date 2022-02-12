package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type scalewayCredentialsDataSourceData struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

type scalewayCredentialsDataSourceType struct{}

func (t scalewayCredentialsDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (t scalewayCredentialsDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return scalewayCredentialsDataSource{
		client: p.(*provider).GetClient(),
	}, nil
}

type scalewayCredentialsDataSource struct {
	client *qovery.APIClient
}

// Read qovery scalewayCredentials data source
func (d scalewayCredentialsDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data scalewayCredentialsDataSourceData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API
	credentials, res, err := d.client.CloudProviderCredentialsApi.
		ListScalewayCredentials(ctx, data.OrganizationId.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := scalewayCredentialsReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	var state *scalewayCredentialsDataSourceData
	for _, creds := range credentials.GetResults() {
		if data.Id.Value == *creds.Id {
			state = &scalewayCredentialsDataSourceData{
				Id:             data.Id,
				OrganizationId: data.OrganizationId,
				Name:           types.String{Value: *creds.Name},
			}
			break
		}
	}

	// If credential id is not in list
	// Returning Not Found error
	if state == nil {
		res.StatusCode = 404
		apiErr := scalewayCredentialsReadAPIError(state.Id.Value, res, nil)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
