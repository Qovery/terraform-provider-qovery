//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
)

// operationKey marks the context of the operation that issues the advanced settings calls.
type operationKey struct{}

// recordingTransport answers every request with an empty JSON object and records its method
// and path, and the operation its context belongs to.
type recordingTransport struct {
	mu         sync.Mutex
	requests   []string
	operations []any
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.requests = append(r.requests, req.Method+" "+req.URL.Path)
	r.operations = append(r.operations, req.Context().Value(operationKey{}))
	r.mu.Unlock()

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Request:    req,
	}, nil
}

func newRecordingConfig() (*qovery.Configuration, *recordingTransport) {
	transport := &recordingTransport{}
	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: "https://api.qovery.test"}}
	cfg.DefaultHeader["Authorization"] = "Token test"
	cfg.HTTPClient = &http.Client{Transport: transport}
	return cfg, transport
}

// TestAdvancedSettingsServices_SendThroughTheConfiguredHTTPClient pins that the advanced
// settings calls go through the http client of the Qovery API configuration, and so get the
// provider's retry policy, with the context of the calling operation so a cancelled operation
// stops its retries.
func TestAdvancedSettingsServices_SendThroughTheConfiguredHTTPClient(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		Call           func(ctx context.Context, cfg *qovery.Configuration) error
		ExpectRequests []string
	}{
		{
			TestName: "cluster_read",
			Call: func(ctx context.Context, cfg *qovery.Configuration) error {
				_, err := NewClusterAdvancedSettingsService(cfg).ReadClusterAdvancedSettings(ctx, "org-1", "cluster-1", "", false)
				return err
			},
			ExpectRequests: []string{
				"GET /defaultClusterAdvancedSettings",
				"GET /organization/org-1/cluster/cluster-1/advancedSettings",
			},
		},
		{
			TestName: "cluster_update",
			Call: func(ctx context.Context, cfg *qovery.Configuration) error {
				return NewClusterAdvancedSettingsService(cfg).UpdateClusterAdvancedSettings(ctx, "org-1", "cluster-1", "{}")
			},
			ExpectRequests: []string{
				"GET /organization/org-1/cluster/cluster-1/advancedSettings",
				"PUT /organization/org-1/cluster/cluster-1/advancedSettings",
			},
		},
		{
			TestName: "service_read",
			Call: func(ctx context.Context, cfg *qovery.Configuration) error {
				_, err := NewServiceAdvancedSettingsService(cfg).ReadServiceAdvancedSettings(ctx, domain.CONTAINER, "container-1", "", "", false)
				return err
			},
			ExpectRequests: []string{
				"GET /container/container-1/advancedSettings",
				"GET /defaultContainerAdvancedSettings",
			},
		},
		{
			TestName: "service_update",
			Call: func(ctx context.Context, cfg *qovery.Configuration) error {
				return NewServiceAdvancedSettingsService(cfg).UpdateServiceAdvancedSettings(ctx, domain.CONTAINER, "container-1", "{}")
			},
			ExpectRequests: []string{
				"GET /container/container-1/advancedSettings",
				"PUT /container/container-1/advancedSettings",
			},
		},
		{
			TestName: "service_unknown_keys",
			Call: func(ctx context.Context, cfg *qovery.Configuration) error {
				_, err := NewServiceAdvancedSettingsService(cfg).UnknownSettingKeys(ctx, domain.CONTAINER, `{"bogus.key": 1}`)
				return err
			},
			ExpectRequests: []string{
				"GET /defaultContainerAdvancedSettings",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			cfg, transport := newRecordingConfig()
			ctx := context.WithValue(context.Background(), operationKey{}, tc.TestName)
			require.NoError(t, tc.Call(ctx, cfg))
			assert.ElementsMatch(t, tc.ExpectRequests, transport.requests)
			for _, operation := range transport.operations {
				assert.Equal(t, tc.TestName, operation, "every request carries the caller's context")
			}
		})
	}
}
