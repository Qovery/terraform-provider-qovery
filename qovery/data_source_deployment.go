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
		Description: "Echoes its arguments. Qovery stores no deployment object: a deployment is an action on an environment, " +
			"so this data source reads nothing from Qovery. Use the qovery_deployment resource to deploy an environment.",
		MarkdownDescription: "Echoes its arguments. Qovery stores no deployment object: a deployment is an action on an environment, " +
			"so this data source reads nothing from Qovery. Use the `qovery_deployment` resource to deploy an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Unique identifier of the deployment (UUID format). Echoed as configured.",
				MarkdownDescription: "Unique identifier of the deployment (UUID format). Echoed as configured.",
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				Description:         "Always null: this data source reads nothing from Qovery.",
				MarkdownDescription: "Always `null`: this data source reads nothing from Qovery.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				Description:         "Version identifier of the deployment. Echoed as configured.",
				MarkdownDescription: "Version identifier of the deployment. Echoed as configured.",
				Optional:            true,
				Computed:            false,
			},
			"desired_state": schema.StringAttribute{
				Description:         "Always null: this data source reads nothing from Qovery.",
				MarkdownDescription: "Always `null`: this data source reads nothing from Qovery.",
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
