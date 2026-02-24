package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (r awsCredentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Qovery AWS credentials.",
		MarkdownDescription: "Use this data source to retrieve information about existing Qovery AWS credentials. This is useful for referencing credentials created outside of Terraform or in a different Terraform configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the AWS credentials.",
				MarkdownDescription: "ID of the AWS credentials to retrieve.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization.",
				MarkdownDescription: "ID of the organization containing the credentials.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the aws credentials.",
				MarkdownDescription: "Name of the AWS credentials.",
				Computed:            true,
			},
		},
	}
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
	creds, err := d.awsCredentialsService.Get(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on aws credentials read", err.Error())
		return
	}

	state := convertDomainCredentialsToAWSCredentialsDataSource(creds)
	tflog.Trace(ctx, "read aws credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
