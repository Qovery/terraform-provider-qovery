package client

import (
	"fmt"
	"net/http"

	"github.com/qovery/qovery-client-go"
)

type Client struct {
	api *qovery.APIClient
}

// New returns a Client that sends its requests through httpClient, normally httpclient.New().
func New(token string, version string, host string, httpClient *http.Client) *Client {
	return &Client{
		NewQoveryAPIClient(token, version, host, httpClient),
	}
}

// NewQoveryAPIClient used for tests only
func NewQoveryAPIClient(token string, version string, host string, httpClient *http.Client) *qovery.APIClient {
	cfg := qovery.NewConfiguration()
	cfg.HTTPClient = httpClient
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", token))

	cfg.UserAgent = fmt.Sprintf("Terraform provider %s", version)
	cfg.Servers = qovery.ServerConfigurations{
		{
			URL:         host,
			Description: "No description provided",
		},
	}

	return qovery.NewAPIClient(cfg)
}

// GetConfig returns the underlying qovery-client configuration. Used by resources that need to
// build a ServiceAdvancedSettingsService for plan-time validation.
func (c *Client) GetConfig() *qovery.Configuration {
	return c.api.GetConfig()
}
