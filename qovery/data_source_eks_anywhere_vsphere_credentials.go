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
var _ datasource.DataSourceWithConfigure = &eksAnywhereVsphereCredentialsDataSource{}

type eksAnywhereVsphereCredentialsDataSource struct {
	eksAnywhereVsphereCredentialsService credentials.EksAnywhereVsphereService
}

func newEksAnywhereVsphereCredentialsDataSource() datasource.DataSource {
	return &eksAnywhereVsphereCredentialsDataSource{}
}

func (d eksAnywhereVsphereCredentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eks_anywhere_vsphere_credentials"
}

func (d *eksAnywhereVsphereCredentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.eksAnywhereVsphereCredentialsService = provider.eksAnywhereVsphereCredentialsService
}

func (d eksAnywhereVsphereCredentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about existing Qovery EKS Anywhere vSphere credentials.",
		MarkdownDescription: "Use this data source to retrieve information about existing Qovery EKS Anywhere vSphere credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the EKS Anywhere vSphere credentials.",
				MarkdownDescription: "ID of the EKS Anywhere vSphere credentials to retrieve.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization.",
				MarkdownDescription: "ID of the organization containing the credentials.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the EKS Anywhere vSphere credentials.",
				MarkdownDescription: "Name of the EKS Anywhere vSphere credentials.",
				Computed:            true,
			},
		},
	}
}

// Read qovery eksAnywhereVsphereCredentials data source
func (d eksAnywhereVsphereCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EksAnywhereVsphereCredentialsDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := d.eksAnywhereVsphereCredentialsService.Get(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on eks anywhere vsphere credentials read", err.Error())
		return
	}

	state := convertDomainCredentialsToEksAnywhereVsphereCredentialsDataSource(creds)
	tflog.Trace(ctx, "read eks anywhere vsphere credentials", map[string]any{"credentials_id": state.Id.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
