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
func newDomainContainerFromQovery(c *qovery.ContainerResponse, deploymentStageId string) (*container.Container, error) {
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

	return container.NewContainer(container.NewContainerParams{
		ContainerID:         c.Id,
		EnvironmentID:       c.Environment.Id,
		RegistryID:          c.Registry.Id,
		Name:                c.Name,
		ImageName:           c.ImageName,
		Tag:                 c.Tag,
		AutoPreview:         c.AutoPreview,
		CPU:                 c.Cpu,
		Memory:              c.Memory,
		MinRunningInstances: c.MinRunningInstances,
		MaxRunningInstances: c.MaxRunningInstances,
		Entrypoint:          c.Entrypoint,
		Arguments:           c.Arguments,
		Ports:               ports,
		Storages:            storages,
		DeploymentStageId:   deploymentStageId,
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

	return &qovery.ContainerRequest{
		RegistryId:          request.RegistryID,
		Name:                request.Name,
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
	}, nil
}
