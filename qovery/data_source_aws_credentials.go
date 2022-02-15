package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"
)

type awsCredentialsDataSourceType struct{}

func (t awsCredentialsDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing aws credentials.",
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
				Description: "Name of the aws credentials.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (t awsCredentialsDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return awsCredentialsDataSource{
		client: p.(*provider).GetClient(),
	}, nil
}

type awsCredentialsDataSource struct {
	client *qovery.APIClient
}

// Read qovery awsCredentials data source
func (d awsCredentialsDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data AWSCredentialsDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credentials from API
	credentials, res, err := d.client.CloudProviderCredentialsApi.
		ListAWSCredentials(ctx, data.OrganizationId.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := awsCredentialsReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	var state AWSCredentialsDataSource
	found := false
	for _, creds := range credentials.GetResults() {
		if data.Id.Value == *creds.Id {
			found = true
			state = convertResponseToAWSCredentialsDataSource(&creds, data)
			break
		}
	}

	// If credential id is not in list
	// Returning Not Found error
	if !found {
		res.StatusCode = 404
		apiErr := awsCredentialsReadAPIError(state.Id.Value, res, nil)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "read aws credentials", "credentials_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
