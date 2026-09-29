//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// nullContainerRegistryConfig returns a config block whose attributes are all null.
func nullContainerRegistryConfig() ContainerRegistryConfig {
	return ContainerRegistryConfig{
		AccessKeyID:                      types.StringNull(),
		SecretAccessKey:                  types.StringNull(),
		Region:                           types.StringNull(),
		ScalewayAccessKey:                types.StringNull(),
		ScalewaySecretKey:                types.StringNull(),
		ScalewayProjectId:                types.StringNull(),
		JsonCredentials:                  types.StringNull(),
		GcpCredentialsType:               types.StringNull(),
		ProjectId:                        types.StringNull(),
		ServiceAccountEmail:              types.StringNull(),
		WorkloadIdentityProviderResource: types.StringNull(),
		TokenLifetimeSeconds:             types.Int64Null(),
		Username:                         types.StringNull(),
		Password:                         types.StringNull(),
	}
}

func withContainerRegistryConfig(change func(*ContainerRegistryConfig)) *ContainerRegistryConfig {
	config := nullContainerRegistryConfig()
	change(&config)
	return &config
}

func TestContainerRegistryConfigFromAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Kind     registry.Kind
		Prior    *ContainerRegistryConfig
		API      registry.Config
		Expect   *ContainerRegistryConfig
	}{
		{
			TestName: "ecr_reads_region_and_access_key_and_keeps_the_secret",
			Kind:     registry.KindECR,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("eu-west-3")
				c.AccessKeyID = types.StringValue("AKIA_STATE")
				c.SecretAccessKey = types.StringValue("secret")
			}),
			API: registry.Config{Region: new("eu-west-1"), AccessKeyID: new("AKIA_CONSOLE")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("eu-west-1")
				c.AccessKeyID = types.StringValue("AKIA_CONSOLE")
				c.SecretAccessKey = types.StringValue("secret")
			}),
		},
		{
			TestName: "ecr_switched_to_a_role_from_the_console_reports_no_access_key",
			Kind:     registry.KindECR,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("eu-west-3")
				c.AccessKeyID = types.StringValue("AKIA_STATE")
				c.SecretAccessKey = types.StringValue("secret")
			}),
			API: registry.Config{Region: new("eu-west-3")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("eu-west-3")
				c.SecretAccessKey = types.StringValue("secret")
			}),
		},
		{
			TestName: "public_ecr_keeps_the_region_it_does_not_store",
			Kind:     registry.KindPublicECR,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("us-east-1")
				c.AccessKeyID = types.StringValue("AKIA")
			}),
			API: registry.Config{AccessKeyID: new("AKIA")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("us-east-1")
				c.AccessKeyID = types.StringValue("AKIA")
			}),
		},
		{
			TestName: "scaleway_keeps_a_region_q_core_canonicalizes",
			Kind:     registry.KindScalewayCR,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("FR-PAR-1")
				c.ScalewayAccessKey = types.StringValue("SCW")
				c.ScalewaySecretKey = types.StringValue("secret")
				c.ScalewayProjectId = types.StringValue("project")
			}),
			API: registry.Config{Region: new("fr-par"), ScalewayAccessKey: new("SCW"), ScalewayProjectID: new("project")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("FR-PAR-1")
				c.ScalewayAccessKey = types.StringValue("SCW")
				c.ScalewaySecretKey = types.StringValue("secret")
				c.ScalewayProjectId = types.StringValue("project")
			}),
		},
		{
			TestName: "scaleway_region_changed_from_the_console_shows_up",
			Kind:     registry.KindScalewayCR,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("fr-par")
				c.ScalewayAccessKey = types.StringValue("SCW")
				c.ScalewayProjectId = types.StringValue("project")
			}),
			API: registry.Config{Region: new("nl-ams"), ScalewayAccessKey: new("SCW"), ScalewayProjectID: new("project")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("nl-ams")
				c.ScalewayAccessKey = types.StringValue("SCW")
				c.ScalewayProjectId = types.StringValue("project")
			}),
		},
		{
			TestName: "docker_hub_username_removed_from_the_console_shows_up",
			Kind:     registry.KindDockerHub,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Username = types.StringValue("user")
				c.Password = types.StringValue("secret")
			}),
			API: registry.Config{},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Password = types.StringValue("secret")
			}),
		},
		{
			TestName: "docker_hub_username_set_from_the_console_creates_the_block",
			Kind:     registry.KindDockerHub,
			Prior:    nil,
			API:      registry.Config{Username: new("console")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Username = types.StringValue("console")
			}),
		},
		{
			TestName: "no_block_and_no_key_stays_null",
			Kind:     registry.KindDockerHub,
			Prior:    nil,
			API:      registry.Config{},
			Expect:   nil,
		},
		{
			TestName: "gcp_json_key_keeps_the_project_q_core_reads_from_the_key",
			Kind:     registry.KindGcpArtifactRegistry,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("europe-west1")
				c.JsonCredentials = types.StringValue(`{"project_id":"from-key"}`)
			}),
			API: registry.Config{Region: new("europe-west1"), ProjectID: new("from-key")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("europe-west1")
				c.JsonCredentials = types.StringValue(`{"project_id":"from-key"}`)
			}),
		},
		{
			TestName: "gcp_workload_identity_reads_its_keys_and_a_default_token_lifetime_keeps_a_null_prior",
			Kind:     registry.KindGcpArtifactRegistry,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("europe-west1")
				c.GcpCredentialsType = types.StringValue("workload_identity_federation")
				c.ProjectId = types.StringValue("project")
				c.ServiceAccountEmail = types.StringValue("state@project.iam.gserviceaccount.com")
				c.WorkloadIdentityProviderResource = types.StringValue("projects/1/locations/global/workloadIdentityPools/p/providers/p")
			}),
			API: registry.Config{
				Region:                           new("europe-west1"),
				GcpCredentialsType:               new("workload_identity_federation"),
				ProjectID:                        new("project"),
				ServiceAccountEmail:              new("console@project.iam.gserviceaccount.com"),
				WorkloadIdentityProviderResource: new("projects/1/locations/global/workloadIdentityPools/p/providers/p"),
				TokenLifetimeSeconds:             new(int32(gcpTokenLifetimeSecondsDefault)),
			},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("europe-west1")
				c.GcpCredentialsType = types.StringValue("workload_identity_federation")
				c.ProjectId = types.StringValue("project")
				c.ServiceAccountEmail = types.StringValue("console@project.iam.gserviceaccount.com")
				c.WorkloadIdentityProviderResource = types.StringValue("projects/1/locations/global/workloadIdentityPools/p/providers/p")
			}),
		},
		{
			TestName: "gcp_switched_to_a_json_key_from_the_console_reports_no_federation_key",
			Kind:     registry.KindGcpArtifactRegistry,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("europe-west1")
				c.GcpCredentialsType = types.StringValue("workload_identity_federation")
				c.ProjectId = types.StringValue("project")
				c.ServiceAccountEmail = types.StringValue("state@project.iam.gserviceaccount.com")
				c.WorkloadIdentityProviderResource = types.StringValue("projects/1/locations/global/workloadIdentityPools/p/providers/p")
				c.TokenLifetimeSeconds = types.Int64Value(3600)
			}),
			API: registry.Config{Region: new("europe-west1"), ProjectID: new("from-key")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("europe-west1")
				c.ProjectId = types.StringValue("project")
			}),
		},
		{
			TestName: "azure_keeps_every_key",
			Kind:     registry.KindAzureCr,
			Prior: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("westeurope")
			}),
			API: registry.Config{Region: new("northeurope")},
			Expect: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
				c.Region = types.StringValue("westeurope")
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expect, containerRegistryConfigFromAPI(tc.Kind, tc.Prior, tc.API))
		})
	}
}

func TestScalewayRegionFromAPI(t *testing.T) {
	t.Parallel()

	assert.Equal(t, types.StringValue("fr-par"), scalewayRegionFromAPI(types.StringValue("fr-par"), new("fr-par")))
	assert.Equal(t, types.StringValue("fr-par-2"), scalewayRegionFromAPI(types.StringValue("fr-par-2"), new("fr-par")), "q-core stores fr-par for fr-par-2")
	assert.Equal(t, types.StringValue("nl-ams"), scalewayRegionFromAPI(types.StringValue("fr-par"), new("nl-ams")))
	assert.Equal(t, types.StringValue("fr-par"), scalewayRegionFromAPI(types.StringNull(), new("fr-par")), "import records the API value")
	assert.Equal(t, types.StringNull(), scalewayRegionFromAPI(types.StringValue("fr-par"), nil))
	assert.Equal(t, types.StringValue(""), scalewayRegionFromAPI(types.StringValue("fr-par"), new("")), "an empty region from the API is not a prefix match")
}

func TestGcpTokenLifetimeSecondsFromAPI(t *testing.T) {
	t.Parallel()

	assert.Equal(t, types.Int64Null(), gcpTokenLifetimeSecondsFromAPI(types.Int64Null(), nil))
	assert.Equal(t, types.Int64Null(), gcpTokenLifetimeSecondsFromAPI(types.Int64Null(), new(int32(gcpTokenLifetimeSecondsDefault))), "the default q-core stores keeps a null prior")
	assert.Equal(t, types.Int64Value(gcpTokenLifetimeSecondsDefault), gcpTokenLifetimeSecondsFromAPI(types.Int64Value(3600), new(int32(gcpTokenLifetimeSecondsDefault))), "a reset from the Console shows up")
	assert.Equal(t, types.Int64Value(7200), gcpTokenLifetimeSecondsFromAPI(types.Int64Null(), new(int32(7200))), "a lifetime set from the Console shows up")
	assert.Equal(t, types.Int64Null(), gcpTokenLifetimeSecondsFromAPI(types.Int64Value(3600), nil))
}

func TestConvertDomainRegistryToContainerRegistry(t *testing.T) {
	t.Parallel()

	registryURL, _ := url.Parse("https://docker.io")
	res := &registry.Registry{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           "registry",
		Kind:           registry.KindDockerHub,
		URL:            *registryURL,
		Config:         registry.Config{Username: new("user")},
	}
	prior := ContainerRegistry{Config: withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
		c.Username = types.StringValue("user")
		c.Password = types.StringValue("secret")
	})}

	got := convertDomainRegistryToContainerRegistry(prior, res)

	assert.Equal(t, types.StringValue(""), got.Description, "a null description reads as the default")
	assert.Equal(t, prior.Config, got.Config)

	res.Description = new("set from the Console")
	got = convertDomainRegistryToContainerRegistry(prior, res)
	assert.Equal(t, types.StringValue("set from the Console"), got.Description)
}
