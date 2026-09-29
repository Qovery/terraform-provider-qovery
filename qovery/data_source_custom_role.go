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
		MarkdownDescription: "Reads an existing Qovery custom role, with its permissions on every cluster and project of the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("custom role"),
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription,
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("custom role"),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("custom role"),
				Computed:            true,
			},
			"cluster_permissions": schema.SetNestedAttribute{
				MarkdownDescription: customRoleClusterPermissionsDescription + " Lists every cluster of the organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_id": schema.StringAttribute{MarkdownDescription: customRoleClusterIDDescription, Computed: true},
						"permission": schema.StringAttribute{MarkdownDescription: customRoleClusterPermissionDescription, Computed: true},
					},
				},
			},
			"project_permissions": schema.SetNestedAttribute{
				MarkdownDescription: customRoleProjectPermissionsDescription + " Lists every project of the organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{MarkdownDescription: customRoleProjectIDDescription, Computed: true},
						"is_admin":   schema.BoolAttribute{MarkdownDescription: customRoleIsAdminDescription, Computed: true},
						"permissions": schema.SetNestedAttribute{
							MarkdownDescription: customRoleEnvironmentPermissionsDescription + " `null` when `is_admin` is `true`.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"environment_type": schema.StringAttribute{MarkdownDescription: customRoleEnvironmentTypeDescription, Computed: true},
									"permission":       schema.StringAttribute{MarkdownDescription: customRoleEnvironmentPermissionDescription, Computed: true},
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
