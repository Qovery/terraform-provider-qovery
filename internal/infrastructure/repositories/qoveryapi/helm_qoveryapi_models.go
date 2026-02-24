package qoveryapi

import (
	"time"

	"github.com/pkg/errors"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type AggregateHelmResponse struct {
	Id                        string
	EnvironmentId             string
	CreatedAt                 time.Time
	UpdatedAt                 *time.Time
	Name                      string
	Description               *string
	IconUri                   string
	TimeoutSec                *int32
	AutoPreview               bool
	AutoDeploy                bool
	Arguments                 []string
	AllowClusterWideResources bool
	Source                    helm.SourceResponse
	ValuesOverride            qovery.HelmResponseAllOfValuesOverride
	Ports                     []qovery.HelmResponseAllOfPorts
}

func getAggregateHelmResponse(helmResponse *qovery.HelmResponse) AggregateHelmResponse {
	source := helm.SourceResponse{}
	if git := helmResponse.Source.HelmResponseAllOfSourceOneOf; git != nil {
		source.Git = &git.Git
	} else if repo := helmResponse.Source.HelmResponseAllOfSourceOneOf1; repo != nil {
		source.Repository = &repo.Repository
	}

	return AggregateHelmResponse{
		Id:                        helmResponse.Id,
		EnvironmentId:             helmResponse.Environment.Id,
		CreatedAt:                 helmResponse.CreatedAt,
		UpdatedAt:                 helmResponse.UpdatedAt,
		Name:                      helmResponse.Name,
		Description:               helmResponse.Description,
		IconUri:                   helmResponse.IconUri,
		TimeoutSec:                helmResponse.TimeoutSec,
		AutoPreview:               helmResponse.AutoPreview,
		AutoDeploy:                helmResponse.AutoDeploy,
		Arguments:                 helmResponse.Arguments,
		AllowClusterWideResources: helmResponse.AllowClusterWideResources,
		Source:                    source,
		ValuesOverride:            helmResponse.ValuesOverride,
		Ports:                     helmResponse.Ports,
	}
}

// newDomainHelmFromQovery converts a Qovery API HelmResponse into the domain model helm.Helm.
func newDomainHelmFromQovery(helmResponse *qovery.HelmResponse, deploymentStageID string, isSkipped bool, advancedSettingsJson string, qoveryCustomDomains *qovery.CustomDomainResponseList) (*helm.Helm, error) {
	if helmResponse == nil {
		return nil, variable.ErrNilVariable
	}

	h := getAggregateHelmResponse(helmResponse)

	var helmSourceGitRepository *helm.NewHelmSourceGitRepository = nil
	if git := h.Source.Git; git != nil {
		gitRepository := git.GitRepository

		var gitTokenId *string = nil
		if gitRepository.GitTokenId.IsSet() && gitRepository.GitTokenId.Get() != nil {
			gitTokenId = gitRepository.GitTokenId.Get()
		}

		helmSourceGitRepository = &helm.NewHelmSourceGitRepository{
			Url:        gitRepository.Url,
			Branch:     gitRepository.Branch,
			RootPath:   *gitRepository.RootPath,
			GitTokenId: gitTokenId,
		}
	}

	var helmSourceHelmRepository *helm.NewHelmSourceHelmRepository = nil
	if repo := h.Source.Repository; repo != nil {
		helmSourceHelmRepository = &helm.NewHelmSourceHelmRepository{
			RepositoryId: repo.Repository.Id,
			ChartName:    repo.ChartName,
			ChartVersion: repo.ChartVersion,
		}
	}

	source := helm.NewHelmSourceParams{
		HelmSourceGitRepository:  helmSourceGitRepository,
		HelmSourceHelmRepository: helmSourceHelmRepository,
	}

	var raw *helm.Raw = nil
	var gitRepository *helm.ValuesOverrideGit = nil
	if h.ValuesOverride.File.IsSet() && h.ValuesOverride.File.Get() != nil {
		file := h.ValuesOverride.File.Get()
		if file.Git.IsSet() && file.Git.Get() != nil {
			git := file.Git.Get()

			var gitToken *string = nil
			if git.GitRepository.GitTokenId.IsSet() && git.GitRepository.GitTokenId.Get() != nil {
				gitToken = git.GitRepository.GitTokenId.Get()
			}

			gitRepository = &helm.ValuesOverrideGit{
				Url:      git.GitRepository.Url,
				Branch:   *git.GitRepository.Branch,
				Paths:    git.Paths,
				GitToken: gitToken,
			}
		}

		if file.Raw.IsSet() && file.Raw.Get() != nil {
			rawInput := file.Raw.Get()

			values := make([]helm.RawValue, 0, len(rawInput.Values))
			for _, value := range rawInput.Values {
				values = append(values, helm.RawValue{Name: value.Name, Content: value.Content})
			}

			raw = &helm.Raw{
				Values: values,
			}
		}
	}

	file := helm.ValuesOverrideFile{
		Raw:           raw,
		GitRepository: gitRepository,
	}

	valuesOverride := helm.NewHelmValuesOverrideParams{
		Set:       h.ValuesOverride.Set,
		SetString: h.ValuesOverride.SetString,
		SetJson:   h.ValuesOverride.SetJson,
		File:      &file,
	}

	ports := make([]helm.NewHelmPortParams, 0, len(h.Ports))
	for _, port := range h.Ports {
		if port.HelmPortResponseWithServiceName != nil {
			portWithServiceName := port.HelmPortResponseWithServiceName
			if portWithServiceName.Name != nil && portWithServiceName.IsDefault != nil {
				protocol := string(portWithServiceName.Protocol)

				pt := helm.NewHelmPortParams{
					Name:         *portWithServiceName.Name,
					InternalPort: portWithServiceName.InternalPort,
					ExternalPort: portWithServiceName.ExternalPort,
					ServiceName:  portWithServiceName.ServiceName,
					Namespace:    portWithServiceName.Namespace,
					Protocol:     protocol,
					IsDefault:    *portWithServiceName.IsDefault,
				}

				ports = append(ports, pt)
			}
		}
	}

	customDomains := make([]*qovery.CustomDomain, 0, len(qoveryCustomDomains.GetResults()))
	for _, v := range qoveryCustomDomains.GetResults() {
		cpy := v
		customDomains = append(customDomains, &cpy)
	}

	return helm.NewHelm(helm.NewHelmParams{
		HelmID:                    h.Id,
		EnvironmentID:             h.EnvironmentId,
		Name:                      h.Name,
		Description:               h.Description,
		IconUri:                   h.IconUri,
		TimeoutSec:                h.TimeoutSec,
		AutoPreview:               h.AutoPreview,
		AutoDeploy:                h.AutoDeploy,
		Arguments:                 h.Arguments,
		AllowClusterWideResources: h.AllowClusterWideResources,
		Source:                    source,
		ValuesOverride:            valuesOverride,
		Ports:                     ports,
		DeploymentStageID:         deploymentStageID,
		IsSkipped:                 isSkipped,
		AdvancedSettingsJson:      advancedSettingsJson,
		CustomDomains:             customDomains,
	})
}

// newQoveryHelmRequestFromDomain takes the domain request helm.UpsertRequest and turns it into a qovery.HelmRequest to make the api call.
func newQoveryHelmRequestFromDomain(request helm.UpsertRepositoryRequest) (*qovery.HelmRequest, error) {
	ports, err := newQoveryHelmPortsRequestFromDomain(request.Ports)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrInvalidUpsertRequest.Error())
	}

	source := newQoveryHelmSourceRequestFromDomain(request.Source)

	fileValuesOverride, err := newQoveryFileValuesOverrideRequestFromDomain(request.ValuesOverride.File)
	if err != nil {
		return nil, errors.Wrap(err, helm.ErrInvalidUpsertRequest.Error())
	}

	return &qovery.HelmRequest{
		Name:                      request.Name,
		Description:               request.Description,
		IconUri:                   request.IconUri,
		TimeoutSec:                request.TimeoutSec,
		AutoPreview:               request.AutoPreview,
		AutoDeploy:                request.AutoDeploy,
		Arguments:                 request.Arguments,
		AllowClusterWideResources: &request.AllowClusterWideResources,
		Source:                    *source,
		Ports:                     *ports,
		ValuesOverride: qovery.HelmRequestAllOfValuesOverride{
			Set:       request.ValuesOverride.Set,
			SetString: request.ValuesOverride.SetString,
			SetJson:   request.ValuesOverride.SetJson,
			File:      *fileValuesOverride,
		},
	}, nil
}

func newQoveryHelmSourceRequestFromDomain(source helm.Source) *qovery.HelmRequestAllOfSource {
	var gitRepositorySource *qovery.HelmRequestAllOfSourceOneOf = nil
	if source.GitRepository != nil {
		gitTokenId := qovery.NullableString{}
		if source.GitRepository.GitTokenId != nil {
			gitTokenId.Set(source.GitRepository.GitTokenId)
		}

		gitRepository := qovery.HelmGitRepositoryRequest{
			Url:        source.GitRepository.Url,
			Branch:     source.GitRepository.Branch,
			RootPath:   &source.GitRepository.RootPath,
			GitTokenId: gitTokenId,
		}

		gitRepositorySource = &qovery.HelmRequestAllOfSourceOneOf{
			GitRepository: &gitRepository,
		}
	}

	var helmRepositorySource *qovery.HelmRequestAllOfSourceOneOf1 = nil
	if source.HelmRepository != nil {
		helmRepositoryId := qovery.NullableString{}
		helmRepositoryId.Set(&source.HelmRepository.RepositoryId)

		helmRepository := qovery.HelmRequestAllOfSourceOneOf1HelmRepository{
			Repository:   helmRepositoryId,
			ChartName:    &source.HelmRepository.ChartName,
			ChartVersion: &source.HelmRepository.ChartVersion,
		}

		helmRepositorySource = &qovery.HelmRequestAllOfSourceOneOf1{
			HelmRepository: &helmRepository,
		}
	}

	s := qovery.HelmRequestAllOfSource{
		HelmRequestAllOfSourceOneOf:  gitRepositorySource,
		HelmRequestAllOfSourceOneOf1: helmRepositorySource,
	}

	return &s
}

func newQoveryFileValuesOverrideRequestFromDomain(file *helm.ValuesOverrideFile) (*qovery.NullableHelmRequestAllOfValuesOverrideFile, error) {
	if file == nil {
		return &qovery.NullableHelmRequestAllOfValuesOverrideFile{}, nil
	}

	f := qovery.NullableHelmRequestAllOfValuesOverrideFile{}
	v := qovery.HelmRequestAllOfValuesOverrideFile{}
	f.Set(&v)

	if file.GitRepository != nil {
		provider, err := detectGitProviderFromURL(file.GitRepository.Url)
		if err != nil {
			return nil, errors.Wrap(err, helm.ErrInvalidUpsertRequest.Error())
		}

		g := qovery.HelmRequestAllOfValuesOverrideFileGit{
			Paths: file.GitRepository.Paths,
			GitRepository: qovery.ApplicationGitRepositoryRequest{
				Url:        file.GitRepository.Url,
				Branch:     &file.GitRepository.Branch,
				GitTokenId: *qovery.NewNullableString(file.GitRepository.GitToken),
				Provider:   provider,
			},
		}
		v.SetGit(g)
	}

	if file.Raw != nil {
		r := qovery.HelmRequestAllOfValuesOverrideFileRaw{}

		values := make([]qovery.HelmRequestAllOfValuesOverrideFileRawValues, 0, len(file.Raw.Values))
		for _, value := range file.Raw.Values {
			name := value.Name
			content := value.Content

			values = append(values, qovery.HelmRequestAllOfValuesOverrideFileRawValues{Name: &name, Content: &content})
		}

		r.SetValues(values)

		v.SetRaw(r)
	}

	return &f, nil
}

func newQoveryHelmPortsRequestFromDomain(ports *[]helm.Port) (*[]qovery.HelmPortRequestPortsInner, error) {
	if ports == nil {
		rv := make([]qovery.HelmPortRequestPortsInner, 0)
		return &rv, nil
	}

	rv := make([]qovery.HelmPortRequestPortsInner, 0, len(*ports))
	for _, port := range *ports {
		protocol, err := qovery.NewHelmPortProtocolEnumFromValue(port.Protocol.String())
		if err != nil {
			return nil, helm.ErrInvalidPortProtocol
		}

		portName := port.Name
		isDefault := port.IsDefault

		helmPort := qovery.HelmPortRequestPortsInner{
			Name:         &portName,
			InternalPort: port.InternalPort,
			ExternalPort: port.ExternalPort,
			ServiceName:  &port.ServiceName,
			Namespace:    port.Namespace,
			Protocol:     protocol,
			IsDefault:    &isDefault,
		}

		rv = append(rv, helmPort)
	}

	return &rv, nil
}
