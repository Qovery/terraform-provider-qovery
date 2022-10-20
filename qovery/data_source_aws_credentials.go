package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &awsCredentialsDataSource{}

type awsCredentialsDataSource struct {
	awsCredentialsService credentials.AwsService
}

func newAwsCredentialsDataSource() datasource.DataSource {
	return &awsCredentialsDataSource{}
}

func (d awsCredentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_credentials"
}

func (d *awsCredentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.awsCredentialsService = provider.awsCredentialsService
}

func (d awsCredentialsDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

// Read qovery awsCredentials data source
func (d awsCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
