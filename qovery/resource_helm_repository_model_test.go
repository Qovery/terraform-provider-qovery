//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// nullHelmRepositoryConfig returns a config block whose attributes are all null.
func nullHelmRepositoryConfig() HelmRepositoryConfig {
	return HelmRepositoryConfig{
		AccessKeyID:       types.StringNull(),
		SecretAccessKey:   types.StringNull(),
		Region:            types.StringNull(),
		ScalewayAccessKey: types.StringNull(),
		ScalewaySecretKey: types.StringNull(),
		ScalewayProjectId: types.StringNull(),
		Username:          types.StringNull(),
		Password:          types.StringNull(),
	}
}

func withHelmRepositoryConfig(change func(*HelmRepositoryConfig)) *HelmRepositoryConfig {
	config := nullHelmRepositoryConfig()
	change(&config)
	return &config
}

func TestHelmRepositoryConfigFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Kind     helmRepository.Kind
		Prior    *HelmRepositoryConfig
		API      registry.Config
		Expect   *HelmRepositoryConfig
	}{
		{
			TestName: "ecr_reads_region_and_access_key_and_keeps_the_secret",
			Kind:     helmRepository.KindECR,
			Prior: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Region = types.StringValue("eu-west-3")
				c.AccessKeyID = types.StringValue("AKIA_STATE")
				c.SecretAccessKey = types.StringValue("secret")
			}),
			API: registry.Config{Region: new("eu-west-1"), AccessKeyID: new("AKIA_CONSOLE")},
			Expect: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Region = types.StringValue("eu-west-1")
				c.AccessKeyID = types.StringValue("AKIA_CONSOLE")
				c.SecretAccessKey = types.StringValue("secret")
			}),
		},
		{
			TestName: "scaleway_keeps_the_project_q_core_does_not_store",
			Kind:     helmRepository.KindScalewayCR,
			Prior: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Region = types.StringValue("fr-par-1")
				c.ScalewayAccessKey = types.StringValue("SCW")
				c.ScalewaySecretKey = types.StringValue("secret")
				c.ScalewayProjectId = types.StringValue("project")
			}),
			API: registry.Config{Region: new("fr-par"), ScalewayAccessKey: new("SCW_CONSOLE")},
			Expect: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Region = types.StringValue("fr-par-1")
				c.ScalewayAccessKey = types.StringValue("SCW_CONSOLE")
				c.ScalewaySecretKey = types.StringValue("secret")
				c.ScalewayProjectId = types.StringValue("project")
			}),
		},
		{
			TestName: "https_username_set_from_the_console_creates_the_block",
			Kind:     helmRepository.KindHttps,
			Prior:    nil,
			API:      registry.Config{Username: new("console")},
			Expect: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Username = types.StringValue("console")
			}),
		},
		{
			TestName: "https_keeps_the_region_it_does_not_store",
			Kind:     helmRepository.KindHttps,
			Prior: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Region = types.StringValue("eu-west-3")
				c.Username = types.StringValue("user")
				c.Password = types.StringValue("secret")
			}),
			API: registry.Config{Username: new("user")},
			Expect: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Region = types.StringValue("eu-west-3")
				c.Username = types.StringValue("user")
				c.Password = types.StringValue("secret")
			}),
		},
		{
			TestName: "oci_docker_hub_username_removed_from_the_console_shows_up",
			Kind:     helmRepository.KindDockerHub,
			Prior: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Username = types.StringValue("user")
				c.Password = types.StringValue("secret")
			}),
			API: registry.Config{},
			Expect: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.Password = types.StringValue("secret")
			}),
		},
		{
			TestName: "oci_public_ecr_keeps_every_key",
			Kind:     helmRepository.KindPublicECR,
			Prior: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.AccessKeyID = types.StringValue("AKIA")
			}),
			API: registry.Config{},
			Expect: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
				c.AccessKeyID = types.StringValue("AKIA")
			}),
		},
		{
			TestName: "no_block_and_no_key_stays_null",
			Kind:     helmRepository.KindHttps,
			Prior:    nil,
			API:      registry.Config{},
			Expect:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expect, helmRepositoryConfigFromAPI(tc.Kind, tc.Prior, tc.API))
		})
	}
}

func TestHelmRepositoryToUpsertRequest_SendsTheScalewayProject(t *testing.T) {
	t.Parallel()

	repository := HelmRepository{
		Name: types.StringValue("repository"),
		Kind: types.StringValue(string(helmRepository.KindScalewayCR)),
		URL:  types.StringValue("oci://rg.fr-par.scw.cloud"),
		Config: withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
			c.Region = types.StringValue("fr-par")
			c.ScalewayAccessKey = types.StringValue("SCW")
			c.ScalewaySecretKey = types.StringValue("secret")
			c.ScalewayProjectId = types.StringValue("project")
		}),
		SkipTlsVerification: types.BoolValue(false),
	}

	request := repository.toUpsertRequest()

	assert.Equal(t, new("project"), request.Config.ScalewayProjectId, "q-core requires the project for OCI_SCALEWAY_CR")
}
