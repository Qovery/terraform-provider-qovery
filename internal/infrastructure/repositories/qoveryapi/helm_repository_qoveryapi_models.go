package qoveryapi

import (
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

func newDomainHelmRepositoryFromQovery(v *qovery.HelmRepositoryResponse, organizationID string) (*helmRepository.HelmRepository, error) {
	if v == nil {
		return nil, registry.ErrNilRegistry
	}

	return helmRepository.NewHelmRepository(helmRepository.NewHelmRepositoryParams{
		RepositoryId:       v.GetId(),
		OrganizationID:     organizationID,
		Name:               v.GetName(),
		Kind:               string(v.GetKind()),
		URL:                v.GetUrl(),
		Description:        v.Description,
		SkiTlsVerification: v.SkipTlsVerification,
	})
}

func newQoveryHelmRepositoryRequestFromDomain(request helmRepository.UpsertRequest) (*qovery.HelmRepositoryRequest, error) {
	kind, err := qovery.NewHelmRepositoryKindEnumFromValue(request.Kind)
	if err != nil {
		return nil, registry.ErrInvalidKindParam
	}

	return &qovery.HelmRepositoryRequest{
		Name:                request.Name,
		Kind:                *kind,
		Url:                 new(request.URL),
		Description:         request.Description,
		SkipTlsVerification: request.SkiTlsVerification,
		Config: qovery.HelmRepositoryRequestConfig{
			AccessKeyId:       request.Config.AccessKeyID,
			SecretAccessKey:   request.Config.SecretAccessKey,
			Region:            request.Config.Region,
			ScalewayAccessKey: request.Config.ScalewayAccessKey,
			ScalewaySecretKey: request.Config.ScalewaySecretKey,
			ScalewayProjectId: request.Config.ScalewayProjectId,
			Username:          request.Config.Username,
			Password:          request.Config.Password,
		},
	}, nil
}
