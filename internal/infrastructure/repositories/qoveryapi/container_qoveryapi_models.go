package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// newDomainCredentialsFromQovery takes a qovery.EnvironmentVariable returned by the API client and turns it into the domain model variable.Variable.
func newDomainContainerFromQovery(
	c *qovery.ContainerResponse,
	deploymentStageID string,
	isSkipped bool,
	advancedSettingsAsJson string,
	qoveryCustomDomains *qovery.CustomDomainResponseList,
) (*container.Container, error) {
	if c == nil {
		return nil, variable.ErrNilVariable
	}

	ports, err := newDomainPortsFromQovery(c.Ports)
	if err != nil {
		return nil, errors.Wrap(err, port.ErrInvalidPorts.Error())
	}

	storages, err := newDomainStoragesFromQovery(c.Storage)
	if err != nil {
		return nil, errors.Wrap(err, storage.ErrInvalidStorages.Error())
	}

	customDomains := make([]*qovery.CustomDomain, 0, len(qoveryCustomDomains.GetResults()))
	for _, v := range qoveryCustomDomains.GetResults() {
		cpy := v
		customDomains = append(customDomains, &cpy)
	}

	annotationsGroupIds := make([]string, 0, len(c.AnnotationsGroups))
	for _, annotationsGroupId := range c.AnnotationsGroups {
		annotationsGroupIds = append(annotationsGroupIds, annotationsGroupId.Id)
	}

	labelsGroupIds := make([]string, 0, len(c.LabelsGroups))
	for _, labelsGroupId := range c.LabelsGroups {
		labelsGroupIds = append(labelsGroupIds, labelsGroupId.Id)
	}

	autoscalingPolicy, err := autoscaling.FromQoveryResponse(c.Autoscaling)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidContainer.Error())
	}

	return container.NewContainer(container.NewContainerParams{
		ContainerID:            c.Id,
		EnvironmentID:          c.Environment.Id,
		RegistryID:             c.Registry.Id,
		Name:                   c.Name,
		IconUri:                c.IconUri,
		ImageName:              c.ImageName,
		Tag:                    c.Tag,
		AutoPreview:            c.AutoPreview,
		CPU:                    c.Cpu,
		Memory:                 c.Memory,
		EphemeralStorage:       c.EphemeralStorageInGib,
		MinRunningInstances:    c.MinRunningInstances,
		MaxRunningInstances:    c.MaxRunningInstances,
		Entrypoint:             c.Entrypoint,
		Arguments:              c.Arguments,
		Ports:                  ports,
		Storages:               storages,
		DeploymentStageID:      deploymentStageID,
		IsSkipped:              isSkipped,
		AdvancedSettingsAsJson: advancedSettingsAsJson,
		CustomDomains:          customDomains,
		Healthchecks:           c.Healthchecks,
		AutoDeploy:             c.AutoDeploy,
		AnnotationsGroupIds:    annotationsGroupIds,
		LabelsGroupIds:         labelsGroupIds,
		Autoscaling:            autoscalingPolicy,
		GitSource:              newDomainContainerGitSourceFromQovery(c.Source),
		BuildSettings:          c.BuildSettings,
	})
}

// newDomainContainerGitSourceFromQovery returns the Git repository source of a container, nil for an
// image source or for a response without source, which an API older than Git containers returns.
func newDomainContainerGitSourceFromQovery(source *qovery.ContainerSourceResponse) *docker.NewDockerParams {
	if source == nil || source.ContainerSourceResponseOneOf1 == nil {
		return nil
	}

	git := source.ContainerSourceResponseOneOf1.GitRepository
	return &docker.NewDockerParams{
		GitRepository: git_repository.NewGitRepositoryParams{
			Url:        git.Url,
			Branch:     git.Branch,
			CommitID:   git.DeployedCommitId,
			RootPath:   git.RootPath,
			GitTokenId: git.GitTokenId.Get(),
			Provider:   gitProviderPointer(git.Provider),
		},
		DockerFilePath:         &git.DockerfilePath,
		DockerTargetBuildStage: git.DockerTargetBuildStage.Get(),
	}
}

// newQoveryContainerRequestFromDomain takes the domain request container.UpsertRequest and turns it into a qovery.ContainerRequest to make the api call.
func newQoveryContainerRequestFromDomain(request container.UpsertRepositoryRequest) (*qovery.ContainerRequest, error) {
	ports, err := newQoveryPortsRequestFromDomain(request.Ports)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	storages, err := newQoveryStoragesRequestFromDomain(request.Storages)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	annotationsGroups, err := NewQoveryServiceAnnotationsGroupRequestFromDomain(request.AnnotationsGroupIds)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	labelsGroups, err := NewQoveryServiceLabelsGroupRequestFromDomain(request.LabelsGroupIds)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	var autoscalingPolicy *qovery.AutoscalingPolicyRequest
	if request.Autoscaling != nil {
		policy, err := autoscaling.ToQoveryRequest(*request.Autoscaling)
		if err != nil {
			return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
		}
		autoscalingPolicy = &policy
	}

	containerRequest := &qovery.ContainerRequest{
		Name:                  request.Name,
		IconUri:               request.IconUri,
		Entrypoint:            request.Entrypoint,
		AutoPreview:           request.AutoPreview,
		Cpu:                   request.CPU,
		Memory:                request.Memory,
		EphemeralStorageInGib: request.EphemeralStorage,
		MinRunningInstances:   request.MinRunningInstances,
		MaxRunningInstances:   request.MaxRunningInstances,
		Arguments:             request.Arguments,
		Storage:               storages,
		Ports:                 ports,
		Healthchecks:          request.Healthchecks,
		AutoDeploy:            request.AutoDeploy,
		AnnotationsGroups:     annotationsGroups,
		LabelsGroups:          labelsGroups,
		Autoscaling:           autoscalingPolicy,
	}

	switch {
	case request.GitSource != nil:
		gitSource, err := newQoveryContainerGitSourceRequestFromDomain(*request.GitSource)
		if err != nil {
			return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
		}
		containerRequest.Source = gitSource
		containerRequest.BuildSettings = request.BuildSettings
	case request.ImageAsSource:
		image := qovery.NewContainerSourceRequestOneOfImage(request.RegistryID, request.ImageName, request.Tag)
		source := qovery.ContainerSourceRequestOneOfAsContainerSourceRequest(qovery.NewContainerSourceRequestOneOf(*image))
		containerRequest.Source = &source
	default:
		containerRequest.RegistryId = &request.RegistryID
		containerRequest.ImageName = &request.ImageName
		containerRequest.Tag = &request.Tag
	}

	return containerRequest, nil
}

func newQoveryContainerGitSourceRequestFromDomain(gitSource docker.Docker) (*qovery.ContainerSourceRequest, error) {
	provider, err := ResolveGitProvider(gitSource.GitRepository.Url, gitSource.GitRepository.Provider)
	if err != nil {
		return nil, err
	}

	dockerfilePath := ""
	if gitSource.DockerFilePath != nil {
		dockerfilePath = *gitSource.DockerFilePath
	}

	gitRepository := qovery.ContainerGitRepositorySourceRequest{
		Url:                    gitSource.GitRepository.Url,
		Branch:                 gitSource.GitRepository.Branch,
		RootPath:               gitSource.GitRepository.RootPath,
		GitTokenId:             *qovery.NewNullableString(gitSource.GitRepository.GitTokenId),
		Provider:               provider,
		DockerfilePath:         dockerfilePath,
		DockerTargetBuildStage: *qovery.NewNullableString(gitSource.DockerTargetBuildStage),
	}
	source := qovery.ContainerSourceRequestOneOf1AsContainerSourceRequest(qovery.NewContainerSourceRequestOneOf1(gitRepository))
	return &source, nil
}
