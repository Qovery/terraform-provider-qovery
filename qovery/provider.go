package qovery

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

const APITokenEnvName = "QOVERY_API_TOKEN"

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ provider.Provider = &qProvider{}
var _ provider.ProviderWithMetadata = &qProvider{}

// qProvider satisfies the provider.Provider interface and usually is included
// with all Resource and DataSource implementations.
type qProvider struct {
	// configured is set to true at the end of the Configure method.
	// This can be used in organizationResource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	// client is set at the end of the Configure method.
	// This is used to make http request to Qovery API.
	client *client.Client

	// organizationService is an instance of an organization.Service that handles the domain logic.
	organizationService organization.Service

	// awsCredentialsService is an instance of a credentials.AwsService that handles the domain logic.
	awsCredentialsService credentials.AwsService

	// scalewayCredentialsService is an instance of a credentials.ScalewayService that handles the domain logic.
	scalewayCredentialsService credentials.ScalewayService

	// projectService is an instance of a project.Service that handles the domain logic.
	projectService project.Service

	// containerRegistryService is an instance of a registry.Service that handles the domain logic.
	containerService container.Service

	// containerRegistryService is an instance of a registry.Service that handles the domain logic.
	containerRegistryService registry.Service

	// environmentService is an instance of an environment.Service that handles the domain logic.
	environmentService environment.Service

	// deploymentStageService is an instance of an deploymentstage.Service that handles the domain logic.
	deploymentStageService deploymentstage.Service
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	Token types.String `tfsdk:"token"`
}

func (p *qProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "qovery"
	resp.Version = p.version
}

func (p *qProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide a token to the provider
	if data.Token.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as token",
		)
		return
	}

	token := data.Token.Value
	if data.Token.Null {
		token = os.Getenv(APITokenEnvName)
	}

	if token == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find token",
			"Token cannot be an empty string",
		)
		return
	}

	// Initialize qovery client
	domainServices, err := services.New(services.WithQoveryRepository(token, p.version))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to initialize domain services",
			err.Error(),
		)
		return
	}

	// Create a new Qovery client and set it to the provider client
	p.configured = true
	p.client = client.New(token, p.version)
	p.organizationService = domainServices.Organization
	p.awsCredentialsService = domainServices.CredentialsAws
	p.scalewayCredentialsService = domainServices.CredentialsScaleway
	p.projectService = domainServices.Project
	p.containerService = domainServices.Container
	p.containerRegistryService = domainServices.ContainerRegistry
	p.environmentService = domainServices.Environment
	p.deploymentStageService = domainServices.DeploymentStage

	resp.DataSourceData = p
	resp.ResourceData = p
}

func (p *qProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newApplicationResource,
		newAwsCredentialsResource,
		newClusterResource,
		newDatabaseResource,
		newEnvironmentResource,
		newOrganizationResource,
		newProjectResource,
		newScalewayCredentialsResource,
		newContainerResource,
		newContainerRegistryResource,
		newDeploymentStageResource,
	}
}

func (p *qProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newApplicationDataSource,
		newAwsCredentialsDataSource,
		newClusterDataSource,
		newContainerDataSource,
		newContainerRegistryDataSource,
		newDatabaseDataSource,
		newEnvironmentDataSource,
		newOrganizationDataSource,
		newProjectDataSource,
		newScalewayCredentialsDataSource,
		newDeploymentStageDataSource,
	}
}

func (p *qProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"token": {
				Description: "The Qovery API Token to use. This can also be specified with the `QOVERY_API_TOKEN` shell environment variable.",
				Type:        types.StringType,
				Optional:    true,
			},
		},
	}, nil
}

func (p qProvider) GetClient() *client.Client {
	return p.client
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &qProvider{
			version: version,
		}
	}
}

func NewQoveryAPIClient(token string, version string) *qovery.APIClient {
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", token))
	cfg.AddDefaultHeader("content-type", "application/json")
	cfg.UserAgent = fmt.Sprintf("terraform-provider-qovery/%s", version)
	return qovery.NewAPIClient(cfg)
}
