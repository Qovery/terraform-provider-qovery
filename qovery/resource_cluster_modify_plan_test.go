//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
)

// newTestClusterAdvancedSettingsService returns a ClusterAdvancedSettingsService backed by a
// test server that serves the given defaults JSON on /defaultClusterAdvancedSettings.
func newTestClusterAdvancedSettingsService(t *testing.T, defaultsJSON string) *advanced_settings.ClusterAdvancedSettingsService {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/defaultClusterAdvancedSettings" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(defaultsJSON))
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	cfg.DefaultHeader["Authorization"] = "Token test"

	return advanced_settings.NewClusterAdvancedSettingsService(cfg)
}

func TestClusterResource_ModifyPlan_WarnsOnUnknownAdvancedSettings(t *testing.T) {
	t.Parallel()

	r := clusterResource{
		clusterAdvancedSettingsService: newTestClusterAdvancedSettingsService(t, `{"aws.iam.admin_group": "Admins"}`),
	}

	req := resource.ModifyPlanRequest{
		Config: testAdvancedSettingsConfig(tftypes.NewValue(tftypes.String, `{"aws.iam.admin_group": "Admins", "bogus.key.qov2034": true}`)),
	}
	var resp resource.ModifyPlanResponse

	r.ModifyPlan(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors, got %v", resp.Diagnostics.Errors())
	}
	warnings := resp.Diagnostics.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("got %d warnings, want 1: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Summary(), "bogus.key.qov2034") {
		t.Errorf("warning summary %q does not mention the unknown key", warnings[0].Summary())
	}
}

func TestClusterResource_ModifyPlan_NoServiceIsNoop(t *testing.T) {
	t.Parallel()

	// Zero-value resource: provider not configured, service is nil.
	var r clusterResource

	req := resource.ModifyPlanRequest{
		Config: testAdvancedSettingsConfig(tftypes.NewValue(tftypes.String, `{"bogus.key.qov2034": true}`)),
	}
	var resp resource.ModifyPlanResponse

	r.ModifyPlan(context.Background(), req, &resp)

	if len(resp.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics with nil service, got %v", resp.Diagnostics)
	}
}
