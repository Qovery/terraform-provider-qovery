package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &organizationMemberDataSource{}

type organizationMemberDataSource struct {
	service member.Service
}

func newOrganizationMemberDataSource() datasource.DataSource {
	return &organizationMemberDataSource{}
}

func (d organizationMemberDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (d *organizationMemberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.service = provider.organizationMemberService
}

func (d organizationMemberDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Provides a Qovery organization member data source. This can be used to read an existing member (or pending invitation) of a Qovery organization by email.",
		MarkdownDescription: "Use this data source to retrieve an existing member (or pending invitation) of a Qovery organization by email.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the member. Invitation id while the invitation is pending; user id once accepted.",
				MarkdownDescription: "Id of the member. Invitation id while the invitation is pending; user id once accepted.",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization.",
				MarkdownDescription: "Id of the organization.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				Description:         "Email of the member.",
				MarkdownDescription: "Email of the member.",
				Required:            true,
			},
			"role_id": schema.StringAttribute{
				Description:         "Id of the role assigned to the member.",
				MarkdownDescription: "Id of the role assigned to the member.",
				Computed:            true,
			},
			"user_id": schema.StringAttribute{
				Description:         "User id of the member. Null until the invitation is accepted.",
				MarkdownDescription: "User id of the member. Null until the invitation is accepted.",
				Computed:            true,
			},
			"invitation_status": schema.StringAttribute{
				Description:         "Status of the invitation: PENDING, EXPIRED or ACCEPTED.",
				MarkdownDescription: "Status of the invitation: `PENDING`, `EXPIRED` or `ACCEPTED`.",
				Computed:            true,
			},
		},
	}
}

// Read qovery organization member data source
func (d organizationMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data OrganizationMember
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get member from API
	domainMember, err := d.service.Get(ctx, data.OrganizationId.ValueString(), data.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on organization member read", err.Error())
		return
	}

	state := convertDomainMemberToOrganizationMember(*domainMember, data.Email)
	tflog.Trace(ctx, "read organization member", map[string]any{"organization_member_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
