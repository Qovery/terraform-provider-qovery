package qovery

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/qovery/apierror"
	"terraform-provider-qovery/qovery/descriptions"
	"terraform-provider-qovery/qovery/validators"
)

const applicationAPIResource = "application"

var (
	// Application Build Mode
	applicationBuildModes       = []string{"BUILDPACKS", "DOCKER"}
	applicationBuildModeDefault = "BUILDPACKS"

	// Application Buildpack
	applicationBuildpackLanguages = []string{"RUBY", "NODE_JS", "CLOJURE", "PYTHON", "JAVA", "GRADLE", "JVM", "GRAILS", "SCALA", "PLAY", "PHP", "GO"}

	// Application CPU
	applicationCPUMin     int64 = 250 // in MB
	applicationCPUDefault int64 = 500 // in MB

	// Application Memory
	applicationMemoryMin     int64 = 1   // in MB
	applicationMemoryDefault int64 = 500 // in MB

	// Application Min Running Instances
	applicationMinRunningInstancesMin     int64 = 0
	applicationMinRunningInstancesDefault int64 = 1

	// Application Max Running Instances
	applicationMaxRunningInstancesMin     int64 = -1
	applicationMaxRunningInstancesDefault int64 = 1

	// Application Env Preview
	applicationEnvPreviewDefault = true

	// Application Storage
	applicationStorageTypes         = []string{"SLOW_HDD", "HDD", "SSD", "FAST_SSD"}
	applicationStorageSizeMin int64 = 1 // in GB

	// Application Port
	applicationPortMin                       int64 = 1
	applicationPortMax                       int64 = 65535
	applicationPortProtocols                       = []string{"HTTPS", "HTTP", "TCP", "UDP"}
	applicationPortProtocolDefault                 = "HTTP"
	applicationPortPubliclyAccessibleDefault       = false

	// Application Git Repository
	applicationGitRepositoryRootPathDefault = "/"
	applicationGitRepositoryBranchDefault   = "main or master (depending on repository)"
)

type applicationData struct {
	Id                  types.String                 `tfsdk:"id"`
	EnvironmentId       types.String                 `tfsdk:"environment_id"`
	Name                types.String                 `tfsdk:"name"`
	Description         types.String                 `tfsdk:"description"`
	GitRepository       applicationGitRepositoryData `tfsdk:"git_repository"`
	BuildMode           types.String                 `tfsdk:"build_mode"`
	DockerfilePath      types.String                 `tfsdk:"dockerfile_path"`
	BuildpackLanguage   types.String                 `tfsdk:"buildpack_language"`
	CPU                 types.Int64                  `tfsdk:"cpu"`
	Memory              types.Int64                  `tfsdk:"memory"`
	MinRunningInstances types.Int64                  `tfsdk:"min_running_instances"`
	MaxRunningInstances types.Int64                  `tfsdk:"max_running_instances"`
	EnvPreview          types.Bool                   `tfsdk:"env_preview"`
	Storage             []applicationStorageData     `tfsdk:"storage"`
	Ports               []applicationPortData        `tfsdk:"ports"`
}

type applicationGitRepositoryData struct {
	URL      types.String `tfsdk:"url"`
	Branch   types.String `tfsdk:"branch"`
	RootPath types.String `tfsdk:"root_path"`
}

type applicationStorageData struct {
	Type       types.String `tfsdk:"type"`
	Size       types.Int64  `tfsdk:"size"`
	MountPoint types.String `tfsdk:"mount_point"`
}

type applicationPortData struct {
	Name               types.String `tfsdk:"name"`
	InternalPort       types.Int64  `tfsdk:"internal_port"`
	ExternalPort       types.Int64  `tfsdk:"external_port"`
	PubliclyAccessible types.Bool   `tfsdk:"publicly_accessible"`
	Protocol           types.String `tfsdk:"protocol"`
}

type applicationResourceType struct{}

func (r applicationResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery application resource. This can be used to create and manage Qovery applications.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the application.",
				Type:        types.StringType,
				Computed:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the application.",
				Type:        types.StringType,
				Required:    true,
			},
			"description": {
				Description: "Description of the application.",
				Type:        types.StringType,
				Optional:    true,
			},
			"git_repository": {
				Description: "Git repository of the application.",
				Required:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"url": {
						Description: "URL of the git repository.",
						Type:        types.StringType,
						Required:    true,
					},
					"branch": {
						Description: descriptions.NewStringDefaultDescription(
							"Branch of the git repository.",
							applicationGitRepositoryBranchDefault,
						),
						Type:     types.StringType,
						Optional: true,
					},
					"root_path": {
						Description: descriptions.NewStringDefaultDescription(
							"Root path of the application.",
							applicationGitRepositoryRootPathDefault,
						),
						Type:     types.StringType,
						Optional: true,
					},
				}),
			},
			"build_mode": {
				Description: descriptions.NewStringEnumDescription(
					"Build Mode of the application.",
					applicationBuildModes,
					&applicationBuildModeDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					validators.StringEnumValidator{Enum: applicationBuildModes},
				},
			},
			"dockerfile_path": {
				Description: "Dockerfile Path of the application.\n\t- Required if: `build_mode=\"DOCKER\"`.",
				Type:        types.StringType,
				Optional:    true,
			},
			"buildpack_language": {
				Description: descriptions.NewStringEnumDescription(
					"Buildpack Language framework.\n\t- Required if: `build_mode=\"BUILDPACKS\"`.",
					applicationBuildpackLanguages,
					nil,
				),
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					validators.StringEnumValidator{Enum: applicationBuildpackLanguages},
				},
			},
			"cpu": {
				Description: descriptions.NewInt64MinDescription(
					"CPU of the application in millicores (m) [1000m = 1 CPU].",
					applicationCPUMin,
					&applicationCPUDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: applicationCPUMin},
				},
			},
			"memory": {
				Description: descriptions.NewInt64MinDescription(
					"RAM of the application in MB [1024MB = 1GB].",
					applicationMemoryMin,
					&applicationMemoryDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: applicationMemoryMin},
				},
			},
			"min_running_instances": {
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of instances running for the application.",
					applicationMinRunningInstancesMin,
					&applicationMinRunningInstancesDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: applicationMinRunningInstancesMin},
				},
			},
			"max_running_instances": {
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of instances running for the application.",
					applicationMaxRunningInstancesMin,
					&applicationMaxRunningInstancesDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: applicationMaxRunningInstancesMin},
				},
			},
			"env_preview": {
				Description: descriptions.NewBoolDefaultDescription(
					"Specify if the environment preview option is activated or not for this application.",
					applicationEnvPreviewDefault,
				),
				Type:     types.BoolType,
				Optional: true,
			},
			"storage": {
				Description: "List of storages linked to this application.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"type": {
						Description: descriptions.NewStringEnumDescription(
							"Type of the storage for the application.",
							applicationStorageTypes,
							nil,
						),
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							validators.StringEnumValidator{Enum: applicationStorageTypes},
						},
					},
					"size": {
						Description: descriptions.NewInt64MinDescription(
							"Size of the storage for the application in GB [1024MB = 1GB].",
							applicationStorageSizeMin,
							nil,
						),
						Type:     types.Int64Type,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							validators.Int64MinValidator{Min: databaseMemoryMin},
						},
					},
					"mount_point": {
						Description: "Mount point of the storage for the application.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"port": {
				Description: "List of storages linked to this application.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Description: "Name of the port.",
						Type:        types.StringType,
						Optional:    true,
					},
					"internal_port": {
						Description: descriptions.NewInt64MinMaxDescription(
							"Internal port of the application.",
							applicationPortMin,
							applicationPortMax,
							nil,
						),
						Type:     types.Int64Type,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							validators.Int64MinMaxValidator{Min: applicationPortMin, Max: applicationPortMax},
						},
					},
					"external_port": {
						Description: descriptions.NewInt64MinMaxDescription(
							"External port of the application.",
							applicationPortMin,
							applicationPortMax,
							nil,
						),
						Type:     types.Int64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							validators.Int64MinMaxValidator{Min: applicationPortMin, Max: applicationPortMax},
						},
					},
					"publicly_accessible": {
						Description: descriptions.NewBoolDefaultDescription(
							"Specify if the port is exposed to the world or not for this application.",
							applicationPortPubliclyAccessibleDefault,
						),
						Type:     types.BoolType,
						Optional: true,
					},
					"protocol": {
						Description: descriptions.NewStringEnumDescription(
							"Protocol used for the port of the application.",
							applicationPortProtocols,
							&applicationPortProtocolDefault,
						),
						Type:     types.StringType,
						Optional: true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (r applicationResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return applicationResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type applicationResource struct {
	client *qovery.APIClient
}

// Create qovery application resource
func (r applicationResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
}

// Read qovery application resource
func (r applicationResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update qovery application resource
func (r applicationResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete qovery application resource
func (r applicationResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

// ImportState imports a qovery application resource using its id
func (r applicationResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStateNotImplemented(ctx, "", resp)
}

func applicationCreateAPIError(applicationName string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(applicationAPIResource, applicationName, apierror.Create, res, err)
}

func applicationReadAPIError(applicationID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(applicationAPIResource, applicationID, apierror.Read, res, err)
}

func applicationUpdateAPIError(applicationID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(applicationAPIResource, applicationID, apierror.Update, res, err)
}

func applicationDeleteAPIError(applicationID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(applicationAPIResource, applicationID, apierror.Delete, res, err)
}
