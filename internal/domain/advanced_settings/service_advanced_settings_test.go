//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
)

// putCapturingTransport answers GET requests with getBody and records the body of the PUT request.
type putCapturingTransport struct {
	getBody string
	putBody []byte
}

func (p *putCapturingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := p.getBody
	if req.Method == http.MethodPut {
		captured, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		p.putBody = captured
		body = `{}`
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func TestUpdateServiceAdvancedSettings_PutBody(t *testing.T) {
	t.Parallel()

	const currentSettings = `{
		"build.timeout_max_sec": 1800,
		"build.cpu_max_in_milli": 4000,
		"build.ram_max_in_gib": 8,
		"build.ephemeral_storage_in_gib": null,
		"build.disable_buildkit_cache": false,
		"build.skip_git_submodules": false,
		"deployment.termination_grace_period_seconds": 60,
		"network.ingress.proxy_body_size_mb": 100
	}`

	testCases := []struct {
		TestName        string
		PlanJson        string
		ExpectedPutBody map[string]any
	}{
		{
			TestName: "build_keys_from_api_are_not_echoed",
			PlanJson: "",
			ExpectedPutBody: map[string]any{
				"deployment.termination_grace_period_seconds": float64(60),
				"network.ingress.proxy_body_size_mb":          float64(100),
			},
		},
		{
			TestName: "build_keys_set_in_plan_are_sent",
			PlanJson: `{"build.timeout_max_sec": 3600, "build.skip_git_submodules": true}`,
			ExpectedPutBody: map[string]any{
				"build.timeout_max_sec":                       float64(3600),
				"build.skip_git_submodules":                   true,
				"deployment.termination_grace_period_seconds": float64(60),
				"network.ingress.proxy_body_size_mb":          float64(100),
			},
		},
		{
			TestName: "build_key_set_to_null_in_plan_is_sent",
			PlanJson: `{"build.ephemeral_storage_in_gib": null}`,
			ExpectedPutBody: map[string]any{
				"build.ephemeral_storage_in_gib":              nil,
				"deployment.termination_grace_period_seconds": float64(60),
				"network.ingress.proxy_body_size_mb":          float64(100),
			},
		},
		{
			TestName: "non_build_keys_from_plan_override_api_values",
			PlanJson: `{"network.ingress.proxy_body_size_mb": 200}`,
			ExpectedPutBody: map[string]any{
				"deployment.termination_grace_period_seconds": float64(60),
				"network.ingress.proxy_body_size_mb":          float64(200),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			transport := &putCapturingTransport{getBody: currentSettings}
			cfg := qovery.NewConfiguration()
			cfg.Servers = qovery.ServerConfigurations{{URL: "https://api.qovery.test"}}
			cfg.HTTPClient = &http.Client{Transport: transport}

			err := NewServiceAdvancedSettingsService(cfg).UpdateServiceAdvancedSettings(context.Background(), domain.APPLICATION, "app-1", tc.PlanJson)
			require.NoError(t, err)

			var putBody map[string]any
			require.NoError(t, json.Unmarshal(transport.putBody, &putBody))
			assert.Equal(t, tc.ExpectedPutBody, putBody)
		})
	}
}
