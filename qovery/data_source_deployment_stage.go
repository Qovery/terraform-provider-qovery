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
		MarkdownDescription: "Reads an existing Qovery deployment stage.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("deployment stage"),
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("deployment stage"),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("deployment stage"),
				Optional:            true,
				Computed:            true,
			},
			"is_after": schema.StringAttribute{
				MarkdownDescription: deploymentStageIsAfterDescription + " The API does not return it, so this data source echoes the configured value.",
				Optional:            true,
				Computed:            true,
			},
			"is_before": schema.StringAttribute{
				MarkdownDescription: deploymentStageIsBeforeDescription + " The API does not return it, so this data source echoes the configured value.",
				Optional:            true,
				Computed:            true,
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
	tflog.Trace(ctx, "read deployment stage", map[string]any{"deployment_stage_id": data.Id.ValueString()})

	// is_after and is_before echo the configuration: they are write-only move instructions, and
	// the API only returns the stage's deployment_order, not its neighbours.
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
