package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &deploymentStageDataSource{}

type deploymentStageDataSource struct {
	deploymentStageService deploymentstage.Service
}

func newDeploymentStageDataSource() datasource.DataSource {
	return &deploymentStageDataSource{}
}

func (d deploymentStageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_stage"
}

func (d *deploymentStageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.deploymentStageService = provider.deploymentStageService
}

func (d deploymentStageDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing deployment stage.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the deployment stage.",
				Type:        types.StringType,
				Required:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the deployment stage.",
				Type:        types.StringType,
				Computed:    true,
			},
			"description": {
				Description: "Description of the deployment stage.",
				Type:        types.StringType,
				Computed:    true,
			},
			"service_ids": {
				Description: "Services associated with the deployment stage",
				Computed:    true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
			},
		},
	}, nil
}

// Read qovery deployment stage data source
func (d deploymentStageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data DeploymentStage
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get deployment stage from API
	deploymentStageDomain, err := d.deploymentStageService.Get(ctx, data.EnvironmentId.Value, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage read", err.Error())
		return
	}

	state := convertDomainDeploymentStageToDeploymentStage(deploymentStageDomain)
	tflog.Trace(ctx, "read deployment stage", map[string]interface{}{"deployment_stage_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
