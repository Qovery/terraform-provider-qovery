package client

import (
	"fmt"

	"github.com/qovery/qovery-client-go"
)

type Client struct {
	api *qovery.APIClient
}

func New(token string, version string, host string) *Client {
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", token))
	cfg.AddDefaultHeader("content-type", "application/json")

	cfg.UserAgent = fmt.Sprintf("terraform-provider-qovery/%s", version)
	cfg.Servers = qovery.ServerConfigurations{
		{
			URL:         host,
			Description: "No description provided",
		},
	}

	return &Client{
		qovery.NewAPIClient(cfg),
	}
}

// NewQoveryApiClient used for tests only
func NewQoveryApiClient(token string, version string, host string) *qovery.APIClient {
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", token))
	cfg.AddDefaultHeader("content-type", "application/json")

	cfg.UserAgent = fmt.Sprintf("terraform-provider-qovery/%s", version)
	cfg.Servers = qovery.ServerConfigurations{
		{
			URL:         host,
			Description: "No description provided",
		},
	}

	return qovery.NewAPIClient(cfg)
}
