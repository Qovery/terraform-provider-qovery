package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

var _ datasource.DataSourceWithConfigure = &blueprintDataSource{}

type blueprintDataSource struct {
	service blueprint.Service
}

// BlueprintData holds what the API returns: icon, spec overrides and secret values are not readable.
type BlueprintData struct {
	ID                  types.String `tfsdk:"id"`
	EnvironmentID       types.String `tfsdk:"environment_id"`
	Blueprint           types.String `tfsdk:"blueprint"`
	Name                types.String `tfsdk:"name"`
	Tag                 types.String `tfsdk:"tag"`
	Variables           types.Map    `tfsdk:"variables"`
	SecretVariableNames types.Set    `tfsdk:"secret_variable_names"`
	ServiceID           types.String `tfsdk:"service_id"`
	ServiceType         types.String `tfsdk:"service_type"`
	CatalogURL          types.String `tfsdk:"catalog_url"`
}

func newBlueprintDataSource() datasource.DataSource {
	return &blueprintDataSource{}
}

func (d blueprintDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (d *blueprintDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.service = provider.blueprintService
}

func (d blueprintDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery blueprint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the blueprint.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Computed:    true,
			},
			"blueprint": schema.StringAttribute{
				Description: "Catalog entry of the blueprint, as `<provider>/<service_family>/<service_version>`.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the blueprint service.",
				Computed:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Catalog tag identifying the blueprint and its version.",
				Computed:    true,
			},
			"variables": schema.MapAttribute{
				Description: "Non-secret blueprint variables, catalog defaults included.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"secret_variable_names": schema.SetAttribute{
				Description: "Names of the secret blueprint variables. The API never returns their values.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"service_id": schema.StringAttribute{
				Description: "Id of the terraform or helm service the blueprint materialized.",
				Computed:    true,
			},
			"service_type": schema.StringAttribute{
				Description: "Type of the service the blueprint materialized: `TERRAFORM` or `HELM`.",
				Computed:    true,
			},
			"catalog_url": schema.StringAttribute{
				Description: "URL of the blueprint catalog entry.",
				Computed:    true,
			},
		},
	}
}

func (d blueprintDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BlueprintData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bp, err := d.service.Get(ctx, ToString(data.ID))
	if err != nil {
		resp.Diagnostics.AddError("Error on blueprint read", err.Error())
		return
	}

	state, diags := convertDomainBlueprintToBlueprintData(ctx, bp)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func convertDomainBlueprintToBlueprintData(ctx context.Context, bp *blueprint.Blueprint) (BlueprintData, diag.Diagnostics) {
	var diags diag.Diagnostics
	variables := map[string]string{}
	secretVariableNames := []string{}
	for _, v := range bp.Variables {
		switch {
		case v.IsSecret:
			secretVariableNames = append(secretVariableNames, v.Name)
		case v.Name != blueprintImportIdentifierVariable && v.Value != nil:
			variables[v.Name] = *v.Value
		}
	}

	variablesValue, d := types.MapValueFrom(ctx, types.StringType, variables)
	diags.Append(d...)
	secretVariableNamesValue, d := types.SetValueFrom(ctx, types.StringType, secretVariableNames)
	diags.Append(d...)

	return BlueprintData{
		ID:                  FromString(bp.ID.String()),
		EnvironmentID:       FromString(bp.EnvironmentID.String()),
		Blueprint:           blueprintVersionValue(types.StringNull(), bp.Tag),
		Name:                FromString(bp.Name),
		Tag:                 FromString(bp.Tag),
		Variables:           variablesValue,
		SecretVariableNames: secretVariableNamesValue,
		ServiceID:           FromStringPointer(bp.ServiceID),
		ServiceType:         FromString(string(bp.ServiceType)),
		CatalogURL:          FromString(bp.CatalogURL),
	}, diags
}
