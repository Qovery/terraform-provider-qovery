//go:build unit && !integration
// +build unit,!integration

package advanced_settings

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/qovery/qovery-client-go"
)

// newTestClusterService returns a ClusterAdvancedSettingsService pointing at a test server
// that serves the given defaults JSON on /defaultClusterAdvancedSettings, plus a counter of
// requests received by the server.
func newTestClusterService(t *testing.T, defaultsJSON string, statusCode int) (*ClusterAdvancedSettingsService, *int64) {
	t.Helper()

	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		if r.URL.Path != "/defaultClusterAdvancedSettings" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(defaultsJSON))
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	cfg.DefaultHeader["Authorization"] = "Token test"

	return NewClusterAdvancedSettingsService(cfg), &requestCount
}

func TestClusterAdvancedSettingsService_UnknownSettingKeys(t *testing.T) {
	t.Parallel()

	defaultsJSON := `{"aws.iam.enable_sso": false, "loki.log_retention_in_week": 12, "registry.image_retention_time": 31536000}`

	testCases := []struct {
		testName             string
		advancedSettingsJSON string
		defaultsJSON         string
		statusCode           int
		expected             []string
		expectError          bool
		expectedRequests     int64
	}{
		{
			testName:             "empty_input_returns_nil_without_fetch",
			advancedSettingsJSON: "",
			defaultsJSON:         defaultsJSON,
			statusCode:           http.StatusOK,
			expected:             nil,
			expectedRequests:     0,
		},
		{
			testName:             "empty_object_returns_nil_without_fetch",
			advancedSettingsJSON: "{}",
			defaultsJSON:         defaultsJSON,
			statusCode:           http.StatusOK,
			expected:             nil,
			expectedRequests:     0,
		},
		{
			testName:             "all_keys_known_returns_empty",
			advancedSettingsJSON: `{"aws.iam.enable_sso": true, "loki.log_retention_in_week": 4}`,
			defaultsJSON:         defaultsJSON,
			statusCode:           http.StatusOK,
			expected:             []string{},
			expectedRequests:     1,
		},
		{
			testName:             "unknown_keys_returned_sorted",
			advancedSettingsJSON: `{"zzz.bogus": 1, "aws.iam.enable_sso": true, "aaa.typo": "x"}`,
			defaultsJSON:         defaultsJSON,
			statusCode:           http.StatusOK,
			expected:             []string{"aaa.typo", "zzz.bogus"},
			expectedRequests:     1,
		},
		{
			testName:             "invalid_json_returns_error",
			advancedSettingsJSON: `{not-json`,
			defaultsJSON:         defaultsJSON,
			statusCode:           http.StatusOK,
			expectError:          true,
		},
		{
			testName:             "defaults_fetch_error_returns_error",
			advancedSettingsJSON: `{"aws.iam.enable_sso": true}`,
			defaultsJSON:         `{}`,
			statusCode:           http.StatusInternalServerError,
			expectError:          true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			svc, requestCount := newTestClusterService(t, tc.defaultsJSON, tc.statusCode)

			result, err := svc.UnknownSettingKeys(tc.advancedSettingsJSON)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("UnknownSettingKeys() = %v, want %v", result, tc.expected)
			}
			if got := atomic.LoadInt64(requestCount); got != tc.expectedRequests {
				t.Errorf("request count = %d, want %d", got, tc.expectedRequests)
			}
		})
	}
}

func TestClusterAdvancedSettingsService_UnknownSettingKeys_CachesDefaults(t *testing.T) {
	t.Parallel()

	svc, requestCount := newTestClusterService(t, `{"aws.iam.enable_sso": false}`, http.StatusOK)

	for i := 0; i < 3; i++ {
		unknown, err := svc.UnknownSettingKeys(`{"bogus.key": 1}`)
		if err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
		if !reflect.DeepEqual(unknown, []string{"bogus.key"}) {
			t.Errorf("call %d: UnknownSettingKeys() = %v, want [bogus.key]", i, unknown)
		}
	}

	if got := atomic.LoadInt64(requestCount); got != 1 {
		t.Errorf("defaults fetched %d times, want 1 (cached)", got)
	}
}
