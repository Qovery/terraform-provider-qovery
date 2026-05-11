package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

const (
	clusterDNSProviderTypeQovery        = "QOVERY"
	clusterDNSProviderTypeCloudflare    = "CLOUDFLARE"
	clusterDNSProviderTypeRoute53       = "ROUTE53"
	clusterDNSProviderCredentialsStatic = "STATIC"
)

type ClusterDNSProvider struct {
	ID           types.String                  `tfsdk:"id"`
	ClusterID    types.String                  `tfsdk:"cluster_id"`
	ProviderType types.String                  `tfsdk:"provider_type"`
	Domain       types.String                  `tfsdk:"domain"`
	Cloudflare   *ClusterDNSProviderCloudflare `tfsdk:"cloudflare"`
	Route53      *ClusterDNSProviderRoute53    `tfsdk:"route53"`
}

type ClusterDNSProviderCloudflare struct {
	Email    types.String `tfsdk:"email"`
	APIToken types.String `tfsdk:"api_token"`
	Proxied  types.Bool   `tfsdk:"proxied"`
}

type ClusterDNSProviderRoute53 struct {
	Credentials  *ClusterDNSProviderRoute53Credentials `tfsdk:"credentials"`
	AWSRegion    types.String                          `tfsdk:"aws_region"`
	HostedZoneID types.String                          `tfsdk:"hosted_zone_id"`
}

type ClusterDNSProviderRoute53Credentials struct {
	Type               types.String `tfsdk:"type"`
	AWSAccessKeyID     types.String `tfsdk:"aws_access_key_id"`
	AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
}

func (p ClusterDNSProvider) toQoveryRequest() (*qovery.ClusterDnsProviderRequest, error) {
	providerType := p.ProviderType.ValueString()
	domain := p.Domain.ValueString()

	switch providerType {
	case clusterDNSProviderTypeQovery:
		if p.Cloudflare != nil || p.Route53 != nil {
			return nil, fmt.Errorf("cloudflare and route53 must not be set when provider_type is %s", clusterDNSProviderTypeQovery)
		}
		provider := qovery.QoveryDnsProviderRequestAsClusterDnsProviderRequestProvider(
			qovery.NewQoveryDnsProviderRequest(clusterDNSProviderTypeQovery, domain),
		)
		return qovery.NewClusterDnsProviderRequest(provider), nil
	case clusterDNSProviderTypeCloudflare:
		if p.Cloudflare == nil {
			return nil, fmt.Errorf("cloudflare must be set when provider_type is %s", clusterDNSProviderTypeCloudflare)
		}
		if p.Route53 != nil {
			return nil, fmt.Errorf("route53 must not be set when provider_type is %s", clusterDNSProviderTypeCloudflare)
		}
		if p.Cloudflare.APIToken.IsNull() || p.Cloudflare.APIToken.IsUnknown() || p.Cloudflare.APIToken.ValueString() == "" {
			return nil, fmt.Errorf("cloudflare.api_token must be set when provider_type is %s", clusterDNSProviderTypeCloudflare)
		}

		cloudflare := qovery.NewCloudflareDnsProviderRequest(
			clusterDNSProviderTypeCloudflare,
			domain,
			p.Cloudflare.Email.ValueString(),
			p.Cloudflare.APIToken.ValueString(),
		)
		if !p.Cloudflare.Proxied.IsNull() && !p.Cloudflare.Proxied.IsUnknown() {
			cloudflare.SetProxied(p.Cloudflare.Proxied.ValueBool())
		}
		provider := qovery.CloudflareDnsProviderRequestAsClusterDnsProviderRequestProvider(cloudflare)
		return qovery.NewClusterDnsProviderRequest(provider), nil
	case clusterDNSProviderTypeRoute53:
		if p.Route53 == nil {
			return nil, fmt.Errorf("route53 must be set when provider_type is %s", clusterDNSProviderTypeRoute53)
		}
		if p.Cloudflare != nil {
			return nil, fmt.Errorf("cloudflare must not be set when provider_type is %s", clusterDNSProviderTypeRoute53)
		}
		if p.Route53.Credentials == nil {
			return nil, fmt.Errorf("route53.credentials must be set when provider_type is %s", clusterDNSProviderTypeRoute53)
		}
		if p.Route53.Credentials.AWSSecretAccessKey.IsNull() || p.Route53.Credentials.AWSSecretAccessKey.IsUnknown() || p.Route53.Credentials.AWSSecretAccessKey.ValueString() == "" {
			return nil, fmt.Errorf("route53.credentials.aws_secret_access_key must be set when provider_type is %s", clusterDNSProviderTypeRoute53)
		}

		staticCredentials := qovery.NewRoute53StaticCredentialsRequest(
			clusterDNSProviderCredentialsStatic,
			p.Route53.Credentials.AWSAccessKeyID.ValueString(),
			p.Route53.Credentials.AWSSecretAccessKey.ValueString(),
		)
		credentials := qovery.Route53StaticCredentialsRequestAsRoute53DnsProviderRequestCredentials(staticCredentials)
		route53 := qovery.NewRoute53DnsProviderRequest(
			clusterDNSProviderTypeRoute53,
			domain,
			credentials,
			p.Route53.AWSRegion.ValueString(),
		)
		if !p.Route53.HostedZoneID.IsNull() && !p.Route53.HostedZoneID.IsUnknown() {
			route53.SetHostedZoneId(p.Route53.HostedZoneID.ValueString())
		}
		provider := qovery.Route53DnsProviderRequestAsClusterDnsProviderRequestProvider(route53)
		return qovery.NewClusterDnsProviderRequest(provider), nil
	default:
		return nil, fmt.Errorf("unsupported provider_type %q", providerType)
	}
}

func convertResponseToClusterDNSProvider(clusterID string, response *qovery.ClusterDnsProviderResponse, previous ClusterDNSProvider) (ClusterDNSProvider, error) {
	if response == nil {
		return ClusterDNSProvider{}, fmt.Errorf("empty cluster DNS provider response")
	}

	provider := response.DnsProvider
	switch {
	case provider.QoveryDnsProviderResponse != nil:
		dnsProvider := provider.QoveryDnsProviderResponse
		return ClusterDNSProvider{
			ID:           types.StringValue(clusterID),
			ClusterID:    types.StringValue(clusterID),
			ProviderType: types.StringValue(dnsProvider.Provider),
			Domain:       types.StringValue(dnsProvider.Domain),
		}, nil
	case provider.CloudflareDnsProviderResponse != nil:
		dnsProvider := provider.CloudflareDnsProviderResponse
		apiToken := types.StringNull()
		if previous.Cloudflare != nil {
			apiToken = previous.Cloudflare.APIToken
		}
		return ClusterDNSProvider{
			ID:           types.StringValue(clusterID),
			ClusterID:    types.StringValue(clusterID),
			ProviderType: types.StringValue(dnsProvider.Provider),
			Domain:       types.StringValue(dnsProvider.Domain),
			Cloudflare: &ClusterDNSProviderCloudflare{
				Email:    types.StringValue(dnsProvider.Email),
				APIToken: apiToken,
				Proxied:  types.BoolValue(dnsProvider.Proxied),
			},
		}, nil
	case provider.Route53DnsProviderResponse != nil:
		dnsProvider := provider.Route53DnsProviderResponse
		secretAccessKey := types.StringNull()
		if previous.Route53 != nil && previous.Route53.Credentials != nil {
			secretAccessKey = previous.Route53.Credentials.AWSSecretAccessKey
		}

		credentials := &ClusterDNSProviderRoute53Credentials{
			Type:               types.StringNull(),
			AWSAccessKeyID:     types.StringNull(),
			AWSSecretAccessKey: secretAccessKey,
		}
		if dnsProvider.Credentials.Route53StaticCredentialsResponse != nil {
			credentials.Type = types.StringValue(dnsProvider.Credentials.Route53StaticCredentialsResponse.Type)
			credentials.AWSAccessKeyID = types.StringValue(dnsProvider.Credentials.Route53StaticCredentialsResponse.AwsAccessKeyId)
		}

		hostedZoneID := types.StringNull()
		if hostedZoneIDValue, ok := dnsProvider.GetHostedZoneIdOk(); ok && hostedZoneIDValue != nil {
			hostedZoneID = types.StringValue(*hostedZoneIDValue)
		}

		return ClusterDNSProvider{
			ID:           types.StringValue(clusterID),
			ClusterID:    types.StringValue(clusterID),
			ProviderType: types.StringValue(dnsProvider.Provider),
			Domain:       types.StringValue(dnsProvider.Domain),
			Route53: &ClusterDNSProviderRoute53{
				Credentials:  credentials,
				AWSRegion:    types.StringValue(dnsProvider.AwsRegion),
				HostedZoneID: hostedZoneID,
			},
		}, nil
	default:
		return ClusterDNSProvider{}, fmt.Errorf("unsupported cluster DNS provider response")
	}
}
