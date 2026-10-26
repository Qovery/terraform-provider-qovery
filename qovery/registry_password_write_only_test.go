//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// The tests below call Create and Update the way the framework does for a write-only password:
// config.password_wo is set in the config and null in the plan, which the framework nulls.

const writeOnlyTestOrganizationID = "00000000-0000-0000-0000-000000000001"

// fakeRegistryService records the requests it receives and returns the registry they describe.
type fakeRegistryService struct {
	registry.Service
	requests []registry.UpsertRequest
}

func (f *fakeRegistryService) Create(_ context.Context, organizationID string, request registry.UpsertRequest) (*registry.Registry, error) {
	f.requests = append(f.requests, request)
	return fakeRegistryFromRequest(organizationID, uuid.NewString(), request), nil
}

func (f *fakeRegistryService) Update(_ context.Context, organizationID string, registryID string, request registry.UpsertRequest) (*registry.Registry, error) {
	f.requests = append(f.requests, request)
	return fakeRegistryFromRequest(organizationID, registryID, request), nil
}

func fakeRegistryFromRequest(organizationID, registryID string, request registry.UpsertRequest) *registry.Registry {
	registryURL, _ := url.Parse(request.URL)
	return &registry.Registry{
		ID:             uuid.MustParse(registryID),
		OrganizationID: uuid.MustParse(organizationID),
		Name:           request.Name,
		Kind:           registry.Kind(request.Kind),
		URL:            *registryURL,
		Description:    request.Description,
		Config:         registry.Config{Username: request.Config.Username},
	}
}

// fakeHelmRepositoryService records the requests it receives and returns the repository they
// describe.
type fakeHelmRepositoryService struct {
	helmRepository.Service
	requests []helmRepository.UpsertRequest
}

func (f *fakeHelmRepositoryService) Create(_ context.Context, organizationID string, request helmRepository.UpsertRequest) (*helmRepository.HelmRepository, error) {
	f.requests = append(f.requests, request)
	return fakeHelmRepositoryFromRequest(organizationID, uuid.NewString(), request), nil
}

func (f *fakeHelmRepositoryService) Update(_ context.Context, organizationID string, repositoryID string, request helmRepository.UpsertRequest) (*helmRepository.HelmRepository, error) {
	f.requests = append(f.requests, request)
	return fakeHelmRepositoryFromRequest(organizationID, repositoryID, request), nil
}

func fakeHelmRepositoryFromRequest(organizationID, repositoryID string, request helmRepository.UpsertRequest) *helmRepository.HelmRepository {
	repositoryURL, _ := url.Parse(request.URL)
	return &helmRepository.HelmRepository{
		ID:                 uuid.MustParse(repositoryID),
		OrganizationID:     uuid.MustParse(organizationID),
		Name:               request.Name,
		Kind:               helmRepository.Kind(request.Kind),
		URL:                *repositoryURL,
		Description:        request.Description,
		Config:             registry.Config{Username: request.Config.Username},
		SkiTlsVerification: &request.SkiTlsVerification,
	}
}

// registryValue builds a value of s with the given top-level attributes and config keys, every
// other attribute null.
func registryValue(t *testing.T, s schema.Schema, attributes map[string]tftypes.Value, config map[string]string) tftypes.Value {
	t.Helper()
	objType := s.Type().TerraformType(context.Background()).(tftypes.Object)
	values := map[string]tftypes.Value{}
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	for name, value := range attributes {
		values[name] = value
	}
	configType := objType.AttributeTypes["config"].(tftypes.Object)
	configValues := map[string]tftypes.Value{}
	for name, attrType := range configType.AttributeTypes {
		configValues[name] = tftypes.NewValue(attrType, nil)
	}
	for name, value := range config {
		configValues[name] = tftypes.NewValue(tftypes.String, value)
	}
	values["config"] = tftypes.NewValue(configType, configValues)
	return tftypes.NewValue(objType, values)
}

// writeOnlyTestStep describes one write: the attributes of the config, and the config keys of the
// config and of the plan. The plan holds what the config sets but password_wo, which the
// framework nulls, and the computed id and description.
type writeOnlyTestStep struct {
	TestName       string
	Description    string
	Config         map[string]string
	ExpectPassword *string
}

// plannedKeys returns the config keys of the plan: those of the config but password_wo.
func plannedKeys(config map[string]string) map[string]string {
	planned := map[string]string{}
	for key, value := range config {
		if key != "password_wo" {
			planned[key] = value
		}
	}
	return planned
}

// writeOnlyTestSteps create the registry with a write-only password, then update it. Each step
// asserts the password the request sends.
var writeOnlyTestSteps = []writeOnlyTestStep{
	{
		TestName:       "create_sends_the_write_only_password",
		Config:         map[string]string{"username": "user", "password_wo": "first", "password_wo_version": "1"},
		ExpectPassword: new("first"),
	},
	{
		TestName:       "version_bump_sends_the_new_password",
		Config:         map[string]string{"username": "user", "password_wo": "second", "password_wo_version": "2"},
		ExpectPassword: new("second"),
	},
	{
		// The trigger-only contract: any update sends the current value.
		TestName:       "unrelated_update_resends_the_current_password",
		Description:    "changed",
		Config:         map[string]string{"username": "user", "password_wo": "current", "password_wo_version": "2"},
		ExpectPassword: new("current"),
	},
	{
		TestName:       "switch_to_the_plain_password",
		Description:    "changed",
		Config:         map[string]string{"username": "user", "password": "plain"},
		ExpectPassword: new("plain"),
	},
	{
		TestName:       "switch_back_to_the_write_only_password",
		Description:    "changed",
		Config:         map[string]string{"username": "user", "password_wo": "third", "password_wo_version": "3"},
		ExpectPassword: new("third"),
	},
}

func TestContainerRegistryResource_WriteOnlyPassword(t *testing.T) {
	t.Parallel()

	service := &fakeRegistryService{}
	r := containerRegistryResource{containerRegistryService: service}
	s := schemaFor(t, r)
	nullState := tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)
	topLevel := func(id tftypes.Value, description string) map[string]tftypes.Value {
		return map[string]tftypes.Value{
			"id":              id,
			"organization_id": tftypes.NewValue(tftypes.String, writeOnlyTestOrganizationID),
			"name":            tftypes.NewValue(tftypes.String, "registry"),
			"kind":            tftypes.NewValue(tftypes.String, "GENERIC_CR"),
			"url":             tftypes.NewValue(tftypes.String, "https://registry.example.com"),
			"description":     tftypes.NewValue(tftypes.String, description),
		}
	}

	var state tfsdk.State
	for i, step := range writeOnlyTestSteps {
		ctx := context.Background()
		config := tfsdk.Config{Schema: s, Raw: registryValue(t, s, topLevel(tftypes.NewValue(tftypes.String, nil), step.Description), step.Config)}
		if i == 0 {
			plan := tfsdk.Plan{Schema: s, Raw: registryValue(t, s, topLevel(tftypes.NewValue(tftypes.String, tftypes.UnknownValue), step.Description), plannedKeys(step.Config))}
			resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: nullState}}
			r.Create(ctx, resource.CreateRequest{Plan: plan, Config: config}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s: %v", step.TestName, resp.Diagnostics)
			state = resp.State
		} else {
			var prior ContainerRegistry
			require.False(t, state.Get(ctx, &prior).HasError())
			plan := tfsdk.Plan{Schema: s, Raw: registryValue(t, s, topLevel(tftypes.NewValue(tftypes.String, prior.Id.ValueString()), step.Description), plannedKeys(step.Config))}
			resp := &resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, Config: config, State: state}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s: %v", step.TestName, resp.Diagnostics)
			state = resp.State
		}

		request := service.requests[len(service.requests)-1]
		assert.Equal(t, step.ExpectPassword, request.Config.Password, step.TestName)

		var got ContainerRegistry
		require.False(t, state.Get(ctx, &got).HasError())
		assert.Equal(t, types.StringNull(), got.Config.PasswordWo, "%s: the write-only password never reaches a model field", step.TestName)
		assert.Equal(t, plannedString(step.Config, "password"), got.Config.Password, step.TestName)
		assert.Equal(t, plannedString(step.Config, "password_wo_version"), got.Config.PasswordWoVersion, "%s: the version follows the plan", step.TestName)
	}
	assert.Len(t, service.requests, len(writeOnlyTestSteps))
}

func TestHelmRepositoryResource_WriteOnlyPassword(t *testing.T) {
	t.Parallel()

	service := &fakeHelmRepositoryService{}
	r := helmRepositoryResource{helmRepositoryService: service}
	s := schemaFor(t, r)
	nullState := tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)
	topLevel := func(id tftypes.Value, description string) map[string]tftypes.Value {
		return map[string]tftypes.Value{
			"id":                    id,
			"organization_id":       tftypes.NewValue(tftypes.String, writeOnlyTestOrganizationID),
			"name":                  tftypes.NewValue(tftypes.String, "repository"),
			"kind":                  tftypes.NewValue(tftypes.String, "HTTPS"),
			"url":                   tftypes.NewValue(tftypes.String, "https://charts.example.com"),
			"description":           tftypes.NewValue(tftypes.String, description),
			"skip_tls_verification": tftypes.NewValue(tftypes.Bool, false),
		}
	}

	var state tfsdk.State
	for i, step := range writeOnlyTestSteps {
		ctx := context.Background()
		config := tfsdk.Config{Schema: s, Raw: registryValue(t, s, topLevel(tftypes.NewValue(tftypes.String, nil), step.Description), step.Config)}
		if i == 0 {
			plan := tfsdk.Plan{Schema: s, Raw: registryValue(t, s, topLevel(tftypes.NewValue(tftypes.String, tftypes.UnknownValue), step.Description), plannedKeys(step.Config))}
			resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: nullState}}
			r.Create(ctx, resource.CreateRequest{Plan: plan, Config: config}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s: %v", step.TestName, resp.Diagnostics)
			state = resp.State
		} else {
			var prior HelmRepository
			require.False(t, state.Get(ctx, &prior).HasError())
			plan := tfsdk.Plan{Schema: s, Raw: registryValue(t, s, topLevel(tftypes.NewValue(tftypes.String, prior.Id.ValueString()), step.Description), plannedKeys(step.Config))}
			resp := &resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, Config: config, State: state}, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s: %v", step.TestName, resp.Diagnostics)
			state = resp.State
		}

		request := service.requests[len(service.requests)-1]
		assert.Equal(t, step.ExpectPassword, request.Config.Password, step.TestName)

		var got HelmRepository
		require.False(t, state.Get(ctx, &got).HasError())
		assert.Equal(t, types.StringNull(), got.Config.PasswordWo, "%s: the write-only password never reaches a model field", step.TestName)
		assert.Equal(t, plannedString(step.Config, "password"), got.Config.Password, step.TestName)
		assert.Equal(t, plannedString(step.Config, "password_wo_version"), got.Config.PasswordWoVersion, "%s: the version follows the plan", step.TestName)
	}
	assert.Len(t, service.requests, len(writeOnlyTestSteps))
}

// plannedString returns the config key as the plan holds it: null when the config does not set it.
func plannedString(config map[string]string, key string) types.String {
	if value, ok := config[key]; ok {
		return types.StringValue(value)
	}
	return types.StringNull()
}

func TestContainerRegistryConfigFromAPI_CarriesTheWriteOnlyVersion(t *testing.T) {
	t.Parallel()

	api := registry.Config{Username: new("user")}
	prior := withContainerRegistryConfig(func(c *ContainerRegistryConfig) {
		c.Username = types.StringValue("user")
		c.PasswordWoVersion = types.StringValue("7")
	})

	got := containerRegistryConfigFromAPI(registry.KindGenericCR, prior, api)
	assert.Equal(t, types.StringValue("7"), got.PasswordWoVersion, "refresh keeps the version from the state")
	assert.Equal(t, types.StringNull(), got.Password)

	imported := containerRegistryConfigFromAPI(registry.KindGenericCR, nil, api)
	assert.Equal(t, types.StringNull(), imported.PasswordWoVersion, "import leaves the version null")
	assert.Equal(t, types.StringValue("user"), imported.Username)
}

func TestHelmRepositoryConfigFromAPI_CarriesTheWriteOnlyVersion(t *testing.T) {
	t.Parallel()

	api := registry.Config{Username: new("user")}
	prior := withHelmRepositoryConfig(func(c *HelmRepositoryConfig) {
		c.Username = types.StringValue("user")
		c.PasswordWoVersion = types.StringValue("7")
	})

	got := helmRepositoryConfigFromAPI(helmRepository.KindHttps, prior, api)
	assert.Equal(t, types.StringValue("7"), got.PasswordWoVersion, "refresh keeps the version from the state")
	assert.Equal(t, types.StringNull(), got.Password)

	imported := helmRepositoryConfigFromAPI(helmRepository.KindHttps, nil, api)
	assert.Equal(t, types.StringNull(), imported.PasswordWoVersion, "import leaves the version null")
	assert.Equal(t, types.StringValue("user"), imported.Username)
}

func TestRegistryPasswordWriteOnlySchema(t *testing.T) {
	t.Parallel()

	for _, r := range []resource.Resource{containerRegistryResource{}, helmRepositoryResource{}} {
		s := schemaFor(t, r)
		config := s.Attributes["config"].(schema.SingleNestedAttribute)

		passwordWo := config.Attributes["password_wo"].(schema.StringAttribute)
		assert.True(t, passwordWo.WriteOnly)
		assert.True(t, passwordWo.Sensitive)
		assert.True(t, passwordWo.Optional)

		version := config.Attributes["password_wo_version"].(schema.StringAttribute)
		assert.False(t, version.WriteOnly)
		assert.False(t, version.Computed, "removing the version from the config plans its removal")
		assert.True(t, version.Optional)

		assert.False(t, config.Attributes["password"].(schema.StringAttribute).Required, "password stays optional")
	}
}
