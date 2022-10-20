package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSource = containerDataSource{}

type containerDataSource struct {
	containerService container.Service
}

func NewContainerDataSource(service container.Service) func() datasource.DataSource {
	return func() datasource.DataSource {
		return containerDataSource{
			containerService: service,
		}
	}
}

func (d containerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

func (d containerDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing container.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the container.",
				Type:        types.StringType,
				Required:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"registry_id": {
				Description: "Id of the registry.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the container.",
				Type:        types.StringType,
				Computed:    true,
			},
			"image_name": {
				Description: "Name of the container image.",
				Type:        types.StringType,
				Computed:    true,
			},
			"tag": {
				Description: "Tag of the container image.",
				Type:        types.StringType,
				Computed:    true,
			},
			"cpu": {
				Description: "CPU of the container in millicores (m) [1000m = 1 CPU].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"memory": {
				Description: "RAM of the container in MB [1024MB = 1GB].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"min_running_instances": {
				Description: "Minimum number of instances running for the container.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"max_running_instances": {
				Description: "Maximum number of instances running for the container.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"auto_preview": {
				Description: "Specify if the environment preview option is activated or not for this container.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"entrypoint": {
				Description: "Entrypoint of the container.",
				Type:        types.StringType,
				Computed:    true,
			},
			"storage": {
				Description: "List of storages linked to this container.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the storage.",
						Type:        types.StringType,
						Computed:    true,
					},
					"type": {
						Description: "Type of the storage for the container.",
						Type:        types.StringType,
						Computed:    true,
					},
					"size": {
						Description: "Size of the storage for the container in GB [1024MB = 1GB].",
						Type:        types.Int64Type,
						Computed:    true,
					},
					"mount_point": {
						Description: "Mount point of the storage for the container.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"ports": {
				Description: "List of storages linked to this container.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the port.",
						Type:        types.StringType,
						Computed:    true,
					},
					"name": {
						Description: "Name of the port.",
						Type:        types.StringType,
						Computed:    true,
					},
					"internal_port": {
						Description: "Internal port of the container.",
						Type:        types.Int64Type,
						Computed:    true,
					},
					"external_port": {
						Description: "External port of the container.",
						Type:        types.Int64Type,
						Computed:    true,
					},
					"publicly_accessible": {
						Description: "Specify if the port is exposed to the world or not for this container.",
						Type:        types.BoolType,
						Computed:    true,
					},
					"protocol": {
						Description: "Protocol used for the port of the container.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this container.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"environment_variables": {
				Description: "List of environment variables linked to this container.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"secrets": {
				Description: "List of secrets linked to this container.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the secret [NOTE: will always be empty].",
						Type:        types.StringType,
						Computed:    true,
						Sensitive:   true,
					},
				}),
			},
			"arguments": {
				Description: "List of arguments of this container.",
				Computed:    true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
			},
			//"custom_domains": {
			//	Description: "List of custom domains linked to this container.",
			//	Computed:    true,
			//	Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//		"id": {
			//			Description: "Id of the custom domain.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//		"domain": {
			//			Description: "Your custom domain.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//		"validation_domain": {
			//			Description: "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//		"status": {
			//			Description: "Status of the custom domain.",
			//			Type:        types.StringType,
			//			Computed:    true,
			//		},
			//	}),
			//},
			"external_host": {
				Description: "The container external FQDN host [NOTE: only if your container is using a publicly accessible port].",
				Type:        types.StringType,
				Computed:    true,
			},
			"internal_host": {
				Description: "The container internal host.",
				Type:        types.StringType,
				Computed:    true,
			},
			"state": {
				Description: "State of the container.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

// Read qovery container data source
func (d containerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Container
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get container from API
	cont, err := d.containerService.Get(ctx, data.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on container read", err.Error())
		return
	}

	state := convertDomainContainerToContainer(data, cont)
	tflog.Trace(ctx, "read container", map[string]interface{}{"container_id": state.ID.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
