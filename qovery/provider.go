package qovery

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

const qoveryAPITokenEnvName = "QOVERY_API_TOKEN"

// provider satisfies the tfsdk.Provider interface and usually is included
// with all organizationResource and DataSource implementations.
type provider struct {
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
	client *qovery.APIClient
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	Token types.String `tfsdk:"token"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
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
		token = os.Getenv(qoveryAPITokenEnvName)
	}

	if token == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find token",
			"Token cannot be an empty string",
		)
		return
	}

	// Create a new Qovery client and set it to the provider client
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", token))

	p.client = qovery.NewAPIClient(cfg)
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"qovery_organization":    organizationResourceType{},
		"qovery_aws_credentials": awsCredentialsResourceType{},
		"qovery_cluster":         clusterResourceType{},
		"qovery_project":         projectResourceType{},
		"qovery_environment":     environmentResourceType{},
		"qovery_application":     applicationResourceType{},
		"qovery_database":        databaseResourceType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"qovery_organization":    organizationDataSourceType{},
		"qovery_aws_credentials": awsCredentialsDataSourceType{},
		"qovery_cluster":         clusterDataSourceType{},
		"qovery_project":         projectDataSourceType{},
	}, nil
}

func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"token": {
				Description: "Qovery API Token",
				Type:        types.StringType,
				Required:    true,
			},
		},
	}, nil
}

func (p provider) GetClient() *qovery.APIClient {
	return p.client
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}
