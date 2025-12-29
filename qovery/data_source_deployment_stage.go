package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
func (r deploymentStageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery deployment stage resource. This can be used to create and manage Qovery deployment stages.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the deployment stage.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the deployment stage.",
				Optional:    true,
				Computed:    true,
			},
			"is_after": schema.StringAttribute{
				Description: "Move the current deployment stage after the target deployment stage",
				Optional:    true,
				Computed:    true,
			},
			"is_before": schema.StringAttribute{
				Description: "Move the current deployment stage before the target deployment stage",
				Optional:    true,
				Computed:    true,
			},
		},
	}
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
	deploymentStageDomain, err := d.deploymentStageService.Get(ctx, data.EnvironmentId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage read", err.Error())
		return
	}

	newState := convertDomainDeploymentStageToDeploymentStage(deploymentStageDomain, data.Description)
	tflog.Trace(ctx, "read deployment stage", map[string]interface{}{"deployment_stage_id": data.Id.ValueString()})

	// We need to keep the 'IsAfter' and 'IsBefore' properties
	newState = DeploymentStage{
		Id:            newState.Id,
		EnvironmentId: newState.EnvironmentId,
		Name:          newState.Name,
		Description:   newState.Description,
		IsAfter:       data.IsAfter,
		IsBefore:      data.IsBefore,
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}
