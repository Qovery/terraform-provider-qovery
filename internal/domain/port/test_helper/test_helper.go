package test_helper

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
)

var (
	DefaultPortName                      = "port-" + uuid.New().String()
	DefaultPortInternalPort        int32 = 8080
	DefaultPortInvalidInternalPort int32 = -1
	DefaultPortExternalPort        int32 = 80
	DefaultPortPubliclyAccessible        = true
	DefaultPortProtocol                  = port.ProtocolHTTP

	DefaultValidPort = port.Port{
		ID:                 uuid.New(),
		InternalPort:       DefaultPortInternalPort,
		ExternalPort:       &DefaultPortExternalPort,
		PubliclyAccessible: DefaultPortPubliclyAccessible,
		Protocol:           &DefaultPortProtocol,
	}

	DefaultValidPortParams = port.NewPortParams{
		PortID:             uuid.New().String(),
		InternalPort:       DefaultPortInternalPort,
		ExternalPort:       &DefaultPortExternalPort,
		PubliclyAccessible: DefaultPortPubliclyAccessible,
		Protocol:           DefaultPortProtocol.String(),
	}

	DefaultInvalidPort = port.Port{
		ID:                 uuid.New(),
		InternalPort:       DefaultPortInvalidInternalPort,
		ExternalPort:       &DefaultPortExternalPort,
		PubliclyAccessible: DefaultPortPubliclyAccessible,
		Protocol:           &DefaultPortProtocol,
	}

	DefaultInvalidPortParams = port.NewPortParams{
		PortID:             uuid.New().String(),
		InternalPort:       DefaultPortInvalidInternalPort,
		ExternalPort:       &DefaultPortExternalPort,
		PubliclyAccessible: DefaultPortPubliclyAccessible,
		Protocol:           DefaultPortProtocol.String(),
	}

	DefaultInvalidPortParamsError = errors.New("invalid internal port param")
)
