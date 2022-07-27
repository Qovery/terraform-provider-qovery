package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
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
		awsCredentialsService: p.(*provider).awsCredentialsService,
	}, nil
}

type awsCredentialsDataSource struct {
	awsCredentialsService credentials.AwsService
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
	creds, err := d.awsCredentialsService.Get(ctx, data.OrganizationId.Value, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials read", err.Error())
		return
	}

	state := convertDomainCredentialsToAWSCredentialsDataSource(creds)
	tflog.Trace(ctx, "read aws credentials", map[string]interface{}{"credentials_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
