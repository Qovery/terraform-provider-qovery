package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

var _ datasource.DataSourceWithConfigure = &customRoleDataSource{}

type customRoleDataSource struct {
	service customrole.Service
}

func newCustomRoleDataSource() datasource.DataSource {
	return &customRoleDataSource{}
}

func (d customRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_role"
}

func (d *customRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.service = provider.customRoleService
}

func (d customRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery organization custom role. Returns the full permission matrix (every cluster and project of the organization).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the custom role.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the custom role.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the custom role.",
				Computed:    true,
			},
			"cluster_permissions": schema.SetNestedAttribute{
				Description: "Cluster permissions of the custom role (every cluster of the organization).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_id": schema.StringAttribute{Computed: true},
						"permission": schema.StringAttribute{Computed: true},
					},
				},
			},
			"project_permissions": schema.SetNestedAttribute{
				Description: "Project permissions of the custom role (every project of the organization).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{Computed: true},
						"is_admin":   schema.BoolAttribute{Computed: true},
						"permissions": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"environment_type": schema.StringAttribute{Computed: true},
									"permission":       schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d customRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CustomRole
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.service.Get(ctx, ToString(data.OrganizationId), ToString(data.Id))
	if err != nil {
		resp.Diagnostics.AddError("Error on custom role read", err.Error())
		return
	}

	state := convertDomainCustomRoleToCustomRole(role, nil, customRoleReadModeKeepAll)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
