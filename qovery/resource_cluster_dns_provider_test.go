//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterSchemaDoesNotExposeDNSProvider(t *testing.T) {
	ctx := context.Background()
	cluster := newClusterResource()
	resp := &resource.SchemaResponse{}

	cluster.Schema(ctx, resource.SchemaRequest{}, resp)

	require.False(t, resp.Diagnostics.HasError())
	_, exists := resp.Schema.Attributes["dns_provider"]
	assert.False(t, exists)
}

func TestClusterDNSProviderResourceIsRegistered(t *testing.T) {
	ctx := context.Background()
	var provider qProvider

	var found bool
	for _, ctor := range provider.Resources(ctx) {
		res := ctor()
		resp := &resource.MetadataResponse{}
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "qovery"}, resp)
		if resp.TypeName == "qovery_cluster_dns_provider" {
			found = true
			break
		}
	}

	assert.True(t, found)
}
