package client

import (
	"fmt"

	"github.com/qovery/qovery-client-go"
)

type Client struct {
	api *qovery.APIClient
}

func New(token string, version string) *Client {
	cfg := qovery.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("Token %s", token))
	cfg.AddDefaultHeader("content-type", "application/json")

	cfg.UserAgent = fmt.Sprintf("terraform-provider-qovery/%s", version)

	return &Client{
		qovery.NewAPIClient(cfg),
	}
}
