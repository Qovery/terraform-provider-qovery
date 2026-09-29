package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSource = &deploymentDataSource{}

// deploymentDataSource reads nothing from Qovery. A deployment is an action on an environment,
// not an object Qovery stores, and the qovery_deployment id is a UUID the provider generates,
// so there is no API value to report and Read echoes the configuration.
type deploymentDataSource struct{}

func newDeploymentDataSource() datasource.DataSource {
	return &deploymentDataSource{}
}

func (d deploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r deploymentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Qovery deployment. Qovery stores no deployment, so this data source only echoes its arguments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: deploymentIDDescription + " Echoed as configured.",
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: deploymentEnvironmentIDDescription + deploymentReadsNothingNote,
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: deploymentVersionDescription + " Echoed as configured.",
				Optional:            true,
				Computed:            false,
			},
			"desired_state": schema.StringAttribute{
				MarkdownDescription: deploymentDesiredStateDescription + deploymentReadsNothingNote,
				Computed:            true,
			},
		},
	}
}

// Read echoes the configuration: there is no deployment object to read (see deploymentDataSource).
func (d deploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NewDeploymentTerraform
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
