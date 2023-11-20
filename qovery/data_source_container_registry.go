package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &containerRegistryDataSource{}

type containerRegistryDataSource struct {
	containerRegistryService registry.Service
}

func newContainerRegistryDataSource() datasource.DataSource {
	return &containerRegistryDataSource{}
}
func (d containerRegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_registry"
}

func (d *containerRegistryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.containerRegistryService = provider.containerRegistryService
}

func (r containerRegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery container registry resource. This can be used to create and manage Qovery container registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Id of the container registry.",
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the container registry.",
				Computed:    true,
			},
			"kind": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Kind of the container registry.",
					registryKinds,
					nil,
				),
				Computed: true,
			},
			"url": schema.StringAttribute{
				Description: "URL of the container registry.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the container registry.",
				Optional:    true,
				Computed:    true,
			},
			"config": schema.SingleNestedAttribute{
				Description: "Configuration needed to authenticate the container registry.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"access_key_id": schema.StringAttribute{
						Description: "Required if kind is `ECR` or `PUBLIC_ECR`.",
						Optional:    true,
						Computed:    true,
					},
					"secret_access_key": schema.StringAttribute{
						Description: "Required if kind is `ECR` or `PUBLIC_ECR`.",
						Optional:    true,
						Computed:    true,
					},
					"region": schema.StringAttribute{
						Description: "Required if kind is `ECR` or `SCALEWAY_CR`.",
						Computed:    true,
						Optional:    true,
					},
					"scaleway_access_key": schema.StringAttribute{
						Description: "Required if kind is `SCALEWAY_CR`.",
						Optional:    true,
						Computed:    true,
					},
					"scaleway_secret_key": schema.StringAttribute{
						Description: "Required if kind is `SCALEWAY_CR`.",
						Optional:    true,
						Computed:    true,
					},
					"username": schema.StringAttribute{
						Description: "Required if kinds are `DOCKER_HUB`, `GITHUB_CR`, `GITLAB`CR`, `GENERIC_CR`.",
						Optional:    true,
						Computed:    true,
					},
					"password": schema.StringAttribute{
						Description: "Required if kinds are `DOCKER_HUB`, `GITHUB_CR`, `GITLAB`CR`, `GENERIC_CR`.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

// Read qovery container registry data source
func (d containerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data ContainerRegistryDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get container registry from API
	reg, err := d.containerRegistryService.Get(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on container registry read", err.Error())
		return
	}

	state := convertDomainRegistryToContainerRegistryDataSource(reg)
	tflog.Trace(ctx, "read container registry", map[string]interface{}{"container_registry_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
