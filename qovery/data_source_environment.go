package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type environmentDataSourceData struct {
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	ClusterId types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`
	Mode      types.String `tfsdk:"mode"`
}

type environmentDataSourceType struct{}

func (t environmentDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing environment.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"project_id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Computed:    true,
			},
			"cluster_id": {
				Description: "Id of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"mode": {
				Description: "Mode of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (t environmentDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return environmentDataSource{
		client: p.(*provider).GetClient(),
	}, nil
}

type environmentDataSource struct {
	client *qovery.APIClient
}

// Read qovery environment data source
func (d environmentDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data environmentDataSourceData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get environment from API
	environment, res, err := d.client.EnvironmentMainCallsApi.
		GetEnvironment(ctx, data.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := environmentReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := environmentDataSourceData{
		Id:        data.Id,
		ProjectId: types.String{Value: environment.Project.Id},
		ClusterId: types.String{Value: environment.ClusterId},
		Name:      types.String{Value: environment.Name},
		Mode:      types.String{Value: environment.Mode},
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
