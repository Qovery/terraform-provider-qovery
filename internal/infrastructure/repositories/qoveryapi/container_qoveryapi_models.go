package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
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
	})
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

	return &qovery.ContainerRequest{
		RegistryId:          request.RegistryID,
		Name:                request.Name,
		IconUri:             request.IconUri,
		ImageName:           request.ImageName,
		Tag:                 request.Tag,
		Entrypoint:          request.Entrypoint,
		AutoPreview:         request.AutoPreview,
		Cpu:                 request.CPU,
		Memory:              request.Memory,
		MinRunningInstances: request.MinRunningInstances,
		MaxRunningInstances: request.MaxRunningInstances,
		Arguments:           request.Arguments,
		Storage:             storages,
		Ports:               ports,
		Healthchecks:        request.Healthchecks,
		AutoDeploy:          request.AutoDeploy,
		AnnotationsGroups:   annotationsGroups,
		LabelsGroups:        labelsGroups,
	}, nil
}
