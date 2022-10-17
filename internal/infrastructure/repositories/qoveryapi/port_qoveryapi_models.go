package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
)

// newQoveryContainerRequestFromDomain takes the domain request container.UpsertRequest and turns it into a qovery.ContainerRequest to make the api call.
func newQoveryPortsRequestFromDomain(requests []port.UpsertRequest) ([]qovery.ServicePortRequestPortsInner, error) {
	ports := make([]qovery.ServicePortRequestPortsInner, 0, len(requests))
	for _, r := range requests {
		newPort, err := newQoveryPortRequestFromDomain(r)
		if err != nil {
			return nil, err
		}

		ports = append(ports, *newPort)
	}

	return ports, nil
}

// newQoveryPortRequestFromDomain takes the domain request port.UpsertRequest and turns it into a qovery.ServicePortRequestPortsInner to make the api call.
func newQoveryPortRequestFromDomain(request port.UpsertRequest) (*qovery.ServicePortRequestPortsInner, error) {
	var portProtocol *qovery.PortProtocolEnum
	if request.Protocol != nil {
		proto, err := qovery.NewPortProtocolEnumFromValue(*request.Protocol)
		if err != nil {
			return nil, errors.Wrap(err, port.ErrInvalidUpsertRequest.Error())
		}

		portProtocol = proto
	}

	return &qovery.ServicePortRequestPortsInner{
		Name:               request.Name,
		Protocol:           portProtocol,
		PubliclyAccessible: request.PubliclyAccessible,
		InternalPort:       request.InternalPort,
		ExternalPort:       request.ExternalPort,
	}, nil

}

func newDomainPortsFromQovery(list []qovery.ServicePort) (port.Ports, error) {
	ports := make(port.Ports, 0, len(list))
	for _, it := range list {
		newPort, err := newDomainPortFromQovery(it)
		if err != nil {
			return nil, err
		}
		ports = append(ports, *newPort)
	}

	return ports, nil
}

func newDomainPortFromQovery(qoveryPort qovery.ServicePort) (*port.Port, error) {
	return port.NewPort(port.NewPortParams{
		PortID:             qoveryPort.Id,
		Name:               qoveryPort.Name,
		PubliclyAccessible: qoveryPort.PubliclyAccessible,
		Protocol:           string(qoveryPort.Protocol),
		InternalPort:       qoveryPort.InternalPort,
		ExternalPort:       qoveryPort.ExternalPort,
	})
}
