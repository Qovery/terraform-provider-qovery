package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type databaseDataSourceData struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	Version       types.String `tfsdk:"version"`
	Mode          types.String `tfsdk:"mode"`
	Accessibility types.String `tfsdk:"accessibility"`
	CPU           types.Int64  `tfsdk:"cpu"`
	Memory        types.Int64  `tfsdk:"memory"`
	Storage       types.Int64  `tfsdk:"storage"`
}

type databaseDataSourceType struct{}

func (t databaseDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing database.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the database.",
				Type:        types.StringType,
				Required:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"type": {
				Description: "Type of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"version": {
				Description: "Version of the database",
				Type:        types.StringType,
				Computed:    true,
			},
			"mode": {
				Description: "Mode of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"accessibility": {
				Description: "Accessibility of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"cpu": {
				Description: "CPU of the database in millicores (m) [1000m = 1 CPU].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"memory": {
				Description: "RAM of the database in MB [1024MB = 1GB].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"storage": {
				Description: "Storage of the database in MB [1024MB = 1GB].",
				Type:        types.Int64Type,
				Computed:    true,
			},
		},
	}, nil
}

func (t databaseDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return databaseDataSource{
		client: p.(*provider).GetClient(),
	}, nil
}

type databaseDataSource struct {
	client *qovery.APIClient
}

// Read qovery database data source
func (d databaseDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data databaseDataSourceData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get database from API
	database, res, err := d.client.DatabaseMainCallsApi.
		GetDatabase(ctx, data.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := &databaseResourceData{
		Id:            data.Id,
		EnvironmentId: types.String{Value: database.Environment.Id},
		Name:          types.String{Value: database.Name},
		Type:          types.String{Value: database.Type},
		Version:       types.String{Value: database.Version},
		Mode:          types.String{Value: database.Mode},
		Accessibility: types.String{Value: *database.Accessibility},
		CPU:           types.Int64{Value: int64(*database.Cpu)},
		Memory:        types.Int64{Value: int64(*database.Memory)},
		Storage:       types.Int64{Value: int64(*database.Storage)},
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
