package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

type HelmRepository struct {
	Id                  types.String             `tfsdk:"id"`
	OrganizationId      types.String             `tfsdk:"organization_id"`
	Name                types.String             `tfsdk:"name"`
	Kind                types.String             `tfsdk:"kind"`
	URL                 types.String             `tfsdk:"url"`
	Description         types.String             `tfsdk:"description"`
	Config              *ContainerRegistryConfig `tfsdk:"config"`
	SkipTlsVerification types.Bool               `tfsdk:"skip_tls_verification"`
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
		Description:         FromStringPointer(res.Description),
		Config:              state.Config,
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
