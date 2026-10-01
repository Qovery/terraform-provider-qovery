package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

type HelmRepository struct {
	Id                  types.String          `tfsdk:"id"`
	OrganizationId      types.String          `tfsdk:"organization_id"`
	Name                types.String          `tfsdk:"name"`
	Kind                types.String          `tfsdk:"kind"`
	URL                 types.String          `tfsdk:"url"`
	Description         types.String          `tfsdk:"description"`
	Config              *HelmRepositoryConfig `tfsdk:"config"`
	SkipTlsVerification types.Bool            `tfsdk:"skip_tls_verification"`
}

// HelmRepositoryConfig holds the authentication config for a helm repository.
// It mirrors the container registry config but without GCP `json_credentials`,
// which is not supported by the helm repository API.
type HelmRepositoryConfig struct {
	AccessKeyID       types.String `tfsdk:"access_key_id"`
	SecretAccessKey   types.String `tfsdk:"secret_access_key"`
	Region            types.String `tfsdk:"region"`
	ScalewayAccessKey types.String `tfsdk:"scaleway_access_key"`
	ScalewaySecretKey types.String `tfsdk:"scaleway_secret_key"`
	ScalewayProjectId types.String `tfsdk:"scaleway_project_id"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
}

type HelmRepositoryDataSource struct {
	Id                  types.String `tfsdk:"id"`
	OrganizationId      types.String `tfsdk:"organization_id"`
	Name                types.String `tfsdk:"name"`
	Kind                types.String `tfsdk:"kind"`
	URL                 types.String `tfsdk:"url"`
	Description         types.String `tfsdk:"description"`
	SkipTlsVerification types.Bool   `tfsdk:"skip_tls_verification"`
}

func (p HelmRepository) toUpsertRequest() helmRepository.UpsertRequest {
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
			Username:          ToStringPointer(p.Config.Username),
			Password:          ToStringPointer(p.Config.Password),
		}
	}
	return helmRepository.UpsertRequest{
		Name:               ToString(p.Name),
		Kind:               ToString(p.Kind),
		URL:                ToString(p.URL),
		Description:        ToStringPointer(p.Description),
		Config:             configRequest,
		SkiTlsVerification: ToBool(p.SkipTlsVerification),
	}
}

func convertDomainHelmRepositoryToHelmRepository(state HelmRepository, res *helmRepository.HelmRepository) HelmRepository {
	return HelmRepository{
		Id:                  FromString(res.ID.String()),
		OrganizationId:      FromString(res.OrganizationID.String()),
		Name:                FromString(res.Name),
		Kind:                FromString(res.Kind.String()),
		URL:                 FromString(res.URL.String()),
		Description:         storedDescriptionFromAPI(res.Description),
		Config:              helmRepositoryConfigFromAPI(res.Kind, state.Config, res.Config),
		SkipTlsVerification: FromBoolPointer(res.SkiTlsVerification),
	}
}

func convertDomainHelmRepositoryToHelmRepositoryDataSource(res *helmRepository.HelmRepository) HelmRepositoryDataSource {
	return HelmRepositoryDataSource{
		Id:                  FromString(res.ID.String()),
		OrganizationId:      FromString(res.OrganizationID.String()),
		Name:                FromString(res.Name),
		Kind:                FromString(res.Kind.String()),
		URL:                 FromString(res.URL.String()),
		Description:         FromStringPointer(res.Description),
		SkipTlsVerification: FromBoolPointer(res.SkiTlsVerification),
	}
}

// helmRepositoryConfigFromAPI builds the config block of a helm repository of the given kind. The
// API returns the non-secret keys the kind stores, so such a key changed or removed outside
// Terraform shows up in the plan. Everything else is kept from prior, the plan on apply and the
// state on refresh:
//   - the secrets (secret_access_key, scaleway_secret_key, password), which the API never returns;
//   - the keys the kind does not store, which the API ignores, such as region on HTTPS;
//   - every key of OCI_PUBLIC_ECR and OCI_DOCR repositories, which store none;
//   - scaleway_project_id on OCI_SCALEWAY_CR, which q-core requires but does not store.
//
// The block stays null when prior has none and the API returns no key the provider manages.
func helmRepositoryConfigFromAPI(kind helmRepository.Kind, prior *HelmRepositoryConfig, api registry.Config) *HelmRepositoryConfig {
	var config HelmRepositoryConfig
	if prior != nil {
		config = *prior
	}

	switch kind {
	case helmRepository.KindECR:
		config.Region = optionalStringFromAPI(config.Region, api.Region)
		config.AccessKeyID = optionalStringFromAPI(config.AccessKeyID, api.AccessKeyID)
	case helmRepository.KindScalewayCR:
		config.Region = scalewayRegionFromAPI(config.Region, api.Region)
		config.ScalewayAccessKey = optionalStringFromAPI(config.ScalewayAccessKey, api.ScalewayAccessKey)
	case helmRepository.KindHttps, helmRepository.KindDockerHub, helmRepository.KindGithubCr, helmRepository.KindGitlabCr, helmRepository.KindGenericCR:
		config.Username = optionalStringFromAPI(config.Username, api.Username)
	}

	// All the attributes are null: the zero value of every types.* value is null.
	if prior == nil && config == (HelmRepositoryConfig{}) {
		return nil
	}
	return &config
}
