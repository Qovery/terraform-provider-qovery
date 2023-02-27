package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &deploymentDataSource{}

type deploymentDataSource struct {
	deploymentService newdeployment.Service
}

func newDeploymentDataSource() datasource.DataSource {
	return &deploymentDataSource{}
}

func (d deploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (d *deploymentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.deploymentService = provider.deploymentService
}

func (d deploymentDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing deployment.",
		Attributes: map[string]tfsdk.Attribute{
			"environment_id": {
				Description: "Id of the environment to deploy.",
				Type:        types.StringType,
				Optional:    true,
			},
			"service_ids": {
				Description: "List of service ids to apply to the deployment.",
				Optional:    true,
				Computed:    true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringSliceDefaultModifier([]string{}),
				},
			},
			"desired_state": {
				Description: "Desired state of the deployment.",
				Type:        types.StringType,
				Optional:    true,
			},
		},
	}, nil
}

// Read qovery deployment data source
func (d deploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data NewDeploymentTerraform
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get deployment from API
	_, err := d.deploymentService.Get(ctx, newdeployment.NewDeploymentParams{
		EnvironmentId: data.EnvironmentId,
		ServiceIds:    data.ServiceIds,
		DesiredState:  data.DesiredState,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment read", err.Error())
		return
	}

	//state := convertDomainDeploymentToDeployment(deploymentDomain)

	// state is not recomputed
	state := data
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
