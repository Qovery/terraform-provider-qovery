//go:build unit && !integration

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetClusterDNSProviderPropagatesForbidden(t *testing.T) {
	c := New("token", "test", "https://api.qovery.test")
	c.api.GetConfig().HTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/cluster/cluster-id/dnsProvider", r.URL.Path)

		return &http.Response{
			StatusCode: http.StatusForbidden,
			Status:     "403 Forbidden",
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"status":403,"detail":"Enterprise plan required"}`)),
			Request:    r,
		}, nil
	})}

	response, apiErr := c.GetClusterDNSProvider(context.Background(), "cluster-id")

	require.Nil(t, response)
	require.NotNil(t, apiErr)
	assert.Equal(t, "Error on cluster DNS provider read", apiErr.Summary())
	assert.Contains(t, apiErr.Detail(), "Enterprise plan required")
}

func TestUpdateClusterDNSProviderUsesPutEndpoint(t *testing.T) {
	c := New("token", "test", "https://api.qovery.test")
	c.api.GetConfig().HTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/cluster/cluster-id/dnsProvider", r.URL.Path)

		var body map[string]map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "CLOUDFLARE", body["dns_provider"]["provider"])
		assert.Equal(t, "example.com", body["dns_provider"]["domain"])
		assert.Equal(t, "token", body["dns_provider"]["api_token"])

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewBufferString(`{"dns_provider":{"provider":"CLOUDFLARE","domain":"example.com","email":"admin@example.com","proxied":false}}`)),
			Request:    r,
		}, nil
	})}
	cloudflare := qovery.NewCloudflareDnsProviderRequest("CLOUDFLARE", "example.com", "admin@example.com", "token")
	request := qovery.NewClusterDnsProviderRequest(qovery.CloudflareDnsProviderRequestAsClusterDnsProviderRequestProvider(cloudflare))

	response, apiErr := c.UpdateClusterDNSProvider(context.Background(), "cluster-id", *request)

	require.Nil(t, apiErr)
	require.NotNil(t, response)
	require.NotNil(t, response.DnsProvider.CloudflareDnsProviderResponse)
	assert.Equal(t, "example.com", response.DnsProvider.CloudflareDnsProviderResponse.Domain)
}
