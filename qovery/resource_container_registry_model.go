package qovery

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

type ContainerRegistry struct {
	Id             types.String             `tfsdk:"id"`
	OrganizationId types.String             `tfsdk:"organization_id"`
	Name           types.String             `tfsdk:"name"`
	Kind           types.String             `tfsdk:"kind"`
	URL            types.String             `tfsdk:"url"`
	Description    types.String             `tfsdk:"description"`
	Config         *ContainerRegistryConfig `tfsdk:"config"`
}

type ContainerRegistryConfig struct {
	AccessKeyID       types.String `tfsdk:"access_key_id"`
	SecretAccessKey   types.String `tfsdk:"secret_access_key"`
	Region            types.String `tfsdk:"region"`
	ScalewayAccessKey types.String `tfsdk:"scaleway_access_key"`
	ScalewaySecretKey types.String `tfsdk:"scaleway_secret_key"`
	ScalewayProjectId types.String `tfsdk:"scaleway_project_id"`
	JsonCredentials   types.String `tfsdk:"json_credentials"`

	GcpCredentialsType               types.String `tfsdk:"gcp_credentials_type"`
	ProjectId                        types.String `tfsdk:"project_id"`
	ServiceAccountEmail              types.String `tfsdk:"service_account_email"`
	WorkloadIdentityProviderResource types.String `tfsdk:"workload_identity_provider_resource"`
	TokenLifetimeSeconds             types.Int64  `tfsdk:"token_lifetime_seconds"`

	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type ContainerRegistryDataSource struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Kind           types.String `tfsdk:"kind"`
	URL            types.String `tfsdk:"url"`
	Description    types.String `tfsdk:"description"`
}

func (p ContainerRegistry) toUpsertRequest() registry.UpsertRequest {
	var configRequest registry.UpsertRequestConfig
	if p.Config == nil {
		configRequest = registry.UpsertRequestConfig{}
	} else {
		configRequest = registry.UpsertRequestConfig{
			AccessKeyID:       ToStringPointer(p.Config.AccessKeyID),
			SecretAccessKey:   ToStringPointer(p.Config.SecretAccessKey),
			Region:            ToStringPointer(p.Config.Region),
			ScalewayAccessKey: ToStringPointer(p.Config.ScalewayAccessKey),
			ScalewaySecretKey: ToStringPointer(p.Config.ScalewaySecretKey),
			ScalewayProjectId: ToStringPointer(p.Config.ScalewayProjectId),
			JsonCredentials:   ToStringPointer(p.Config.JsonCredentials),

			GcpCredentialsType:               ToStringPointer(p.Config.GcpCredentialsType),
			ProjectId:                        ToStringPointer(p.Config.ProjectId),
			ServiceAccountEmail:              ToStringPointer(p.Config.ServiceAccountEmail),
			WorkloadIdentityProviderResource: ToStringPointer(p.Config.WorkloadIdentityProviderResource),
			TokenLifetimeSeconds:             ToInt32Pointer(p.Config.TokenLifetimeSeconds),

			Username: ToStringPointer(p.Config.Username),
			Password: ToStringPointer(p.Config.Password),
		}
	}
	return registry.UpsertRequest{
		Name:        ToString(p.Name),
		Kind:        ToString(p.Kind),
		URL:         ToString(p.URL),
		Description: ToStringPointer(p.Description),
		Config:      configRequest,
	}
}

func convertDomainRegistryToContainerRegistry(state ContainerRegistry, res *registry.Registry) ContainerRegistry {
	return ContainerRegistry{
		Id:             FromString(res.ID.String()),
		OrganizationId: FromString(res.OrganizationID.String()),
		Name:           FromString(res.Name),
		Kind:           FromString(res.Kind.String()),
		URL:            FromString(res.URL.String()),
		Description:    storedDescriptionFromAPI(res.Description),
		Config:         containerRegistryConfigFromAPI(res.Kind, state.Config, res.Config),
	}
}

// containerRegistryConfigFromAPI builds the config block of a registry of the given kind. The API
// returns the non-secret keys the kind stores, so such a key changed or removed outside Terraform
// shows up in the plan. Everything else is kept from prior, the plan on apply and the state on
// refresh:
//   - the secrets (secret_access_key, scaleway_secret_key, json_credentials, password), which the
//     API never returns;
//   - the keys the kind does not store, which the API ignores, such as region on PUBLIC_ECR;
//   - every key of DOCR and AZURE_CR registries, which the provider cannot manage;
//   - project_id on a GCP registry that uses a JSON key, which q-core reads from the key. The
//     API reports gcp_credentials_type only for Workload Identity Federation.
//
// The block stays null when prior has none and the API returns no key the provider manages.
func containerRegistryConfigFromAPI(kind registry.Kind, prior *ContainerRegistryConfig, api registry.Config) *ContainerRegistryConfig {
	var config ContainerRegistryConfig
	if prior != nil {
		config = *prior
	}

	switch kind {
	case registry.KindECR:
		config.Region = optionalStringFromAPI(config.Region, api.Region)
		config.AccessKeyID = optionalStringFromAPI(config.AccessKeyID, api.AccessKeyID)
	case registry.KindPublicECR:
		config.AccessKeyID = optionalStringFromAPI(config.AccessKeyID, api.AccessKeyID)
	case registry.KindScalewayCR:
		config.Region = scalewayRegionFromAPI(config.Region, api.Region)
		config.ScalewayAccessKey = optionalStringFromAPI(config.ScalewayAccessKey, api.ScalewayAccessKey)
		config.ScalewayProjectId = optionalStringFromAPI(config.ScalewayProjectId, api.ScalewayProjectID)
	case registry.KindDockerHub, registry.KindGithubCr, registry.KindGithubEnterpriseCr, registry.KindGitlabCr, registry.KindGenericCR:
		config.Username = optionalStringFromAPI(config.Username, api.Username)
	case registry.KindGcpArtifactRegistry:
		config.Region = optionalStringFromAPI(config.Region, api.Region)
		config.GcpCredentialsType = optionalStringFromAPI(config.GcpCredentialsType, api.GcpCredentialsType)
		config.ServiceAccountEmail = optionalStringFromAPI(config.ServiceAccountEmail, api.ServiceAccountEmail)
		config.WorkloadIdentityProviderResource = optionalStringFromAPI(config.WorkloadIdentityProviderResource, api.WorkloadIdentityProviderResource)
		config.TokenLifetimeSeconds = gcpTokenLifetimeSecondsFromAPI(config.TokenLifetimeSeconds, api.TokenLifetimeSeconds)
		if api.GcpCredentialsType != nil {
			config.ProjectId = optionalStringFromAPI(config.ProjectId, api.ProjectID)
		}
	}

	// All the attributes are null: the zero value of every types.* value is null.
	if prior == nil && config == (ContainerRegistryConfig{}) {
		return nil
	}
	return &config
}

// scalewayRegionFromAPI reads the region of a Scaleway registry. q-core stores the canonical
// region the request starts with (fr-par for fr-par-1 or FR-PAR), so prior is kept when it names
// the region the API returns.
func scalewayRegionFromAPI(prior types.String, apiVal *string) types.String {
	if apiVal != nil && *apiVal != "" && !prior.IsNull() && !prior.IsUnknown() && strings.HasPrefix(strings.ToLower(prior.ValueString()), *apiVal) {
		return prior
	}
	return optionalStringFromAPI(prior, apiVal)
}

// gcpTokenLifetimeSecondsDefault is the token lifetime q-core stores for a Workload Identity
// Federation registry whose request omits it.
const gcpTokenLifetimeSecondsDefault = 14400

// gcpTokenLifetimeSecondsFromAPI reads token_lifetime_seconds. The API value wins, except that nil
// or the q-core default keeps a null prior: config has no schema Default, since the attribute
// only applies to Workload Identity Federation.
func gcpTokenLifetimeSecondsFromAPI(prior types.Int64, apiVal *int32) types.Int64 {
	if prior.IsNull() && (apiVal == nil || *apiVal == gcpTokenLifetimeSecondsDefault) {
		return types.Int64Null()
	}
	return FromInt32Pointer(apiVal)
}

func convertDomainRegistryToContainerRegistryDataSource(res *registry.Registry) ContainerRegistryDataSource {
	return ContainerRegistryDataSource{
		Id:             FromString(res.ID.String()),
		OrganizationId: FromString(res.OrganizationID.String()),
		Name:           FromString(res.Name),
		Kind:           FromString(res.Kind.String()),
		URL:            FromString(res.URL.String()),
		Description:    FromStringPointer(res.Description),
	}
}
