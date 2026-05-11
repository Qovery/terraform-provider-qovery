//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterDNSProviderToQoveryRequest_RequiresCloudflareSecret(t *testing.T) {
	model := ClusterDNSProvider{
		ProviderType: types.StringValue(clusterDNSProviderTypeCloudflare),
		Domain:       types.StringValue("example.com"),
		Cloudflare: &ClusterDNSProviderCloudflare{
			Email:    types.StringValue("admin@example.com"),
			APIToken: types.StringNull(),
			Proxied:  types.BoolValue(false),
		},
	}

	request, err := model.toQoveryRequest()

	require.Error(t, err)
	assert.Nil(t, request)
	assert.Contains(t, err.Error(), "cloudflare.api_token")
}

func TestClusterDNSProviderToQoveryRequest_RequiresRoute53Secret(t *testing.T) {
	model := ClusterDNSProvider{
		ProviderType: types.StringValue(clusterDNSProviderTypeRoute53),
		Domain:       types.StringValue("example.com"),
		Route53: &ClusterDNSProviderRoute53{
			Credentials: &ClusterDNSProviderRoute53Credentials{
				Type:               types.StringValue(clusterDNSProviderCredentialsStatic),
				AWSAccessKeyID:     types.StringValue("access-key"),
				AWSSecretAccessKey: types.StringValue(""),
			},
			AWSRegion: types.StringValue("eu-west-3"),
		},
	}

	request, err := model.toQoveryRequest()

	require.Error(t, err)
	assert.Nil(t, request)
	assert.Contains(t, err.Error(), "route53.credentials.aws_secret_access_key")
}

func TestConvertResponseToClusterDNSProvider_PreservesCloudflareAPIToken(t *testing.T) {
	response := qovery.NewClusterDnsProviderResponse(
		qovery.CloudflareDnsProviderResponseAsClusterDnsProviderResponseProvider(
			qovery.NewCloudflareDnsProviderResponse(clusterDNSProviderTypeCloudflare, "example.com", "admin@example.com", true),
		),
	)
	previous := ClusterDNSProvider{
		Cloudflare: &ClusterDNSProviderCloudflare{
			APIToken: types.StringValue("secret-token"),
		},
	}

	state, err := convertResponseToClusterDNSProvider("cluster-id", response, previous)

	require.NoError(t, err)
	require.NotNil(t, state.Cloudflare)
	assert.Equal(t, "cluster-id", state.ID.ValueString())
	assert.Equal(t, clusterDNSProviderTypeCloudflare, state.ProviderType.ValueString())
	assert.Equal(t, "secret-token", state.Cloudflare.APIToken.ValueString())
	assert.True(t, state.Cloudflare.Proxied.ValueBool())
}

func TestConvertResponseToClusterDNSProvider_PreservesRoute53SecretAccessKey(t *testing.T) {
	credentials := qovery.Route53StaticCredentialsResponseAsRoute53DnsProviderResponseCredentials(
		qovery.NewRoute53StaticCredentialsResponse(clusterDNSProviderCredentialsStatic, "access-key"),
	)
	route53 := qovery.NewRoute53DnsProviderResponse(clusterDNSProviderTypeRoute53, "example.com", credentials, "eu-west-3")
	route53.SetHostedZoneId("Z123")
	response := qovery.NewClusterDnsProviderResponse(
		qovery.Route53DnsProviderResponseAsClusterDnsProviderResponseProvider(route53),
	)
	previous := ClusterDNSProvider{
		Route53: &ClusterDNSProviderRoute53{
			Credentials: &ClusterDNSProviderRoute53Credentials{
				AWSSecretAccessKey: types.StringValue("secret-key"),
			},
		},
	}

	state, err := convertResponseToClusterDNSProvider("cluster-id", response, previous)

	require.NoError(t, err)
	require.NotNil(t, state.Route53)
	require.NotNil(t, state.Route53.Credentials)
	assert.Equal(t, clusterDNSProviderTypeRoute53, state.ProviderType.ValueString())
	assert.Equal(t, "access-key", state.Route53.Credentials.AWSAccessKeyID.ValueString())
	assert.Equal(t, "secret-key", state.Route53.Credentials.AWSSecretAccessKey.ValueString())
	assert.Equal(t, "Z123", state.Route53.HostedZoneID.ValueString())
}

func TestConvertResponseToClusterDNSProvider_ReadsQovery(t *testing.T) {
	response := qovery.NewClusterDnsProviderResponse(
		qovery.QoveryDnsProviderResponseAsClusterDnsProviderResponseProvider(
			qovery.NewQoveryDnsProviderResponse(clusterDNSProviderTypeQovery, "cluster.qovery.dev"),
		),
	)

	state, err := convertResponseToClusterDNSProvider("cluster-id", response, ClusterDNSProvider{})

	require.NoError(t, err)
	assert.Equal(t, "cluster-id", state.ID.ValueString())
	assert.Equal(t, "cluster-id", state.ClusterID.ValueString())
	assert.Equal(t, clusterDNSProviderTypeQovery, state.ProviderType.ValueString())
	assert.Equal(t, "cluster.qovery.dev", state.Domain.ValueString())
	assert.Nil(t, state.Cloudflare)
	assert.Nil(t, state.Route53)
}
