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

// NewQoveryAPIClient used for tests only
func NewQoveryAPIClient(token string, version string, host string) *qovery.APIClient {
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
