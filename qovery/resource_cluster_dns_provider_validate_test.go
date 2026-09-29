//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validateClusterDNSProviderConfig runs ValidateConfig on the configuration the model describes.
func validateClusterDNSProviderConfig(t *testing.T, model ClusterDNSProvider) *resource.ValidateConfigResponse {
	t.Helper()
	ctx := context.Background()
	r := newClusterDNSProviderResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
	diags := state.Set(ctx, &model)
	require.False(t, diags.HasError(), diags)

	resp := &resource.ValidateConfigResponse{}
	r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: state.Raw},
	}, resp)
	return resp
}

func cloudflareDNSProvider(apiToken types.String) ClusterDNSProvider {
	return ClusterDNSProvider{
		ID:           types.StringUnknown(),
		ClusterID:    types.StringValue("cluster-id"),
		ProviderType: types.StringValue(clusterDNSProviderTypeCloudflare),
		Domain:       types.StringValue("example.com"),
		Cloudflare: &ClusterDNSProviderCloudflare{
			Email:    types.StringValue("admin@example.com"),
			APIToken: apiToken,
			Proxied:  types.BoolValue(false),
		},
	}
}

func route53DNSProvider(secretAccessKey types.String) ClusterDNSProvider {
	return ClusterDNSProvider{
		ID:           types.StringUnknown(),
		ClusterID:    types.StringValue("cluster-id"),
		ProviderType: types.StringValue(clusterDNSProviderTypeRoute53),
		Domain:       types.StringValue("example.com"),
		Route53: &ClusterDNSProviderRoute53{
			AWSRegion:    types.StringValue("us-east-1"),
			HostedZoneID: types.StringValue("Z123"),
			Credentials: &ClusterDNSProviderRoute53Credentials{
				Type:               types.StringValue(clusterDNSProviderCredentialsStatic),
				AWSAccessKeyID:     types.StringValue("AKIA"),
				AWSSecretAccessKey: secretAccessKey,
			},
		},
	}
}

func TestClusterDNSProviderValidateConfig_Secrets(t *testing.T) {
	tests := []struct {
		name      string
		model     ClusterDNSProvider
		wantError bool
	}{
		// An unknown secret, such as a variable at validate time or another resource's output at
		// plan time, is only checked once it is known.
		{name: "cloudflare_unknown_token", model: cloudflareDNSProvider(types.StringUnknown()), wantError: false},
		{name: "cloudflare_token", model: cloudflareDNSProvider(types.StringValue("token")), wantError: false},
		{name: "cloudflare_null_token", model: cloudflareDNSProvider(types.StringNull()), wantError: true},
		{name: "cloudflare_empty_token", model: cloudflareDNSProvider(types.StringValue("")), wantError: true},
		{name: "route53_unknown_secret", model: route53DNSProvider(types.StringUnknown()), wantError: false},
		{name: "route53_secret", model: route53DNSProvider(types.StringValue("secret")), wantError: false},
		{name: "route53_null_secret", model: route53DNSProvider(types.StringNull()), wantError: true},
		{name: "route53_empty_secret", model: route53DNSProvider(types.StringValue("")), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validateClusterDNSProviderConfig(t, tt.model)
			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError(), resp.Diagnostics)
		})
	}
}
