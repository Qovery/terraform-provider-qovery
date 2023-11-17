package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &clusterDataSource{}

type clusterDataSource struct {
	client *client.Client
}

func newClusterDataSource() datasource.DataSource {
	return &clusterDataSource{}
}

func (d clusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *clusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = provider.client
}

func (r clusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery cluster resource. This can be used to create and manage Qovery cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the cluster.",
				Computed:    true,
			},
			"credentials_id": schema.StringAttribute{
				Description: "Id of the credentials.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Required:    true,
			},
			"cloud_provider": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Cloud provider of the cluster.",
					cloudProviders,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(cloudProviders),
				},
			},
			"region": schema.StringAttribute{
				Description: "Region of the cluster.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: descriptions.NewStringDefaultDescription(
					"Description of the cluster.",
					clusterDescriptionDefault,
				),
				Optional: true,
			},
			"kubernetes_mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Kubernetes mode of the cluster.",
					clusterKubernetesModes,
					&clusterKubernetesModeDefault,
				),
				Optional: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterKubernetesModes),
				},
			},
			"instance_type": schema.StringAttribute{
				Description: "Instance type of the cluster. I.e: For Aws `t3a.xlarge`, for Scaleway `DEV-L`",
				Required:    true,
			},
			"min_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters].",
					clusterMinRunningNodesMin,
					&clusterMinRunningNodesDefault,
				),
				Optional: true,
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: clusterMinRunningNodesMin},
				},
			},
			"max_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters]",
					clusterMaxRunningNodesMin,
					&clusterMaxRunningNodesDefault,
				),
				Optional: true,
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: clusterMaxRunningNodesMin},
				},
			},
			"features": schema.SetNestedAttribute{
				Description: "Features of the cluster.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"vpc_subnet": schema.StringAttribute{
							Description: descriptions.NewStringDefaultDescription(
								"Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].",
								clusterFeatureVpcSubnetDefault,
							),
							Optional: true,
						},
						"static_ip": schema.BoolAttribute{
							Description: descriptions.NewBoolDefaultDescription(
								"Static IP (AWS only) [NOTE: can't be updated after creation].",
								clusterFeatureStaticIPDefault,
							),
							Optional: true,
						},
					},
				},
			},
			"routing_table": schema.SetNestedAttribute{
				Description: "List of routes of the cluster.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Description: "Description of the route.",
							Required:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination of the route.",
							Required:    true,
						},
						"target": schema.StringAttribute{
							Description: "Target of the route.",
							Required:    true,
						},
					},
				},
			},
			"state": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"State of the cluster.",
					clusterStates,
					&clusterStateDefault,
				),
				Optional: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterStates),
				},
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings of the cluster.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Read qovery cluster data source
func (d clusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Cluster
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from the API
	cluster, apiErr := d.client.GetCluster(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToCluster(ctx, cluster)
	tflog.Trace(ctx, "read cluster", map[string]interface{}{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
